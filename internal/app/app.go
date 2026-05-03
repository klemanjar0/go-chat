package app

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type App struct {
	Server *fiber.App
	PgPool *pgxpool.Pool
	Redis  *redis.Client
}

func New(
	Server *fiber.App,
	PgPool *pgxpool.Pool,
	Redis *redis.Client,
) *App {

	return &App{
		PgPool: PgPool,
		Redis:  Redis,
		Server: Server,
	}
}
