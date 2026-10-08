package middlewares

import (
	"net/http"

	"github.com/alexedwards/scs/v2"
	"github.com/duddy57/sperium/internal/plataform/json_utils"
)

const AuthenticatedUserIDKey = "user_id"

func AuthMiddleware(sessions *scs.SessionManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !sessions.Exists(r.Context(), AuthenticatedUserIDKey) {
				json_utils.WriteError(w, http.StatusUnauthorized, "unauthorized")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
