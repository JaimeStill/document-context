package logger

import (
	"fmt"
	"io"
	"log/slog"

	"github.com/JaimeStill/document-context/pkg/config"
)

// Slogger implements the Logger interface using Go's standard log/slog package.
//
// Slogger provides structured logging with configurable output format (text or JSON)
// and log level filtering. It wraps slog.Logger to provide a consistent interface
// across the application.
//
// Instances are safe for concurrent use by multiple goroutines.
type Slogger struct {
	logger *slog.Logger
}

// NewSlogger creates a new Logger implementation using log/slog.
//
// This function implements the Configuration Transformation Pattern, validating
// the provided configuration and returning a Logger interface. Invalid configurations
// are rejected at this boundary before the logger is created.
//
// Configuration:
//   - Level: Filters log output by minimum severity (debug, info, warn, error, disabled)
//   - Format: Output format - "text" for human-readable or "json" for structured
//   - Output: Destination for log output (typically os.Stderr, file, or custom writer)
//
// Special handling:
//   - LogLevelDisabled automatically replaces output with io.Discard for zero overhead
//   - Configuration is finalized (defaults applied) before validation
//
// Example usage:
//
//	cfg := config.LoggerConfig{
//	    Level:  config.LogLevelInfo,
//	    Format: "json",
//	}
//	logger, err := logger.NewSlogger(cfg, os.Stderr)
//	if err != nil {
//	    return fmt.Errorf("failed to create logger: %w", err)
//	}
//
// Returns an error if the log level or format is invalid.
func NewSlogger(cfg config.LoggerConfig, output io.Writer) (Logger, error) {
	cfg.Finalize()

	if cfg.Level == config.LogLevelDisabled {
		output = io.Discard
	}

	var level slog.Level
	switch cfg.Level {
	case config.LogLevelDebug:
		level = slog.LevelDebug
	case config.LogLevelInfo:
		level = slog.LevelInfo
	case config.LogLevelWarn:
		level = slog.LevelWarn
	case config.LogLevelError:
		level = slog.LevelError
	case config.LogLevelDisabled:
		level = slog.Level(1000)
	default:
		return nil, fmt.Errorf("invalid log level: %s", cfg.Level)
	}

	opts := &slog.HandlerOptions{Level: level}

	var handler slog.Handler
	switch cfg.Format {
	case "json":
		handler = slog.NewJSONHandler(output, opts)
	case "text":
		handler = slog.NewTextHandler(output, opts)
	default:
		return nil, fmt.Errorf("invalid format: %s", cfg.Format)
	}

	return &Slogger{
		logger: slog.New(handler),
	}, nil
}

func (l *Slogger) Debug(msg string, args ...any) {
	l.logger.Debug(msg, args...)
}

func (l *Slogger) Info(msg string, args ...any) {
	l.logger.Info(msg, args...)
}

func (l *Slogger) Warn(msg string, args ...any) {
	l.logger.Warn(msg, args...)
}

func (l *Slogger) Error(msg string, args ...any) {
	l.logger.Error(msg, args...)
}
