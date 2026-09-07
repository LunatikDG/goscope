package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/LunatikDG/goscope/internal/engine"
)

func TestToWireSceneMapsFields(t *testing.T) {
	scene := engine.Scene{
		Name:        "demo",
		Description: "a scene",
		Steps: []engine.Step{
			{Event: engine.Spawn, Goroutine: 1, Label: "worker-1"},
			{Event: engine.Block, Goroutine: 1, Chan: 2},
			{Event: engine.GCStart, Goroutine: 0, Label: "GC concurrent mark phase"},
		},
	}

	ws := toWireScene(scene)
	if ws.Name != scene.Name || ws.Description != scene.Description {
		t.Fatalf("name/description not carried over: %+v", ws)
	}
	if len(ws.Steps) != len(scene.Steps) {
		t.Fatalf("step count = %d, want %d", len(ws.Steps), len(scene.Steps))
	}

	want := []wireStep{
		{Event: "spawn", Goroutine: 1, Label: "worker-1"},
		{Event: "block", Goroutine: 1, Chan: 2},
		{Event: "gcstart", Goroutine: 0, Label: "GC concurrent mark phase"},
	}
	for i, w := range want {
		if ws.Steps[i] != w {
			t.Errorf("step %d = %+v, want %+v", i, ws.Steps[i], w)
		}
	}
}

// repoRoot finds the repo root relative to this file so the mux's static
// file serving has real assets to serve, matching internal/server's tests.
func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller could not resolve this file's path")
	}
	return filepath.Join(filepath.Dir(file), "..", "..")
}

func TestNewTraceMuxServesSceneAndStatic(t *testing.T) {
	scene := engine.Scene{Name: "demo", Description: "d", Steps: []engine.Step{{Event: engine.Spawn, Goroutine: 1}}}
	body, err := json.Marshal(toWireScene(scene))
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	ts := httptest.NewServer(newTraceMux(filepath.Join(repoRoot(t), "web"), body))
	defer ts.Close()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, ts.URL+"/api/trace", http.NoBody)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET /api/trace: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}

	got, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	var decoded wireScene
	if err := json.Unmarshal(got, &decoded); err != nil {
		t.Fatalf("response body didn't decode as wireScene: %v", err)
	}
	if decoded.Name != "demo" || len(decoded.Steps) != 1 {
		t.Errorf("decoded scene = %+v, want name=demo with 1 step", decoded)
	}
}
