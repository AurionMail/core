package handlers

import (
    "strings"

    "github.com/gin-gonic/gin"
)

func extractBearer(c *gin.Context) string {
    header := c.GetHeader("Authorization")
    if header == "" {
        return ""
    }

    // Format attendu : "Bearer <token>"
    parts := strings.SplitN(header, " ", 2)
    if len(parts) != 2 {
        return ""
    }

    if strings.ToLower(parts[0]) != "bearer" {
        return ""
    }

    return parts[1]
}
