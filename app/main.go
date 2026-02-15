package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	httpadapter "github.com/Vera-Kovaleva/subscriptions-service/internal/adapters/http"
	"github.com/Vera-Kovaleva/subscriptions-service/internal/domain"
	"github.com/Vera-Kovaleva/subscriptions-service/internal/infra/database"
	"github.com/Vera-Kovaleva/subscriptions-service/internal/infra/noerr"
	"github.com/Vera-Kovaleva/subscriptions-service/internal/repository"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

type Config struct {
	DBConnection    string
	ServerPort      string
	ShutdownTimeout time.Duration
}

func loadConfig() (*Config, error) {
	_ = godotenv.Load() // It's ok if .env doesn't exist

	cfg := &Config{
		DBConnection:    os.Getenv("DB_CONNECTION"),
		ServerPort:      getEnvOrDefault("SERVER_PORT", ":8080"),
		ShutdownTimeout: 10 * time.Second,
	}

	if cfg.DBConnection == "" {
		return nil, errors.New("DB_CONNECTION environment variable is required")
	}

	return cfg, nil
}

func getEnvOrDefault(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func setupLogger() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)
}

func main() {
	setupLogger()

	cfg, err := loadConfig()
	if err != nil {
		slog.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	provider := database.NewPostgresProvider(
		noerr.Must(pgxpool.New(ctx, cfg.DBConnection)),
	)
	defer provider.Close()

	subscriptionService := domain.NewSubscriptionService(provider, repository.NewSubscription())
	server := httpadapter.NewServer(subscriptionService)
	strictHandler := httpadapter.NewStrictHandler(server, nil)

	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		if err := provider.Execute(ctx, func(ctx context.Context, c domain.Connection) error {
			_, err := c.ExecContext(ctx, "SELECT 1")
			return err
		}); err != nil {
			slog.Error("Health check failed", "error", err)
			http.Error(w, "unhealthy", http.StatusServiceUnavailable)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("healthy"))
	})

	handler := httpadapter.HandlerWithOptions(strictHandler, httpadapter.StdHTTPServerOptions{
		BaseRouter: mux,
	})

	httpServer := &http.Server{
		Addr:           cfg.ServerPort,
		Handler:        handler,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1 MB
	}

	go func() {
		slog.Info("Server started", "port", cfg.ServerPort)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()

	slog.Info("Shutting down server gracefully...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
	}

	slog.Info("Server stopped gracefully")
}
