package handlers

import (
    "github.com/gin-gonic/gin"
    "aurion/core/internal/db/repository"
)

type KeyHandler struct {
    Identities   *repository.IdentityRepository
    PublicKeys   *repository.IdentityPublicKeyRepository
    PrivateKeys  *repository.IdentityPrivateKeyRepository
    Members      *repository.IdentityMemberRepository
}

func NewKeyHandler(
    identities *repository.IdentityRepository,
    pub *repository.IdentityPublicKeyRepository,
    priv *repository.IdentityPrivateKeyRepository,
    members *repository.IdentityMemberRepository,
) *KeyHandler {
    return &KeyHandler{identities, pub, priv, members}
}

type UploadPublicKeyRequest struct {
    Email      string `json:"email"`
    ArmoredKey string `json:"armored_key"`
    WKDHash    string `json:"wkd_hash"`
}

func (h *KeyHandler) UploadPublicKey(c *gin.Context) {
    var req UploadPublicKeyRequest
    if err := c.BindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": "invalid request"})
        return
    }

    userID := c.GetString("user_id")

    // 1. Trouver ou créer l’identité
    identity, err := h.Identities.GetByEmail(c, req.Email)
    if err != nil {
        // identité n’existe pas → on la crée
        identity, err = h.Identities.CreateIdentity(c, req.Email, "primary")
        if err != nil {
            c.JSON(500, gin.H{"error": "failed to create identity"})
            return
        }
    }

    // 2. Ajouter le user comme membre
    _ = h.Members.AddMember(c, identity.ID, userID)

    // 3. Insérer la clé publique
    key, err := h.PublicKeys.InsertPublicKey(
        c,
        identity.ID,
        req.ArmoredKey,
        req.WKDHash,
        true,
    )
    if err != nil {
        c.JSON(500, gin.H{"error": "failed to store public key"})
        return
    }

    c.JSON(200, gin.H{"id": key.ID})
}


type UploadPrivateKeyRequest struct {
    IdentityEmail       string `json:"identity_email"`
    EncryptedPrivateKey string `json:"encrypted_private_key"`
}

func (h *KeyHandler) UploadPrivateKey(c *gin.Context) {
    var req UploadPrivateKeyRequest
    if err := c.BindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": "invalid request"})
        return
    }

    userID := c.GetString("user_id")

    // 1. Trouver l’identité
    identity, err := h.Identities.GetByEmail(c, req.IdentityEmail)
    if err != nil {
        c.JSON(404, gin.H{"error": "identity not found"})
        return
    }

    // 2. Stocker la clé privée chiffrée
    key, err := h.PrivateKeys.InsertPrivateKey(
        c,
        identity.ID,
        userID,
        req.EncryptedPrivateKey,
    )
    if err != nil {
        c.JSON(500, gin.H{"error": "failed to store private key"})
        return
    }

    c.JSON(200, gin.H{"id": key.ID})
}


func (h *KeyHandler) GetPublicKey(c *gin.Context) {
    email := c.Param("email")

    identity, err := h.Identities.GetByEmail(c, email)
    if err != nil {
        c.JSON(404, gin.H{"error": "identity not found"})
        return
    }

    keys, err := h.PublicKeys.GetActiveKeysByIdentity(c, identity.ID)
    if err != nil || len(keys) == 0 {
        c.JSON(404, gin.H{"error": "no active public key"})
        return
    }

    key := keys[0] // la plus récente

    c.JSON(200, gin.H{
        "email":       email,
        "armored_key": key.ArmoredKey,
    })
}

func (h *KeyHandler) GetPrivateKey(c *gin.Context) {
    userID := c.GetString("user_id")

    // 1. Trouver les identités du user
    identities, err := h.Members.ListIdentitiesForUser(c, userID)
    if err != nil || len(identities) == 0 {
        c.JSON(404, gin.H{"error": "no identity for user"})
        return
    }

    identity := identities[0]

    // 2. Récupérer la clé privée chiffrée
    key, err := h.PrivateKeys.GetForUserIdentity(c, identity.ID, userID)
    if err != nil {
        c.JSON(404, gin.H{"error": "private key not found"})
        return
    }

    c.JSON(200, gin.H{
        "identity_email":       identity.Email,
        "encrypted_private_key": key.EncryptedPrivateKey,
    })
}

