package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"time"

	"github.com/LunatikDG/goscope/internal/engine"
	"github.com/LunatikDG/goscope/internal/traceparse"
)

// wireStep and wireScene are the JSON shape served at GET /api/trace.
// cmd/webdemo decodes this independently with its own matching types — the
// same deliberate decoupling internal/instrument.Event and webdemo's
// liveEvent already use for the live-streaming wire format, so the two
// binaries never need to share a Go package just to agree on a schema.
type wireStep struct {
	Event     string `json:"event"`
	Label     string `json:"label,omitempty"`
	Goroutine int    `json:"goroutine"`
	Chan      int    `json:"chan,omitempty"`
}

type wireScene struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Steps       []wireStep `json:"steps"`
}

func toWireScene(s engine.Scene) wireScene {
	steps := make([]wireStep, len(s.Steps))
	for i, st := range s.Steps {
		steps[i] = wireStep{Event: st.Event.String(), Label: st.Label, Goroutine: st.Goroutine, Chan: st.Chan}
	}
	return wireScene{Name: s.Name, Description: s.Description, Steps: steps}
}

// newTraceMux serves the static WASM demo from webDir on "/" and the given
// pre-encoded scene JSON on GET /api/trace — the same static assets
// `make serve` uses, just with a trace scene available for cmd/webdemo to
// pick up instead of the built-in pattern gallery.
func newTraceMux(webDir string, sceneJSON []byte) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(webDir)))
	mux.HandleFunc("GET /api/trace", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write(sceneJSON); err != nil {
			return // client went away mid-response; nothing left to do
		}
	})
	return mux
}

// runTrace parses path as a runtime/trace file, serves it as a scene for the
// WASM demo, opens the result in the browser, and blocks until interrupted.
func runTrace(path string) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	//nolint:gosec // path is the CLI's own positional argument — a local file the caller named on purpose, not attacker input
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open trace: %w", err)
	}
	defer f.Close()

	steps, err := traceparse.Parse(f)
	if err != nil {
		return fmt.Errorf("parse trace: %w", err)
	}
	if len(steps) == 0 {
		return fmt.Errorf("%s contains no goroutine/GC events to visualize", path)
	}

	scene := engine.Scene{
		Name:        filepath.Base(path),
		Description: fmt.Sprintf("Real runtime/trace capture — %d events", len(steps)),
		Steps:       steps,
	}
	sceneJSON, err := json.Marshal(toWireScene(scene))
	if err != nil {
		return fmt.Errorf("encode scene: %w", err)
	}

	webDir := os.Getenv("GOSCOPE_WEB_DIR")
	if webDir == "" {
		webDir = "web"
	}
	addr := os.Getenv("GOSCOPE_ADDR")
	if addr == "" {
		addr = "127.0.0.1:0" // ephemeral port: avoid clashing with a `make serve` already on :8080
	}

	var lc net.ListenConfig
	ln, err := lc.Listen(ctx, "tcp", addr)
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}

	url := "http://" + ln.Addr().String() + "/"
	srv := &http.Server{Handler: newTraceMux(webDir, sceneJSON), ReadHeaderTimeout: 5 * time.Second}

	errCh := make(chan error, 1)
	go func() { errCh <- srv.Serve(ln) }()

	fmt.Printf("goscope: serving %s at %s (Ctrl+C to stop)\n", path, url)
	if err := openBrowser(ctx, url); err != nil {
		fmt.Fprintf(os.Stderr, "goscope: couldn't open a browser automatically (%v) — open %s yourself\n", err, url)
	}

	select {
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	}
}

// openBrowser best-effort opens url with the OS's default browser.
//
//nolint:gosec // url is our own freshly-bound localhost server address, not user input
func openBrowser(ctx context.Context, url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.CommandContext(ctx, "open", url)
	case "windows":
		cmd = exec.CommandContext(ctx, "cmd", "/c", "start", "", url)
	default:
		cmd = exec.CommandContext(ctx, "xdg-open", url)
	}
	return cmd.Start()
}
