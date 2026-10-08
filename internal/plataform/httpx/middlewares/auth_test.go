package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alexedwards/scs/v2"
	"github.com/stretchr/testify/assert"
)

func TestAuthMiddlewareWithoutUserIDReturnsUnauthorizedJSON(t *testing.T) {
	sessions := scs.New()
	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/users/me", nil)
	recorder := httptest.NewRecorder()

	sessions.LoadAndSave(AuthMiddleware(sessions)(next)).ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
	assert.Equal(t, `{"error":"unauthorized"}
`, recorder.Body.String())
	assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))
	assert.False(t, nextCalled)
}
