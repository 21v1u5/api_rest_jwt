package config

import (
	"fmt"
	"os"
	"time"
)

type Config struct {
	Port		string
	DatabaseURL	string
	JWTSecret	string
	AcessTTL	time.Duration
	RefreshTTL	time.Duration
}

func Load() (Config, error) {
	cfg := Config{
		Port:			getEnv("PORT", "8080"),
		DatabaseURL:	os.Getenv("DATABASE_URL"),
		JWTSecret:		os.Getenv("JWT_SECRET"),
		AcessTTL:		15 * time.Minute,
		RefreshTTL:		7 * 24 * time.Hour,
	}

	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}
	if len(cfg.JWT_Secret) < 32 {
		return Config{}, fmt.Errorf("JWT_SECRET must be at least 32 characteres")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}