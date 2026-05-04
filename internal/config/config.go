package config

import (
	"fmt"
	"time"

	"go-chat/pkg/config"
	"go-chat/pkg/postgres"
	"go-chat/pkg/redis"
)

type Environment string

const (
	EnvProd  Environment = "PRODUCTION"
	EnvStage Environment = "STAGING"
	EnvDev   Environment = "DEVELOPMENT"
)

type Auth struct {
	JWTSecret  string
	JWTIssuer  string
	AccessTTL  time.Duration
	RefreshTTL time.Duration
	BcryptCost int
}

type App struct {
	Env      Environment
	HTTPPort string
	Postgres postgres.Config
	Redis    redis.Config
	Auth     Auth
}

func (cfg *App) IsProd() bool { return cfg.Env == EnvProd }

func parseEnvironment(s string) (Environment, error) {
	switch s {
	case string(EnvDev):
		return EnvDev, nil
	case string(EnvProd):
		return EnvProd, nil
	case string(EnvStage):
		return EnvStage, nil
	default:
		return EnvProd, fmt.Errorf("unknown environment: %s", s)
	}
}

func Load() *App {
	envStr := config.GetEnv("ENV", string(EnvProd))
	env, err := parseEnvironment(envStr)
	if err != nil {
		fmt.Println("Error reading environment mode, applying production...")
	}

	pg := postgres.Config{
		Host:     config.GetEnv("DB_HOST", "localhost"),
		Port:     config.GetEnvInt("DB_PORT", 5432),
		User:     config.GetEnv("DB_USER", "chat-db-admin"),
		Password: config.GetEnv("DB_PASSWORD", "chat-db-admin-password"),
		DBName:   config.GetEnv("DB_NAME", "chat-db"),
		SSLMode:  config.GetEnv("DB_SSL_MODE", "disable"),
		MaxConns: int32(config.GetEnvInt("DB_MAX_CONNS", 10)),
	}

	rd := redis.Config{
		Host:     config.GetEnv("REDIS_HOST", "localhost"),
		Port:     config.GetEnvInt("REDIS_PORT", 6379),
		Password: config.GetEnv("REDIS_PASSWORD", ""),
		DB:       config.GetEnvInt("REDIS_DB", 0),
	}

	auth := Auth{
		JWTSecret:  config.GetEnv("JWT_SECRET", ""),
		JWTIssuer:  config.GetEnv("JWT_ISSUER", "go-chat"),
		AccessTTL:  time.Duration(config.GetEnvInt("ACCESS_TOKEN_TTL_SECONDS", 900)) * time.Second,
		RefreshTTL: time.Duration(config.GetEnvInt("REFRESH_TOKEN_TTL_SECONDS", 60*60*24*30)) * time.Second,
		BcryptCost: config.GetEnvInt("BCRYPT_COST", 12),
	}

	if env == EnvProd && auth.JWTSecret == "" {
		panic("JWT_SECRET must be set in production")
	}
	if auth.JWTSecret == "" {
		auth.JWTSecret = "dev-insecure-secret-change-me"
	}

	return &App{
		Env:      env,
		HTTPPort: config.GetEnv("HTTP_PORT", "8080"),
		Postgres: pg,
		Redis:    rd,
		Auth:     auth,
	}
}
