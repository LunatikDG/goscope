// Package server содержит транспортный слой goscope: раздачу статики (веб-демо
// на WASM) и SSE-эндпоинт, который запускает инструментированные примеры из
// examples/ подпроцессом и стримит их события в браузер. Ничего не знает про
// internal/engine или internal/render — тем и раскладку кадров считает WASM
// в браузере, сервер лишь довозит сырые события.
package server

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"time"
)

// Server — HTTP-транспорт goscope: конфиг + логгер + белый список примеров,
// обнаруженных на диске.
type Server struct {
	logger   *slog.Logger
	examples map[string]bool
	cfg      Config
}

// New строит Server: сканирует ExamplesDir на предмет подпапок-примеров.
// Отсутствие или пустота каталога — не фатальная ошибка: статика продолжит
// раздаваться, а /api/run/* будет отвечать 404 всем именам.
//
//nolint:gocritic // Config копируется только на старте/в тестах, не в горячем пути
func New(cfg Config, logger *slog.Logger) *Server {
	examples, err := discoverExamples(cfg.ExamplesDir)
	if err != nil {
		logger.Warn("examples dir not readable — /api/run/* will 404 for every name",
			slog.String("dir", cfg.ExamplesDir), slog.Any("error", err))
		examples = map[string]bool{}
	} else {
		logger.Info("discovered instrumented examples", slog.Int("count", len(examples)))
	}

	return &Server{cfg: cfg, logger: logger, examples: examples}
}

func discoverExamples(dir string) (map[string]bool, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	examples := make(map[string]bool, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			examples[e.Name()] = true
		}
	}
	return examples, nil
}

// Handler собирает маршруты: статика веб-демо на "/" и live-стрим на /api/run/{name}.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(s.cfg.WebDir)))
	mux.HandleFunc("GET /api/run/{name}", s.handleRun)
	return s.withLogging(mux)
}

// Run поднимает HTTP-сервер и блокируется, пока не отменят ctx — тогда даёт
// текущим запросам (в том числе live-стримам) до 5с на корректное завершение.
func (s *Server) Run(ctx context.Context) error {
	httpSrv := &http.Server{
		Addr:              s.cfg.Addr,
		Handler:           s.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		s.logger.Info("listening", slog.String("addr", s.cfg.Addr))
		if err := httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		s.logger.Info("shutting down")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return httpSrv.Shutdown(shutdownCtx)
	}
}

// statusRecorder перехватывает код ответа для логирования, оставаясь
// http.Flusher — это критично для /api/run/{name}: без проброса Flush
// SSE-хендлер за этой обёрткой не смог бы стримить построчно.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *statusRecorder) Flush() {
	if f, ok := r.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func (s *Server) withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(rec, r)

		s.logger.Info("http request",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Int("status", rec.status),
			slog.Duration("elapsed", time.Since(start)),
		)
	})
}
