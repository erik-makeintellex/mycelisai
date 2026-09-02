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

	"github.com/mycelis/framework-runs/internal/auth"
	"github.com/mycelis/framework-runs/internal/config"
	"github.com/mycelis/framework-runs/internal/controller"
	"github.com/mycelis/framework-runs/internal/httpapi"
	"github.com/mycelis/framework-runs/internal/journal"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil)).With("component", "framework-runs")
	settings, err := config.FromEnv()
	if err != nil {
		logger.Error("configuration rejected", "phase", "startup")
		os.Exit(1)
	}
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	repository, err := journal.OpenPostgres(ctx, settings.DatabaseURL)
	if err != nil {
		logger.Error("database unavailable", "phase", "startup", "error_type", "database_unready")
		os.Exit(1)
	}
	defer repository.Close()
	credential, err := auth.NewCredential("mycelis-core", settings.CoreToken, "runs:api")
	if err != nil {
		logger.Error("service credential rejected", "phase", "startup")
		os.Exit(1)
	}
	authenticator, _ := auth.New(credential)

	// Slice B intentionally has no production executor. The controller and
	// journal are ready for authenticated inspection, while POST create fails
	// before durable acceptance until a certified adapter is injected later.
	service := controller.New(repository, nil)
	service.MaxRuns = settings.MaxRuns
	service.LeaseDuration = settings.LeaseDuration
	server := &http.Server{
		Addr:              settings.ListenAddress,
		Handler:           httpapi.New(service, authenticator),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      0,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    16 * 1024,
	}
	go func() {
		<-ctx.Done()
		shutdown, stop := context.WithTimeout(context.Background(), 10*time.Second)
		defer stop()
		_ = server.Shutdown(shutdown)
	}()
	logger.Info("service listening", "address", settings.ListenAddress, "production_ready", false)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("service stopped unexpectedly", "error_type", "http_server_failure")
		os.Exit(1)
	}
}
