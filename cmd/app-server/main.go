package main

import (
	"github.com/joho/godotenv"
    "aurion/core/internal/config"
    "aurion/core/internal/http"
    "aurion/core/internal/log"
    "aurion/core/internal/db"
    "aurion/core/internal/db/generated"
    "aurion/core/internal/db/repository"
)

func main() {
    // Load config
	_ = godotenv.Load()
    cfg := config.Load()
    logger := log.New(cfg.Env)

    // Connect to PostgreSQL
    dbConn, err := db.Connect(
        cfg.DBHost,
        cfg.DBPort,
        cfg.DBUser,
        cfg.DBPass,
        cfg.DBName,
    )
    if err != nil {
        logger.Error("failed to connect to PostgreSQL", "error", err)
        return
    }
    logger.Info("Connected to PostgreSQL")

    // Initialize SQLC
    queries := generated.New(dbConn)

    // Initialize repositories
    userRepo := repository.NewUserRepository(queries)
    publicKeyRepo := repository.NewPublicKeyRepository(queries)
    privateKeyRepo := repository.NewPrivateKeyRepository(queries)
    sessionRepo := repository.NewSessionRepository(queries)

    // Pass repositories to the router (dependency injection)
    router := http.NewRouter(logger,
        userRepo,
        publicKeyRepo,
        privateKeyRepo,
        sessionRepo,
    )

    logger.Info("Starting Boson app-server", "port", cfg.AppPort)
    router.Run(":" + cfg.AppPort)
}
