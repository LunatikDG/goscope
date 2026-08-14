package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/LunatikDG/goscope/internal/server"
)

func main() {
	os.Exit(run())
}

// run returns an exit code instead of calling os.Exit directly — otherwise
// defer stop() (unsubscribing from signals) wouldn't get a chance to run.
func run() int {
	cfg, err := server.LoadConfig(os.Getenv)
	if err != nil {
		// there's no logger yet: the config is invalid before we've even decided what to log to
		slog.Error("invalid config", slog.Any("error", err))
		return 1
	}

	logger := newLogger(cfg)
	srv := server.New(cfg, logger)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := srv.Run(ctx); err != nil {
		logger.Error("server stopped", slog.Any("error", err))
		return 1
	}
	return 0
}

//nolint:gocritic // Config is only copied once at process startup, not on a hot path
func newLogger(cfg server.Config) *slog.Logger {
	opts := &slog.HandlerOptions{Level: cfg.LogLevel}

	var handler slog.Handler
	if cfg.LogFormat == "json" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}
	return slog.New(handler)
}
