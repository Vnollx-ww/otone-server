package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

const devJWTSecret = "dev-only-insecure-secret-change-me"

type Config struct {
	Env             string
	HTTPAddr        string
	MySQLDSN        string
	RedisAddr       string
	RedisPassword   string
	RedisDB         int
	JWTSecret       string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
	LogLevel        string
	ShutdownTimeout time.Duration
}

// Load reads configuration from environment variables, falling back to .env
// in the current working directory when present. Environment variables always
// take precedence over .env values.
func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		Env:           getEnv("APP_ENV", "dev"),
		HTTPAddr:      getEnv("HTTP_ADDR", ":8090"),
		MySQLDSN:      getEnv("MYSQL_DSN", "otome:otome@tcp(127.0.0.1:3307)/otome?parseTime=true&charset=utf8mb4&loc=UTC&timeout=5s&readTimeout=10s&writeTimeout=10s"),
		RedisAddr:     getEnv("REDIS_ADDR", "127.0.0.1:6380"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		JWTSecret:     getEnv("JWT_SECRET", ""),
		LogLevel:      getEnv("LOG_LEVEL", "info"),
	}

	var err error
	if cfg.RedisDB, err = strconv.Atoi(getEnv("REDIS_DB", "0")); err != nil {
		return nil, fmt.Errorf("invalid REDIS_DB: %w", err)
	}
	if cfg.AccessTokenTTL, err = time.ParseDuration(getEnv("ACCESS_TOKEN_TTL", "15m")); err != nil {
		return nil, fmt.Errorf("invalid ACCESS_TOKEN_TTL: %w", err)
	}
	if cfg.RefreshTokenTTL, err = time.ParseDuration(getEnv("REFRESH_TOKEN_TTL", "720h")); err != nil {
		return nil, fmt.Errorf("invalid REFRESH_TOKEN_TTL: %w", err)
	}
	if cfg.ShutdownTimeout, err = time.ParseDuration(getEnv("SHUTDOWN_TIMEOUT", "10s")); err != nil {
		return nil, fmt.Errorf("invalid SHUTDOWN_TIMEOUT: %w", err)
	}

	if strings.TrimSpace(cfg.JWTSecret) == "" {
		if cfg.IsProd() {
			return nil, errors.New("JWT_SECRET is required in production")
		}
		cfg.JWTSecret = devJWTSecret
	}
	if cfg.IsProd() && len(cfg.JWTSecret) < 32 {
		return nil, errors.New("JWT_SECRET must be at least 32 bytes in production")
	}
	return cfg, nil
}

func (c *Config) IsProd() bool {
	return strings.EqualFold(c.Env, "prod") || strings.EqualFold(c.Env, "production")
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && strings.TrimSpace(value) != "" {
		return value
	}
	return fallback
}
