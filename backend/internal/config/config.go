// Package config читает настройки из переменных окружения (.env.example).
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	HTTPAddr     string
	DatabaseURL  string
	JWTSecret    string
	JWTTTL       time.Duration
	DemoAuth     bool
	MockURL      string
	MockTimeout  time.Duration
	StorageDir   string
	CORSOrigin   string
	WorkerEnable bool
}

func Load() (Config, error) {
	c := Config{
		HTTPAddr:     env("HTTP_ADDR", ":8080"),
		DatabaseURL:  env("DATABASE_URL", ""),
		JWTSecret:    env("JWT_SECRET", ""),
		JWTTTL:       duration("JWT_TTL", 12*time.Hour),
		DemoAuth:     boolean("DEMO_AUTH", true),
		MockURL:      env("MOCK_URL", "http://localhost:8090/mock"),
		MockTimeout:  duration("MOCK_TIMEOUT", 3*time.Second),
		StorageDir:   env("STORAGE_DIR", "./data/documents"),
		CORSOrigin:   env("CORS_ORIGIN", "http://localhost:3000"),
		WorkerEnable: boolean("WORKER_ENABLED", true),
	}
	if c.DatabaseURL == "" || c.JWTSecret == "" {
		return c, fmt.Errorf("DATABASE_URL и JWT_SECRET обязательны")
	}
	return c, nil
}

func env(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}

func duration(key string, def time.Duration) time.Duration {
	if d, err := time.ParseDuration(env(key, "")); err == nil {
		return d
	}
	return def
}

func boolean(key string, def bool) bool {
	if b, err := strconv.ParseBool(env(key, "")); err == nil {
		return b
	}
	return def
}
