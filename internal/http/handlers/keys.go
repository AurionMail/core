package handlers

import (
    "github.com/gin-gonic/gin"
    "aurion/core/internal/db/repository"
)

type KeyHandler struct {
    PublicKeys  *repository.PublicKeyRepository
    PrivateKeys *repository.PrivateKeyRepository
}

func NewKeyHandler(pub *repository.PublicKeyRepository, priv *repository.PrivateKeyRepository) *KeyHandler {
    return &KeyHandler{pub, priv}
}

type UploadPublicKeyRequest struct {
    Email      string `json:"email"`
    ArmoredKey string `json:"armored_key"`
}

func (h *KeyHandler) UploadPublicKey(c *gin.Context) {
    var req UploadPublicKeyRequest
    if err := c.BindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": "invalid request"})
        return
    }

    userID := c.GetString("user_id") // injecté par middleware auth

    key, err := h.PublicKeys.InsertPublicKey(c, userID, req.Email, req.ArmoredKey, true)
    if err != nil {
        c.JSON(500, gin.H{"error": "failed to store public key"})
        return
    }

    c.JSON(200, gin.H{"id": key.ID})
}


type UploadPrivateKeyRequest struct {
    ArmoredEncryptedKey string `json:"armored_encrypted_key"`
}

func (h *KeyHandler) UploadPrivateKey(c *gin.Context) {
    var req UploadPrivateKeyRequest
    if err := c.BindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": "invalid request"})
        return
    }

    userID := c.GetString("user_id")

    key, err := h.PrivateKeys.InsertPrivateKey(c, userID, req.ArmoredEncryptedKey)
    if err != nil {
        c.JSON(500, gin.H{"error": "failed to store private key"})
        return
    }

    c.JSON(200, gin.H{"id": key.ID})
}

func (h *KeyHandler) GetPublicKey(c *gin.Context) {
    email := c.Param("email")

    key, err := h.PublicKeys.GetPrimaryPublicKeyByEmail(c, email)
    if err != nil {
        c.JSON(404, gin.H{"error": "not found"})
        return
    }

    c.JSON(200, gin.H{
        "email":       key.Email,
        "armored_key": key.ArmoredKey,
    })
}

func (h *KeyHandler) GetPrivateKey(c *gin.Context) {
    userID := c.GetString("user_id")

    key, err := h.PrivateKeys.GetLatestPrivateKey(c, userID)
    if err != nil {
        c.JSON(404, gin.H{"error": "not found"})
        return
    }

    c.JSON(200, gin.H{
        "armored_encrypted_key": key.ArmoredEncryptedKey,
    })
}
