package config

import "maps"

// CacheConfig defines configuration for cache implementations.
//
// This structure uses a name-based approach where Name identifies the cache
// implementation type, and Options provides implementation-specific settings.
// The flexible Options map allows different cache implementations to accept
// different configuration parameters without requiring schema changes.
//
// Validation of Name and Options values is performed by the consuming package.
type CacheConfig struct {
	Name    string         `json:"name"` // Cache implementation name (e.g., "memory", "filesystem")
	Logger  LoggerConfig   `json:"logger"`
	Options map[string]any `json:"options,omitempty"` // Implementation-specific options
}

// DefaultCacheConfig returns a CacheConfig with empty default values.
//
// The default configuration has no cache implementation selected (empty Name)
// and an initialized but empty Options map.
func DefaultCacheConfig() CacheConfig {
	return CacheConfig{
		Name:    "",
		Logger:  DefaultLoggerConfig(),
		Options: make(map[string]any),
	}
}

// Merge overlays non-empty values from source onto the receiver.
//
// Merge semantics:
//   - Name: only merge if source is non-empty
//   - Options: merge using maps.Copy, which overlays source entries onto receiver map
//
// For Options, existing keys are overwritten by source values, and new keys are added.
// This enables layered configuration and option overrides.
func (c *CacheConfig) Merge(source *CacheConfig) {
	if source == nil {
		return
	}

	if source.Name != "" {
		c.Name = source.Name
	}

	c.Logger.Merge(&source.Logger)

	if source.Options != nil {
		if c.Options == nil {
			c.Options = make(map[string]any)
		}
		maps.Copy(c.Options, source.Options)
	}
}
