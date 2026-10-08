package redis

import (
	"context"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"
)

func TestConnect_Errors(t *testing.T) {
	t.Run("invalid DSN", func(t *testing.T) {
		pool, err := Connect("redis://invalid uri:::")
		require.Error(t, err)
		assert.Nil(t, pool)
		assert.Contains(t, err.Error(), "failed to parse DSN")
	})

	t.Run("unreachable connection is reported by Ping", func(t *testing.T) {
		pool, err := Connect("redis://localhost:0")
		require.NoError(t, err)
		require.NotNil(t, pool)
		defer func() { _ = pool.Close() }()

		assert.Error(t, pool.Ping(context.Background()).Err())
	})

}
func TestConnect_Success_And_Metrics(t *testing.T) {
	ctx := context.Background()
	redisContainer, err := tcredis.Run(ctx, "redis:7-alpine")
	if err != nil {
		t.Skipf("skipping testcontainer test: %v", err)
	}
	defer func() {
		_ = redisContainer.Terminate(ctx)
	}()

	connStr, err := redisContainer.ConnectionString(ctx)
	require.NoError(t, err)

	pool, err := Connect(connStr)
	require.NoError(t, err)
	assert.NotNil(t, pool)
	defer func() { _ = pool.Close() }()

	assert.NoError(t, pool.Ping(ctx).Err())

	assert.NotPanics(t, func() {
		RegisterRedisMetrics(pool.PoolStats())
	})

	_, err = prometheus.DefaultGatherer.Gather()
	assert.NoError(t, err)
}
