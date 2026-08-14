package server

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// repoRoot finds the repo root relative to this file, not go test's current
// directory (which is internal/server, while web/ and examples/ live above it).
func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller не смог определить путь к этому файлу")
	}
	return filepath.Join(filepath.Dir(file), "..", "..")
}

func testConfig(t *testing.T) Config {
	t.Helper()
	root := repoRoot(t)

	cfg := DefaultConfig()
	cfg.WebDir = filepath.Join(root, "web")
	cfg.ExamplesDir = filepath.Join(root, "examples")
	cfg.RunTimeout = 10 * time.Second
	return cfg
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func getURL(t *testing.T, url string) *http.Response {
	t.Helper()
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, http.NoBody)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET %s: %v", url, err)
	}
	return resp
}

func TestStaticFileServing(t *testing.T) {
	srv := New(testConfig(t), testLogger())
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	resp := getURL(t, ts.URL+"/")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("статус = %d, ожидалось 200", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("чтение тела: %v", err)
	}
	if !strings.Contains(string(body), "goscope") {
		t.Errorf("тело / не упоминает goscope: %s", body)
	}
}

func TestRunUnknownExample(t *testing.T) {
	srv := New(testConfig(t), testLogger())
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	for _, name := range []string{"nope", "does-not-exist", ".."} {
		t.Run(name, func(t *testing.T) {
			resp := getURL(t, ts.URL+"/api/run/"+name)
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusNotFound {
				t.Errorf("статус = %d, ожидалось 404 (имя не из белого списка)", resp.StatusCode)
			}
		})
	}
}

// TestRunLiveExample is an end-to-end transport test: a real subprocess, a
// real SSE stream. pipeline was picked as the fastest of the five examples
// and the most deterministic in its set of events.
func TestRunLiveExample(t *testing.T) {
	srv := New(testConfig(t), testLogger())
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ts.URL+"/api/run/pipeline", http.NoBody)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("GET /api/run/pipeline: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("статус = %d, ожидалось 200", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); ct != "text/event-stream" {
		t.Errorf("Content-Type = %q, ожидалось text/event-stream", ct)
	}

	var dataLines []string
	sawEnd := false

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case strings.HasPrefix(line, "data: "):
			dataLines = append(dataLines, strings.TrimPrefix(line, "data: "))
		case line == "event: end":
			sawEnd = true
		}
		if sawEnd {
			// The server holds the connection open until the client disconnects
			// (otherwise the browser's EventSource would reconnect on its own) —
			// we close it ourselves, exactly like the browser does upon receiving "end".
			cancel()
			break
		}
	}

	if !sawEnd {
		t.Fatal("не дождались event: end")
	}
	if len(dataLines) == 0 {
		t.Fatal("не получили ни одной строки данных до event: end")
	}

	knownEvents := map[string]bool{"spawn": true, "block": true, "unblock": true, "send": true, "done": true}
	for _, line := range dataLines {
		var payload struct {
			Event string `json:"event"`
		}
		if err := json.Unmarshal([]byte(line), &payload); err != nil {
			t.Fatalf("data-строка %q не распарсилась как JSON: %v", line, err)
		}
		if !knownEvents[payload.Event] {
			t.Errorf("неизвестный тип события %q в строке %q", payload.Event, line)
		}
	}
}
