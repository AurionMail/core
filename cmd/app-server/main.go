package main

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/pressly/goose/v3"

	"aurion/core/internal/config"
	"aurion/core/internal/db"
	"aurion/core/internal/db/generated"
	"aurion/core/internal/db/repository"
	"aurion/core/internal/http"
	"aurion/core/internal/log"
	"aurion/core/internal/mail"
	"aurion/core/internal/sync"
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

	// -------------------------------
	//  AUTOMATIC MIGRATIONS (Goose)
	// -------------------------------
	logger.Info("Checking and applying database migrations...")
	goose.SetBaseFS(db.EmbedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		logger.Error("failed to set goose dialect", "error", err)
		return
	}

	// Remplace le chemin par le dossier exact où se trouvent tes fichiers .sql
	if err := goose.Up(dbConn, "migrations"); err != nil {
		logger.Error("failed to run database migrations", "error", err)
		return
	}
	logger.Info("Database migrations applied successfully")

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
	serverSecret := []byte(os.Getenv("AUTH_FAKE_SALT_SECRET"))
	backendType := os.Getenv("MAIL_BACKEND")

	var mailBackend mail.MailBackend

	switch backendType {
	case "smtp":
		// Utilise l'adresse du serveur SMTP (ex: mail.domaine.com:465)
		mailBackend = mail.NewSMTPBackend(os.Getenv("SMTP_URL"))
	case "imap":
		mailBackend = mail.NewIMAPBackend(os.Getenv("IMAP_URL"))
	default:
		logger.Warn("No MAIL_BACKEND configured, external credential check will be skipped")
	}

	// -------------------------------
	//  BACKGROUND SYNC WORKER
	// -------------------------------
	stalwartJmapURL := os.Getenv("STALWART_URL")
	stalwartAdminKey := os.Getenv("STALWART_API_KEY")

	if stalwartJmapURL != "" && stalwartAdminKey != "" {
		logger.Info("Initializing Stalwart identity synchronizer via JMAP Admin API")

		stalwartConn := sync.NewStalwartJMAPConnector(stalwartJmapURL, stalwartAdminKey)

		syncService := sync.NewSyncService(
			stalwartConn,
			userRepo,
			identityRepo,
			identityMemberRepo,
			logger,
			5*time.Minute,
		)

		ctx := context.Background()
		syncService.Start(ctx)

		logger.Info("Stalwart identity synchronizer started background loop")
	} else {
		logger.Warn("Stalwart sync bypassed: JMAP_URL or STALWART_API_KEY missing in env")
	}

	// -------------------------------
	//  ROUTER
	// -------------------------------
	allowedOriginsRaw := os.Getenv("CORS_ALLOWED_ORIGINS")
	var allowedOrigins []string

	if allowedOriginsRaw != "" {
		// Découpage par virgule pour obtenir []string{"url1", "url2"}
		allowedOrigins = strings.Split(allowedOriginsRaw, ",")
	} else {
		// Fallback de sécurité par défaut si l'env est vide
		allowedOrigins = []string{"http://localhost:5173"}
		logger.Warn("CORS_ALLOWED_ORIGINS not set, falling back to localhost")
	}

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
		allowedOrigins,
	)

	logger.Info("Starting Boson app-server", "port", cfg.AppPort)
	router.Run(":" + cfg.AppPort)
}
