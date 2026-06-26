package handlers

import (
	"aurion/core/internal/db/repository"

	"aurion/core/internal/wkd"

	"github.com/gin-gonic/gin"
)

type KeyHandler struct {
	Identities  *repository.IdentityRepository
	PublicKeys  *repository.IdentityPublicKeyRepository
	PrivateKeys *repository.IdentityPrivateKeyRepository
	Members     *repository.IdentityMemberRepository
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
	wkd := wkd.WKDHash(req.Email)
	key, err := h.PublicKeys.InsertPublicKey(
		c,
		identity.ID,
		req.ArmoredKey,
		wkd,
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

type PrivateKeyResponseItem struct {
	IdentityEmail       string `json:"identity_email"`
	EncryptedPrivateKey string `json:"encrypted_private_key"`
}

// GetPrivateKey renvoie désormais TOUTES les clés privées chiffrées des identités de l'utilisateur.
func (h *KeyHandler) GetPrivateKey(c *gin.Context) {
	userID := c.GetString("user_id")

	// 1. Trouver TOUTES les identités auxquelles le user appartient (alias, groupes, primary, etc.)
	identities, err := h.Members.ListIdentitiesForUser(c, userID)
	if err != nil || len(identities) == 0 {
		c.JSON(404, gin.H{"error": "no identity found for this user"})
		return
	}

	var privateKeysList []PrivateKeyResponseItem

	// 2. Boucler sur chaque identité pour collecter sa clé privée correspondante
	for _, identity := range identities {
		key, err := h.PrivateKeys.GetForUserIdentity(c, identity.ID, userID)
		if err != nil {
			// Si une clé privée n'est pas encore générée ou disponible pour une identité spécifique,
			// on l'ignore silencieusement pour ne pas bloquer la récupération des autres clés.
			continue
		}

		privateKeysList = append(privateKeysList, PrivateKeyResponseItem{
			IdentityEmail:       identity.Email,
			EncryptedPrivateKey: key.EncryptedPrivateKey,
		})
	}

	// Si aucune clé privée n'a pu être trouvée dans toute la liste
	if len(privateKeysList) == 0 {
		c.JSON(404, gin.H{"error": "no private keys found"})
		return
	}

	c.JSON(200, gin.H{
		"keys": privateKeysList,
	})
}
