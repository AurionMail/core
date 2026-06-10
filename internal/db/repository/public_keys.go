package repository

import (
    "context"
    "aurion/core/internal/db/generated"
	"github.com/google/uuid"
)

type PublicKeyRepository struct {
    q *generated.Queries
}

func NewPublicKeyRepository(q *generated.Queries) *PublicKeyRepository {
    return &PublicKeyRepository{q}
}

func (r *PublicKeyRepository) InsertPublicKey(
    ctx context.Context,
    userID string,
    email string,
    armoredKey string,
    isPrimary bool,
) (generated.PublicKey, error) {
    uid, err := uuid.Parse(userID)
	if err != nil {
		return generated.PublicKey{}, err
	}

	return r.q.InsertPublicKey(ctx, generated.InsertPublicKeyParams{
		UserID:     uid,
		Email:      email,
		ArmoredKey: armoredKey,
		IsPrimary:  isPrimary,
	})
}

func (r *PublicKeyRepository) GetPrimaryPublicKey(ctx context.Context, email string) (generated.PublicKey, error) {
    return r.q.GetPrimaryPublicKey(ctx, email)
}
