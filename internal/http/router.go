package http

import (
    "github.com/gin-gonic/gin"
    "log/slog"
	"aurion/core/internal/db/repository"
	"aurion/core/internal/http/handlers"
)

func NewRouter(
    logger *slog.Logger,
    users *repository.UserRepository,
    publicKeys *repository.PublicKeyRepository,
    privateKeys *repository.PrivateKeyRepository,
    sessions *repository.SessionRepository,
) *gin.Engine {

    r := gin.New()
    r.Use(gin.Recovery())

    auth := handlers.NewAuthHandler(users, sessions)

    r.POST("/auth/signup", auth.Signup)
    r.POST("/auth/login", auth.Login)
    r.GET("/auth/session", auth.Session)

    r.GET("/health", func(c *gin.Context) {
		logger.Info("healthcheck")
        c.JSON(200, gin.H{"status": "ok"})
    })

    return r
}


