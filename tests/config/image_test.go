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

	if cfg.Brightness != nil {
		t.Errorf("expected Brightness nil, got %v", *cfg.Brightness)
	}
	if cfg.Contrast != nil {
		t.Errorf("expected Contrast nil, got %v", *cfg.Contrast)
	}
	if cfg.Saturation != nil {
		t.Errorf("expected Saturation nil, got %v", *cfg.Saturation)
	}
	if cfg.Rotation != nil {
		t.Errorf("expected Rotation nil, got %v", *cfg.Rotation)
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

func TestImageConfig_Merge_FilterFields(t *testing.T) {
	tests := []struct {
		name    string
		base    config.ImageConfig
		source  config.ImageConfig
		checkFn func(t *testing.T, result config.ImageConfig)
	}{
		{
			name: "merge all filter fields",
			base: config.DefaultImageConfig(),
			source: config.ImageConfig{
				Brightness: intPtr(10),
				Contrast:   intPtr(5),
				Saturation: intPtr(-10),
				Rotation:   intPtr(90),
			},
			checkFn: func(t *testing.T, result config.ImageConfig) {
				if result.Brightness == nil || *result.Brightness != 10 {
					t.Errorf("expected Brightness 10, got %v", result.Brightness)
				}
				if result.Contrast == nil || *result.Contrast != 5 {
					t.Errorf("expected Contrast 5, got %v", result.Contrast)
				}
				if result.Saturation == nil || *result.Saturation != -10 {
					t.Errorf("expected Saturation -10, got %v", result.Saturation)
				}
				if result.Rotation == nil || *result.Rotation != 90 {
					t.Errorf("expected Rotation 90, got %v", result.Rotation)
				}
			},
		},
		{
			name: "nil source fields ignored",
			base: config.ImageConfig{
				Brightness: intPtr(20),
				Contrast:   intPtr(15),
			},
			source: config.ImageConfig{
				Brightness: nil,
				Saturation: intPtr(5),
			},
			checkFn: func(t *testing.T, result config.ImageConfig) {
				if result.Brightness == nil || *result.Brightness != 20 {
					t.Errorf("expected Brightness 20 (unchanged), got %v", result.Brightness)
				}
				if result.Contrast == nil || *result.Contrast != 15 {
					t.Errorf("expected Contrast 15 (unchanged), got %v", result.Contrast)
				}
				if result.Saturation == nil || *result.Saturation != 5 {
					t.Errorf("expected Saturation 5, got %v", result.Saturation)
				}
			},
		},
		{
			name: "explicit zero overrides",
			base: config.ImageConfig{
				Brightness: intPtr(20),
			},
			source: config.ImageConfig{
				Brightness: intPtr(0),
			},
			checkFn: func(t *testing.T, result config.ImageConfig) {
				if result.Brightness == nil || *result.Brightness != 0 {
					t.Errorf("expected Brightness 0 (explicit override), got %v", result.Brightness)
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
				Format:     "jpg",
				Quality:    90,
				DPI:        150,
				Brightness: intPtr(10),
			},
			expected: config.ImageConfig{
				Format:     "jpg",
				Quality:    90,
				DPI:        150,
				Brightness: intPtr(10),
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

			if tt.expected.Brightness != nil {
				if tt.input.Brightness == nil || *tt.input.Brightness != *tt.expected.Brightness {
					t.Errorf("Brightness mismatch: expected %v, got %v", tt.expected.Brightness, tt.input.Brightness)
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
			name: "full config with filters",
			config: config.ImageConfig{
				Format:     "jpg",
				Quality:    90,
				DPI:        150,
				Brightness: intPtr(10),
				Contrast:   intPtr(5),
				Saturation: intPtr(-10),
				Rotation:   intPtr(90),
			},
			expected: `{"format":"jpg","quality":90,"dpi":150,"brightness":10,"contrast":5,"saturation":-10,"rotation":90}`,
		},
		{
			name: "partial filters",
			config: config.ImageConfig{
				Format:     "png",
				DPI:        300,
				Brightness: intPtr(15),
			},
			expected: `{"format":"png","dpi":300,"brightness":15}`,
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
				if cfg.Brightness != nil {
					t.Errorf("expected Brightness nil, got %v", *cfg.Brightness)
				}
			},
		},
		{
			name: "with filter fields",
			json: `{"format":"png","dpi":300,"brightness":10,"rotation":90}`,
			checkFn: func(t *testing.T, cfg config.ImageConfig) {
				if cfg.Format != "png" {
					t.Errorf("expected Format 'png', got %q", cfg.Format)
				}
				if cfg.DPI != 300 {
					t.Errorf("expected DPI 300, got %d", cfg.DPI)
				}
				if cfg.Brightness == nil || *cfg.Brightness != 10 {
					t.Errorf("expected Brightness 10, got %v", cfg.Brightness)
				}
				if cfg.Rotation == nil || *cfg.Rotation != 90 {
					t.Errorf("expected Rotation 90, got %v", cfg.Rotation)
				}
				if cfg.Contrast != nil {
					t.Errorf("expected Contrast nil, got %v", *cfg.Contrast)
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
		Format:     "jpg",
		Quality:    90,
		DPI:        150,
		Brightness: intPtr(10),
		Contrast:   intPtr(5),
		Saturation: intPtr(-10),
		Rotation:   intPtr(90),
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

	if decoded.Brightness == nil || *decoded.Brightness != *original.Brightness {
		t.Errorf("Brightness mismatch: expected %v, got %v", original.Brightness, decoded.Brightness)
	}
	if decoded.Contrast == nil || *decoded.Contrast != *original.Contrast {
		t.Errorf("Contrast mismatch: expected %v, got %v", original.Contrast, decoded.Contrast)
	}
	if decoded.Saturation == nil || *decoded.Saturation != *original.Saturation {
		t.Errorf("Saturation mismatch: expected %v, got %v", original.Saturation, decoded.Saturation)
	}
	if decoded.Rotation == nil || *decoded.Rotation != *original.Rotation {
		t.Errorf("Rotation mismatch: expected %v, got %v", original.Rotation, decoded.Rotation)
	}
}
