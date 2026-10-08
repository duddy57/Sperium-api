package users

import (
	"context"
	"errors"
	"uuid"

	"github.com/alexedwards/scs/v2"
	usersRepo "github.com/duddy57/sperium/internal/users/queries"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	repo    Repository
	logger  *zap.Logger
	session *scs.SessionManager
}

func NewUsersServices(repo Repository, logger *zap.Logger, session *scs.SessionManager) *Service {
	return &Service{repo, logger, session}
}

func (s *Service) Create(ctx context.Context, req CreateUserReq) (uuid.UUID, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		s.logger.Warn("failed to hash password", zap.Error(err))
		return uuid.Nil(), err
	}
	args := usersRepo.CreateUserQueryParams{
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: hash,
	}

	userId, err := s.repo.CreateUserQuery(ctx, args)
	if err != nil {
		s.logger.Warn("failed to create user", zap.Error(err))

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return uuid.Nil(), ErrUserAlreadyExists
		}
		return uuid.Nil(), err
	}

	err = s.session.RenewToken(ctx)
	if err != nil {
		s.logger.Warn("failed to hash password", zap.Error(err))
		return uuid.Nil(), err
	}

	s.session.Put(ctx, "user_id", userId)

	return userId, nil
}
func (s *Service) Login(ctx context.Context, payload LoginUserReq) error {
	user, err := s.repo.GetUserByEmailQuery(ctx, payload.Email)
	if err != nil {
		s.logger.Warn("[ERROR] find user on database", zap.Error(err))
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrUserNotFound
		}

		return ErrUserAlreadyExists
	}

	err = bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(payload.Password))
	if err != nil {
		s.logger.Warn("[ERROR] User compare password", zap.Error(err))
		return ErrCredentialsNotFound
	}

	err = s.session.RenewToken(ctx)
	if err != nil {
		s.logger.Warn("Failed to renew token", zap.Error(err))
		return ErrUserAlreadyExists
	}

	s.session.Put(ctx, "user_id", user.ID)

	return nil
}
func (s *Service) Details(ctx context.Context) (Users, error) {
	userID, ok := s.session.Get(ctx, "user_id").(uuid.UUID)
	if !ok {
		s.logger.Warn("[ERROR] failed to get user ID")
		return Users{}, errors.New("failed to get user ID")
	}
	user, err := s.repo.GetUserByIDQuery(ctx, userID)
	if err != nil {
		s.logger.Warn("[ERROR] failed to find user", zap.Error(err))
		if errors.Is(err, pgx.ErrNoRows) {
			return Users{}, ErrUserNotFound
		}
		return Users{}, ErrUserAlreadyExists
	}

	return Users{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}
func (s *Service) Logout(ctx context.Context) error {
	err := s.session.RenewToken(ctx)
	if err != nil {
		s.logger.Warn("[ERROR] Failed to renew token", zap.Error(err))
		return ErrUserAlreadyExists
	}

	s.session.Remove(ctx, "user_id")

	return nil
}
func (s *Service) Delete(ctx context.Context) error {
	userID, ok := s.session.Get(ctx, "user_id").(uuid.UUID)
	if !ok {
		s.logger.Warn("[ERROR] failed to get user ID")
		return errors.New("failed to get user ID")
	}

	if err := s.repo.DeleteUserQuery(ctx, userID); err != nil {
		s.logger.Warn("[ERROR] failed to delete user", zap.Error(err))
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrUserNotFound
		}
		return err
	}

	if err := s.session.Destroy(ctx); err != nil {
		s.logger.Error(
			"user deleted but failed to destroy session",
			zap.Error(err),
		)
		return err
	}

	return nil
}
func (s *Service) Update(ctx context.Context, payload UpdateUserReq) error {
	userID, ok := s.session.Get(ctx, "user_id").(uuid.UUID)
	if !ok {
		s.logger.Warn("[ERROR] failed to get user ID")
		return errors.New("failed to get user ID")
	}
	user, err := s.repo.GetUserByIDQuery(ctx, userID)
	if err != nil {
		s.logger.Warn("[ERROR] failed to find user", zap.Error(err))
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrUserNotFound
		}
		return ErrUserAlreadyExists
	}

	if payload.Name == "" {
		payload.Name = user.Name
	}

	if err := s.repo.UpdateUserQuery(ctx, usersRepo.UpdateUserQueryParams{
		ID:   userID,
		Name: payload.Name,
	}); err != nil {
		s.logger.Warn("[ERROR] failed to update user", zap.Error(err))
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrUserNotFound
		}
		return err
	}

	return nil
}
