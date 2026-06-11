package handlers

import (
    "strings"
    "strconv"

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

func parseIntOrDefault(s string, def int) int {
    if s == "" {
        return def
    }
    v, err := strconv.Atoi(s)
    if err != nil {
        return def
    }
    return v
}
