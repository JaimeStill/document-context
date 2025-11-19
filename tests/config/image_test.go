package config_test

import (
	"encoding/json"
	"testing"

	"github.com/JaimeStill/document-context/pkg/config"
)

func intPtr(i int) *int {
	return &i
}

func TestDefaultImageConfig(t *testing.T) {
	cfg := config.DefaultImageConfig()

	if cfg.Format != "png" {
		t.Errorf("expected Format 'png', got %q", cfg.Format)
	}
	if cfg.Quality != 0 {
		t.Errorf("expected Quality 0, got %d", cfg.Quality)
	}
	if cfg.DPI != 300 {
		t.Errorf("expected DPI 300, got %d", cfg.DPI)
	}

	if cfg.Options == nil {
		t.Error("expected Options map to be initialized")
	}
	if len(cfg.Options) != 0 {
		t.Errorf("expected Options map to be empty, got %d entries", len(cfg.Options))
	}
}

func TestImageConfig_Merge_BaseFields(t *testing.T) {
	tests := []struct {
		name     string
		base     config.ImageConfig
		source   config.ImageConfig
		expected config.ImageConfig
	}{
		{
			name: "merge all base fields",
			base: config.DefaultImageConfig(),
			source: config.ImageConfig{
				Format:  "jpg",
				Quality: 90,
				DPI:     150,
			},
			expected: config.ImageConfig{
				Format:  "jpg",
				Quality: 90,
				DPI:     150,
			},
		},
		{
			name: "ignore zero DPI",
			base: config.ImageConfig{
				Format: "png",
				DPI:    300,
			},
			source: config.ImageConfig{
				Format: "jpg",
				DPI:    0,
			},
			expected: config.ImageConfig{
				Format: "jpg",
				DPI:    300,
			},
		},
		{
			name: "ignore empty format",
			base: config.ImageConfig{
				Format: "png",
				DPI:    300,
			},
			source: config.ImageConfig{
				Format: "",
				DPI:    150,
			},
			expected: config.ImageConfig{
				Format: "png",
				DPI:    150,
			},
		},
		{
			name: "ignore zero quality",
			base: config.ImageConfig{
				Format:  "jpg",
				Quality: 85,
			},
			source: config.ImageConfig{
				Quality: 0,
			},
			expected: config.ImageConfig{
				Format:  "jpg",
				Quality: 85,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.base.Merge(&tt.source)

			if tt.base.Format != tt.expected.Format {
				t.Errorf("Format: expected %q, got %q", tt.expected.Format, tt.base.Format)
			}
			if tt.base.Quality != tt.expected.Quality {
				t.Errorf("Quality: expected %d, got %d", tt.expected.Quality, tt.base.Quality)
			}
			if tt.base.DPI != tt.expected.DPI {
				t.Errorf("DPI: expected %d, got %d", tt.expected.DPI, tt.base.DPI)
			}
		})
	}
}

func TestImageConfig_Merge_Options(t *testing.T) {
	tests := []struct {
		name    string
		base    config.ImageConfig
		source  config.ImageConfig
		checkFn func(t *testing.T, result config.ImageConfig)
	}{
		{
			name: "merge options into empty base",
			base: config.DefaultImageConfig(),
			source: config.ImageConfig{
				Options: map[string]any{
					"brightness": 10,
					"contrast":   5,
				},
			},
			checkFn: func(t *testing.T, result config.ImageConfig) {
				if result.Options["brightness"] != 10 {
					t.Errorf("expected brightness 10, got %v", result.Options["brightness"])
				}
				if result.Options["contrast"] != 5 {
					t.Errorf("expected contrast 5, got %v", result.Options["contrast"])
				}
			},
		},
		{
			name: "merge options with existing values",
			base: config.ImageConfig{
				Options: map[string]any{
					"brightness": 20,
					"saturation": 15,
				},
			},
			source: config.ImageConfig{
				Options: map[string]any{
					"brightness": 10,
					"contrast":   5,
				},
			},
			checkFn: func(t *testing.T, result config.ImageConfig) {
				if result.Options["brightness"] != 10 {
					t.Errorf("expected brightness 10 (overridden), got %v", result.Options["brightness"])
				}
				if result.Options["contrast"] != 5 {
					t.Errorf("expected contrast 5, got %v", result.Options["contrast"])
				}
				if result.Options["saturation"] != 15 {
					t.Errorf("expected saturation 15 (unchanged), got %v", result.Options["saturation"])
				}
			},
		},
		{
			name: "nil source options ignored",
			base: config.ImageConfig{
				Options: map[string]any{
					"brightness": 20,
				},
			},
			source: config.ImageConfig{
				Options: nil,
			},
			checkFn: func(t *testing.T, result config.ImageConfig) {
				if result.Options["brightness"] != 20 {
					t.Errorf("expected brightness 20 (unchanged), got %v", result.Options["brightness"])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.base.Merge(&tt.source)
			tt.checkFn(t, tt.base)
		})
	}
}

func TestImageConfig_Finalize(t *testing.T) {
	tests := []struct {
		name     string
		input    config.ImageConfig
		expected config.ImageConfig
	}{
		{
			name:     "empty config gets all defaults",
			input:    config.ImageConfig{},
			expected: config.DefaultImageConfig(),
		},
		{
			name: "partial config merges with defaults",
			input: config.ImageConfig{
				Format: "jpg",
			},
			expected: config.ImageConfig{
				Format:  "jpg",
				Quality: 0,
				DPI:     300,
			},
		},
		{
			name: "full config unchanged",
			input: config.ImageConfig{
				Format:  "jpg",
				Quality: 90,
				DPI:     150,
				Options: map[string]any{
					"brightness": 10,
				},
			},
			expected: config.ImageConfig{
				Format:  "jpg",
				Quality: 90,
				DPI:     150,
				Options: map[string]any{
					"brightness": 10,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.input.Finalize()

			if tt.input.Format != tt.expected.Format {
				t.Errorf("Format: expected %q, got %q", tt.expected.Format, tt.input.Format)
			}
			if tt.input.Quality != tt.expected.Quality {
				t.Errorf("Quality: expected %d, got %d", tt.expected.Quality, tt.input.Quality)
			}
			if tt.input.DPI != tt.expected.DPI {
				t.Errorf("DPI: expected %d, got %d", tt.expected.DPI, tt.input.DPI)
			}

			if tt.expected.Options != nil {
				for key, expectedVal := range tt.expected.Options {
					if gotVal, ok := tt.input.Options[key]; !ok {
						t.Errorf("Options[%q] missing", key)
					} else if gotVal != expectedVal {
						t.Errorf("Options[%q]: expected %v, got %v", key, expectedVal, gotVal)
					}
				}
			}
		})
	}
}

func TestImageConfig_JSON_Marshal(t *testing.T) {
	tests := []struct {
		name     string
		config   config.ImageConfig
		expected string
	}{
		{
			name:     "default config",
			config:   config.DefaultImageConfig(),
			expected: `{"format":"png","dpi":300}`,
		},
		{
			name: "full config with options",
			config: config.ImageConfig{
				Format:  "jpg",
				Quality: 90,
				DPI:     150,
				Options: map[string]any{
					"brightness": 10,
					"contrast":   5,
					"saturation": -10,
					"rotation":   90,
				},
			},
			expected: `{"format":"jpg","quality":90,"dpi":150,"options":{"brightness":10,"contrast":5,"rotation":90,"saturation":-10}}`,
		},
		{
			name: "partial options",
			config: config.ImageConfig{
				Format: "png",
				DPI:    300,
				Options: map[string]any{
					"brightness": 15,
				},
			},
			expected: `{"format":"png","dpi":300,"options":{"brightness":15}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.config)
			if err != nil {
				t.Fatalf("Marshal failed: %v", err)
			}

			got := string(data)
			if got != tt.expected {
				t.Errorf("JSON mismatch:\nexpected: %s\ngot:      %s", tt.expected, got)
			}
		})
	}
}

func TestImageConfig_JSON_Unmarshal(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		checkFn func(t *testing.T, cfg config.ImageConfig)
	}{
		{
			name: "base fields only",
			json: `{"format":"jpg","quality":85,"dpi":150}`,
			checkFn: func(t *testing.T, cfg config.ImageConfig) {
				if cfg.Format != "jpg" {
					t.Errorf("expected Format 'jpg', got %q", cfg.Format)
				}
				if cfg.Quality != 85 {
					t.Errorf("expected Quality 85, got %d", cfg.Quality)
				}
				if cfg.DPI != 150 {
					t.Errorf("expected DPI 150, got %d", cfg.DPI)
				}
				if cfg.Options != nil && len(cfg.Options) > 0 {
					t.Errorf("expected empty Options, got %v", cfg.Options)
				}
			},
		},
		{
			name: "with options",
			json: `{"format":"png","dpi":300,"options":{"brightness":10,"rotation":90}}`,
			checkFn: func(t *testing.T, cfg config.ImageConfig) {
				if cfg.Format != "png" {
					t.Errorf("expected Format 'png', got %q", cfg.Format)
				}
				if cfg.DPI != 300 {
					t.Errorf("expected DPI 300, got %d", cfg.DPI)
				}
				if cfg.Options == nil {
					t.Fatal("expected Options to be set")
				}
				brightness, ok := cfg.Options["brightness"].(float64)
				if !ok || brightness != 10 {
					t.Errorf("expected brightness 10, got %v", cfg.Options["brightness"])
				}
				rotation, ok := cfg.Options["rotation"].(float64)
				if !ok || rotation != 90 {
					t.Errorf("expected rotation 90, got %v", cfg.Options["rotation"])
				}
			},
		},
		{
			name: "empty JSON object",
			json: `{}`,
			checkFn: func(t *testing.T, cfg config.ImageConfig) {
				if cfg.Format != "" {
					t.Errorf("expected Format empty, got %q", cfg.Format)
				}
				if cfg.Quality != 0 {
					t.Errorf("expected Quality 0, got %d", cfg.Quality)
				}
				if cfg.DPI != 0 {
					t.Errorf("expected DPI 0, got %d", cfg.DPI)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cfg config.ImageConfig
			err := json.Unmarshal([]byte(tt.json), &cfg)
			if err != nil {
				t.Fatalf("Unmarshal failed: %v", err)
			}

			tt.checkFn(t, cfg)
		})
	}
}

func TestImageConfig_JSON_RoundTrip(t *testing.T) {
	original := config.ImageConfig{
		Format:  "jpg",
		Quality: 90,
		DPI:     150,
		Options: map[string]any{
			"brightness": 10,
			"contrast":   5,
			"saturation": -10,
			"rotation":   90,
		},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded config.ImageConfig
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded.Format != original.Format {
		t.Errorf("Format: expected %q, got %q", original.Format, decoded.Format)
	}
	if decoded.Quality != original.Quality {
		t.Errorf("Quality: expected %d, got %d", original.Quality, decoded.Quality)
	}
	if decoded.DPI != original.DPI {
		t.Errorf("DPI: expected %d, got %d", original.DPI, decoded.DPI)
	}

	if decoded.Options == nil {
		t.Fatal("expected Options to be set")
	}

	for key, originalVal := range original.Options {
		decodedVal, ok := decoded.Options[key]
		if !ok {
			t.Errorf("Options[%q] missing in decoded config", key)
			continue
		}

		origFloat, origOk := originalVal.(int)
		decFloat, decOk := decodedVal.(float64)
		if origOk && decOk {
			if float64(origFloat) != decFloat {
				t.Errorf("Options[%q]: expected %v, got %v", key, originalVal, decodedVal)
			}
		} else if originalVal != decodedVal {
			t.Errorf("Options[%q]: expected %v, got %v", key, originalVal, decodedVal)
		}
	}
}
