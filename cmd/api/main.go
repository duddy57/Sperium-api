package main

import (
	"context"
	"encoding/gob"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"uuid"

	"github.com/duddy57/sperium/internal/plataform/config"
	postgres "github.com/duddy57/sperium/internal/plataform/database"
	"github.com/duddy57/sperium/internal/plataform/httpx"
	"github.com/duddy57/sperium/internal/plataform/logging"
	"github.com/duddy57/sperium/internal/plataform/redis"
	"github.com/joho/godotenv"
	"github.com/klauspost/compress/gzhttp"
	"go.uber.org/zap"
)

func main() {
	gob.Register(uuid.UUID{})
	logger, err := logging.New("[SPERIUM_API]")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if err := godotenv.Load(); err != nil {
		logger.Error("service stopped", zap.String("error", err.Error()))
		os.Exit(1)
	}

	if err := run(logger); err != nil {
		logger.Error("service stopped", zap.String("error", err.Error()))
		os.Exit(1)
	}
}

func run(logger *zap.Logger) error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	dsnDb := config.String("DATABASE_URL",
		"postgres://postgres:postgres@localhost:5432/payments?sslmode=disable")
	dsnRedis := config.String("REDIS_URL",
		"redis://localhost:6379")

	pool, err := postgres.Connect(ctx, dsnDb)
	if err != nil {
		return err
	}
	defer pool.Close()
	postgres.RegisterPoolMetrics(pool)

	redisClient, err := redis.Connect(dsnRedis)
	if err != nil {
		return err
	}
	redis.RegisterRedisMetrics(redisClient.PoolStats())

	router := httpx.StartHTTPServer(logger, ctx, pool, redisClient)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", config.Int("SERVER_PORT", 8080)),
		Handler:      gzhttp.GzipHandler(router),
		IdleTimeout:  config.Duration("SERVER_IDLE_TIMEOUT", 60*time.Second),
		ReadTimeout:  config.Duration("SERVER_READ_TIMEOUT", 60*time.Second),
		WriteTimeout: config.Duration("SERVER_WRITE_TIMEOUT", 60*time.Second),
		ErrorLog:     zap.NewStdLog(logger.Named("sperium_http_error_logger")),
	}

	errChan := make(chan error, 2)
	go func() {
		logger.Info("starting api server", zap.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errChan <- err
		}
	}()

	select {
	case <-ctx.Done():
		logger.Info("shutdown signal received")
	case err := <-errChan:
		if err != nil {
			logger.Error("server stopped", zap.String("error", err.Error()))
		}
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()
		if shutdownErr := srv.Shutdown(shutdownCtx); shutdownErr != nil {
			return shutdownErr
		}
		return err
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return err
	}
	return nil

}
