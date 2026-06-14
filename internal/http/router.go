package http

import (
    "github.com/gin-gonic/gin"
    "log/slog"
    "aurion/core/internal/db/repository"
    "aurion/core/internal/http/handlers"
    "aurion/core/internal/http/middleware"
    "aurion/core/internal/mail"
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
    mailService *mail.MailService,
    serverSecret []byte,
) *gin.Engine {

    r := gin.New()
    r.Use(gin.Recovery())

    //
    // AUTH
    //
    auth := handlers.NewAuthHandler(users, sessions, serverSecret)

    r.GET("/health", func(c *gin.Context) {
        logger.Info("healthcheck")
        c.JSON(200, gin.H{"status": "ok"})
    })

    r.POST("/auth/signup", auth.Signup)
    r.POST("/auth/login", auth.Login)
    r.POST("/auth/salt", auth.GetSalt)
    r.GET("/auth/session", auth.Session)

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
    // AUTHENTICATED ROUTES
    //
    authGroup := r.Group("/")
    authGroup.Use(middleware.AuthMiddleware(sessions))

    // Upload keys
    authGroup.POST("/keys/public", keyHandler.UploadPublicKey)
    authGroup.POST("/keys/private", keyHandler.UploadPrivateKey)
    authGroup.GET("/keys/private/me", keyHandler.GetPrivateKey)

    //
    // MAIL
    //
    mailHandler := handlers.NewMailHandler(mailService)

    authGroup.POST("/mail/send", mailHandler.SendMail)
    authGroup.GET("/mail/messages", mailHandler.ListMessages)
    authGroup.GET("/mail/message/:id", mailHandler.GetMessage)
    authGroup.DELETE("/mail/message/:id", mailHandler.DeleteMessage)
    authGroup.POST("/mail/message/:id/seen", mailHandler.SetSeen)
    authGroup.POST("/mail/message/:id/tags", mailHandler.UpdateTags)

    //
    // MAILBOXES
    //
    authGroup.GET("/mail/mailboxes", mailHandler.ListMailboxes)
    authGroup.POST("/mail/mailbox/create", mailHandler.CreateMailbox)
    authGroup.POST("/mail/mailbox/rename", mailHandler.RenameMailbox)
    authGroup.POST("/mail/mailbox/delete", mailHandler.DeleteMailbox)

    //
    // DRAFTS
    //
    authGroup.POST("/mail/draft", mailHandler.CreateDraft)
    authGroup.PUT("/mail/draft/:id", mailHandler.UpdateDraft)
    authGroup.DELETE("/mail/draft/:id", mailHandler.DeleteDraft)

    // INTERNAL
    routingHandler := handlers.NewRoutingHandler(
        identities,
        members,
        publicKeys,
        catchall,
    )

    internal := r.Group("/internal")
    internal.POST("/routing/resolve", routingHandler.Resolve)
    
    return r
}
