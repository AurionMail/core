package repository

import (
    "context"

    "aurion/core/internal/db/generated"
)

type RoutingCatchallRepository struct {
    q *generated.Queries
}

func NewRoutingCatchallRepository(q *generated.Queries) *RoutingCatchallRepository {
    return &RoutingCatchallRepository{q}
}

func (r *RoutingCatchallRepository) ResolveDomain(ctx context.Context, domain string) (generated.Identity, error) {
    return r.q.GetCatchallIdentity(ctx, domain)
}
