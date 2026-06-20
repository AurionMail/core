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
    // backendType := os.Getenv("MAIL_BACKEND")
    serverSecret := []byte(os.Getenv("AUTH_FAKE_SALT_SECRET"))

   backendType := os.Getenv("MAIL_BACKEND")

    var mailBackend mail.MailBackend

    switch backendType {
    case "jmap":
        mailBackend = mail.NewJMAPBackend(os.Getenv("JMAP_URL"))
    case "imap":
        mailBackend = mail.NewIMAPBackend(os.Getenv("IMAP_SERVER"))
    default:
        logger.Warn("No MAIL_BACKEND configured, external credential check will be skipped")
    }

    
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
        mailBackend,
        serverSecret,
    )

    logger.Info("Starting Boson app-server", "port", cfg.AppPort)
    router.Run(":" + cfg.AppPort)
}
