# document-context Usage Guide

This guide provides comprehensive documentation for using the document-context library, including detailed examples, configuration patterns, caching strategies, and troubleshooting.

For a quick introduction, see [README.md](./README.md). For technical architecture details, see [ARCHITECTURE.md](./ARCHITECTURE.md).

## Table of Contents

- [Basic Usage](#basic-usage)
- [Configuration](#configuration)
- [Caching](#caching)
- [Image Enhancement Filters](#image-enhancement-filters)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)

## Basic Usage

### Opening and Converting a PDF

```go
package main

import (
    "fmt"
    "log"

    "github.com/JaimeStill/document-context/pkg/config"
    "github.com/JaimeStill/document-context/pkg/document"
    "github.com/JaimeStill/document-context/pkg/image"
)

func main() {
    // Create renderer with default settings (PNG, 300 DPI)
    renderer, err := image.NewImageMagickRenderer(config.DefaultImageConfig())
    if err != nil {
        log.Fatalf("Failed to create renderer: %v", err)
    }

    // Open PDF document
    doc, err := document.OpenPDF("example.pdf")
    if err != nil {
        log.Fatalf("Failed to open PDF: %v", err)
    }
    defer doc.Close()

    // Extract and convert first page
    page, err := doc.ExtractPage(1)
    if err != nil {
        log.Fatalf("Failed to extract page: %v", err)
    }

    // Convert to image (no caching)
    imageData, err := page.ToImage(renderer, nil)
    if err != nil {
        log.Fatalf("Failed to convert page: %v", err)
    }

    fmt.Printf("Rendered page 1: %d bytes\n", len(imageData))
}
```

### Processing Multiple Pages

```go
// Extract all pages
pages, err := doc.ExtractAllPages()
if err != nil {
    log.Fatalf("Failed to extract pages: %v", err)
}

// Convert each page
for i, page := range pages {
    imageData, err := page.ToImage(renderer, nil)
    if err != nil {
        log.Printf("Failed to convert page %d: %v", page.Number(), err)
        continue
    }

    // Save to file
    filename := fmt.Sprintf("page-%d.png", page.Number())
    if err := os.WriteFile(filename, imageData, 0644); err != nil {
        log.Printf("Failed to save page %d: %v", page.Number(), err)
    }
}
```

### Concurrent Page Processing

```go
var wg sync.WaitGroup
errors := make(chan error, doc.PageCount())
results := make(chan struct{
    pageNum int
    data    []byte
}, doc.PageCount())

// Process pages concurrently
for pageNum := 1; pageNum <= doc.PageCount(); pageNum++ {
    wg.Add(1)
    go func(num int) {
        defer wg.Done()

        page, err := doc.ExtractPage(num)
        if err != nil {
            errors <- fmt.Errorf("page %d extraction: %w", num, err)
            return
        }

        imageData, err := page.ToImage(renderer, cache)
        if err != nil {
            errors <- fmt.Errorf("page %d rendering: %w", num, err)
            return
        }

        results <- struct{
            pageNum int
            data    []byte
        }{num, imageData}
    }(pageNum)
}

// Wait for completion
wg.Wait()
close(errors)
close(results)

// Handle errors
for err := range errors {
    log.Printf("Error: %v", err)
}

// Process results
for result := range results {
    fmt.Printf("Page %d: %d bytes\n", result.pageNum, len(result.data))
}
```

## Configuration

### Image Format and Quality

```go
// High-quality PNG (lossless, larger files)
pngCfg := config.ImageConfig{
    Format: "png",
    DPI:    300,  // 300 DPI for high quality
}

// JPEG with quality control (lossy, smaller files)
jpegCfg := config.ImageConfig{
    Format:  "jpg",
    DPI:     300,
    Quality: 85,  // 1-100, higher = better quality
}

// Low-resolution preview
previewCfg := config.ImageConfig{
    Format: "png",
    DPI:    150,  // Lower DPI for faster rendering
}

// Print quality
printCfg := config.ImageConfig{
    Format: "png",
    DPI:    600,  // High DPI for printing
}
```

### Configuration Defaults and Merging

```go
// Get default configuration
defaultCfg := config.DefaultImageConfig()
// Returns: Format="png", DPI=300, Quality=0

// Create partial configuration
customCfg := config.ImageConfig{
    Format: "jpg",
    // DPI and Quality not specified
}

// Apply defaults via Finalize()
customCfg.Finalize()
// Now: Format="jpg", DPI=300, Quality=90 (JPEG default)

// Merge configurations (overlay pattern)
baseCfg := config.ImageConfig{
    Format: "png",
    DPI:    150,
}

overrideCfg := &config.ImageConfig{
    DPI: 300,  // Override DPI only
}

baseCfg.Merge(overrideCfg)
// Result: Format="png" (unchanged), DPI=300 (overridden)
```

## Caching

### Basic Cache Setup

```go
import "github.com/JaimeStill/document-context/pkg/cache"

// Create cache configuration
cacheCfg := &config.CacheConfig{
    Name: "filesystem",
    Logger: config.LoggerConfig{
        Level:  config.LogLevelInfo,
        Output: config.LoggerOutputStdout,
    },
    Options: map[string]any{
        "directory": "/var/cache/document-context",
    },
}

// Create cache instance
c, err := cache.Create(cacheCfg)
if err != nil {
    log.Fatalf("Cache creation failed: %v", err)
}

// Use cache with ToImage()
imageData, err := page.ToImage(renderer, c)
```

### Cache Behavior Examples

```go
// First call: Cache miss - renders and stores (~500ms)
start := time.Now()
img1, err := page.ToImage(renderer, c)
fmt.Printf("First call: %v (%d bytes)\n", time.Since(start), len(img1))
// Output: First call: 523ms (146019 bytes)

// Second call: Cache hit - returns immediately (~1ms)
start = time.Now()
img2, err := page.ToImage(renderer, c)
fmt.Printf("Second call: %v (%d bytes)\n", time.Since(start), len(img2))
// Output: Second call: 1ms (146019 bytes)

// Verify data matches
if bytes.Equal(img1, img2) {
    fmt.Println("Cache hit returned identical data")
}
```

### Cache Key Behavior

Cache keys are generated from all rendering parameters. The same configuration always produces the same key.

**Different pages produce different keys:**
```go
page1, _ := doc.ExtractPage(1)
page2, _ := doc.ExtractPage(2)

img1, _ := page1.ToImage(renderer, c)  // Cache miss, stores with key A
img2, _ := page2.ToImage(renderer, c)  // Cache miss, stores with key B

img1Again, _ := page1.ToImage(renderer, c)  // Cache hit, key A
```

**Different formats produce different keys:**
```go
pngRenderer, _ := image.NewImageMagickRenderer(config.ImageConfig{Format: "png", DPI: 300})
jpgRenderer, _ := image.NewImageMagickRenderer(config.ImageConfig{Format: "jpg", DPI: 300, Quality: 85})

pngImage, _ := page.ToImage(pngRenderer, c)  // Cache miss, stores with key A
jpgImage, _ := page.ToImage(jpgRenderer, c)  // Cache miss, stores with key B (different format)
```

**Different DPI values produce different keys:**
```go
renderer150, _ := image.NewImageMagickRenderer(config.ImageConfig{Format: "png", DPI: 150})
renderer300, _ := image.NewImageMagickRenderer(config.ImageConfig{Format: "png", DPI: 300})

img150, _ := page.ToImage(renderer150, c)  // Cache miss, stores with key A
img300, _ := page.ToImage(renderer300, c)  // Cache miss, stores with key B (different DPI)
```

**Filter changes produce different keys:**
```go
baseCfg := config.ImageConfig{Format: "png", DPI: 300}
baseRenderer, _ := image.NewImageMagickRenderer(baseCfg)

filteredCfg := config.ImageConfig{
    Format: "png",
    DPI:    300,
    Options: map[string]any{
        "brightness": 110,  // 10% brighter
    },
}
filteredRenderer, _ := image.NewImageMagickRenderer(filteredCfg)

baseImg, _ := page.ToImage(baseRenderer, c)      // Cache miss, key A (no filters)
filteredImg, _ := page.ToImage(filteredRenderer, c)  // Cache miss, key B (brightness=110)
```

### Cache Directory Structure

The filesystem cache uses a directory-per-key structure:

```
/var/cache/document-context/
├── a3a6788c43b16d73b83cc01f34ea39e416bf1fcbff5cbaccceb818b1118f06ed/
│   └── document.1.png
├── d4e6f8a1b2c3f9e5d7a8c4b6f1e3d5a7c9b8e6f4d2a1c3b5f7e9d8c6a4b2f1e3/
│   └── document.2.png
└── f9b2c5d3e4a6f8b7d9e1c3a5b7d9f1e3c5a7b9d1f3e5c7a9b1d3f5e7c9a1b3d5/
    └── report.1.jpg
```

Each cache key directory contains exactly one file (the rendered image). This structure enables:
- **Visual inspection**: Browse cache to verify rendered images
- **Simple invalidation**: Delete a directory to invalidate a specific entry
- **Corruption detection**: Multiple files in a directory indicates corruption
- **Concurrent access**: Different keys can be accessed concurrently

### Cache Management

```go
// Clear all cache entries
err := c.Clear()
if err != nil {
    log.Printf("Cache clear failed: %v", err)
}

// Invalidate specific entry (requires knowing the key)
// Note: Usually you'd regenerate the key from document/renderer
page, _ := doc.ExtractPage(1)
key, _ := cache.GenerateKey(/* build key string */)
err := c.Invalidate(key)

// Manual cache directory management
cacheDir := "/var/cache/document-context"

// Remove specific document's cache entries
// (requires knowing cache keys or clearing all)
os.RemoveAll(cacheDir)  // Nuclear option: removes everything
```

### Cache with Updated Documents

Cache keys are based on document **path** and rendering parameters, not document **content**. If a PDF file is updated in place, the cache key remains the same and stale data may be returned.

**Problem:**
```go
// Render and cache original document
img1, _ := page.ToImage(renderer, c)  // Renders original.pdf

// Update PDF file in place (same path)
updatePDFFile("document.pdf")

// Cache returns stale data!
img2, _ := page.ToImage(renderer, c)  // Returns cached version of original
```

**Solutions:**

1. **Clear cache after updates:**
```go
updatePDFFile("document.pdf")
c.Clear()  // Remove all cached entries
```

2. **Invalidate specific entries** (if you can generate the cache key)

3. **Use versioned paths:**
```go
// Don't update files in place - use versioned filenames
doc1, _ := document.OpenPDF("report-v1.pdf")
doc2, _ := document.OpenPDF("report-v2.pdf")  // Different path = different cache key
```

## Image Enhancement Filters

ImageMagick filters can be applied to improve document clarity and readability.

### Brightness Adjustment

```go
cfg := config.ImageConfig{
    Format: "png",
    DPI:    300,
    Options: map[string]any{
        "brightness": 110,  // Range: 0-200, default 100 (neutral)
    },
}

renderer, _ := image.NewImageMagickRenderer(cfg)
imageData, _ := page.ToImage(renderer, nil)

// brightness=90  → 10% darker
// brightness=100 → neutral (no change)
// brightness=110 → 10% brighter
// brightness=150 → 50% brighter
```

### Contrast Adjustment

```go
cfg := config.ImageConfig{
    Format: "png",
    DPI:    300,
    Options: map[string]any{
        "contrast": 10,  // Range: -100 to +100, default 0 (neutral)
    },
}

// contrast=-20 → reduce contrast
// contrast=0   → neutral (no change)
// contrast=10  → increase contrast
// contrast=50  → strong contrast boost
```

### Saturation Adjustment

```go
cfg := config.ImageConfig{
    Format: "png",
    DPI:    300,
    Options: map[string]any{
        "saturation": 120,  // Range: 0-200, default 100 (neutral)
    },
}

// saturation=0   → grayscale
// saturation=100 → neutral (original colors)
// saturation=120 → 20% more saturated
// saturation=200 → maximum saturation
```

### Rotation

```go
cfg := config.ImageConfig{
    Format: "png",
    DPI:    300,
    Options: map[string]any{
        "rotation": 90,  // Range: 0-360 degrees, clockwise
    },
}

// rotation=90  → rotate 90° clockwise
// rotation=180 → rotate 180° (upside down)
// rotation=270 → rotate 270° clockwise (90° counter-clockwise)
```

### Background Color

```go
cfg := config.ImageConfig{
    Format: "png",
    DPI:    300,
    Options: map[string]any{
        "background": "white",  // Default: "white"
    },
}

// Common values:
// "white"  → white background
// "black"  → black background
// "#f0f0f0" → light gray
// "transparent" → transparent background (PNG only)
```

### Combining Multiple Filters

```go
cfg := config.ImageConfig{
    Format: "png",
    DPI:    300,
    Options: map[string]any{
        "brightness": 110,      // Slightly brighter
        "contrast":   5,        // Slight contrast boost
        "saturation": 105,      // Slightly more saturated
        "rotation":   0,        // No rotation
        "background": "white",  // White background
    },
}

renderer, _ := image.NewImageMagickRenderer(cfg)
enhancedImage, _ := page.ToImage(renderer, cache)
```

**Filter Application Order** (ImageMagick processing order):
1. Rotation
2. Brightness and Saturation (via `-modulate`)
3. Contrast (via `-brightness-contrast`)
4. Background flattening

## Best Practices

### Use Absolute Paths

Cache keys include the absolute document path. Using relative paths can cause cache misses.

```go
// ❌ Relative path (inconsistent cache keys)
doc, _ := document.OpenPDF("../documents/report.pdf")

// ✅ Absolute path (consistent cache keys)
absPath, _ := filepath.Abs("../documents/report.pdf")
doc, _ := document.OpenPDF(absPath)
```

### Defer Document Cleanup

Always close documents to free resources:

```go
doc, err := document.OpenPDF("document.pdf")
if err != nil {
    return err
}
defer doc.Close()  // ✅ Ensures cleanup even on errors
```

### Error Handling

Wrap errors with context for better debugging:

```go
page, err := doc.ExtractPage(pageNum)
if err != nil {
    return fmt.Errorf("failed to extract page %d: %w", pageNum, err)
}

imageData, err := page.ToImage(renderer, cache)
if err != nil {
    return fmt.Errorf("failed to render page %d: %w", pageNum, err)
}
```

### Pre-verify External Dependencies

Check for ImageMagick at application startup:

```go
func init() {
    if _, err := exec.LookPath("magick"); err != nil {
        log.Fatal("ImageMagick not installed - required for PDF rendering")
    }
}
```

### Cache Configuration

```go
// ✅ Good: Appropriate cache directory
cacheCfg := &config.CacheConfig{
    Name: "filesystem",
    Options: map[string]any{
        "directory": "/var/cache/myapp",  // Dedicated cache directory
    },
}

// ❌ Bad: System temp directory (may be cleared)
cacheCfg := &config.CacheConfig{
    Name: "filesystem",
    Options: map[string]any{
        "directory": os.TempDir(),  // ❌ Temp dirs may be cleared
    },
}
```

### Renderer Reuse

Renderers are immutable and thread-safe - create once and reuse:

```go
// ✅ Good: Create once, reuse
renderer, _ := image.NewImageMagickRenderer(cfg)

for pageNum := 1; pageNum <= doc.PageCount(); pageNum++ {
    page, _ := doc.ExtractPage(pageNum)
    imageData, _ := page.ToImage(renderer, cache)  // Reuse renderer
}

// ❌ Bad: Creating renderer in loop
for pageNum := 1; pageNum <= doc.PageCount(); pageNum++ {
    renderer, _ := image.NewImageMagickRenderer(cfg)  // ❌ Unnecessary
    page, _ := doc.ExtractPage(pageNum)
    imageData, _ := page.ToImage(renderer, cache)
}
```

## Troubleshooting

### Cache Issues

**Problem:** Cache misses when cache hits expected

**Diagnosis:**
- Verify document path is absolute (cache keys use absolute paths)
- Check that all rendering parameters match exactly
- Confirm filter values haven't changed

**Solution:**
```go
// Use absolute paths for consistent cache keys
absPath, err := filepath.Abs("relative/path/document.pdf")
doc, _ := document.OpenPDF(absPath)

// Debug cache key components
renderer, _ := image.NewImageMagickRenderer(cfg)
settings := renderer.Settings()
params := renderer.Parameters()
fmt.Printf("Settings: Format=%s, DPI=%d, Quality=%d\n",
    settings.Format, settings.DPI, settings.Quality)
fmt.Printf("Parameters: %v\n", params)
```

---

**Problem:** Cache errors ("cache storage unavailable", "cache write failed")

**Diagnosis:**
- Verify cache directory exists and has write permissions
- Check disk space availability
- Confirm directory path is absolute

**Solution:**
```go
// Ensure cache directory exists
cacheDir := "/var/cache/document-context"
if err := os.MkdirAll(cacheDir, 0755); err != nil {
    return fmt.Errorf("failed to create cache directory: %w", err)
}

// Test write permissions
testFile := filepath.Join(cacheDir, "test.txt")
if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
    return fmt.Errorf("cache directory not writable: %w", err)
}
os.Remove(testFile)
```

---

**Problem:** Cache directory growing too large

**Solution:**
```go
// Implement cache size monitoring
func getCacheSize(dir string) (int64, error) {
    var size int64
    err := filepath.Walk(dir, func(_ string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        if !info.IsDir() {
            size += info.Size()
        }
        return nil
    })
    return size, err
}

// Clear cache if size exceeds threshold
size, _ := getCacheSize("/var/cache/document-context")
maxSize := int64(10 * 1024 * 1024 * 1024)  // 10 GB
if size > maxSize {
    cache.Clear()
}
```

### Rendering Issues

**Problem:** ImageMagick not found

**Error:** `exec: "magick": executable file not found in $PATH`

**Solution:**
```bash
# macOS (Homebrew)
brew install imagemagick

# Ubuntu/Debian
apt-get update && apt-get install -y imagemagick

# Alpine Linux (Docker)
apk add --no-cache imagemagick

# Verify installation
magick --version
```

---

**Problem:** Poor image quality

**Solution:**
```go
// High-quality settings for text-heavy documents
cfg := config.ImageConfig{
    Format:  "png",    // PNG for text (lossless)
    DPI:     300,      // 300 DPI minimum for quality
}

// If using JPEG, use high quality
jpegCfg := config.ImageConfig{
    Format:  "jpg",
    DPI:     300,
    Quality: 90,  // 90-100 for high quality
}
```

---

**Problem:** Rendering too slow

**Diagnosis:** High DPI and filters increase rendering time

**Solution:**
```go
// Use lower DPI for preview/development
previewCfg := config.ImageConfig{
    Format: "png",
    DPI:    150,  // Faster rendering
}

// Or use JPEG with lower quality
fastCfg := config.ImageConfig{
    Format:  "jpg",
    DPI:     150,
    Quality: 70,
}
```

### Concurrency Issues

**Problem:** Race conditions

**Diagnosis:** Run tests with race detector:
```bash
go test -race ./...
```

**Solution:** The library is thread-safe. Ensure proper synchronization in your code:

```go
// ✅ Good: Proper synchronization
var wg sync.WaitGroup
var mu sync.Mutex
results := make([][]byte, doc.PageCount())

for i := 1; i <= doc.PageCount(); i++ {
    wg.Add(1)
    go func(pageNum int) {
        defer wg.Done()

        page, _ := doc.ExtractPage(pageNum)
        imageData, _ := page.ToImage(renderer, cache)

        mu.Lock()
        results[pageNum-1] = imageData
        mu.Unlock()
    }(i)
}

wg.Wait()
```

---

**Problem:** Too many open files

**Error:** `too many open files`

**Solution:**
```go
// Limit concurrent operations
const maxConcurrent = 10
sem := make(chan struct{}, maxConcurrent)

for pageNum := 1; pageNum <= doc.PageCount(); pageNum++ {
    sem <- struct{}{}  // Acquire

    go func(num int) {
        defer func() { <-sem }()  // Release

        page, _ := doc.ExtractPage(num)
        imageData, _ := page.ToImage(renderer, cache)
        // Process imageData...
    }(pageNum)
}

// Wait for all to complete
for i := 0; i < cap(sem); i++ {
    sem <- struct{}{}
}
```

### Configuration Issues

**Problem:** Filter values out of range

**Error:** `brightness must be 0-200, got 250`

**Solution:**
```go
// Validate filter values before creating renderer
brightness := 250
if brightness < 0 || brightness > 200 {
    brightness = 100  // Reset to neutral
}

cfg := config.ImageConfig{
    Options: map[string]any{
        "brightness": brightness,
    },
}
```

---

**Problem:** Configuration not being applied

**Diagnosis:** Forgot to call Finalize() or values are zero

**Solution:**
```go
// ✅ Explicit Finalize()
cfg := config.ImageConfig{Format: "jpg"}
cfg.Finalize()  // Applies defaults for DPI and Quality

// Or use default and merge
cfg := config.DefaultImageConfig()
cfg.Format = "jpg"  // Override specific fields
```

## Additional Resources

- **Architecture Details**: See [ARCHITECTURE.md](./ARCHITECTURE.md) for technical specifications
- **Project Roadmap**: See [PROJECT.md](./PROJECT.md) for development status and future plans
- **Examples**: See `examples/` directory for complete working examples
- **API Documentation**: Run `go doc` in the project directory

## Getting Help

For issues or questions:
1. Check this guide's troubleshooting section
2. Review the [ARCHITECTURE.md](./ARCHITECTURE.md) for technical details
3. Check existing [GitHub Issues](https://github.com/JaimeStill/document-context/issues)
4. Create a new issue with a minimal reproducible example
