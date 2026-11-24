package config

import (
	"os"
	"time"
)

type Config struct {
	Env            string
	Port           string
	RequestTimeout time.Duration
}

func Load() Config {
	return Config{
		Env:            getEnv("APP_ENV", "development"),
		Port:           getEnv("APP_PORT", "8080"),
		RequestTimeout: getDuration("APP_REQUEST_TIMEOUT", 2*time.Second),
	}
}

func getDuration(key string, fallback time.Duration) time.Duration {
	d, err := time.ParseDuration(getEnv(key, ""))
	if err != nil || d <= 0 {
		return fallback
	}
	return d
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return fallback
}
