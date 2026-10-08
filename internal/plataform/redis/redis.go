package redis

import (
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

func Connect(dsn string) (*redis.Client, error) {
	cfg, err := redis.ParseURL(dsn)
	if err != nil {
		return nil, fmt.Errorf("redis: failed to parse DSN: %w", err)
	}

	cfg.MaxActiveConns = 10
	cfg.ConnMaxLifetime = 30 * time.Minute

	return redis.NewClient(cfg), nil
}
