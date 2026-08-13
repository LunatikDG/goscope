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

// run возвращает код выхода вместо того, чтобы звать os.Exit напрямую —
// иначе defer stop() (отписка от сигналов) не успевал бы отработать.
func run() int {
	cfg, err := server.LoadConfig(os.Getenv)
	if err != nil {
		// логгера ещё нет: конфиг невалиден до того, как мы решили, во что логировать
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

//nolint:gocritic // Config копируется только один раз на старте процесса, не в горячем пути
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
