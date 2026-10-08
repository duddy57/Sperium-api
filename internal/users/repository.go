package users

import (
	"context"

	"uuid"

	usersRepo "github.com/duddy57/sperium/internal/users/queries"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	CreateUserQuery(ctx context.Context, arg usersRepo.CreateUserQueryParams) (uuid.UUID, error)
	DeleteUserQuery(ctx context.Context, id uuid.UUID) error
	GetUserByEmailQuery(ctx context.Context, email string) (usersRepo.User, error)
	GetUserByIDQuery(ctx context.Context, id uuid.UUID) (usersRepo.User, error)
	UpdateUserQuery(ctx context.Context, arg usersRepo.UpdateUserQueryParams) error
	WithTx(tx pgx.Tx) *usersRepo.Queries
}

var _ Repository = (*usersRepo.Queries)(nil)

func NewRepository(pool *pgxpool.Pool) Repository {
	return usersRepo.New(pool)
}
