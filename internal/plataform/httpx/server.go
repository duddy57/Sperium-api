package httpx

import (
	"context"
	"net/http"
	"time"

	"github.com/alexedwards/scs/goredisstore"
	"github.com/alexedwards/scs/v2"
	"github.com/duddy57/sperium/internal/machines"
	"github.com/duddy57/sperium/internal/organizations"
	"github.com/duddy57/sperium/internal/users"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func StartHTTPServer(logger *zap.Logger, ctx context.Context, pool *pgxpool.Pool, redis *redis.Client) http.Handler {
	session := scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = true
	session.Store = goredisstore.New(redis)

	validate := validator.New(validator.WithRequiredStructEnabled())

	usersHandlers := users.NewUserHandler(
		users.NewUsersServices(
			users.NewRepository(pool),
			logger,
			session),
		validate,
	)

	organizationsHandlers := organizations.NewOrganizationHandler(
		organizations.NewOrganizationServices(
			organizations.NewRepository(pool),
			pool,
			logger,
			session),
		validate,
	)

	machinesHandlers := machines.NewMachineHandler(
		machines.NewMachineServices(
			machines.NewRepository(pool),
			pool,
			logger,
			session,
		),
		validate,
	)

	r := NewRouter(
		logger,
		session,
		usersHandlers,
		organizationsHandlers,
		machinesHandlers,
	)

	return r
}
