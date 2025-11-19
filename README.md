# Document Context

A Go library for converting documents into context-friendly formats suitable for LLM consumption and analysis.

## Status: Pre-Release Development

**document-context** is currently in active development and has not yet reached v0.1.0. The API is subject to change as additional features and format support are added.

## Documentation

- **[ARCHITECTURE.md](./ARCHITECTURE.md)**: Technical specifications and implementation details
- **[PROJECT.md](./PROJECT.md)**: Project scope, philosophy, and roadmap
- **[CLAUDE.md](./CLAUDE.md)**: Development principles and conventions

## Overview

This library provides format-agnostic interfaces for document processing with extensible format support. It was created as a tooling extension for the [go-agents](https://github.com/JaimeStill/go-agents) project but can be used standalone for document processing needs.

**Current Capabilities**:
- PDF document processing
- Page-level image extraction
- Multiple output formats (PNG, JPEG)
- Configurable quality and resolution
- Persistent filesystem caching for rendered images
- Structured logging infrastructure
- Base64 data URI encoding for LLM APIs

## Prerequisites

### Go Version

- Go 1.25.4 or later

### External Dependencies

**ImageMagick** (Required):
- Used for high-quality PDF page rendering
- Must use version 7.0+ with the `magick` command
- Installation varies by platform:

**Verify Installation**:
```bash
magick --version
```

## Installation

```bash
go get github.com/JaimeStill/document-context
```

## Usage Examples

### Basic PDF to Image Conversion

```go
package main

import (
    "fmt"
    "os"

    "github.com/JaimeStill/document-context/pkg/config"
    "github.com/JaimeStill/document-context/pkg/document"
    "github.com/JaimeStill/document-context/pkg/image"
)

func main() {
    // Create configuration (or use config.DefaultImageConfig() for PNG, 300 DPI)
    cfg := config.ImageConfig{
        Format:  "png",    // "png" or "jpg"
        DPI:     300,      // Resolution
        Quality: 85,       // JPEG quality (1-100, ignored for PNG)
        Options: map[string]any{
            "brightness": 110,  // 0-200, 100=neutral
            "contrast":   10,   // -100 to +100, 0=neutral
        },
    }

    // Transform configuration to renderer
    renderer, err := image.NewImageMagickRenderer(cfg)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Invalid configuration: %v\n", err)
        return
    }

    // Open PDF and extract page
    doc, err := document.OpenPDF("report.pdf")
    if err != nil {
        fmt.Fprintf(os.Stderr, "Failed to open PDF: %v\n", err)
        return
    }
    defer doc.Close()

    page, err := doc.ExtractPage(1)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Failed to extract page: %v\n", err)
        return
    }

    // Convert to image (nil = no caching)
    imageData, err := page.ToImage(renderer, nil)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Failed to convert page: %v\n", err)
        return
    }

    // Save image
    err = os.WriteFile("page-1.png", imageData, 0644)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Failed to write image: %v\n", err)
        return
    }

    fmt.Println("Successfully converted page to image")
}
```

**Processing Multiple Pages**:
```go
// Extract all pages
pages, err := doc.ExtractAllPages()
if err != nil {
    return err
}

// Convert each page
for _, page := range pages {
    imageData, err := page.ToImage(renderer, nil)
    // Handle imageData...
}
```

### Using the Filesystem Cache

Enable persistent caching for faster repeated conversions:

```go
// Create cache configuration
cacheCfg := &config.CacheConfig{
    Name: "filesystem",
    Logger: config.LoggerConfig{Level: config.LogLevelInfo},
    Options: map[string]any{
        "directory": "/var/cache/document-context",
    },
}

// Create cache instance
c, err := cache.Create(cacheCfg)
if err != nil {
    return err
}

// Use cache with ToImage()
page, _ := doc.ExtractPage(1)
imageData, err := page.ToImage(renderer, c)  // First call renders and caches
cachedData, err := page.ToImage(renderer, c)  // Second call returns cached data
```

Cache keys are generated from document path, page number, and all rendering parameters (format, DPI, quality, filters). The same configuration always produces the same cache key.

For detailed cache behavior, troubleshooting, and advanced usage, see [GUIDE.md](./GUIDE.md#caching)

### Data URI Encoding for LLM APIs

Convert pages to base64 data URIs for LLM vision APIs:

```go
import "github.com/JaimeStill/document-context/pkg/encoding"

// After converting page to image (as shown above)
imageData, _ := page.ToImage(renderer, nil)

// Encode as data URI
dataURI, err := encoding.EncodeImageDataURI(imageData, document.PNG)
if err != nil {
    return err
}

// Use dataURI with LLM vision API
// response := llm.Vision("Analyze this document", []string{dataURI})
```

For integration with [go-agents](https://github.com/JaimeStill/go-agents), see the go-agents documentation for vision API usage patterns.

## Configuration

The library uses a configuration-to-renderer transformation pattern where configuration (data) transforms into renderers (behavior):

```go
// Create configuration
cfg := config.ImageConfig{
    Format:  "png",    // "png" or "jpg"
    Quality: 85,       // JPEG quality (1-100), ignored for PNG
    DPI:     300,      // Resolution (72/150/300/600)
    Options: map[string]any{  // ImageMagick filters
        "brightness": 110,     // 0-200, 100=neutral
        "contrast":   10,      // -100 to +100, 0=neutral
        "saturation": 100,     // 0-200, 100=neutral
        "rotation":   0,       // 0-360 degrees
        "background": "white", // Color name for alpha channel
    },
}

// Transform to renderer (validates configuration)
renderer, err := image.NewImageMagickRenderer(cfg)

// Or use defaults: PNG, 300 DPI, no filters
renderer, _ := image.NewImageMagickRenderer(config.DefaultImageConfig())
```

**Format Selection**: PNG (lossless, larger) vs JPEG (lossy, smaller). **DPI**: 72 (screen), 150 (web), 300 (print/default), 600 (professional).

## Testing

The library includes comprehensive unit tests. Tests requiring ImageMagick will be skipped if the binary is not available.

### Run All Tests

```bash
go test ./tests/... -v
```

### Run Tests for Specific Package

```bash
# Test document package
go test ./tests/document/... -v

# Test encoding package
go test ./tests/encoding/... -v
```

## Error Handling

All operations return descriptive errors with context:

```go
// Common error scenarios
doc, err := document.OpenPDF("file.pdf")     // File not found, invalid/corrupted PDF
page, err := doc.ExtractPage(999)            // Page out of range
imageData, err := page.ToImage(renderer, c)  // ImageMagick not installed, config invalid
dataURI, err := encoding.EncodeImageDataURI(data, format)  // Empty data, unsupported format
```

Error messages include operation context and external command output for debugging.

## Deployment

**Container Deployment** - Ensure ImageMagick is available:

```dockerfile
FROM golang:1.25-alpine
RUN apk add --no-cache imagemagick
COPY . /app
WORKDIR /app
RUN go build -o service .
CMD ["./service"]
```

**Startup Verification**:
```go
if _, err := exec.LookPath("magick"); err != nil {
    log.Fatal("ImageMagick not installed")
}
```

## Limitations

### Current Limitations

- **PDF Only**: Only PDF format currently supported
- **Image Output Only**: Cannot extract raw text (planned for future)
- **Sequential Processing**: Pages processed one at a time (parallel processing planned)
- **No OCR**: Cannot extract text from image-based PDFs (OCR support planned)
- **ImageMagick Required**: External binary dependency for PDF rendering

## Roadmap

Planned features include additional document formats (Office, HTML, Markdown), alternative outputs (text extraction, structured content), and processing enhancements (parallel processing, streaming). See [PROJECT.md](./PROJECT.md) for the complete roadmap and current development status.

## License

This project is licensed under the MIT License.

## Related Projects

- **[go-agents](https://github.com/JaimeStill/go-agents)**: Go library for building LLM-powered applications
