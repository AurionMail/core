package config

import (
    "os"
)

type Config struct {
    AppPort string
}

func Load() Config {
    return Config{
        AppPort: getEnv("APP_PORT", "8080"),
    }
}

func getEnv(key, fallback string) string {
    if v, ok := os.LookupEnv(key); ok {
        return v
    }
    return fallback
}
