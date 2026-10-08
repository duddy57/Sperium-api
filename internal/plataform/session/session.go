package session

import (
	"net/http"
	"time"

	"github.com/alexedwards/scs/goredisstore"
	"github.com/alexedwards/scs/v2"
	"github.com/redis/go-redis/v9"
)

func Client(redis *redis.Client) *scs.SessionManager {

	session := scs.New()
	session.Store = goredisstore.New(redis)
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = true

	return session
}
