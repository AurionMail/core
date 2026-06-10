package http

import (
    "github.com/gin-gonic/gin"
    "log/slog"
	"aurion/core/internal/db/repository"
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

    r.GET("/health", func(c *gin.Context) {
        logger.Info("healthcheck")
        c.JSON(200, gin.H{"status": "ok"})
    })

    return r
}

