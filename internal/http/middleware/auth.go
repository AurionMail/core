package middleware

import (
    "net/http"
    "strings"
    "time"

    "github.com/gin-gonic/gin"
    "aurion/core/internal/db/repository"
)

func AuthMiddleware(sessions *repository.SessionRepository) gin.HandlerFunc {
    return func(c *gin.Context) {

        header := c.GetHeader("Authorization")
        if header == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
            c.Abort()
            return
        }

        // 2. extract token
        parts := strings.SplitN(header, " ", 2)
        if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization format"})
            c.Abort()
            return
        }

        token := parts[1]

        // 3. check session in DB
        session, err := sessions.GetSessionByToken(c, token)
        if err != nil || session.ExpiresAt.Before(time.Now()) {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired session"})
            c.Abort()
            return
        }

        // 4. inject user_id in context
        c.Set("user_id", session.UserID.String())

        c.Next()
    }
}
