package logger_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/JaimeStill/document-context/pkg/config"
	"github.com/JaimeStill/document-context/pkg/logger"
)

func TestNewSlogger_DefaultConfig(t *testing.T) {
	var buf bytes.Buffer
	cfg := config.DefaultLoggerConfig()

	log, err := logger.NewSlogger(cfg, &buf)
	if err != nil {
		t.Fatalf("NewSlogger failed: %v", err)
	}

	if log == nil {
		t.Fatal("expected non-nil logger")
	}
}

func TestNewSlogger_DisabledConfig(t *testing.T) {
	var buf bytes.Buffer
	cfg := config.DisabledLoggerConfig()

	log, err := logger.NewSlogger(cfg, &buf)
	if err != nil {
		t.Fatalf("NewSlogger failed: %v", err)
	}

	if log == nil {
		t.Fatal("expected non-nil logger")
	}

	// Log something - should go to io.Discard, not buf
	log.Info("test message")

	if buf.Len() != 0 {
		t.Errorf("expected no output with disabled logger, got %d bytes", buf.Len())
	}
}

func TestNewSlogger_TextFormat(t *testing.T) {
	var buf bytes.Buffer
	cfg := config.LoggerConfig{
		Level:  config.LogLevelInfo,
		Format: "text",
	}

	log, err := logger.NewSlogger(cfg, &buf)
	if err != nil {
		t.Fatalf("NewSlogger failed: %v", err)
	}

	log.Info("test message", "key", "value")

	output := buf.String()
	if !strings.Contains(output, "test message") {
		t.Errorf("expected log output to contain message, got: %s", output)
	}
}

func TestNewSlogger_JSONFormat(t *testing.T) {
	var buf bytes.Buffer
	cfg := config.LoggerConfig{
		Level:  config.LogLevelInfo,
		Format: "json",
	}

	log, err := logger.NewSlogger(cfg, &buf)
	if err != nil {
		t.Fatalf("NewSlogger failed: %v", err)
	}

	log.Info("test message", "key", "value")

	output := buf.String()
	if !strings.Contains(output, `"msg":"test message"`) {
		t.Errorf("expected JSON log output to contain message field, got: %s", output)
	}
}

func TestNewSlogger_InvalidLevel(t *testing.T) {
	var buf bytes.Buffer
	cfg := config.LoggerConfig{
		Level:  config.LogLevel("invalid"),
		Format: "text",
	}

	_, err := logger.NewSlogger(cfg, &buf)
	if err == nil {
		t.Error("expected error for invalid log level")
	}
}

func TestNewSlogger_InvalidFormat(t *testing.T) {
	var buf bytes.Buffer
	cfg := config.LoggerConfig{
		Level:  config.LogLevelInfo,
		Format: "invalid",
	}

	_, err := logger.NewSlogger(cfg, &buf)
	if err == nil {
		t.Error("expected error for invalid format")
	}
}

func TestSlogger_LogLevels(t *testing.T) {
	tests := []struct {
		name      string
		level     config.LogLevel
		logFunc   func(logger.Logger)
		shouldLog bool
	}{
		{
			name:      "debug level logs debug",
			level:     config.LogLevelDebug,
			logFunc:   func(l logger.Logger) { l.Debug("debug message") },
			shouldLog: true,
		},
		{
			name:      "info level skips debug",
			level:     config.LogLevelInfo,
			logFunc:   func(l logger.Logger) { l.Debug("debug message") },
			shouldLog: false,
		},
		{
			name:      "info level logs info",
			level:     config.LogLevelInfo,
			logFunc:   func(l logger.Logger) { l.Info("info message") },
			shouldLog: true,
		},
		{
			name:      "warn level skips info",
			level:     config.LogLevelWarn,
			logFunc:   func(l logger.Logger) { l.Info("info message") },
			shouldLog: false,
		},
		{
			name:      "warn level logs warn",
			level:     config.LogLevelWarn,
			logFunc:   func(l logger.Logger) { l.Warn("warn message") },
			shouldLog: true,
		},
		{
			name:      "error level skips warn",
			level:     config.LogLevelError,
			logFunc:   func(l logger.Logger) { l.Warn("warn message") },
			shouldLog: false,
		},
		{
			name:      "error level logs error",
			level:     config.LogLevelError,
			logFunc:   func(l logger.Logger) { l.Error("error message") },
			shouldLog: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			cfg := config.LoggerConfig{
				Level:  tt.level,
				Format: "text",
			}

			log, err := logger.NewSlogger(cfg, &buf)
			if err != nil {
				t.Fatalf("NewSlogger failed: %v", err)
			}

			tt.logFunc(log)

			gotOutput := buf.Len() > 0
			if gotOutput != tt.shouldLog {
				t.Errorf("expected shouldLog=%v, got output=%v", tt.shouldLog, gotOutput)
			}
		})
	}
}

func TestSlogger_LogArguments(t *testing.T) {
	var buf bytes.Buffer
	cfg := config.LoggerConfig{
		Level:  config.LogLevelInfo,
		Format: "text",
	}

	log, err := logger.NewSlogger(cfg, &buf)
	if err != nil {
		t.Fatalf("NewSlogger failed: %v", err)
	}

	log.Info("test message", "key1", "value1", "key2", 42)

	output := buf.String()
	if !strings.Contains(output, "key1") || !strings.Contains(output, "value1") {
		t.Errorf("expected log output to contain key-value pairs, got: %s", output)
	}
}

func TestSlogger_Interface(t *testing.T) {
	var buf bytes.Buffer
	cfg := config.DefaultLoggerConfig()

	// Should satisfy logger.Logger interface
	var log logger.Logger
	log, err := logger.NewSlogger(cfg, &buf)
	if err != nil {
		t.Fatalf("NewSlogger failed: %v", err)
	}

	if log == nil {
		t.Fatal("expected non-nil logger")
	}

	// Should be able to call all interface methods
	log.Debug("debug")
	log.Info("info")
	log.Warn("warn")
	log.Error("error")
}
