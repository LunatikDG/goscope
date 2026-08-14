package server

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os/exec"
	"path/filepath"
)

// handleRun runs an instrumented example from ExamplesDir as a subprocess and
// streams its stdout (NDJSON, one line per event) line by line to the browser
// over Server-Sent Events. Deadlock examples naturally crash their subprocess
// via the Go runtime — that's an expected outcome, not a stream error.
func (s *Server) handleRun(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if !s.examples[name] {
		http.Error(w, "unknown example", http.StatusNotFound)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), s.cfg.RunTimeout)
	defer cancel()

	logger := s.logger.With(slog.String("example", name))

	examplePath := filepath.Join(s.cfg.ExamplesDir, name)
	if !filepath.IsAbs(examplePath) {
		examplePath = "./" + examplePath // without "./" go run looks for an import path, not a local folder
	}

	//nolint:gosec // name is clean: checked against the s.examples whitelist above, not raw input
	cmd := exec.CommandContext(ctx, "go", "run", examplePath)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	cmd.Stderr = &slogLineWriter{logger: s.logger, example: name} // the example's compile errors/panics go to the server log, not the client

	if err := cmd.Start(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	logger.Info("example started")

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)

	events := 0
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		events++
		fmt.Fprintf(w, "data: %s\n\n", scanner.Text())
		flusher.Flush()
	}

	waitErr := cmd.Wait() // the subprocess may have crashed with a deadlock — that's expected, not a stream error
	logger.Info("example finished", slog.Int("events", events), slog.Any("error", waitErr))

	fmt.Fprint(w, "event: end\ndata: {}\n\n")
	flusher.Flush()

	// The browser's EventSource reconnects on any connection drop, including a
	// clean close from the server — but we want exactly one run. So we hold the
	// connection open and wait for the client to disconnect itself (or for the
	// overall ctx timeout) — that way the close is client-initiated, no reconnect.
	<-ctx.Done()
}

// slogLineWriter turns bytes written to it into structured slog records, one
// per line. It buffers up to the next newline: a subprocess's Stderr delivers
// data in arbitrary chunks, not necessarily aligned to line boundaries.
type slogLineWriter struct {
	logger  *slog.Logger
	example string
	buf     []byte
}

func (w *slogLineWriter) Write(p []byte) (int, error) {
	w.buf = append(w.buf, p...)
	for {
		i := bytes.IndexByte(w.buf, '\n')
		if i < 0 {
			break
		}
		line := string(w.buf[:i])
		w.buf = w.buf[i+1:]
		if line != "" {
			w.logger.Warn("example stderr", slog.String("example", w.example), slog.String("line", line))
		}
	}
	return len(p), nil
}
