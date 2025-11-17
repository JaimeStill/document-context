package logger_test

import (
	"testing"

	"github.com/JaimeStill/document-context/pkg/config"
	"github.com/JaimeStill/document-context/pkg/logger"
)

func TestNewSlogger_DefaultConfig(t *testing.T) {
	cfg := config.DefaultLoggerConfig()

	log, err := logger.NewSlogger(cfg)
	if err != nil {
		t.Fatalf("NewSlogger failed: %v", err)
	}

	if log == nil {
		t.Fatal("expected non-nil logger")
	}
}

func TestNewSlogger_DisabledConfig(t *testing.T) {
	cfg := config.DisabledLoggerConfig()

	log, err := logger.NewSlogger(cfg)
	if err != nil {
		t.Fatalf("NewSlogger failed: %v", err)
	}

	if log == nil {
		t.Fatal("expected non-nil logger")
	}

	log.Info("test message")
}

func TestNewSlogger_TextFormat(t *testing.T) {
	cfg := config.LoggerConfig{
		Level:  config.LogLevelInfo,
		Format: "text",
		Output: config.LoggerOutputStdout,
	}

	log, err := logger.NewSlogger(cfg)
	if err != nil {
		t.Fatalf("NewSlogger failed: %v", err)
	}

	log.Info("test message", "key", "value")
}

func TestNewSlogger_JSONFormat(t *testing.T) {
	cfg := config.LoggerConfig{
		Level:  config.LogLevelInfo,
		Format: "json",
		Output: config.LoggerOutputStdout,
	}

	log, err := logger.NewSlogger(cfg)
	if err != nil {
		t.Fatalf("NewSlogger failed: %v", err)
	}

	log.Info("test message", "key", "value")
}

func TestNewSlogger_InvalidLevel(t *testing.T) {
	cfg := config.LoggerConfig{
		Level:  config.LogLevel("invalid"),
		Format: "text",
	}

	_, err := logger.NewSlogger(cfg)
	if err == nil {
		t.Error("expected error for invalid log level")
	}
}

func TestNewSlogger_InvalidFormat(t *testing.T) {
	cfg := config.LoggerConfig{
		Level:  config.LogLevelInfo,
		Format: "invalid",
	}

	_, err := logger.NewSlogger(cfg)
	if err == nil {
		t.Error("expected error for invalid format")
	}
}

func TestNewSlogger_InvalidOutput(t *testing.T) {
	cfg := config.LoggerConfig{
		Level:  config.LogLevelInfo,
		Format: "text",
		Output: config.LoggerOutput("invalid"),
	}

	_, err := logger.NewSlogger(cfg)
	if err == nil {
		t.Error("expected error for invalid output")
	}
}

func TestNewSlogger_OutputVariations(t *testing.T) {
	tests := []struct {
		name   string
		output config.LoggerOutput
	}{
		{"discard", config.LoggerOutputDiscard},
		{"stdout", config.LoggerOutputStdout},
		{"stderr", config.LoggerOutputStderr},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.LoggerConfig{
				Level:  config.LogLevelInfo,
				Format: "text",
				Output: tt.output,
			}

			log, err := logger.NewSlogger(cfg)
			if err != nil {
				t.Fatalf("NewSlogger failed: %v", err)
			}

			if log == nil {
				t.Fatal("expected non-nil logger")
			}

			log.Info("test message")
		})
	}
}

func TestSlogger_LogLevels(t *testing.T) {
	tests := []struct {
		name    string
		level   config.LogLevel
		logFunc func(logger.Logger)
	}{
		{
			name:    "debug level logs debug",
			level:   config.LogLevelDebug,
			logFunc: func(l logger.Logger) { l.Debug("debug message") },
		},
		{
			name:    "info level skips debug",
			level:   config.LogLevelInfo,
			logFunc: func(l logger.Logger) { l.Debug("debug message") },
		},
		{
			name:    "info level logs info",
			level:   config.LogLevelInfo,
			logFunc: func(l logger.Logger) { l.Info("info message") },
		},
		{
			name:    "warn level skips info",
			level:   config.LogLevelWarn,
			logFunc: func(l logger.Logger) { l.Info("info message") },
		},
		{
			name:    "warn level logs warn",
			level:   config.LogLevelWarn,
			logFunc: func(l logger.Logger) { l.Warn("warn message") },
		},
		{
			name:    "error level skips warn",
			level:   config.LogLevelError,
			logFunc: func(l logger.Logger) { l.Warn("warn message") },
		},
		{
			name:    "error level logs error",
			level:   config.LogLevelError,
			logFunc: func(l logger.Logger) { l.Error("error message") },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.LoggerConfig{
				Level:  tt.level,
				Format: "text",
				Output: config.LoggerOutputDiscard,
			}

			log, err := logger.NewSlogger(cfg)
			if err != nil {
				t.Fatalf("NewSlogger failed: %v", err)
			}

			tt.logFunc(log)
		})
	}
}

func TestSlogger_LogArguments(t *testing.T) {
	cfg := config.LoggerConfig{
		Level:  config.LogLevelInfo,
		Format: "text",
		Output: config.LoggerOutputDiscard,
	}

	log, err := logger.NewSlogger(cfg)
	if err != nil {
		t.Fatalf("NewSlogger failed: %v", err)
	}

	log.Info("test message", "key1", "value1", "key2", 42)
}

func TestSlogger_Interface(t *testing.T) {
	cfg := config.DefaultLoggerConfig()

	// Should satisfy logger.Logger interface
	var log logger.Logger
	log, err := logger.NewSlogger(cfg)
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
