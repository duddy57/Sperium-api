package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
)

func TestConnect_Errors(t *testing.T) {
	ctx := context.Background()

	t.Run("invalid DSN", func(t *testing.T) {
		pool, err := Connect(ctx, "postgres://invalid uri:::")
		assert.Error(t, err)
		assert.Nil(t, pool)
		assert.Contains(t, err.Error(), "failed to parse dsn")
	})

	t.Run("unreachable connection error", func(t *testing.T) {
		ctxTimeout, cancel := context.WithTimeout(ctx, 1*time.Second)
		defer cancel()

		pool, err := Connect(ctxTimeout, "postgres://postgres:postgres@127.0.0.1:59999/testdb?sslmode=disable")
		assert.Error(t, err)
		assert.Nil(t, pool)
	})
}

func TestConnect_Success_And_Metrics(t *testing.T) {
	ctx := context.Background()
	pgContainer, err := tcpostgres.Run(ctx, "postgres:16-alpine",
		tcpostgres.WithDatabase("testdb"),
		tcpostgres.WithUsername("postgres"),
		tcpostgres.WithPassword("postgres"),
		tcpostgres.BasicWaitStrategies(),
	)
	if err != nil {
		t.Skipf("skipping testcontainer test: %v", err)
	}
	defer func() {
		_ = pgContainer.Terminate(ctx)
	}()

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	pool, err := Connect(ctx, connStr)
	require.NoError(t, err)
	require.NotNil(t, pool)
	defer pool.Close()

	assert.NoError(t, pool.Ping(ctx))

	// Test RegisterPoolMetrics
	assert.NotPanics(t, func() {
		RegisterPoolMetrics(pool)
	})

	_, err = prometheus.DefaultGatherer.Gather()
	assert.NoError(t, err)
}
