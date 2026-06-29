package http

import (
	"aurion/core/internal/db/repository"
	"aurion/core/internal/http/handlers"
	"aurion/core/internal/http/middleware"
	"aurion/core/internal/mail"
	"log/slog"

	"github.com/gin-contrib/cors" // <--- 1. AJOUTE CET IMPORT
	"github.com/gin-gonic/gin"
)

func NewRouter(
	logger *slog.Logger,
	users *repository.UserRepository,
	identities *repository.IdentityRepository,
	publicKeys *repository.IdentityPublicKeyRepository,
	privateKeys *repository.IdentityPrivateKeyRepository,
	members *repository.IdentityMemberRepository,
	sessions *repository.SessionRepository,
	catchall *repository.RoutingCatchallRepository,
	mailBackend mail.MailBackend,
	serverSecret []byte,
	allowedOrigins []string,
) *gin.Engine {

	r := gin.New()
	r.Use(gin.Recovery())

	// -------------------------------------------------------------------------
	// 2. CONFIGURATION CORS GLOBALE (CRUCIAL)
	// -------------------------------------------------------------------------
	// Ce middleware intercepte TOUTES les requêtes, y compris les requêtes OPTIONS (Preflight)
	r.Use(cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	//
	// AUTH
	//
	auth := handlers.NewAuthHandler(users, sessions, mailBackend, serverSecret)

	r.GET("/health", func(c *gin.Context) {
		logger.Info("healthcheck")
		c.JSON(200, gin.H{"status": "ok"})
	})

	r.POST("/auth/signup", auth.Signup)
	r.POST("/auth/login", auth.Login)
	r.POST("/auth/salt", auth.GetSalt)
	r.GET("/auth/session", auth.Session)
	r.POST("/auth/verify", auth.VerifyCredentials)

	//
	// KEYS
	//
	keyHandler := handlers.NewKeyHandler(
		identities,
		publicKeys,
		privateKeys,
		members,
	)

	// Public no auth
	r.GET("/keys/public/:email", keyHandler.GetPublicKey)

	// WKD no auth of course
	wkdHandler := handlers.NewWKDHandler(publicKeys)
	r.GET("/.well-known/openpgpkey/policy", wkdHandler.GetPolicy)
	r.GET("/.well-known/openpgpkey/hu/:hash", wkdHandler.GetWKDKey)

	//
	// SYNC HANDLER (Nouveau)
	//
	syncHandler := handlers.NewSyncHandler(
		identities,
		publicKeys,
		privateKeys,
		members,
	)

	//
	// AUTHENTICATED ROUTES
	//
	authGroup := r.Group("/")
	authGroup.Use(middleware.AuthMiddleware(sessions))

	// Upload keys
	authGroup.POST("/keys/public", keyHandler.UploadPublicKey)
	authGroup.POST("/keys/private", keyHandler.UploadPrivateKey)
	authGroup.GET("/keys/private/me", keyHandler.GetPrivateKey)

	// Routings & Key share synchronization (Ajouté sous protection auth)
	authGroup.GET("/sync/routing", syncHandler.SyncRouting)
	authGroup.POST("/keys/sync", syncHandler.UploadSyncKeys)

	authGroup.GET("/server", auth.GetServerLogin)

	// INTERNAL
	routingHandler := handlers.NewRoutingHandler(
		identities,
		publicKeys,
		catchall,
	)

	internal := r.Group("/internal")
	internal.POST("/routing/resolve", routingHandler.Resolve)

	return r
}
