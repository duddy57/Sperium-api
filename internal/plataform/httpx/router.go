package httpx

import (
	"net/http"

	"github.com/alexedwards/scs/v2"
	"github.com/duddy57/sperium/internal/machines"
	"github.com/duddy57/sperium/internal/organizations"
	"github.com/duddy57/sperium/internal/plataform/config"
	myhttputils "github.com/duddy57/sperium/internal/plataform/httpx/middlewares"
	"github.com/duddy57/sperium/internal/users"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/phenpessoa/gutils/netutils/httputils"
	"go.uber.org/zap"
)

func NewRouter(
	logger *zap.Logger,
	session *scs.SessionManager,
	usersHandler *users.Handler,
	organizationsHandler *organizations.Handler,
	machinesHandler *machines.Handler,
) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(session.LoadAndSave)
	r.Use(httputils.ChiLogger(logger))
	r.Use(cors.Handler(cors.Options{
		// AllowedOrigins:   []string{"https://foo.com"}, // Use this to allow specific origin hosts
		AllowedOrigins: []string{config.String("CORS_ORIGINS", "http://localhost:3000")},
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link", "Set-Cookie"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	r.Route("/api", func(r chi.Router) {
		r.Route("/users", func(r chi.Router) {
			usersHandler.RegisterRoutes(r, myhttputils.AuthMiddleware(session))
		})

		r.Route("/organizations", func(r chi.Router) {
			organizationsHandler.RegisterRoutes(r, myhttputils.AuthMiddleware(session))
		})

		r.Route("/machines", func(r chi.Router) {
			machinesHandler.RegisterRoutes(r, myhttputils.AuthMiddleware(session))
		})
	})

	return r
}
