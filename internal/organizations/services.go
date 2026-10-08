package organizations

import (
	"context"
	"errors"
	"strings"
	"uuid"

	"github.com/alexedwards/scs/v2"
	organizationsRepo "github.com/duddy57/sperium/internal/organizations/queries"
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

func NewOrganizationServices(repo Repository, pool *pgxpool.Pool, logger *zap.Logger, session *scs.SessionManager) *Service {
	return &Service{repo, pool, logger, session}
}

func (s *Service) Create(ctx context.Context, req CreateOrganizationReq) (uuid.UUID, string, error) {
	userID, ok := s.session.Get(ctx, "user_id").(uuid.UUID)
	if !ok {
		s.logger.Warn("[ERROR] failed to get user ID")
		return uuid.Nil(), "", errors.New("failed to get user ID")
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		s.logger.Warn("[ERROR] failed to init transaction", zap.Error(err))
		return uuid.Nil(), "", ErrOrganizationInternalServerError
	}
	defer func() { _ = tx.Rollback(ctx) }()

	slug := generateSlug(req.Name)

	qtx := s.repo.WithTx(tx)

	args := organizationsRepo.CreateOrganizationQueryParams{
		Name:   req.Name,
		Domain: req.Domain,
		Plan:   organizationsRepo.PlanstatusFREE,
		Slug:   slug,
		OwnerID: pgtype.UUID{
			Bytes: userID,
			Valid: true,
		},
	}

	organizationId, err := qtx.CreateOrganizationQuery(ctx, args)
	if err != nil {
		s.logger.Warn("failed to create organization", zap.Error(err))

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return uuid.Nil(), "", ErrOrganizationAlreadyExists
		}
		return uuid.Nil(), "", err
	}

	if err := qtx.CreateMemberOwnerQuery(ctx, organizationsRepo.CreateMemberOwnerQueryParams{
		UserID: pgtype.UUID{
			Bytes: userID,
			Valid: true,
		},
		OrganizationID: pgtype.UUID{
			Bytes: organizationId,
			Valid: true,
		},
		Role: organizationsRepo.RolestatusOWNER,
	}); err != nil {
		s.logger.Warn("failed to create member owner", zap.Error(err))
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return uuid.Nil(), "", ErrOrganizationInternalServerError
		}
	}

	if err := tx.Commit(ctx); err != nil {
		s.logger.Warn("failed to commit transaction", zap.Error(err))
		return uuid.Nil(), "", ErrOrganizationInternalServerError
	}

	return organizationId, slug, nil
}
func (s *Service) Details(ctx context.Context, organizationId uuid.UUID) (Organization, error) {
	organizations, err := s.repo.GetOrganizationQuery(ctx, organizationId)
	if err != nil {
		s.logger.Warn("[ERROR] failed to find user", zap.Error(err))
		if errors.Is(err, pgx.ErrNoRows) {
			return Organization{}, ErrOrganizationNotFound
		}
		return Organization{}, ErrOrganizationInternalServerError
	}

	rows, err := s.repo.ListMembersQuery(ctx, pgtype.UUID{
		Bytes: organizationId,
		Valid: true,
	})
	if err != nil {
		s.logger.Warn("[ERROR] failed to find user", zap.Error(err))
		if errors.Is(err, pgx.ErrNoRows) {
			return Organization{}, ErrOrganizationNotFound
		}
		return Organization{}, ErrOrganizationInternalServerError
	}

	members := Map[organizationsRepo.ListMembersQueryRow, Member](rows, func(row organizationsRepo.ListMembersQueryRow) Member {
		return Member{
			MemberID:  row.Memberid,
			ID:        row.ID.Bytes,
			Name:      row.Name.String,
			Email:     row.Email.String,
			Role:      string(row.Role),
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
		}
	})

	return Organization{
		ID:        organizations.ID,
		Name:      organizations.Name,
		Domain:    organizations.Domain,
		Slug:      organizations.Slug,
		Plan:      string(organizations.Plan),
		Members:   members,
		CreatedAt: organizations.CreatedAt,
		UpdatedAt: organizations.UpdatedAt,
	}, nil
}
func (s *Service) Delete(ctx context.Context, organizationId uuid.UUID) error {
	if err := s.repo.DeleteOrganizationQuery(ctx, organizationId); err != nil {
		s.logger.Warn("[ERROR] failed to delete organization", zap.Error(err))
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrOrganizationNotFound
		}
		return ErrOrganizationInternalServerError
	}

	return nil
}
func (s *Service) GetDefault(ctx context.Context) (Organization, error) {
	userID, ok := s.session.Get(ctx, "user_id").(uuid.UUID)
	if !ok {
		s.logger.Warn("[ERROR] failed to get user ID")
		return Organization{}, errors.New("failed to get user ID")
	}

	org, err := s.repo.GetDefaultOrganizationQuery(ctx, pgtype.UUID{Bytes: userID, Valid: true})
	if err != nil {
		s.logger.Warn("[ERROR] failed to get default organization", zap.Error(err))
		if errors.Is(err, pgx.ErrNoRows) {
			return Organization{}, ErrOrganizationNotFound

		}
	}

	rows, err := s.repo.ListMembersQuery(ctx, pgtype.UUID{
		Bytes: org.ID,
		Valid: true,
	})
	if err != nil {
		s.logger.Warn("[ERROR] failed to find user", zap.Error(err))
		if errors.Is(err, pgx.ErrNoRows) {
			return Organization{}, ErrOrganizationNotFound
		}
		return Organization{}, ErrOrganizationInternalServerError
	}

	members := Map[organizationsRepo.ListMembersQueryRow, Member](rows, func(row organizationsRepo.ListMembersQueryRow) Member {
		return Member{
			MemberID:  row.Memberid,
			ID:        row.ID.Bytes,
			Name:      row.Name.String,
			Email:     row.Email.String,
			Role:      string(row.Role),
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
		}
	})

	return Organization{
		ID:        org.ID,
		Name:      org.Name,
		Domain:    org.Domain,
		Slug:      org.Slug,
		Plan:      string(org.Plan),
		Members:   members,
		CreatedAt: org.CreatedAt,
		UpdatedAt: org.UpdatedAt,
	}, nil
}
func (s *Service) Update(ctx context.Context, organizationId uuid.UUID, req UpdateOrganizationReq) error {
	org, err := s.repo.GetOrganizationQuery(ctx, organizationId)
	if err != nil {
		s.logger.Warn("[ERROR] failed to find user", zap.Error(err))
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrOrganizationNotFound
		}
		return ErrOrganizationInternalServerError
	}

	var slug string
	if req.Name == "" {
		req.Name = org.Name
		slug = org.Slug
	} else {
		slug = generateSlug(req.Name)
	}
	if req.Domain == "" {
		req.Name = org.Domain
	}

	arg := organizationsRepo.UpdateOrganizationQueryParams{
		ID:     organizationId,
		Name:   req.Name,
		Domain: req.Domain,
		Slug:   slug,
	}

	if err := s.repo.UpdateOrganizationQuery(ctx, arg); err != nil {
		s.logger.Warn("[ERROR] failed to delete organization", zap.Error(err))
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrOrganizationNotFound
		}
		return ErrOrganizationInternalServerError
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

func generateSlug(name string) string {
	replacer := strings.NewReplacer(
		"á", "a", "à", "a", "â", "a", "ã", "a", "ä", "a",
		"é", "e", "è", "e", "ê", "e", "ë", "e",
		"í", "i", "ì", "i", "î", "i", "ï", "i",
		"ó", "o", "ò", "o", "ô", "o", "õ", "o", "ö", "o",
		"ú", "u", "ù", "u", "û", "u", "ü", "u",
		"ç", "c", "ñ", "n",
	)

	name = replacer.Replace(strings.ToLower(name))

	var slug strings.Builder
	var separator bool

	for _, char := range name {
		if (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') {
			if separator && slug.Len() > 0 {
				slug.WriteByte('-')
			}

			slug.WriteRune(char)
			separator = false
		} else {
			separator = true
		}
	}

	return slug.String()
}
