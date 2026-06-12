package repository

import (
    "context"

    "aurion/core/internal/db/generated"
    "github.com/google/uuid"
)

type IdentityPrivateKeyRepository struct {
    q *generated.Queries
}

func NewIdentityPrivateKeyRepository(q *generated.Queries) *IdentityPrivateKeyRepository {
    return &IdentityPrivateKeyRepository{q}
}

func (r *IdentityPrivateKeyRepository) InsertPrivateKey(
    ctx context.Context,
    identityID uuid.UUID,
    userID string,
    encryptedPrivateKey string,
) (generated.IdentityPrivateKey, error) {

    
    uid, err := uuid.Parse(userID)
    if err != nil {
        return generated.IdentityPrivateKey{}, err
    }

    return r.q.InsertIdentityPrivateKey(ctx, generated.InsertIdentityPrivateKeyParams{
        IdentityID:          identityID,
        UserID:              uid,
        EncryptedPrivateKey: encryptedPrivateKey,
    })
}

func (r *IdentityPrivateKeyRepository) GetForUserIdentity(
    ctx context.Context,
    identityID uuid.UUID,
    userID string,
) (generated.IdentityPrivateKey, error) {

   
    uid, err := uuid.Parse(userID)
    if err != nil {
        return generated.IdentityPrivateKey{}, err
    }

    return r.q.GetIdentityPrivateKey(ctx, generated.GetIdentityPrivateKeyParams{
        IdentityID: identityID,
        UserID:     uid,
    })
}
