package config_test

import (
	"encoding/json"
	"testing"

	"github.com/JaimeStill/document-context/pkg/config"
)

func TestDefaultCacheConfig(t *testing.T) {
	cfg := config.DefaultCacheConfig()

	if cfg.Name != "" {
		t.Errorf("expected Name empty, got %q", cfg.Name)
	}

	if cfg.Options == nil {
		t.Error("expected Options map initialized, got nil")
	}

	if len(cfg.Options) != 0 {
		t.Errorf("expected Options empty, got %d entries", len(cfg.Options))
	}
}

func TestCacheConfig_Merge(t *testing.T) {
	tests := []struct {
		name    string
		base    config.CacheConfig
		source  config.CacheConfig
		checkFn func(t *testing.T, result config.CacheConfig)
	}{
		{
			name: "merge name",
			base: config.CacheConfig{
				Name:    "filesystem",
				Options: map[string]any{"directory": "/tmp"},
			},
			source: config.CacheConfig{
				Name: "memory",
			},
			checkFn: func(t *testing.T, result config.CacheConfig) {
				if result.Name != "memory" {
					t.Errorf("expected Name 'memory', got %q", result.Name)
				}
				if result.Options["directory"] != "/tmp" {
					t.Errorf("expected directory '/tmp', got %v", result.Options["directory"])
				}
			},
		},
		{
			name: "ignore empty name",
			base: config.CacheConfig{
				Name: "filesystem",
			},
			source: config.CacheConfig{
				Name: "",
			},
			checkFn: func(t *testing.T, result config.CacheConfig) {
				if result.Name != "filesystem" {
					t.Errorf("expected Name 'filesystem' (unchanged), got %q", result.Name)
				}
			},
		},
		{
			name: "merge options into empty base",
			base: config.DefaultCacheConfig(),
			source: config.CacheConfig{
				Options: map[string]any{
					"directory": "/var/cache",
					"size":      1000,
				},
			},
			checkFn: func(t *testing.T, result config.CacheConfig) {
				if result.Options["directory"] != "/var/cache" {
					t.Errorf("expected directory '/var/cache', got %v", result.Options["directory"])
				}
				if result.Options["size"] != 1000 {
					t.Errorf("expected size 1000, got %v", result.Options["size"])
				}
			},
		},
		{
			name: "merge options with override",
			base: config.CacheConfig{
				Options: map[string]any{
					"directory": "/tmp",
					"size":      500,
				},
			},
			source: config.CacheConfig{
				Options: map[string]any{
					"directory": "/var/cache",
					"ttl":       3600,
				},
			},
			checkFn: func(t *testing.T, result config.CacheConfig) {
				if result.Options["directory"] != "/var/cache" {
					t.Errorf("expected directory '/var/cache', got %v", result.Options["directory"])
				}
				if result.Options["size"] != 500 {
					t.Errorf("expected size 500 (unchanged), got %v", result.Options["size"])
				}
				if result.Options["ttl"] != 3600 {
					t.Errorf("expected ttl 3600, got %v", result.Options["ttl"])
				}
			},
		},
		{
			name: "ignore nil options",
			base: config.CacheConfig{
				Options: map[string]any{"directory": "/tmp"},
			},
			source: config.CacheConfig{
				Options: nil,
			},
			checkFn: func(t *testing.T, result config.CacheConfig) {
				if result.Options["directory"] != "/tmp" {
					t.Errorf("expected directory '/tmp' (unchanged), got %v", result.Options["directory"])
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

func TestCacheConfig_JSON_Marshal(t *testing.T) {
	tests := []struct {
		name     string
		config   config.CacheConfig
		expected string
	}{
		{
			name: "full config",
			config: config.CacheConfig{
				Name: "filesystem",
				Options: map[string]any{
					"directory": "/var/cache",
				},
			},
			expected: `{"name":"filesystem","options":{"directory":"/var/cache"}}`,
		},
		{
			name: "name only",
			config: config.CacheConfig{
				Name: "memory",
			},
			expected: `{"name":"memory"}`,
		},
		{
			name:     "empty config",
			config:   config.CacheConfig{},
			expected: `{"name":""}`,
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

func TestCacheConfig_JSON_Unmarshal(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		checkFn func(t *testing.T, cfg config.CacheConfig)
	}{
		{
			name: "full config",
			json: `{"name":"filesystem","options":{"directory":"/var/cache","size":1000}}`,
			checkFn: func(t *testing.T, cfg config.CacheConfig) {
				if cfg.Name != "filesystem" {
					t.Errorf("expected Name 'filesystem', got %q", cfg.Name)
				}
				if cfg.Options["directory"] != "/var/cache" {
					t.Errorf("expected directory '/var/cache', got %v", cfg.Options["directory"])
				}
				if cfg.Options["size"] != float64(1000) {
					t.Errorf("expected size 1000, got %v", cfg.Options["size"])
				}
			},
		},
		{
			name: "name only",
			json: `{"name":"memory"}`,
			checkFn: func(t *testing.T, cfg config.CacheConfig) {
				if cfg.Name != "memory" {
					t.Errorf("expected Name 'memory', got %q", cfg.Name)
				}
				if cfg.Options != nil {
					t.Errorf("expected Options nil, got %v", cfg.Options)
				}
			},
		},
		{
			name: "empty object",
			json: `{}`,
			checkFn: func(t *testing.T, cfg config.CacheConfig) {
				if cfg.Name != "" {
					t.Errorf("expected Name empty, got %q", cfg.Name)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cfg config.CacheConfig
			err := json.Unmarshal([]byte(tt.json), &cfg)
			if err != nil {
				t.Fatalf("Unmarshal failed: %v", err)
			}

			tt.checkFn(t, cfg)
		})
	}
}

func TestCacheConfig_JSON_RoundTrip(t *testing.T) {
	original := config.CacheConfig{
		Name: "filesystem",
		Options: map[string]any{
			"directory": "/var/cache",
			"size":      1000,
			"ttl":       3600,
		},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded config.CacheConfig
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded.Name != original.Name {
		t.Errorf("Name: expected %q, got %q", original.Name, decoded.Name)
	}

	if decoded.Options["directory"] != original.Options["directory"] {
		t.Errorf("directory: expected %v, got %v", original.Options["directory"], decoded.Options["directory"])
	}

	if decoded.Options["size"] != float64(1000) {
		t.Errorf("size: expected 1000, got %v", decoded.Options["size"])
	}
	if decoded.Options["ttl"] != float64(3600) {
		t.Errorf("ttl: expected 3600, got %v", decoded.Options["ttl"])
	}
}
