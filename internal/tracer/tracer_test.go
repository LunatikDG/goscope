package tracer

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestStartToFileProducesATrace(t *testing.T) {
	path := filepath.Join(t.TempDir(), "trace.out")

	stop, err := StartToFile(path)
	if err != nil {
		t.Fatalf("StartToFile: %v", err)
	}

	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
		}()
	}
	wg.Wait()

	if stopErr := stop(); stopErr != nil {
		t.Fatalf("stop: %v", stopErr)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat trace file: %v", err)
	}
	if info.Size() == 0 {
		t.Error("trace file is empty")
	}
}

func TestStartToFileBadPath(t *testing.T) {
	// A directory that doesn't exist can't be created into — os.Create must fail.
	path := filepath.Join(t.TempDir(), "no-such-dir", "trace.out")

	if _, err := StartToFile(path); err == nil {
		t.Fatal("expected an error for an unwritable path, got nil")
	}
}
