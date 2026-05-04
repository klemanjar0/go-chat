// Package app is the composition root: it wires concrete repositories,
// stores, use cases and HTTP handlers and exposes a Run() that owns the
// process lifecycle. main.go stays a thin entry point.
package app

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-chat/internal/config"
	domainauth "go-chat/internal/domain/auth"
	domainuser "go-chat/internal/domain/user"
	transporthttp "go-chat/internal/transport/http"
	authhttp "go-chat/internal/transport/http/handlers/auth"
	userhttp "go-chat/internal/transport/http/handlers/user"
	"go-chat/internal/transport/http/middleware"
	authuc "go-chat/internal/usecase/auth"
	useruc "go-chat/internal/usecase/user"
	pkgauth "go-chat/pkg/auth"
	"go-chat/pkg/clock"
	"go-chat/pkg/fiberutil"
	"go-chat/pkg/httputil"
	"go-chat/pkg/logger"
	pkgpostgres "go-chat/pkg/postgres"
	pkgredis "go-chat/pkg/redis"
)

func Run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	cfg := config.Load()
	logger.Init("chat", !cfg.IsProd())
	logger.Info("starting chat service", "env", cfg.Env)

	pgPool, err := pkgpostgres.NewPool(ctx, cfg.Postgres)
	if err != nil {
		return err
	}
	defer pgPool.Close()

	rdb, err := pkgredis.NewClient(ctx, cfg.Redis)
	if err != nil {
		return err
	}
	defer func() { _ = rdb.Close() }()

	userRepo := domainuser.NewRepository(pgPool)
	refreshRepo := domainauth.NewRefreshTokenRepository(pgPool)
	accessStore := domainauth.NewAccessTokenStore(rdb)
	jwt := pkgauth.NewJWTIssuer(cfg.Auth.JWTSecret, cfg.Auth.JWTIssuer, cfg.Auth.AccessTTL)

	authUC := authuc.NewUseCase(cfg.Auth, userRepo, refreshRepo, accessStore, jwt, pkgauth.Service{}, clock.System{})
	userUC := useruc.NewUseCase(userRepo, pkgauth.Service{}, cfg.Auth.BcryptCost)

	registerErrorMappings()

	app := fiberutil.NewApp()
	app.Use(middleware.RequestID(), fiberutil.Logging())
	transporthttp.RegisterRoutes(app, transporthttp.Handlers{
		Auth: authhttp.NewHandler(authUC),
		User: userhttp.NewHandler(userUC),
	}, middleware.Auth(authUC))

	srvErr := make(chan error, 1)
	go func() {
		addr := ":" + cfg.HTTPPort
		logger.Info("http listening", "addr", addr)
		if err := app.Listen(addr); err != nil && !errors.Is(err, http.ErrServerClosed) {
			srvErr <- err
			return
		}
		srvErr <- nil
	}()

	select {
	case <-ctx.Done():
		logger.Info("shutdown requested, draining...")
	case err := <-srvErr:
		if err != nil {
			return err
		}
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutdownCancel()
	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		logger.Error("http shutdown error", "err", err)
	}
	return nil
}

func registerErrorMappings() {
	httputil.RegisterErrorMapping(domainauth.ErrInvalidPayload, http.StatusBadRequest)
	httputil.RegisterErrorMapping(domainauth.ErrInvalidCredentials, http.StatusUnauthorized)
	httputil.RegisterErrorMapping(domainauth.ErrUnauthorized, http.StatusUnauthorized)
	httputil.RegisterErrorMapping(domainauth.ErrInvalidToken, http.StatusUnauthorized)
	httputil.RegisterErrorMapping(domainauth.ErrTokenRevoked, http.StatusUnauthorized)
	httputil.RegisterErrorMapping(domainauth.ErrTokenExpired, http.StatusUnauthorized)
	httputil.RegisterErrorMapping(domainuser.ErrUsernameTaken, http.StatusConflict)
	httputil.RegisterErrorMapping(domainuser.ErrUserNotFound, http.StatusNotFound)
	httputil.RegisterErrorMapping(domainuser.ErrInvalidUserID, http.StatusBadRequest)
}
