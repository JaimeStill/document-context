package cache

import (
	"fmt"
	"sort"
	"sync"

	"github.com/JaimeStill/document-context/pkg/config"
)

// Factory is a function that creates a Cache instance from configuration.
//
// Cache implementations register their factory functions using Register.
// The factory receives a CacheConfig containing implementation-specific
// options and returns a configured Cache instance or an error if the
// configuration is invalid.
type Factory func(c *config.CacheConfig) (Cache, error)

type registry struct {
	factories map[string]Factory
	mu        sync.RWMutex
}

var register = &registry{
	factories: make(map[string]Factory),
}

// Register registers a cache factory function under the given name.
//
// Cache implementations should call Register in their init() functions to
// make themselves available for creation via Create. If a factory is
// registered multiple times with the same name, the latest registration
// silently overwrites the previous one.
//
// Register panics if name is empty or factory is nil.
//
// Example:
//
//	func init() {
//	    cache.Register("filesystem", NewFilesystem)
//	}
func Register(name string, factory Factory) {
	if name == "" {
		panic("cache: Register name is empty")
	}
	if factory == nil {
		panic("cache: Register factory is nil")
	}

	register.mu.Lock()
	defer register.mu.Unlock()
	register.factories[name] = factory
}

// Create instantiates a Cache using the registered factory for the given configuration.
//
// The configuration's Name field must match a registered cache implementation.
// Create returns an error if the name is empty or unknown, or if the factory
// function returns an error during cache creation.
//
// Example:
//
//	cfg := &config.CacheConfig{
//	    Name: "filesystem",
//	    Options: map[string]any{"directory": "/var/cache"},
//	}
//	cache, err := cache.Create(cfg)
func Create(c *config.CacheConfig) (Cache, error) {
	if c.Name == "" {
		return nil, fmt.Errorf("cache name cannot be empty")
	}

	register.mu.RLock()
	factory, ok := register.factories[c.Name]
	register.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("unknown cache type: %s", c.Name)
	}

	return factory(c)
}

// ListCaches returns the names of all registered cache implementations
// in alphabetical order.
//
// This function is useful for discovering available cache types and for
// validation or help text generation.
func ListCaches() []string {
	register.mu.RLock()
	defer register.mu.RUnlock()

	names := make([]string, 0, len(register.factories))
	for name := range register.factories {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
