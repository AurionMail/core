package handlers

import (
    "net/http"
    "strings"

    "aurion/core/internal/db/repository"
    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
)

type MemberKeyInfo struct {
    UserID    string `json:"user_id"`
    PublicKey string `json:"public_key"` // La clé publique armored du membre
}

type RoutingSyncItem struct {
    IdentityID    string          `json:"identity_id"`
    Email         string          `json:"email"`
    Type          string          `json:"type"`
    NeedsKeyGen   bool            `json:"needs_key_gen"`
    NeedsKeyFetch bool            `json:"needs_key_fetch"`
    EncryptedKey  string          `json:"encrypted_private_key,omitempty"`
    WkdHash       string          `json:"wkd_hash,omitempty"`
    Members       []MemberKeyInfo `json:"members,omitempty"` // Structure enrichie avec les clés publiques
}

type SyncResponse struct {
    Identities []RoutingSyncItem `json:"identities"`
}

type KeyUploadPayload struct {
    IdentityID       string            `json:"identity_id"`
    ArmoredPublicKey string            `json:"armored_public_key"`
    WkdHash          string            `json:"wkd_hash"`
    Shares           []KeySharePayload `json:"shares"`
}

type KeySharePayload struct {
    UserID              string `json:"user_id"`
    EncryptedPrivateKey string `json:"encrypted_private_key"`
}

type SyncHandler struct {
    Identities  *repository.IdentityRepository
    PublicKeys  *repository.IdentityPublicKeyRepository
    PrivateKeys *repository.IdentityPrivateKeyRepository
    Members     *repository.IdentityMemberRepository
}

func NewSyncHandler(
    identities *repository.IdentityRepository,
    publicKeys *repository.IdentityPublicKeyRepository,
    privateKeys *repository.IdentityPrivateKeyRepository,
    members *repository.IdentityMemberRepository,
) *SyncHandler {
    return &SyncHandler{identities, publicKeys, privateKeys, members}
}

func (h *SyncHandler) SyncRouting(c *gin.Context) {
    // 1. Récupérer l'ID de l'utilisateur connecté via le middleware
    userID, exists := c.Get("user_id")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
        return
    }
    uidStr := userID.(string)

    // 2. Récupérer les identités rattachées au user via sa méthode existante : ListIdentitiesForUser
    userIdentities, err := h.Members.ListIdentitiesForUser(c, uidStr)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch identities"})
        return
    }

    var response SyncResponse
    response.Identities = make([]RoutingSyncItem, 0)

    for _, idnt := range userIdentities {
        item := RoutingSyncItem{
            IdentityID: idnt.ID.String(),
            Email:      idnt.Email,
            Type:       idnt.Type,
        }

        // 3. Vérifier les clés publiques actives via GetActiveKeysByIdentity
        pubKeys, err := h.PublicKeys.GetActiveKeysByIdentity(c, idnt.ID)
        if err != nil || len(pubKeys) == 0 {
            // Aucune clé publique trouvée -> Le SDK doit générer la paire de clés
            item.NeedsKeyGen = true

            // Récupérer la liste des membres associés pour en extraire leurs clés publiques personnelles
            dbMembers, _ := h.Members.ListMembersForIdentity(c, idnt.ID)
            item.Members = make([]MemberKeyInfo, 0, len(dbMembers))

            for _, m := range dbMembers {
                // Trouver l'identité 'primary' (personnelle) de chaque membre via son e-mail de connexion
                memberIdentity, err := h.Identities.GetByEmail(c, m.Email)
                if err != nil {
                    continue // Si l'identité n'est pas encore créée pour ce membre, on passe au suivant
                }

                // Récupérer la clé publique active de ce membre
                mPubKeys, err := h.PublicKeys.GetActiveKeysByIdentity(c, memberIdentity.ID)
                if err == nil && len(mPubKeys) > 0 {
                    item.Members = append(item.Members, MemberKeyInfo{
                        UserID:    m.ID.String(),
                        PublicKey: mPubKeys[0].ArmoredKey,
                    })
                }
            }
        } else {
            // Une clé publique existe, on prend la première active
            activeKey := pubKeys[0]
            item.WkdHash = activeKey.WkdHash

            // 4. Vérifier si l'enveloppe de clé privée existe pour cet utilisateur
            privKey, err := h.PrivateKeys.GetForUserIdentity(c, idnt.ID, uidStr)
            if err != nil {
                // L'enveloppe n'existe pas encore pour lui (un autre membre l'a peut-être générée sans l'inclure)
                item.NeedsKeyFetch = true
            } else {
                // Clé privée récupérée avec succès
                item.EncryptedKey = privKey.EncryptedPrivateKey
            }
        }

        response.Identities = append(response.Identities, item)
    }

    c.JSON(http.StatusOK, response)
}

func (h *SyncHandler) UploadSyncKeys(c *gin.Context) {
    var req KeyUploadPayload
    if err := c.BindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
        return
    }

    identityUUID, err := uuid.Parse(req.IdentityID)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid identity id format"})
        return
    }

    // 1. Sauvegarder la clé publique transmise par le SDK
    _, err = h.PublicKeys.InsertPublicKey(c, identityUUID, req.ArmoredPublicKey, req.WkdHash, true)
    if err != nil {
        // Si la clé existe déjà, on log ou on ignore selon la stratégie de verrouillage
        if !strings.Contains(err.Error(), "unique constraint") {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save public key"})
            return
        }
    }

    // 2. Insérer les enveloppes chiffrées pour chaque membre ciblé par le SDK
    for _, share := range req.Shares {
        _, _ = h.PrivateKeys.InsertPrivateKey(c, identityUUID, share.UserID, share.EncryptedPrivateKey)
    }

    c.JSON(http.StatusOK, gin.H{"status": "keys_synchronized"})
}