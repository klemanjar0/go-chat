package configuration

import (
	"fmt"
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

type GlobalConfig struct {
	Env      Environment
	PgCfg    *postgres.Config
	RedisCfg *redis.Config
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

	return &GlobalConfig{
		Env:      env,
		PgCfg:    pgCfg,
		RedisCfg: redisCfg,
		HttpPort: config.GetEnv("HTTP_PORT", "8080"),
	}
}

// utils

func (cfg *GlobalConfig) IsProd() bool {
	return cfg.Env == EnvProd
}
