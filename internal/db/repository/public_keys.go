package repository

import (
    "context"
    "aurion/core/internal/db/generated"
	"github.com/google/uuid"
)

type IdentityPublicKeyRepository struct {
    q *generated.Queries
}

func NewIdentityPublicKeyRepository(q *generated.Queries) *IdentityPublicKeyRepository {
    return &IdentityPublicKeyRepository{q}
}

func (r *IdentityPublicKeyRepository) InsertPublicKey(
    ctx context.Context,
    identityID uuid.UUID,
    armoredKey string,
    wkdHash string,
    isActive bool,
) (generated.IdentityPublicKey, error) {


    return r.q.InsertIdentityPublicKey(ctx, generated.InsertIdentityPublicKeyParams{
        IdentityID: identityID,
        ArmoredKey: armoredKey,
        WkdHash:    wkdHash,
        IsActive:   isActive,
    })
}

func (r *IdentityPublicKeyRepository) GetActiveKeysByIdentity(ctx context.Context, identityID uuid.UUID) ([]generated.IdentityPublicKey, error) {
  
    return r.q.GetActiveIdentityPublicKeys(ctx, identityID)
}

func (r *IdentityPublicKeyRepository) GetByWKDHash(ctx context.Context, hash string) (*generated.IdentityPublicKey, error) {
    key, err := r.q.GetIdentityPublicKeyByWKDHash(ctx, hash)
    if err != nil {
        return nil, err
    }
    return &key, nil
}
