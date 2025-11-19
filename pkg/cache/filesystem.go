package cache

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/JaimeStill/document-context/pkg/config"
	"github.com/JaimeStill/document-context/pkg/logger"
)

// FilesystemCache implements Cache using a directory-per-key storage structure.
//
// Each cache entry is stored in its own subdirectory under the cache root, with
// the subdirectory name matching the cache key. The entry's data is written to
// a file within that subdirectory using the entry's filename.
//
// Storage structure: <cache_root>/<key>/<filename>
//
// FilesystemCache is thread-safe for concurrent operations on different keys,
// but operations on the same key are not atomic across multiple goroutines.
type FilesystemCache struct {
	directory string
	logger    logger.Logger
}

// NewFilesystem creates a FilesystemCache from configuration.
//
// This function serves as the factory registered with the cache registry.
// It requires a "directory" option in the configuration's Options map specifying
// the cache root directory. The directory will be created if it doesn't exist.
//
// Configuration options:
//   - directory (string, required): Path to the cache root directory
//
// Returns an error if the directory option is missing, invalid, or if the
// directory cannot be created.
func NewFilesystem(c *config.CacheConfig) (Cache, error) {
	fsConfig, err := parseFilesystemConfig(c.Options)
	if err != nil {
		return nil, err
	}

	log, err := logger.NewSlogger(c.Logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}

	abs, err := filepath.Abs(fsConfig.Directory)
	if err != nil {
		return nil, fmt.Errorf("failed to normalize directory path: %w", err)
	}

	if err := os.MkdirAll(abs, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	return &FilesystemCache{
		directory: abs,
		logger:    log,
	}, nil
}

// Get retrieves a cache entry by key.
//
// Returns ErrCacheEntryNotFound if the key directory doesn't exist.
// Returns an error if the cache is corrupted (wrong number of files or
// directory found instead of file) or if reading the cached data fails.
func (fc *FilesystemCache) Get(key string) (*CacheEntry, error) {
	keyDir := filepath.Join(fc.directory, key)

	entries, err := os.ReadDir(keyDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			fc.logger.Debug("cache.get", "key", key, "found", false)
			return nil, ErrCacheEntryNotFound
		}
		return nil, fmt.Errorf("failed to read cache directory: %w", err)
	}

	if len(entries) != 1 {
		return nil, fmt.Errorf("cache corruption: expected 1 file, found %d", len(entries))
	}

	entry := entries[0]
	if entry.IsDir() {
		return nil, fmt.Errorf("cache corruption: expected file, found directory")
	}

	filename := entry.Name()
	filePath := filepath.Join(keyDir, filename)

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read cache file: %w", err)
	}

	fc.logger.Debug("cache.get", "key", key, "found", true)

	return &CacheEntry{
		Key:      key,
		Data:     data,
		Filename: filename,
	}, nil
}

// Set stores a cache entry.
//
// Creates the key directory if it doesn't exist and writes the entry's
// data to a file with the entry's filename. If a file already exists for
// this key, it will be overwritten.
func (fc *FilesystemCache) Set(entry *CacheEntry) error {
	keyDir := filepath.Join(fc.directory, entry.Key)

	if err := os.MkdirAll(keyDir, 0755); err != nil {
		return fmt.Errorf("failed to create cache key directory: %w", err)
	}

	filePath := filepath.Join(keyDir, entry.Filename)

	if err := os.WriteFile(filePath, entry.Data, 0644); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	fc.logger.Debug("cache.set", "key", entry.Key, "size", len(entry.Data))

	return nil
}

// Invalidate removes a cache entry by key.
//
// Deletes the entire key directory and its contents. This operation is
// idempotent - invalidating a non-existent key succeeds without error.
func (fc *FilesystemCache) Invalidate(key string) error {
	keyDir := filepath.Join(fc.directory, key)

	err := os.RemoveAll(keyDir)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("failed to invalidate cache entry: %w", err)
	}

	fc.logger.Debug("cache.invalidate", "key", key)

	return nil
}

// Clear removes all cache entries.
//
// Iterates through all subdirectories in the cache root and removes them.
// If individual removals fail, they are logged as warnings but don't stop
// the operation. Always returns nil.
func (fc *FilesystemCache) Clear() error {
	entries, err := os.ReadDir(fc.directory)
	if err != nil {
		return fmt.Errorf("failed to read cache directory: %w", err)
	}

	removedCount := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		path := filepath.Join(fc.directory, entry.Name())
		if err := os.RemoveAll(path); err != nil {
			fc.logger.Warn("cache.clear failed to remove entry", "path", path, "error", err)
			continue
		}
		removedCount++
	}

	fc.logger.Info("cache.clear", "file_count", removedCount)

	return nil
}

func init() {
	Register("filesystem", NewFilesystem)
}

// FilesystemCacheConfig contains configuration specific to FilesystemCache.
//
// This typed configuration is parsed from the generic Options map in CacheConfig,
// following the Configuration Composition Pattern for interfaces with multiple
// implementations.
type FilesystemCacheConfig struct {
	Directory string
}

// parseFilesystemConfig extracts and validates filesystem-specific configuration
// from the generic Options map.
//
// Validates that the "directory" option exists, is a string, and is non-empty.
// Returns a typed FilesystemCacheConfig or an error if validation fails.
func parseFilesystemConfig(options map[string]any) (*FilesystemCacheConfig, error) {
	dir, ok := options["directory"]
	if !ok {
		return nil, fmt.Errorf("directory option is required")
	}

	directory, ok := dir.(string)
	if !ok {
		return nil, fmt.Errorf("directory option must be a string")
	}

	if directory == "" {
		return nil, fmt.Errorf("directory option cannot be empty")
	}

	return &FilesystemCacheConfig{
		Directory: directory,
	}, nil
}
