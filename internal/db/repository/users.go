package repository

import (
	"aurion/core/internal/db/generated"
	"context"

	"github.com/google/uuid"
)

type UserRepository struct {
	q *generated.Queries
}

func NewUserRepository(q *generated.Queries) *UserRepository {
	return &UserRepository{q}
}

func (r *UserRepository) CreateUser(ctx context.Context, email, passwordHash, salt_s, salt_c string) (generated.User, error) {
	return r.q.CreateUser(ctx, generated.CreateUserParams{
		Email:        email,
		PasswordHash: passwordHash,
		SaltServer:   salt_s,
		SaltClient:   salt_c,
	})
}

func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (generated.User, error) {
	return r.q.GetUserByEmail(ctx, email)
}

func (r *UserRepository) GetUserById(ctx context.Context, id uuid.UUID) (generated.User, error) {
	return r.q.GetUserById(ctx, id)
}

func (r *UserRepository) UpdateUserByEmail(ctx context.Context, user generated.User) (generated.User, error) {
	return r.q.UpdateUserByEmail(ctx, generated.UpdateUserByEmailParams{
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
		SaltServer:   user.SaltServer,
		SaltClient:   user.SaltClient,
	})
}
