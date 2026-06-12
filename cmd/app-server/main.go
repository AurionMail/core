package main

import (
    "os"

    "github.com/joho/godotenv"

    "aurion/core/internal/config"
    "aurion/core/internal/http"
    "aurion/core/internal/log"
    "aurion/core/internal/db"
    "aurion/core/internal/db/generated"
    "aurion/core/internal/db/repository"
    "aurion/core/internal/mail"
    "aurion/core/internal/mail/backends/jmap"
)

func main() {
    // Load .env
    _ = godotenv.Load()

    // Load config
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

    // -------------------------------
    //  REPOSITORIES (nouveau modèle)
    // -------------------------------
    userRepo := repository.NewUserRepository(queries)

    identityRepo := repository.NewIdentityRepository(queries)
    identityPublicKeyRepo := repository.NewIdentityPublicKeyRepository(queries)
    identityPrivateKeyRepo := repository.NewIdentityPrivateKeyRepository(queries)
    identityMemberRepo := repository.NewIdentityMemberRepository(queries)
    catchallRepo := repository.NewRoutingCatchallRepository(queries)

    sessionRepo := repository.NewSessionRepository(queries)

    // -------------------------------
    //  MAIL BACKEND
    // -------------------------------
    backendType := os.Getenv("MAIL_BACKEND")

    var mailBackend mail.MailBackend

    switch backendType {
    case "jmap":
        mailBackend = jmap.NewJMAPBackend(
            os.Getenv("JMAP_URL"),
            os.Getenv("JMAP_USERNAME"),
            os.Getenv("JMAP_PASSWORD"),
        )
    default:
        logger.Error("Unknown MAIL_BACKEND", "backend", backendType)
        return
    }

    // MailService
    mailService := mail.NewMailService(
        mailBackend,
        identityRepo,
        identityPublicKeyRepo,
        identityPrivateKeyRepo,
    )

    // -------------------------------
    //  ROUTER
    // -------------------------------
    router := http.NewRouter(
        logger,
        userRepo,
        identityRepo,
        identityPublicKeyRepo,
        identityPrivateKeyRepo,
        identityMemberRepo,
        sessionRepo,
        catchallRepo,
        mailService,
    )

    logger.Info("Starting Boson app-server", "port", cfg.AppPort)
    router.Run(":" + cfg.AppPort)
}
