package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	AppName         string
	DatabaseURL     string
	RedisURL        string
	JWTSecret       string
	JWTExpiration   time.Duration
	RefreshExp      time.Duration
}

func Load() *Config {
	return &Config{
		AppName:       getEnv("APP_NAME", "User CRUD API"),
		DatabaseURL:   getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/users_db"),
		RedisURL:      getEnv("REDIS_URL", "redis://localhost:6379"),
		JWTSecret:     getEnv("JWT_SECRET", "your-super-secret-key-change-in-production"),
		JWTExpiration: getDurationEnv("JWT_EXPIRATION_HOURS", 24),
		RefreshExp:   getDurationEnv("JWT_REFRESH_EXPIRATION_DAYS", 7),
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getDurationEnv(key string, defaultVal time.Duration) time.Duration {
	if val := os.Getenv(key); val != "" {
		if hours, err := strconv.Atoi(val); err == nil {
			return time.Duration(hours)
		}
	}
	return defaultVal
}