package machines

import (
	"context"

	"uuid"

	machinesRepo "github.com/duddy57/sperium/internal/machines/queries"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	WithTx(tx pgx.Tx) *machinesRepo.Queries
	DeleteMachineQuery(ctx context.Context, id uuid.UUID) error
	GetMachineQuery(ctx context.Context, id uuid.UUID) (machinesRepo.Machine, error)
	ConsumeMachineTokenQuery(ctx context.Context, arg machinesRepo.ConsumeMachineTokenQueryParams) (machinesRepo.ConsumeMachineTokenQueryRow, error)
	ListMachineQuery(ctx context.Context, organizationID uuid.UUID) ([]machinesRepo.Machine, error)
	RegisterMachineQuery(ctx context.Context, arg machinesRepo.RegisterMachineQueryParams) (uuid.UUID, error)
	RegisterTokenMachineQuery(ctx context.Context, arg machinesRepo.RegisterTokenMachineQueryParams) error
	UpdateMachineNameQuery(ctx context.Context, arg machinesRepo.UpdateMachineNameQueryParams) error
	UpdateMachineQuery(ctx context.Context, arg machinesRepo.UpdateMachineQueryParams) error
}

var _ Repository = (*machinesRepo.Queries)(nil)

func NewRepository(pool *pgxpool.Pool) Repository {
	return machinesRepo.New(pool)
}
