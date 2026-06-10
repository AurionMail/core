package http

import (
    "github.com/gin-gonic/gin"
    "log/slog"
)

func NewRouter(logger *slog.Logger) *gin.Engine {
    r := gin.New()
    r.Use(gin.Recovery())

    r.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "ok"})
    })

    return r
}
