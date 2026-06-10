package repository

import (
    "context"
    "time"
    "aurion/core/internal/db/generated"
	"github.com/google/uuid"
)

type SessionRepository struct {
    q *generated.Queries
}

func NewSessionRepository(q *generated.Queries) *SessionRepository {
    return &SessionRepository{q}
}

func (r *SessionRepository) CreateSession(
    ctx context.Context,
    userID string,
    token string,
    expiresAt time.Time,
) (generated.Session, error) {
    uid, err := uuid.Parse(userID)
	if err != nil {
		return generated.Session{}, err
	}

	return r.q.CreateSession(ctx, generated.CreateSessionParams{
		UserID:    uid,
		Token:     token,
		ExpiresAt: expiresAt,
	})

}

func (r *SessionRepository) GetSessionByToken(ctx context.Context, token string) (generated.Session, error) {
    return r.q.GetSessionByToken(ctx, token)
}
