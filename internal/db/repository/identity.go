package repository

import (
    "context"
    "github.com/google/uuid"
    "aurion/core/internal/db/generated"
)


type IdentityRepository struct {
    q *generated.Queries
}

func NewIdentityRepository(q *generated.Queries) *IdentityRepository {
    return &IdentityRepository{q}
}

func (r *IdentityRepository) CreateIdentity(ctx context.Context, email, typ string) (generated.Identity, error) {
    return r.q.CreateIdentity(ctx, generated.CreateIdentityParams{
        Email: email,
        Type:  typ, // "primary" ou "shared"
    })
}

func (r *IdentityRepository) GetByEmail(ctx context.Context, email string) (generated.Identity, error) {
    return r.q.GetIdentityByEmail(ctx, email)
}


type IdentityMemberRepository struct {
    q *generated.Queries
}

func NewIdentityMemberRepository(q *generated.Queries) *IdentityMemberRepository {
    return &IdentityMemberRepository{q}
}

func (r *IdentityMemberRepository) AddMember(ctx context.Context, identityID uuid.UUID, userID string) error {
    uid, err := uuid.Parse(userID)
    if err != nil {
        return err
    }

    return r.q.AddIdentityMember(ctx, generated.AddIdentityMemberParams{
        IdentityID: identityID,
        UserID:     uid,
    })
}

func (r *IdentityMemberRepository) ListIdentitiesForUser(ctx context.Context, userID string) ([]generated.Identity, error) {
    uid, err := uuid.Parse(userID)
    if err != nil {
        return nil, err
    }

    return r.q.ListIdentitiesForUser(ctx, uid)
}
