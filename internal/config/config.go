package config

import (
    "os"
)


type Config struct {
    AppPort string
    Env     string

    DBHost string
    DBPort string
    DBUser string
    DBPass string
    DBName string
}

func Load() Config {
    return Config{
        AppPort: getEnv("APP_PORT", "8080"),
        Env:     getEnv("APP_ENV", "dev"),

        DBHost: getEnv("DB_HOST", "localhost"),
        DBPort: getEnv("DB_PORT", "5432"),
        DBUser: getEnv("DB_USER", "postgress"),
        DBPass: getEnv("DB_PASS", ""),
        DBName: getEnv("DB_NAME", "boson_dev"),
    }
}


func getEnv(key, fallback string) string {
    if v, ok := os.LookupEnv(key); ok {
        return v
    }
    return fallback
}
