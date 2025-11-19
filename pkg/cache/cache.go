package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
)

// CacheEntry represents a cached image with associated metadata.
//
// Cache entries store both the binary image data and metadata needed for
// serving or storing the cached image. The metadata enables proper content
// type handling and meaningful filename suggestions when retrieving cached
// images.
type CacheEntry struct {
	// Key is the unique cache key generated from document path, page number,
	// and rendering settings. This is a SHA256 hash in hexadecimal format.
	Key string

	// Data contains the raw image bytes (PNG, JPEG, etc.).
	Data []byte

	// Filename is a suggested filename for the cached image, formatted as
	// "basename.pagenum.ext" (e.g., "document.1.png").
	Filename string
}

// Cache defines the interface for storing and retrieving rendered document images.
//
// Cache implementations provide persistent storage for rendered images to avoid
// redundant conversions. The interface is implementation-agnostic, supporting
// filesystem, blob storage, databases, or in-memory caches.
//
// Implementations must be safe for concurrent use by multiple goroutines.
//
// Cache keys are generated using GenerateKey() from normalized inputs (document
// path, page number, and rendering configuration). The same inputs always produce
// the same key, enabling deterministic cache lookups.
//
// Example usage:
//
//	// Check cache before rendering
//	entry, err := cache.Get(key)
//	if err == nil {
//	    return entry.Data, nil // Cache hit
//	}
//	if !errors.Is(err, cache.ErrCacheEntryNotFound) {
//	    return nil, err // Real error
//	}
//
//	// Cache miss - render and store
//	imageData := renderPage(...)
//	entry = &cache.CacheEntry{
//	    Key:      key,
//	    Data:     imageData,
//	    Filename: "document.1.png",
//	}
//	cache.Set(entry)
type Cache interface {
	// Get retrieves a cache entry by key.
	//
	// Returns the cache entry if found. If the key does not exist, returns
	// ErrCacheEntryNotFound. Other errors indicate storage failures.
	//
	// Callers should distinguish between cache misses (expected) and real
	// errors using errors.Is(err, ErrCacheEntryNotFound).
	Get(key string) (*CacheEntry, error)

	// Set stores a cache entry.
	//
	// The entry's Key field must be populated. If an entry with the same key
	// already exists, it is replaced.
	//
	// Returns an error if the cache entry cannot be stored.
	Set(entry *CacheEntry) error

	// Invalidate removes a cache entry by key.
	//
	// If the key does not exist, Invalidate returns nil (idempotent operation).
	// This allows safe cleanup without checking if entries exist first.
	//
	// Returns an error only if the invalidation operation fails.
	Invalidate(key string) error

	// Clear removes all cache entries.
	//
	// This operation should be used carefully as it removes all cached data.
	// Implementations should ensure this operation is safe and atomic where possible.
	//
	// Returns an error if the cache cannot be cleared.
	Clear() error
}

// ErrCacheEntryNotFound indicates a cache key was not found.
//
// This error distinguishes cache misses (expected behavior) from actual storage
// failures. Callers should use errors.Is(err, ErrCacheEntryNotFound) to detect
// cache misses and continue with fallback operations.
var ErrCacheEntryNotFound = errors.New("cache entry not found")

// GenerateKey creates a deterministic cache key from an input string.
//
// The input string should be a normalized representation of all factors that
// uniquely identify a cached image, typically formatted as a path with query
// parameters:
//
//	/absolute/path/to/document.pdf/1.png?dpi=300&quality=90&format=png
//
// The key generation process:
//  1. Takes the input string (caller must normalize parameters)
//  2. Computes SHA256 hash of the UTF-8 bytes
//  3. Returns hexadecimal encoding (64 characters)
//
// The same input always produces the same key, enabling reliable cache lookups.
// Different inputs produce different keys with high probability (cryptographic hash).
//
// Note: The caller is responsible for normalizing the input string (e.g., sorting
// query parameters alphabetically) to ensure consistent key generation.
func GenerateKey(input string) string {
	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:])
}
