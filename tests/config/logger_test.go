package config_test

import (
	"testing"

	"github.com/JaimeStill/document-context/pkg/config"
)

func TestDefaultLoggerConfig(t *testing.T) {
	cfg := config.DefaultLoggerConfig()

	if cfg.Level != config.LogLevelInfo {
		t.Errorf("expected default level %q, got %q", config.LogLevelInfo, cfg.Level)
	}

	if cfg.Format != "text" {
		t.Errorf("expected default format %q, got %q", "text", cfg.Format)
	}
}

func TestDisabledLoggerConfig(t *testing.T) {
	cfg := config.DisabledLoggerConfig()

	if cfg.Level != config.LogLevelDisabled {
		t.Errorf("expected disabled level %q, got %q", config.LogLevelDisabled, cfg.Level)
	}

	if cfg.Format != "text" {
		t.Errorf("expected format %q, got %q", "text", cfg.Format)
	}
}

func TestLoggerConfig_Finalize_EmptyConfig(t *testing.T) {
	cfg := config.LoggerConfig{}
	cfg.Finalize()

	defaults := config.DefaultLoggerConfig()

	if cfg.Level != defaults.Level {
		t.Errorf("expected level %q, got %q", defaults.Level, cfg.Level)
	}

	if cfg.Format != defaults.Format {
		t.Errorf("expected format %q, got %q", defaults.Format, cfg.Format)
	}
}

func TestLoggerConfig_Finalize_PartialConfig(t *testing.T) {
	tests := []struct {
		name     string
		input    config.LoggerConfig
		expected config.LoggerConfig
	}{
		{
			name: "only level set",
			input: config.LoggerConfig{
				Level: config.LogLevelDebug,
			},
			expected: config.LoggerConfig{
				Level:  config.LogLevelDebug,
				Format: "text",
			},
		},
		{
			name: "only format set",
			input: config.LoggerConfig{
				Format: "json",
			},
			expected: config.LoggerConfig{
				Level:  config.LogLevelInfo,
				Format: "json",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := tt.input
			cfg.Finalize()

			if cfg.Level != tt.expected.Level {
				t.Errorf("expected level %q, got %q", tt.expected.Level, cfg.Level)
			}

			if cfg.Format != tt.expected.Format {
				t.Errorf("expected format %q, got %q", tt.expected.Format, cfg.Format)
			}
		})
	}
}

func TestLoggerConfig_Finalize_FullConfig(t *testing.T) {
	cfg := config.LoggerConfig{
		Level:  config.LogLevelWarn,
		Format: "json",
	}

	cfg.Finalize()

	if cfg.Level != config.LogLevelWarn {
		t.Errorf("expected level preserved as %q, got %q", config.LogLevelWarn, cfg.Level)
	}

	if cfg.Format != "json" {
		t.Errorf("expected format preserved as %q, got %q", "json", cfg.Format)
	}
}

func TestLogLevel_Constants(t *testing.T) {
	levels := []config.LogLevel{
		config.LogLevelDebug,
		config.LogLevelInfo,
		config.LogLevelWarn,
		config.LogLevelError,
		config.LogLevelDisabled,
	}

	expectedValues := []string{
		"debug",
		"info",
		"warn",
		"error",
		"disabled",
	}

	if len(levels) != len(expectedValues) {
		t.Fatalf("expected %d log levels, got %d", len(expectedValues), len(levels))
	}

	for i, level := range levels {
		if string(level) != expectedValues[i] {
			t.Errorf("level %d: expected %q, got %q", i, expectedValues[i], string(level))
		}
	}
}
