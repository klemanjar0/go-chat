package cmd

import (
	"context"
	"go-chat/internal/app"
	"go-chat/internal/configuration"
	"go-chat/pkg/logger"
	"net/http"

	pkgpostgres "go-chat/pkg/postgres"
	pkgredis "go-chat/pkg/redis"
)

func main() {
	ctx := context.Background()
	global := configuration.Load()
	logger.Init("chat", !global.IsProd())

	logger.Info("Starting chat service...")

	// --- PostgreSQL ---
	pgPool, err := pkgpostgres.NewPool(ctx, *global.PgCfg)
	if err != nil {
		logger.Fatal("failed to connect to postgres", "err", err)
	}

	redis, redisErr := pkgredis.NewClient(ctx, *global.RedisCfg)
	if redisErr != nil {
		logger.Fatal("failed to connect to redis", "err", err)
	}

	app := app.New(pgPool, redis)

}
