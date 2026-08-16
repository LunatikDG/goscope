// Package tracer wraps runtime/trace's global Start/Stop into a small,
// file-oriented helper: open a file, start tracing into it, and get back a
// single stop function that stops tracing and closes the file.
//
// This is the input side for a future trace-parsing/visualization phase —
// for now it just produces a standard runtime/trace file, inspectable with
// `go tool trace <path>`.
package tracer

import (
	"fmt"
	"os"
	"runtime/trace"
)

// StartToFile creates path and starts a runtime/trace into it. The returned
// stop function stops the trace and closes the file; call it exactly once
// (typically via defer) before the program exits, or the trace will be
// incomplete or unreadable.
func StartToFile(path string) (stop func() error, err error) {
	f, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("create trace file: %w", err)
	}

	if err := trace.Start(f); err != nil {
		f.Close()
		return nil, fmt.Errorf("start trace: %w", err)
	}

	return func() error {
		trace.Stop()
		return f.Close()
	}, nil
}
