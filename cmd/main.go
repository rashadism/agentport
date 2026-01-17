package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"rca.agent/test/internal/config"
	"rca.agent/test/internal/handler"
	"rca.agent/test/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("Failed to load config", "error", err)
		os.Exit(1)
	}

	setupLogging(cfg.LogLevel)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Create service
	svc, err := service.NewAnalysisService(ctx, cfg)
	if err != nil {
		slog.Error("Failed to create service", "error", err)
		os.Exit(1)
	}

	// Create handler and routes
	h := handler.New(svc, cfg.AnalysisTimeout)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	// Create server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.ServerPort),
		Handler:      mux,
		ReadTimeout:  cfg.ReadTimeout, // TODO: update these
		WriteTimeout: cfg.WriteTimeout,
	}

	// Start server
	go func() {
		slog.Info("Server starting", "port", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Server error", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt
	<-ctx.Done()

	// Graceful shutdown
	slog.Info("Shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	svc.Close()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("Shutdown error", "error", err)
	}
}

func setupLogging(level string) {
	var logLevel slog.Level
	switch level {
	case "DEBUG":
		logLevel = slog.LevelDebug
	case "INFO":
		logLevel = slog.LevelInfo
	case "WARN", "WARNING":
		logLevel = slog.LevelWarn
	case "ERROR":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	h := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel})
	slog.SetDefault(slog.New(h))
	slog.Debug("Logging initialized", "level", level)
}
