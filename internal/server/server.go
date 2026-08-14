// Package server holds goscope's transport layer: serving the static WASM demo
// and the SSE endpoint that runs instrumented examples from examples/ as a
// subprocess and streams their events to the browser. It knows nothing about
// internal/engine or internal/render — the WASM layer in the browser is what
// folds those events into frames; the server just delivers the raw events.
package server

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"time"
)

// Server — goscope's HTTP transport: config + logger + the whitelist of
// examples discovered on disk.
type Server struct {
	logger   *slog.Logger
	examples map[string]bool
	cfg      Config
}

// New builds a Server: scans ExamplesDir for example subdirectories. A
// missing or empty directory isn't fatal — static files keep being served,
// and /api/run/* just returns 404 for every name.
//
//nolint:gocritic // Config is only copied at startup/in tests, not on a hot path
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

// Handler wires up the routes: the WASM demo's static files on "/" and the
// live stream on /api/run/{name}.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(s.cfg.WebDir)))
	mux.HandleFunc("GET /api/run/{name}", s.handleRun)
	return s.withLogging(mux)
}

// Run starts the HTTP server and blocks until ctx is canceled — at which
// point it gives in-flight requests (including live streams) up to 5s to finish cleanly.
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

// statusRecorder captures the response status for logging while staying an
// http.Flusher — that matters for /api/run/{name}: without forwarding Flush,
// the SSE handler behind this wrapper couldn't stream line by line.
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
