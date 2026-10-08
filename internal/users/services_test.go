package users

import (
	"context"
	"testing"
	"time"
	"uuid"

	"github.com/alexedwards/scs/v2"
	usersRepo "github.com/duddy57/sperium/internal/users/queries"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type fakeRepository struct {
	createID   uuid.UUID
	createErr  error
	createArgs usersRepo.CreateUserQueryParams
	userByMail usersRepo.User
	mailErr    error
	userByID   usersRepo.User
	idErr      error
	updateArgs usersRepo.UpdateUserQueryParams
	updateErr  error
	deleteID   uuid.UUID
	deleteErr  error
}

func (f *fakeRepository) CreateUserQuery(_ context.Context, args usersRepo.CreateUserQueryParams) (uuid.UUID, error) {
	f.createArgs = args
	return f.createID, f.createErr
}
func (f *fakeRepository) GetUserByEmailQuery(_ context.Context, _ string) (usersRepo.User, error) {
	return f.userByMail, f.mailErr
}
func (f *fakeRepository) GetUserByIDQuery(_ context.Context, _ uuid.UUID) (usersRepo.User, error) {
	return f.userByID, f.idErr
}
func (f *fakeRepository) UpdateUserQuery(_ context.Context, args usersRepo.UpdateUserQueryParams) error {
	f.updateArgs = args
	return f.updateErr
}
func (f *fakeRepository) DeleteUserQuery(_ context.Context, id uuid.UUID) error {
	f.deleteID = id
	return f.deleteErr
}

func (f *fakeRepository) WithTx(_ pgx.Tx) *usersRepo.Queries {
	return nil
}

func testService(t *testing.T, repo *fakeRepository) (*Service, context.Context) {
	t.Helper()
	session := scs.New()
	ctx, err := session.Load(context.Background(), "")
	require.NoError(t, err)
	return NewUsersServices(repo, zap.NewNop(), session), ctx
}

func TestServiceCreateSuccess(t *testing.T) {
	id := uuid.New()
	repo := &fakeRepository{createID: id}
	service, ctx := testService(t, repo)

	got, err := service.Create(ctx, CreateUserReq{Name: "Ada", Email: "ada@example.com", Password: "password"})
	require.NoError(t, err)
	assert.Equal(t, id, got)
	assert.Equal(t, "Ada", repo.createArgs.Name)
	assert.Equal(t, "ada@example.com", repo.createArgs.Email)
	require.NoError(t, bcrypt.CompareHashAndPassword(repo.createArgs.PasswordHash, []byte("password")))
	cost, err := bcrypt.Cost(repo.createArgs.PasswordHash)
	require.NoError(t, err)
	assert.Equal(t, 12, cost)
	assert.Equal(t, id, service.session.Get(ctx, "user_id"))
}

func TestServiceCreateDuplicate(t *testing.T) {
	repo := &fakeRepository{createErr: &pgconn.PgError{Code: "23505"}}
	service, ctx := testService(t, repo)

	id, err := service.Create(ctx, CreateUserReq{Name: "Ada", Email: "ada@example.com", Password: "password"})
	assert.ErrorIs(t, err, ErrUserAlreadyExists)
	assert.Equal(t, uuid.Nil(), id)
}

func TestServiceLoginSuccess(t *testing.T) {
	id := uuid.New()
	hash, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.MinCost)
	require.NoError(t, err)
	repo := &fakeRepository{userByMail: usersRepo.User{ID: id, PasswordHash: hash}}
	service, ctx := testService(t, repo)

	require.NoError(t, service.Login(ctx, LoginUserReq{Email: "ada@example.com", Password: "password"}))
	assert.Equal(t, id, service.session.Get(ctx, "user_id"))
}

func TestServiceLoginUserAbsent(t *testing.T) {
	repo := &fakeRepository{mailErr: pgx.ErrNoRows}
	service, ctx := testService(t, repo)

	assert.ErrorIs(t, service.Login(ctx, LoginUserReq{Email: "missing@example.com", Password: "password"}), ErrUserNotFound)
}

func TestServiceLoginWrongPassword(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.MinCost)
	require.NoError(t, err)
	repo := &fakeRepository{userByMail: usersRepo.User{PasswordHash: hash}}
	service, ctx := testService(t, repo)

	assert.ErrorIs(t, service.Login(ctx, LoginUserReq{Email: "ada@example.com", Password: "wrongpass"}), ErrCredentialsNotFound)
}

func TestServiceDetailsWithoutSession(t *testing.T) {
	service, ctx := testService(t, &fakeRepository{})
	_, err := service.Details(ctx)
	assert.EqualError(t, err, "failed to get user ID")
}

func TestServiceDetailsUserAbsent(t *testing.T) {
	id := uuid.New()
	service, ctx := testService(t, &fakeRepository{idErr: pgx.ErrNoRows})
	service.session.Put(ctx, "user_id", id)

	_, err := service.Details(ctx)
	assert.ErrorIs(t, err, ErrUserNotFound)
}

func TestServiceUpdatePreservesNameWhenPayloadEmpty(t *testing.T) {
	id := uuid.New()
	repo := &fakeRepository{userByID: usersRepo.User{ID: id, Name: "Existing"}}
	service, ctx := testService(t, repo)
	service.session.Put(ctx, "user_id", id)

	require.NoError(t, service.Update(ctx, UpdateUserReq{}))
	assert.Equal(t, usersRepo.UpdateUserQueryParams{ID: id, Name: "Existing"}, repo.updateArgs)
}

func TestServiceDelete(t *testing.T) {
	id := uuid.New()
	repo := &fakeRepository{}
	service, ctx := testService(t, repo)
	service.session.Put(ctx, "user_id", id)

	require.NoError(t, service.Delete(ctx))
	assert.Equal(t, id, repo.deleteID)
}

func TestServiceDetailsMapsUser(t *testing.T) {
	id := uuid.New()
	now := time.Now()
	repo := &fakeRepository{userByID: usersRepo.User{ID: id, Name: "Ada", Email: "ada@example.com", CreatedAt: now, UpdatedAt: now}}
	service, ctx := testService(t, repo)
	service.session.Put(ctx, "user_id", id)

	got, err := service.Details(ctx)
	require.NoError(t, err)
	assert.Equal(t, Users{ID: id, Name: "Ada", Email: "ada@example.com", CreatedAt: now, UpdatedAt: now}, got)
}

var _ Repository = (*fakeRepository)(nil)
