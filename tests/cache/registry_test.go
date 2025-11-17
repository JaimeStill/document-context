package cache_test

import (
	"sync"
	"testing"

	"github.com/JaimeStill/document-context/pkg/cache"
	"github.com/JaimeStill/document-context/pkg/config"
)

// mockCache implements cache.Cache for testing
type mockCache struct {
	name string
}

func (m *mockCache) Get(key string) (*cache.CacheEntry, error) {
	return nil, cache.ErrCacheEntryNotFound
}

func (m *mockCache) Set(entry *cache.CacheEntry) error {
	return nil
}

func (m *mockCache) Invalidate(key string) error {
	return nil
}

func (m *mockCache) Clear() error {
	return nil
}

func mockFactory(c *config.CacheConfig) (cache.Cache, error) {
	return &mockCache{name: c.Name}, nil
}

func TestRegister_ValidFactory(t *testing.T) {
	// Register a test cache
	cache.Register("test-cache", mockFactory)

	// Should not panic
	caches := cache.ListCaches()

	found := false
	for _, name := range caches {
		if name == "test-cache" {
			found = true
			break
		}
	}

	if !found {
		t.Error("expected test-cache to be registered")
	}
}

func TestRegister_EmptyName_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for empty name")
		}
	}()

	cache.Register("", mockFactory)
}

func TestRegister_NilFactory_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for nil factory")
		}
	}()

	cache.Register("nil-factory", nil)
}

func TestCreate_ValidCache(t *testing.T) {
	cache.Register("mock-valid", mockFactory)

	cfg := &config.CacheConfig{
		Name: "mock-valid",
	}

	c, err := cache.Create(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if c == nil {
		t.Fatal("expected non-nil cache")
	}
}

func TestCreate_UnknownCache(t *testing.T) {
	cfg := &config.CacheConfig{
		Name: "nonexistent-cache",
	}

	_, err := cache.Create(cfg)
	if err == nil {
		t.Fatal("expected error for unknown cache type")
	}

	expectedMsg := "unknown cache type: nonexistent-cache"
	if err.Error() != expectedMsg {
		t.Errorf("expected error %q, got %q", expectedMsg, err.Error())
	}
}

func TestCreate_EmptyName(t *testing.T) {
	cfg := &config.CacheConfig{
		Name: "",
	}

	_, err := cache.Create(cfg)
	if err == nil {
		t.Fatal("expected error for empty cache name")
	}

	expectedMsg := "cache name cannot be empty"
	if err.Error() != expectedMsg {
		t.Errorf("expected error %q, got %q", expectedMsg, err.Error())
	}
}

func TestListCaches_Sorted(t *testing.T) {
	cache.Register("zebra-cache", mockFactory)
	cache.Register("alpha-cache", mockFactory)
	cache.Register("beta-cache", mockFactory)

	caches := cache.ListCaches()

	if len(caches) < 3 {
		t.Fatalf("expected at least 3 caches, got %d", len(caches))
	}

	// Check that list is sorted
	for i := 1; i < len(caches); i++ {
		if caches[i-1] >= caches[i] {
			t.Errorf("caches not sorted: %q >= %q", caches[i-1], caches[i])
		}
	}
}

func TestRegister_Concurrent(t *testing.T) {
	var wg sync.WaitGroup

	// Register 10 caches concurrently
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			// Using unique names to avoid conflicts
			cache.Register("concurrent-"+string(rune('a'+id)), mockFactory)
		}(i)
	}

	wg.Wait()

	// All should be registered
	caches := cache.ListCaches()
	if len(caches) < 10 {
		t.Errorf("expected at least 10 caches after concurrent registration, got %d", len(caches))
	}
}

func TestCreate_Concurrent(t *testing.T) {
	cache.Register("concurrent-create", mockFactory)

	var wg sync.WaitGroup
	errors := make(chan error, 100)

	// Create 100 cache instances concurrently
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cfg := &config.CacheConfig{
				Name: "concurrent-create",
			}
			_, err := cache.Create(cfg)
			if err != nil {
				errors <- err
			}
		}()
	}

	wg.Wait()
	close(errors)

	// Check for any errors
	for err := range errors {
		t.Errorf("concurrent create failed: %v", err)
	}
}
