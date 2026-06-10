package repository

import (
    "context"
    "aurion/core/internal/db/generated"
)

type UserRepository struct {
    q *generated.Queries
}

func NewUserRepository(q *generated.Queries) *UserRepository {
    return &UserRepository{q}
}

func (r *UserRepository) CreateUser(ctx context.Context, email, passwordHash string) (generated.User, error) {
    return r.q.CreateUser(ctx, generated.CreateUserParams{
        Email:        email,
        PasswordHash: passwordHash,
    })
}

func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (generated.User, error) {
    return r.q.GetUserByEmail(ctx, email)
}
