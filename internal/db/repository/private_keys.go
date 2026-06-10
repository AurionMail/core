package repository

import (
    "context"

    "aurion/core/internal/db/generated"
    "github.com/google/uuid"
)

type PrivateKeyRepository struct {
    q *generated.Queries
}

func NewPrivateKeyRepository(q *generated.Queries) *PrivateKeyRepository {
    return &PrivateKeyRepository{q}
}

func (r *PrivateKeyRepository) InsertPrivateKey(
    ctx context.Context,
    userID string,
    armoredEncryptedKey string,
) (generated.PrivateKey, error) {

    uid, err := uuid.Parse(userID)
    if err != nil {
        return generated.PrivateKey{}, err
    }

    return r.q.InsertPrivateKey(ctx, generated.InsertPrivateKeyParams{
        UserID:              uid,
        ArmoredEncryptedKey: armoredEncryptedKey,
    })
}

func (r *PrivateKeyRepository) GetLatestPrivateKey(ctx context.Context, userID string) (generated.PrivateKey, error) {

    uid, err := uuid.Parse(userID)
    if err != nil {
        return generated.PrivateKey{}, err
    }

    return r.q.GetLatestPrivateKey(ctx, uid)
}
