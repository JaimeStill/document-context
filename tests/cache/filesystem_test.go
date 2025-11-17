package cache_test

import (
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/JaimeStill/document-context/pkg/cache"
	"github.com/JaimeStill/document-context/pkg/config"
)

func TestNewFilesystem_ValidDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.CacheConfig{
		Name: "filesystem",
		Options: map[string]any{
			"directory": tmpDir,
		},
	}

	c, err := cache.Create(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if c == nil {
		t.Fatal("expected non-nil cache")
	}
}

func TestNewFilesystem_DirectoryCreation(t *testing.T) {
	tmpDir := t.TempDir()
	cacheDir := filepath.Join(tmpDir, "nonexistent", "cache")

	cfg := &config.CacheConfig{
		Name: "filesystem",
		Options: map[string]any{
			"directory": cacheDir,
		},
	}

	c, err := cache.Create(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if c == nil {
		t.Fatal("expected non-nil cache")
	}

	// Verify directory was created
	info, err := os.Stat(cacheDir)
	if err != nil {
		t.Fatalf("cache directory not created: %v", err)
	}

	if !info.IsDir() {
		t.Error("expected cache directory to be a directory")
	}
}

func TestNewFilesystem_MissingDirectory(t *testing.T) {
	cfg := &config.CacheConfig{
		Name: "filesystem",
		Options: map[string]any{},
	}

	_, err := cache.Create(cfg)
	if err == nil {
		t.Fatal("expected error for missing directory option")
	}

	expectedMsg := "directory option is required"
	if err.Error() != expectedMsg {
		t.Errorf("expected error %q, got %q", expectedMsg, err.Error())
	}
}

func TestNewFilesystem_EmptyDirectory(t *testing.T) {
	cfg := &config.CacheConfig{
		Name: "filesystem",
		Options: map[string]any{
			"directory": "",
		},
	}

	_, err := cache.Create(cfg)
	if err == nil {
		t.Fatal("expected error for empty directory")
	}

	expectedMsg := "directory option cannot be empty"
	if err.Error() != expectedMsg {
		t.Errorf("expected error %q, got %q", expectedMsg, err.Error())
	}
}

func TestNewFilesystem_InvalidDirectoryType(t *testing.T) {
	cfg := &config.CacheConfig{
		Name: "filesystem",
		Options: map[string]any{
			"directory": 12345,
		},
	}

	_, err := cache.Create(cfg)
	if err == nil {
		t.Fatal("expected error for invalid directory type")
	}

	expectedMsg := "directory option must be a string"
	if err.Error() != expectedMsg {
		t.Errorf("expected error %q, got %q", expectedMsg, err.Error())
	}
}

func TestFilesystemCache_SetAndGet(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.CacheConfig{
		Name: "filesystem",
		Options: map[string]any{
			"directory": tmpDir,
		},
	}

	c, err := cache.Create(cfg)
	if err != nil {
		t.Fatalf("unexpected error creating cache: %v", err)
	}

	// Set an entry
	entry := &cache.CacheEntry{
		Key:      "test-key",
		Data:     []byte("test data"),
		Filename: "test.png",
	}

	if err := c.Set(entry); err != nil {
		t.Fatalf("unexpected error setting entry: %v", err)
	}

	// Get the entry
	retrieved, err := c.Get("test-key")
	if err != nil {
		t.Fatalf("unexpected error getting entry: %v", err)
	}

	if retrieved.Key != entry.Key {
		t.Errorf("expected key %q, got %q", entry.Key, retrieved.Key)
	}

	if string(retrieved.Data) != string(entry.Data) {
		t.Errorf("expected data %q, got %q", entry.Data, retrieved.Data)
	}

	if retrieved.Filename != entry.Filename {
		t.Errorf("expected filename %q, got %q", entry.Filename, retrieved.Filename)
	}
}

func TestFilesystemCache_GetNotFound(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.CacheConfig{
		Name: "filesystem",
		Options: map[string]any{
			"directory": tmpDir,
		},
	}

	c, err := cache.Create(cfg)
	if err != nil {
		t.Fatalf("unexpected error creating cache: %v", err)
	}

	// Try to get non-existent entry
	_, err = c.Get("nonexistent")
	if err == nil {
		t.Fatal("expected error for non-existent entry")
	}

	if !errors.Is(err, cache.ErrCacheEntryNotFound) {
		t.Errorf("expected ErrCacheEntryNotFound, got %v", err)
	}
}

func TestFilesystemCache_Invalidate(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.CacheConfig{
		Name: "filesystem",
		Options: map[string]any{
			"directory": tmpDir,
		},
	}

	c, err := cache.Create(cfg)
	if err != nil {
		t.Fatalf("unexpected error creating cache: %v", err)
	}

	// Set an entry
	entry := &cache.CacheEntry{
		Key:      "test-key",
		Data:     []byte("test data"),
		Filename: "test.png",
	}

	if err := c.Set(entry); err != nil {
		t.Fatalf("unexpected error setting entry: %v", err)
	}

	// Verify it exists
	if _, err := c.Get("test-key"); err != nil {
		t.Fatalf("unexpected error getting entry: %v", err)
	}

	// Invalidate the entry
	if err := c.Invalidate("test-key"); err != nil {
		t.Fatalf("unexpected error invalidating entry: %v", err)
	}

	// Verify it no longer exists
	_, err = c.Get("test-key")
	if !errors.Is(err, cache.ErrCacheEntryNotFound) {
		t.Errorf("expected ErrCacheEntryNotFound after invalidation, got %v", err)
	}
}

func TestFilesystemCache_InvalidateNonexistent(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.CacheConfig{
		Name: "filesystem",
		Options: map[string]any{
			"directory": tmpDir,
		},
	}

	c, err := cache.Create(cfg)
	if err != nil {
		t.Fatalf("unexpected error creating cache: %v", err)
	}

	// Invalidate non-existent entry (should be idempotent)
	if err := c.Invalidate("nonexistent"); err != nil {
		t.Errorf("expected nil error for invalidating nonexistent entry, got %v", err)
	}
}

func TestFilesystemCache_Clear(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.CacheConfig{
		Name: "filesystem",
		Options: map[string]any{
			"directory": tmpDir,
		},
	}

	c, err := cache.Create(cfg)
	if err != nil {
		t.Fatalf("unexpected error creating cache: %v", err)
	}

	// Set multiple entries
	for i := 0; i < 5; i++ {
		entry := &cache.CacheEntry{
			Key:      cache.GenerateKey("test-key-" + string(rune('a'+i))),
			Data:     []byte("test data"),
			Filename: "test.png",
		}
		if err := c.Set(entry); err != nil {
			t.Fatalf("unexpected error setting entry %d: %v", i, err)
		}
	}

	// Clear all entries
	if err := c.Clear(); err != nil {
		t.Fatalf("unexpected error clearing cache: %v", err)
	}

	// Verify directory is empty (only directories should remain, and they should be gone)
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error reading cache directory: %v", err)
	}

	if len(entries) != 0 {
		t.Errorf("expected empty cache directory, found %d entries", len(entries))
	}
}

func TestFilesystemCache_DirectoryPerKey(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.CacheConfig{
		Name: "filesystem",
		Options: map[string]any{
			"directory": tmpDir,
		},
	}

	c, err := cache.Create(cfg)
	if err != nil {
		t.Fatalf("unexpected error creating cache: %v", err)
	}

	key := cache.GenerateKey("test")

	entry := &cache.CacheEntry{
		Key:      key,
		Data:     []byte("test data"),
		Filename: "document.1.png",
	}

	if err := c.Set(entry); err != nil {
		t.Fatalf("unexpected error setting entry: %v", err)
	}

	// Verify directory structure: cache_root/[key]/[filename]
	keyDir := filepath.Join(tmpDir, key)
	info, err := os.Stat(keyDir)
	if err != nil {
		t.Fatalf("expected key directory to exist: %v", err)
	}

	if !info.IsDir() {
		t.Error("expected key path to be a directory")
	}

	// Verify file exists
	filePath := filepath.Join(keyDir, "document.1.png")
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("expected cache file to exist: %v", err)
	}

	if string(data) != "test data" {
		t.Errorf("expected file content %q, got %q", "test data", string(data))
	}
}

func TestFilesystemCache_ConcurrentAccess(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.CacheConfig{
		Name: "filesystem",
		Options: map[string]any{
			"directory": tmpDir,
		},
	}

	c, err := cache.Create(cfg)
	if err != nil {
		t.Fatalf("unexpected error creating cache: %v", err)
	}

	var wg sync.WaitGroup
	errors := make(chan error, 100)

	// Concurrent writes
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			entry := &cache.CacheEntry{
				Key:      cache.GenerateKey("concurrent-" + string(rune('a'+id%26))),
				Data:     []byte("test data"),
				Filename: "test.png",
			}
			if err := c.Set(entry); err != nil {
				errors <- err
			}
		}(i)
	}

	// Concurrent reads
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			key := cache.GenerateKey("concurrent-" + string(rune('a'+id%26)))
			_, _ = c.Get(key) // May or may not exist, just test for races
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for any errors
	for err := range errors {
		t.Errorf("concurrent operation failed: %v", err)
	}
}

func TestFilesystemCache_GetCorruption_MultipleFiles(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.CacheConfig{
		Name: "filesystem",
		Options: map[string]any{
			"directory": tmpDir,
		},
	}

	c, err := cache.Create(cfg)
	if err != nil {
		t.Fatalf("unexpected error creating cache: %v", err)
	}

	key := cache.GenerateKey("corruption-test")

	// Manually create corrupted state: multiple files in key directory
	keyDir := filepath.Join(tmpDir, key)
	os.MkdirAll(keyDir, 0755)
	os.WriteFile(filepath.Join(keyDir, "file1.png"), []byte("data1"), 0644)
	os.WriteFile(filepath.Join(keyDir, "file2.png"), []byte("data2"), 0644)

	// Get should detect corruption
	_, err = c.Get(key)
	if err == nil {
		t.Fatal("expected error for corrupted cache (multiple files)")
	}

	expectedMsg := "cache corruption: expected 1 file, found 2"
	if err.Error() != expectedMsg {
		t.Errorf("expected error %q, got %q", expectedMsg, err.Error())
	}
}

func TestFilesystemCache_GetCorruption_DirectoryInsteadOfFile(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.CacheConfig{
		Name: "filesystem",
		Options: map[string]any{
			"directory": tmpDir,
		},
	}

	c, err := cache.Create(cfg)
	if err != nil {
		t.Fatalf("unexpected error creating cache: %v", err)
	}

	key := cache.GenerateKey("corruption-test-dir")

	// Manually create corrupted state: directory instead of file
	keyDir := filepath.Join(tmpDir, key)
	subDir := filepath.Join(keyDir, "subdir")
	os.MkdirAll(subDir, 0755)

	// Get should detect corruption
	_, err = c.Get(key)
	if err == nil {
		t.Fatal("expected error for corrupted cache (directory instead of file)")
	}

	expectedMsg := "cache corruption: expected file, found directory"
	if err.Error() != expectedMsg {
		t.Errorf("expected error %q, got %q", expectedMsg, err.Error())
	}
}
