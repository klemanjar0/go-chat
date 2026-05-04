package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-chat/internal/configuration"
	"go-chat/internal/domain/user"
	transporthttp "go-chat/internal/transport/http"
	"go-chat/pkg/fiberutil"
	"go-chat/pkg/httputil"
	"go-chat/pkg/logger"

	pkgpostgres "go-chat/pkg/postgres"
	pkgredis "go-chat/pkg/redis"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	global := configuration.Load()
	logger.Init("chat", !global.IsProd())
	logger.Info("Starting chat service...")

	pgPool, err := pkgpostgres.NewPool(ctx, *global.PgCfg)
	if err != nil {
		logger.Fatal("failed to connect to postgres", "err", err)
	}
	defer pgPool.Close()

	rdb, err := pkgredis.NewClient(ctx, *global.RedisCfg)
	if err != nil {
		logger.Fatal("failed to connect to redis", "err", err)
	}
	defer func() { _ = rdb.Close() }()

	userRepo := user.NewRepository(pgPool)
	tokenStore := user.NewTokenStore(rdb)
	userUC := user.NewUseCase(global.AuthCfg, userRepo, tokenStore)
	userHandler := user.NewHttpHandler(userUC)
	authMW := user.AuthMiddleware(userUC)

	httputil.RegisterErrorMapping(user.ErrInvalidPayload, http.StatusBadRequest)
	httputil.RegisterErrorMapping(user.ErrInvalidCredentials, http.StatusUnauthorized)
	httputil.RegisterErrorMapping(user.ErrUnauthorized, http.StatusUnauthorized)
	httputil.RegisterErrorMapping(user.ErrInvalidToken, http.StatusUnauthorized)
	httputil.RegisterErrorMapping(user.ErrTokenRevoked, http.StatusUnauthorized)
	httputil.RegisterErrorMapping(user.ErrTokenExpired, http.StatusUnauthorized)
	httputil.RegisterErrorMapping(user.ErrUsernameTaken, http.StatusConflict)
	httputil.RegisterErrorMapping(user.ErrUserNotFound, http.StatusNotFound)

	app := fiberutil.NewApp()
	app.Use(fiberutil.Logging())
	transporthttp.RegisterRoutes(app, userHandler, authMW)

	go func() {
		addr := ":" + global.HttpPort
		logger.Info("http listening", "addr", addr)
		if err := app.Listen(addr); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal("http server error", "err", err)
		}
	}()

	<-ctx.Done()
	logger.Info("shutdown requested, draining...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutdownCancel()
	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		logger.Error("http shutdown error", "err", err)
	}
}
