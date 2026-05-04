package configuration

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

type AuthConfig struct {
	JWTSecret       string
	JWTIssuer       string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
	BcryptCost      int
}

type GlobalConfig struct {
	Env      Environment
	PgCfg    *postgres.Config
	RedisCfg *redis.Config
	AuthCfg  *AuthConfig
	HttpPort string
}

func parseEnvironment(s string) (Environment, error) {
	switch s {
	case "DEVELOPMENT":
		return EnvDev, nil
	case "PRODUCTION":
		return EnvProd, nil
	case "STAGING":
		return EnvStage, nil
	default:
		return EnvProd, fmt.Errorf("unknown environment: %s", s)
	}
}

func Load() *GlobalConfig {
	envStr := config.GetEnv("ENV", string(EnvProd))
	env, err := parseEnvironment(envStr)

	if err != nil {
		fmt.Println("Error reading environment mode, applying production...")
	}

	// -- postgres config --

	pgCfg := &postgres.Config{
		Host:     config.GetEnv("DB_HOST", "localhost"),
		Port:     config.GetEnvInt("DB_PORT", 5432),
		User:     config.GetEnv("DB_USER", "chat-db-admin"),
		Password: config.GetEnv("DB_PASSWORD", "chat-db-admin-password"),
		DBName:   config.GetEnv("DB_NAME", "chat-db"),
		SSLMode:  config.GetEnv("DB_SSL_MODE", "disable"),
		MaxConns: int32(config.GetEnvInt("DB_MAX_CONNS", 10)),
	}

	// -- redis config --

	redisCfg := &redis.Config{
		Host:     config.GetEnv("REDIS_HOST", "localhost"),
		Port:     config.GetEnvInt("REDIS_PORT", 6379),
		Password: config.GetEnv("REDIS_PASSWORD", ""),
		DB:       config.GetEnvInt("REDIS_DB", 0),
	}

	// -- auth config --

	authCfg := &AuthConfig{
		JWTSecret:       config.GetEnv("JWT_SECRET", ""),
		JWTIssuer:       config.GetEnv("JWT_ISSUER", "go-chat"),
		AccessTokenTTL:  time.Duration(config.GetEnvInt("ACCESS_TOKEN_TTL_SECONDS", 900)) * time.Second,
		RefreshTokenTTL: time.Duration(config.GetEnvInt("REFRESH_TOKEN_TTL_SECONDS", 60*60*24*30)) * time.Second,
		BcryptCost:      config.GetEnvInt("BCRYPT_COST", 12),
	}

	if env == EnvProd && authCfg.JWTSecret == "" {
		panic("JWT_SECRET must be set in production")
	}
	if authCfg.JWTSecret == "" {
		authCfg.JWTSecret = "dev-insecure-secret-change-me"
	}

	return &GlobalConfig{
		Env:      env,
		PgCfg:    pgCfg,
		RedisCfg: redisCfg,
		AuthCfg:  authCfg,
		HttpPort: config.GetEnv("HTTP_PORT", "8080"),
	}
}

// utils

func (cfg *GlobalConfig) IsProd() bool {
	return cfg.Env == EnvProd
}
