package machines

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"time"
	"uuid"

	"github.com/alexedwards/scs/v2"
	machinesRepo "github.com/duddy57/sperium/internal/machines/queries"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"
)

type Service struct {
	repo    Repository
	pool    *pgxpool.Pool
	logger  *zap.Logger
	session *scs.SessionManager
}

func NewMachineServices(repo Repository, pool *pgxpool.Pool, logger *zap.Logger, session *scs.SessionManager) *Service {
	return &Service{repo: repo, pool: pool, logger: logger, session: session}
}

func (s *Service) Create(ctx context.Context, req CreateMachineReq, orgID uuid.UUID) (uuid.UUID, string, error) {
	userID, ok := s.session.Get(ctx, "user_id").(uuid.UUID)
	if !ok {
		s.logger.Warn("failed to get user ID")
		return uuid.Nil(), "", errors.New("failed to get user ID")
	}

	rawToken := make([]byte, 32)
	if _, err := rand.Read(rawToken); err != nil {
		s.logger.Warn("failed to generate machine enrollment token", zap.Error(err))
		return uuid.Nil(), "", ErrMachineInternalServerError
	}
	token := base64.RawURLEncoding.EncodeToString(rawToken)
	hash := sha256.Sum256(rawToken)
	tokenHash := hex.EncodeToString(hash[:])
	expiresAt := time.Now().Add(3 * time.Hour)

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		s.logger.Warn("[ERROR] failed to init transaction", zap.Error(err))
		return uuid.Nil(), "", ErrMachineInternalServerError
	}
	defer func() { _ = tx.Rollback(ctx) }()

	qtx := s.repo.WithTx(tx)

	args := machinesRepo.RegisterMachineQueryParams{
		Name:           req.Name,
		OrganizationID: orgID,
	}

	machineID, err := qtx.RegisterMachineQuery(ctx, args)
	if err != nil {
		return uuid.Nil(), "", machineCreateError(s.logger, err)
	}
	if err := qtx.RegisterTokenMachineQuery(ctx, machinesRepo.RegisterTokenMachineQueryParams{
		MachineID:       machineID,
		TokenHash:       tokenHash,
		ExpiresAt:       expiresAt,
		CreatedByUserID: userID,
	}); err != nil {
		return uuid.Nil(), "", machineCreateError(s.logger, err)
	}

	if err := tx.Commit(ctx); err != nil {
		s.logger.Warn("failed to commit transaction", zap.Error(err))
		return uuid.Nil(), "", ErrMachineInternalServerError
	}
	return machineID, token, nil
}

func (s *Service) Details(ctx context.Context, machineID uuid.UUID) (Machine, error) {
	machine, err := s.repo.GetMachineQuery(ctx, machineID)
	if err != nil {
		s.logger.Warn("failed to find machine", zap.Error(err))
		if errors.Is(err, pgx.ErrNoRows) {
			return Machine{}, ErrMachineNotFound
		}
		return Machine{}, ErrMachineInternalServerError
	}

	return Machine{
		ID:             machine.ID,
		OrganizationID: machine.OrganizationID,
		Name:           machine.Name,
		Hostname:       machine.Hostname.String,
		AgentVersion:   machine.AgentVersion.String,
		LastSeenAt:     machine.LastSeenAt.Time,
		CreatedAt:      machine.CreatedAt,
		UpdatedAt:      machine.UpdatedAt,
	}, nil
}
func (s *Service) Delete(ctx context.Context, machine uuid.UUID) error {
	if err := s.repo.DeleteMachineQuery(ctx, machine); err != nil {
		s.logger.Warn("failed to delete machine", zap.Error(err))
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrMachineNotFound
		}
		return ErrMachineInternalServerError
	}

	return nil
}
func (s *Service) List(ctx context.Context, orgID uuid.UUID) ([]Machine, error) {

	rows, err := s.repo.ListMachineQuery(ctx, orgID)
	if err != nil {
		s.logger.Warn("failed to list machines", zap.Error(err))
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrMachineNotFound
		}
		return nil, ErrMachineInternalServerError
	}

	machines := Map[machinesRepo.Machine, Machine](rows, func(row machinesRepo.Machine) Machine {
		return Machine{
			ID:             row.ID,
			OrganizationID: row.OrganizationID,
			Name:           row.Name,
			Hostname:       row.Hostname.String,
			AgentVersion:   row.AgentVersion.String,
			LastSeenAt:     row.LastSeenAt.Time,
			CreatedAt:      row.CreatedAt,
			UpdatedAt:      row.UpdatedAt,
		}
	})

	return machines, nil
}
func (s *Service) Update(ctx context.Context, machineId uuid.UUID, req UpdateMachineReq) error {
	machine, err := s.repo.GetMachineQuery(ctx, machineId)
	if err != nil {
		s.logger.Warn("failed to find machine", zap.Error(err))
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrMachineNotFound
		}
		return ErrMachineInternalServerError
	}
	if req.Name == "" {
		req.Name = machine.Name
	}

	arg := machinesRepo.UpdateMachineNameQueryParams{
		ID:   machineId,
		Name: req.Name,
	}

	if err := s.repo.UpdateMachineNameQuery(ctx, arg); err != nil {
		s.logger.Warn("failed to update machine", zap.Error(err))
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrMachineNotFound
		}
		return ErrMachineInternalServerError
	}

	return nil
}
func (s *Service) Activate(ctx context.Context, machineID uuid.UUID, req ActivateMachineReq) error {
	rawToken, err := base64.RawURLEncoding.DecodeString(req.Token)
	if err != nil {
		s.logger.Warn("failed to decode base64 token", zap.Error(err))
		return ErrMachineNotFound
	}
	hash := sha256.Sum256(rawToken)

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		s.logger.Warn("failed to begin machine activation transaction", zap.Error(err))
		return ErrMachineInternalServerError
	}
	defer func() { _ = tx.Rollback(ctx) }()

	qtx := s.repo.WithTx(tx)
	if _, err := qtx.ConsumeMachineTokenQuery(ctx, machinesRepo.ConsumeMachineTokenQueryParams{
		TokenHash: hex.EncodeToString(hash[:]),
		MachineID: machineID,
	}); err != nil {
		s.logger.Warn("failed to consume token db", zap.Error(err))
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrMachineNotFound
		}
		return ErrMachineInternalServerError
	}

	if err := qtx.UpdateMachineQuery(ctx, machinesRepo.UpdateMachineQueryParams{
		ID:           machineID,
		Hostname:     pgtype.Text{String: req.Hostname, Valid: true},
		AgentVersion: pgtype.Text{String: req.AgentVersion, Valid: true},
		LastSeenAt:   pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}); err != nil {

		s.logger.Warn("failed to update machine", zap.Error(err))
		return ErrMachineInternalServerError
	}

	if err := tx.Commit(ctx); err != nil {
		s.logger.Warn("failed to commit machine activation", zap.Error(err))
		return ErrMachineInternalServerError
	}

	return nil
}

func Map[In any, Out any](
	data []In,
	mapper func(In) Out,
) []Out {
	list := make([]Out, 0, len(data))

	for _, item := range data {
		list = append(list, mapper(item))
	}

	return list
}

func machineCreateError(logger *zap.Logger, err error) error {
	logger.Warn("failed to create machine", zap.Error(err))
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return ErrMachineAlreadyExists
	}
	return ErrMachineInternalServerError
}
