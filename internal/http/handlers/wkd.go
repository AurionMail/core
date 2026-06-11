package handlers

import (
    "github.com/gin-gonic/gin"
    "aurion/core/internal/db/repository"
)

type WKDHandler struct {
    publicKeys *repository.PublicKeyRepository
}

func NewWKDHandler(pub *repository.PublicKeyRepository) *WKDHandler {
    return &WKDHandler{pub}
}

func (h *WKDHandler) GetWKDKey(c *gin.Context) {
    hash := c.Param("hash")

    key, err := h.publicKeys.GetByWKDHash(c, hash)
    if err != nil {
        c.JSON(404, gin.H{"error": "not found"})
        return
    }

    c.Data(200, "application/octet-stream", []byte(key.ArmoredKey))
}


func (h *WKDHandler) GetPolicy(c *gin.Context) {
    c.String(200, "policy: openpgpkey")
}
