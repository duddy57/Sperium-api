package session

import (
	"net/http"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func TestClient(t *testing.T) {
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	manager := Client(redisClient)

	assert.NotNil(t, manager)
	assert.NotNil(t, manager.Store)

	assert.Equal(t, 24*time.Hour, manager.Lifetime)
	assert.True(t, manager.Cookie.Persist)
	assert.True(t, manager.Cookie.Secure)
	assert.Equal(t, http.SameSiteLaxMode, manager.Cookie.SameSite)
}
