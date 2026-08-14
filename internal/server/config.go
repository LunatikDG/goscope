package server

import (
	"fmt"
	"log/slog"
	"time"
)

// Config gathers everything the server needs to start: the address, where to
// serve static files and examples from, the timeout for one live run, and
// logging settings. The defaults assume running from the repo root (make serve).
type Config struct {
	Addr        string
	WebDir      string
	ExamplesDir string
	LogFormat   string // "text" or "json"
	RunTimeout  time.Duration
	LogLevel    slog.Level
}

// DefaultConfig — settings for a local run with zero environment variables set.
func DefaultConfig() Config {
	return Config{
		Addr:        ":8080",
		WebDir:      "web",
		ExamplesDir: "examples",
		RunTimeout:  15 * time.Second,
		LogLevel:    slog.LevelInfo,
		LogFormat:   "text",
	}
}

// LoadConfig takes DefaultConfig and overrides fields from the environment
// variables GOSCOPE_ADDR / GOSCOPE_WEB_DIR / GOSCOPE_EXAMPLES_DIR /
// GOSCOPE_RUN_TIMEOUT / GOSCOPE_LOG_LEVEL / GOSCOPE_LOG_FORMAT — all optional.
// getenv is taken as a parameter (rather than calling os.Getenv directly) so
// the function is testable without side effects on the real process environment.
func LoadConfig(getenv func(string) string) (Config, error) {
	cfg := DefaultConfig()

	if v := getenv("GOSCOPE_ADDR"); v != "" {
		cfg.Addr = v
	}
	if v := getenv("GOSCOPE_WEB_DIR"); v != "" {
		cfg.WebDir = v
	}
	if v := getenv("GOSCOPE_EXAMPLES_DIR"); v != "" {
		cfg.ExamplesDir = v
	}
	if v := getenv("GOSCOPE_RUN_TIMEOUT"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			return Config{}, fmt.Errorf("GOSCOPE_RUN_TIMEOUT: %w", err)
		}
		cfg.RunTimeout = d
	}
	if v := getenv("GOSCOPE_LOG_LEVEL"); v != "" {
		lvl, err := parseLogLevel(v)
		if err != nil {
			return Config{}, err
		}
		cfg.LogLevel = lvl
	}
	if v := getenv("GOSCOPE_LOG_FORMAT"); v != "" {
		if v != "text" && v != "json" {
			return Config{}, fmt.Errorf("GOSCOPE_LOG_FORMAT: неизвестный формат %q, ожидался text или json", v)
		}
		cfg.LogFormat = v
	}

	return cfg, nil
}

func parseLogLevel(s string) (slog.Level, error) {
	switch s {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return 0, fmt.Errorf("GOSCOPE_LOG_LEVEL: неизвестный уровень %q", s)
	}
}
