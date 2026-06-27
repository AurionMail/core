package handlers

import (
	"strings"

	"github.com/gin-gonic/gin"

	"aurion/core/internal/db/repository"
)

type RoutingHandler struct {
	Identities *repository.IdentityRepository
	PublicKeys *repository.IdentityPublicKeyRepository
	Catchall   *repository.RoutingCatchallRepository
}

func NewRoutingHandler(
	identities *repository.IdentityRepository,
	pub *repository.IdentityPublicKeyRepository,
	catchall *repository.RoutingCatchallRepository,
) *RoutingHandler {
	return &RoutingHandler{identities, pub, catchall}
}

type ResolveRequest struct {
	Email string `json:"email"`
}

func (h *RoutingHandler) Resolve(c *gin.Context) {
	var req ResolveRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request"})
		return
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))

	// 1. Try direct identity match
	identity, err := h.Identities.GetByEmail(c, email)
	if err != nil {
		// 2. Try catch-all
		parts := strings.Split(email, "@")
		if len(parts) != 2 {
			c.JSON(400, gin.H{"error": "invalid email"})
			return
		}

		domain := parts[1]

		identity, err = h.Catchall.ResolveDomain(c, domain)
		if err != nil {
			c.JSON(404, gin.H{"error": "unknown address"})
			return
		}
	}

	// 3. Get public key
	pubKeys, err := h.PublicKeys.GetActiveKeysByIdentity(c, identity.ID)
	if err != nil || len(pubKeys) == 0 {
		c.JSON(500, gin.H{"error": "identity has no active public key"})
		return
	}
	pubKey := pubKeys[0]

	c.JSON(200, gin.H{
		"identity_email": identity.Email,
		"public_key":     pubKey.ArmoredKey,
	})
}
