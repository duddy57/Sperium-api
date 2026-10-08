package organizations

import (
	"context"

	"uuid"

	organizationsRepo "github.com/duddy57/sperium/internal/organizations/queries"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	CreateOrganizationQuery(ctx context.Context, arg organizationsRepo.CreateOrganizationQueryParams) (uuid.UUID, error)
	DeleteOrganizationQuery(ctx context.Context, id uuid.UUID) error
	GetOrganizationQuery(ctx context.Context, id uuid.UUID) (organizationsRepo.GetOrganizationQueryRow, error)
	UpdateOrganizationQuery(ctx context.Context, arg organizationsRepo.UpdateOrganizationQueryParams) error
	WithTx(tx pgx.Tx) *organizationsRepo.Queries
	CreateMemberOwnerQuery(ctx context.Context, arg organizationsRepo.CreateMemberOwnerQueryParams) error
	GetMemberByIDQuery(ctx context.Context, id uuid.UUID) (organizationsRepo.GetMemberByIDQueryRow, error)
	ListMembersQuery(ctx context.Context, organizationID pgtype.UUID) ([]organizationsRepo.ListMembersQueryRow, error)
	RemoveMemberQuery(ctx context.Context, id uuid.UUID) error
	GetDefaultOrganizationQuery(ctx context.Context, ownerID pgtype.UUID) (organizationsRepo.GetDefaultOrganizationQueryRow, error)
}

var _ Repository = (*organizationsRepo.Queries)(nil)

func NewRepository(pool *pgxpool.Pool) Repository {
	return organizationsRepo.New(pool)
}
