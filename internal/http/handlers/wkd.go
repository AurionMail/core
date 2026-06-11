package handlers

import (
    "github.com/gin-gonic/gin"
    "aurion/core/internal/wkd"
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

    // 1. Récupérer toutes les clés publiques
    keys, err := h.publicKeys.ListAllPublicKeys(c)
    if err != nil {
        c.JSON(500, gin.H{"error": "server error"})
        return
    }

    // 2. Trouver celle dont le hash correspond
    for _, k := range keys {
        if wkd.WKDHash(k.Email) == hash {
            c.Data(200, "application/octet-stream", []byte(k.ArmoredKey))
            return
        }
    }

    c.JSON(404, gin.H{"error": "not found"})
}

func (h *WKDHandler) GetPolicy(c *gin.Context) {
    c.String(200, "policy: openpgpkey")
}
