// Package cache provides interfaces and implementations for persistent image caching.
//
// This package defines the Cache interface for storing and retrieving rendered document
// images, along with deterministic cache key generation. The interface enables multiple
// storage backends (filesystem, blob storage, databases, in-memory) to be used
// interchangeably.
//
// # Design Philosophy
//
// The cache package follows the Interface-Based Layer Interconnection principle:
//   - Cache interface defines the contract for persistent storage
//   - Implementations provide concrete storage backends
//   - Applications can provide custom implementations if needed
//
// The interface ensures caching abstraction throughout the application, preventing
// direct dependencies on specific storage mechanisms.
//
// # Cache Key Generation
//
// Cache keys are generated using GenerateKey() from normalized input strings that
// uniquely identify a cached image:
//
//	/absolute/path/to/document.pdf/1.png?dpi=300&quality=90
//
// The key generation process:
//  1. Takes normalized input string (caller must sort parameters alphabetically)
//  2. Computes SHA256 hash of the UTF-8 bytes
//  3. Returns hexadecimal encoding (64 characters)
//
// The same input always produces the same key, enabling deterministic cache lookups.
// Different inputs produce different keys with high probability (cryptographic hash).
//
// # Usage Example
//
// Basic cache-aware rendering with proper error handling:
//
//	// Generate cache key from normalized parameters
//	key := cache.GenerateKey("/path/to/doc.pdf/1.png?dpi=300&quality=90")
//
//	// Check cache before rendering
//	entry, err := cache.Get(key)
//	if err == nil {
//	    return entry.Data, nil // Cache hit
//	}
//	if !errors.Is(err, cache.ErrCacheEntryNotFound) {
//	    return nil, err // Real storage error
//	}
//
//	// Cache miss - render and store
//	imageData := renderPage(...)
//	entry = &cache.CacheEntry{
//	    Key:      key,
//	    Data:     imageData,
//	    Filename: "document.1.png",
//	    MimeType: "image/png",
//	}
//	if err := cache.Set(entry); err != nil {
//	    return nil, err
//	}
//	return imageData, nil
//
// Cache invalidation:
//
//	// Remove specific entry
//	err := cache.Invalidate(key)
//	if err != nil {
//	    log.Error("failed to invalidate cache", "error", err)
//	}
//
//	// Clear entire cache
//	err := cache.Clear()
//	if err != nil {
//	    log.Error("failed to clear cache", "error", err)
//	}
//
// # Cache Entry Metadata
//
// CacheEntry stores both image data and metadata needed for serving cached images:
//   - Key: SHA256 hash identifying the cache entry
//   - Data: Raw image bytes (PNG, JPEG, etc.)
//   - Filename: Suggested filename in format "basename.pagenum.ext"
//   - MimeType: MIME content type (e.g., "image/png", "image/jpeg")
//
// The metadata enables proper content type handling and meaningful filenames when
// retrieving cached images for HTTP responses or file storage.
//
// # Error Handling
//
// The package distinguishes between cache misses (expected) and storage failures:
//
//	entry, err := cache.Get(key)
//	if errors.Is(err, cache.ErrCacheEntryNotFound) {
//	    // Expected: cache miss, proceed with fallback
//	} else if err != nil {
//	    // Unexpected: storage failure, handle error
//	}
//
// Use errors.Is() to detect cache misses and continue with fallback operations.
// Other errors indicate actual storage failures requiring error handling.
//
// # Thread Safety
//
// All Cache implementations must be safe for concurrent use by multiple goroutines.
// The interface contract requires implementations to handle concurrent Get, Set,
// Invalidate, and Clear operations without corruption or race conditions.
//
// # Implementation Requirements
//
// Cache implementations should:
//   - Support concurrent access from multiple goroutines
//   - Return ErrCacheEntryNotFound for missing keys (not other error types)
//   - Make Invalidate idempotent (no error if key doesn't exist)
//   - Ensure atomic operations where possible
//   - Handle storage failures gracefully with meaningful errors
package cache
