package organizations

import (
	"context"
	"errors"
	"testing"
	"time"
	"uuid"

	"github.com/alexedwards/scs/v2"
	organizationsRepo "github.com/duddy57/sperium/internal/organizations/queries"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type fakeOrganizationRepository struct {
	org          organizationsRepo.GetOrganizationQueryRow
	orgErr       error
	defaultOrg   organizationsRepo.GetDefaultOrganizationQueryRow
	defaultErr   error
	members      []organizationsRepo.ListMembersQueryRow
	membersErr   error
	deleteErr    error
	deletedID    uuid.UUID
	updateErr    error
	updateParams organizationsRepo.UpdateOrganizationQueryParams
}

func (f *fakeOrganizationRepository) CreateOrganizationQuery(context.Context, organizationsRepo.CreateOrganizationQueryParams) (uuid.UUID, error) {
	return uuid.Nil(), nil
}
func (f *fakeOrganizationRepository) DeleteOrganizationQuery(_ context.Context, id uuid.UUID) error {
	f.deletedID = id
	return f.deleteErr
}
func (f *fakeOrganizationRepository) GetOrganizationQuery(context.Context, uuid.UUID) (organizationsRepo.GetOrganizationQueryRow, error) {
	return f.org, f.orgErr
}
func (f *fakeOrganizationRepository) UpdateOrganizationQuery(_ context.Context, params organizationsRepo.UpdateOrganizationQueryParams) error {
	f.updateParams = params
	return f.updateErr
}
func (f *fakeOrganizationRepository) WithTx(_ pgx.Tx) *organizationsRepo.Queries { return nil }
func (f *fakeOrganizationRepository) CreateMemberOwnerQuery(context.Context, organizationsRepo.CreateMemberOwnerQueryParams) error {
	return nil
}
func (f *fakeOrganizationRepository) GetMemberByIDQuery(context.Context, uuid.UUID) (organizationsRepo.GetMemberByIDQueryRow, error) {
	return organizationsRepo.GetMemberByIDQueryRow{}, nil
}
func (f *fakeOrganizationRepository) ListMembersQuery(context.Context, pgtype.UUID) ([]organizationsRepo.ListMembersQueryRow, error) {
	return f.members, f.membersErr
}
func (f *fakeOrganizationRepository) RemoveMemberQuery(context.Context, uuid.UUID) error { return nil }
func (f *fakeOrganizationRepository) GetDefaultOrganizationQuery(context.Context, pgtype.UUID) (organizationsRepo.GetDefaultOrganizationQueryRow, error) {
	return f.defaultOrg, f.defaultErr
}

func newOrganizationService(t *testing.T, repo Repository) (*Service, context.Context) {
	t.Helper()
	session := scs.New()
	ctx, err := session.Load(context.Background(), "")
	require.NoError(t, err)
	return NewOrganizationServices(repo, nil, zap.NewNop(), session), ctx
}

func TestServiceDetailsMapsOrganizationAndMembers(t *testing.T) {
	orgID, memberID, userID := uuid.New(), uuid.New(), uuid.New()
	now := time.Now()
	repo := &fakeOrganizationRepository{
		org:     organizationsRepo.GetOrganizationQueryRow{ID: orgID, Name: "Acme", Domain: "acme.test", Slug: "acme", Plan: organizationsRepo.PlanstatusPRO, CreatedAt: now, UpdatedAt: now},
		members: []organizationsRepo.ListMembersQueryRow{{Memberid: memberID, ID: pgtype.UUID{Bytes: userID, Valid: true}, Name: pgtype.Text{String: "Ada", Valid: true}, Email: pgtype.Text{String: "ada@acme.test", Valid: true}, Role: organizationsRepo.RolestatusOWNER, CreatedAt: now, UpdatedAt: now}},
	}
	service, ctx := newOrganizationService(t, repo)

	got, err := service.Details(ctx, orgID)
	require.NoError(t, err)
	assert.Equal(t, orgID, got.ID)
	assert.Equal(t, "Acme", got.Name)
	require.Len(t, got.Members, 1)
	assert.Equal(t, Member{MemberID: memberID, ID: userID, Name: "Ada", Email: "ada@acme.test", Role: "OWNER", CreatedAt: now, UpdatedAt: now}, got.Members[0])
}

func TestServiceDetailsMapsRepositoryErrors(t *testing.T) {
	service, ctx := newOrganizationService(t, &fakeOrganizationRepository{orgErr: pgx.ErrNoRows})
	_, err := service.Details(ctx, uuid.New())
	assert.ErrorIs(t, err, ErrOrganizationNotFound)

	service, ctx = newOrganizationService(t, &fakeOrganizationRepository{orgErr: errors.New("database unavailable")})
	_, err = service.Details(ctx, uuid.New())
	assert.ErrorIs(t, err, ErrOrganizationInternalServerError)
}

func TestServiceDetailsMapsMemberErrors(t *testing.T) {
	service, ctx := newOrganizationService(t, &fakeOrganizationRepository{org: organizationsRepo.GetOrganizationQueryRow{ID: uuid.New()}, membersErr: pgx.ErrNoRows})
	_, err := service.Details(ctx, uuid.New())
	assert.ErrorIs(t, err, ErrOrganizationNotFound)
}

func TestServiceDeleteMapsErrorsAndPassesID(t *testing.T) {
	id := uuid.New()
	repo := &fakeOrganizationRepository{}
	service, ctx := newOrganizationService(t, repo)
	require.NoError(t, service.Delete(ctx, id))
	assert.Equal(t, id, repo.deletedID)

	repo.deleteErr = pgx.ErrNoRows
	assert.ErrorIs(t, service.Delete(ctx, id), ErrOrganizationNotFound)
	repo.deleteErr = errors.New("database unavailable")
	assert.ErrorIs(t, service.Delete(ctx, id), ErrOrganizationInternalServerError)
}

func TestServiceUpdateBuildsSlugAndMapsErrors(t *testing.T) {
	id := uuid.New()
	repo := &fakeOrganizationRepository{org: organizationsRepo.GetOrganizationQueryRow{ID: id, Name: "Old", Domain: "old.test", Slug: "old"}}
	service, ctx := newOrganizationService(t, repo)
	require.NoError(t, service.Update(ctx, id, UpdateOrganizationReq{Name: "São José", Domain: "new.test"}))
	assert.Equal(t, organizationsRepo.UpdateOrganizationQueryParams{ID: id, Name: "São José", Domain: "new.test", Slug: "sao-jose"}, repo.updateParams)

	repo.orgErr = pgx.ErrNoRows
	assert.ErrorIs(t, service.Update(ctx, id, UpdateOrganizationReq{Name: "New", Domain: "new.test"}), ErrOrganizationNotFound)
	repo.orgErr = nil
	repo.updateErr = errors.New("database unavailable")
	assert.ErrorIs(t, service.Update(ctx, id, UpdateOrganizationReq{Name: "New", Domain: "new.test"}), ErrOrganizationInternalServerError)
}

func TestServiceSessionRequired(t *testing.T) {
	service, ctx := newOrganizationService(t, &fakeOrganizationRepository{})
	_, _, err := service.Create(ctx, CreateOrganizationReq{Name: "Acme", Domain: "acme.test"})
	assert.EqualError(t, err, "failed to get user ID")
	_, err = service.GetDefault(ctx)
	assert.EqualError(t, err, "failed to get user ID")
}

func TestGenerateSlug(t *testing.T) {
	assert.Equal(t, "sao-jose-42", generateSlug("São José 42"))
}

var _ Repository = (*fakeOrganizationRepository)(nil)
