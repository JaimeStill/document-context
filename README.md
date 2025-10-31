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
- Base64 data URI encoding for LLM APIs

## Prerequisites

### Go Version

- Go 1.25.2 or later

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

Convert a PDF page to PNG image:

```go
package main

import (
    "fmt"
    "os"

    "github.com/JaimeStill/document-context/pkg/document"
)

func main() {
    // Open PDF document
    doc, err := document.OpenPDF("report.pdf")
    if err != nil {
        fmt.Fprintf(os.Stderr, "Failed to open PDF: %v\n", err)
        return
    }
    defer doc.Close()

    // Get first page
    page, err := doc.ExtractPage(1)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Failed to extract page: %v\n", err)
        return
    }

    // Convert to PNG image (300 DPI)
    imageData, err := page.ToImage(document.ImageOptions{
        Format: document.PNG,
        DPI:    300,
    })
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

### JPEG Output with Custom Quality

Convert PDF page to JPEG with specific quality setting:

```go
package main

import (
    "fmt"
    "os"

    "github.com/JaimeStill/document-context/pkg/document"
)

func main() {
    doc, err := document.OpenPDF("photo-document.pdf")
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        return
    }
    defer doc.Close()

    page, err := doc.ExtractPage(1)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        return
    }

    // Convert to JPEG with 85% quality at 150 DPI
    imageData, err := page.ToImage(document.ImageOptions{
        Format:  document.JPEG,
        Quality: 85,
        DPI:     150,
    })
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        return
    }

    err = os.WriteFile("page-1.jpg", imageData, 0644)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        return
    }

    fmt.Println("Successfully converted page to JPEG")
}
```

### Using Default Options

Simplify conversion with sensible defaults:

```go
package main

import (
    "fmt"
    "os"

    "github.com/JaimeStill/document-context/pkg/document"
)

func main() {
    doc, err := document.OpenPDF("contract.pdf")
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        return
    }
    defer doc.Close()

    page, err := doc.ExtractPage(1)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        return
    }

    // Use defaults: PNG format, 300 DPI
    imageData, err := page.ToImage(document.DefaultImageOptions())
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        return
    }

    err = os.WriteFile("page-1.png", imageData, 0644)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        return
    }

    fmt.Println("Successfully converted page with defaults")
}
```

### Data URI Encoding for LLM APIs

Convert document pages to data URIs for direct LLM consumption:

```go
package main

import (
    "fmt"

    "github.com/JaimeStill/document-context/pkg/document"
    "github.com/JaimeStill/document-context/pkg/encoding"
)

func main() {
    doc, err := document.OpenPDF("analysis.pdf")
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        return
    }
    defer doc.Close()

    page, err := doc.ExtractPage(1)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        return
    }

    // Convert page to image
    imageData, err := page.ToImage(document.DefaultImageOptions())
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        return
    }

    // Encode as data URI
    dataURI, err := encoding.EncodeImageDataURI(imageData, document.PNG)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        return
    }

    // Data URI is now ready for LLM API
    fmt.Printf("Data URI length: %d bytes\n", len(dataURI))
    fmt.Printf("Data URI prefix: %s...\n", dataURI[:50])
    
    // Use dataURI with LLM vision API
    // response := llm.Vision("Analyze this document", []string{dataURI})
}
```

### Processing All Pages

Convert entire PDF to images:

```go
package main

import (
    "fmt"
    "os"
    "path/filepath"

    "github.com/JaimeStill/document-context/pkg/document"
)

func main() {
    doc, err := document.OpenPDF("multi-page.pdf")
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        return
    }
    defer doc.Close()

    fmt.Printf("Processing %d pages...\n", doc.PageCount())

    // Extract all pages
    pages, err := doc.ExtractAllPages()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        return
    }

    // Convert each page
    for _, page := range pages {
        imageData, err := page.ToImage(document.DefaultImageOptions())
        if err != nil {
            fmt.Fprintf(os.Stderr, "Failed to convert page %d: %v\n", 
                page.Number(), err)
            continue
        }

        filename := filepath.Join("output", 
            fmt.Sprintf("page-%d.png", page.Number()))
        
        err = os.WriteFile(filename, imageData, 0644)
        if err != nil {
            fmt.Fprintf(os.Stderr, "Failed to write page %d: %v\n", 
                page.Number(), err)
            continue
        }

        fmt.Printf("Converted page %d\n", page.Number())
    }

    fmt.Println("Processing complete")
}
```

### Integration with go-agents

Use document-context with go-agents for LLM analysis:

```go
package main

import (
    "context"
    "fmt"

    "github.com/JaimeStill/document-context/pkg/document"
    "github.com/JaimeStill/document-context/pkg/encoding"
    "github.com/JaimeStill/go-agents/pkg/agent"
    "github.com/JaimeStill/go-agents/pkg/config"
)

func main() {
    // Load agent configuration
    cfg, err := config.LoadAgentConfig("config.json")
    if err != nil {
        fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
        return
    }

    // Create agent
    agent, err := agent.New(cfg)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Failed to create agent: %v\n", err)
        return
    }

    // Open and process document
    doc, err := document.OpenPDF("contract.pdf")
    if err != nil {
        fmt.Fprintf(os.Stderr, "Failed to open PDF: %v\n", err)
        return
    }
    defer doc.Close()

    // Convert pages to data URIs
    pages, err := doc.ExtractAllPages()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Failed to extract pages: %v\n", err)
        return
    }

    var images []string
    for _, page := range pages {
        imageData, err := page.ToImage(document.DefaultImageOptions())
        if err != nil {
            fmt.Fprintf(os.Stderr, "Failed to convert page %d: %v\n", 
                page.Number(), err)
            continue
        }

        dataURI, err := encoding.EncodeImageDataURI(imageData, document.PNG)
        if err != nil {
            fmt.Fprintf(os.Stderr, "Failed to encode page %d: %v\n", 
                page.Number(), err)
            continue
        }

        images = append(images, dataURI)
    }

    // Send to LLM for analysis
    ctx := context.Background()
    response, err := agent.Vision(ctx, 
        "Analyze this contract and summarize the key terms", 
        images,
    )
    if err != nil {
        fmt.Fprintf(os.Stderr, "LLM analysis failed: %v\n", err)
        return
    }

    // Process response
    fmt.Println("Analysis:", response.Choices[0].Message.Content)
}
```

## Configuration

### Image Options

```go
type ImageOptions struct {
    Format  ImageFormat  // PNG or JPEG
    Quality int          // JPEG quality (1-100), ignored for PNG
    DPI     int          // Resolution in dots per inch
}
```

**Format Selection**:
- **PNG**: Best for text-heavy documents, no quality loss, larger files
- **JPEG**: Best for photo-heavy documents, smaller files, configurable quality

**Quality Settings** (JPEG only):
- 1-50: Low quality, small files, visible artifacts
- 50-75: Medium quality, balanced size/quality
- 75-90: High quality, larger files, minimal artifacts
- 90-100: Very high quality, large files, nearly lossless

**DPI Recommendations**:
- 72: Screen viewing only
- 150: Web images, low resolution
- 300: Standard print quality, high resolution (default)
- 600: Professional print, very high resolution

### Default Options

```go
opts := document.DefaultImageOptions()
// Returns: PNG format, 300 DPI
```

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

### Common Errors

**PDF Opening Errors**:
```go
doc, err := document.OpenPDF("missing.pdf")
if err != nil {
    // File not found, invalid PDF, corrupted file
}
```

**Page Range Errors**:
```go
page, err := doc.ExtractPage(999)
if err != nil {
    // Page number out of range [1-N]
}
```

**Image Conversion Errors**:
```go
imageData, err := page.ToImage(opts)
if err != nil {
    // ImageMagick not installed
    // Invalid format or quality settings
    // Insufficient disk space for temp files
}
```

**Encoding Errors**:
```go
dataURI, err := encoding.EncodeImageDataURI(imageData, format)
if err != nil {
    // Empty image data
    // Unsupported format
}
```

### Error Messages

The library provides detailed error messages with context:

```
imagemagick failed for page 1: exit status 1
Output: convert-im6.q16: unable to read font...
```

These messages include:
- Which operation failed
- Which page (if applicable)
- The external command output (for debugging)

## Deployment

### Container Deployment

When deploying services that use document-context, ensure ImageMagick is available:

**Dockerfile Example**:
```dockerfile
FROM golang:1.25-alpine

# Install ImageMagick
RUN apk add --no-cache imagemagick

# Copy application
COPY . /app
WORKDIR /app

# Build
RUN go build -o service .

CMD ["./service"]
```

**Ubuntu-based Container**:
```dockerfile
FROM golang:1.25

# Install ImageMagick
RUN apt-get update && \
    apt-get install -y imagemagick && \
    rm -rf /var/lib/apt/lists/*

# Copy and build application
COPY . /app
WORKDIR /app
RUN go build -o service .

CMD ["./service"]
```

### Binary Verification at Startup

Check for required binaries when your application starts:

```go
func checkDependencies() error {
    if _, err := exec.LookPath("magick"); err != nil {
        return fmt.Errorf("ImageMagick not installed: %w", err)
    }
    return nil
}

func main() {
    if err := checkDependencies(); err != nil {
        log.Fatal(err)
    }
    
    // Application startup continues...
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

### Planned Features

**Additional Document Formats**:
- Office documents (.docx, .xlsx, .pptx) via OpenXML
- Image formats (.png, .jpg) via OCR
- HTML and Markdown documents

**Alternative Outputs**:
- Raw text extraction
- Structured text with formatting
- Hybrid text + images
- Metadata extraction

**Processing Enhancements**:
- Parallel page processing
- Streaming for large documents
- Intelligent chunking strategies
- Caching layer for repeated operations

See [PROJECT.md](./PROJECT.md) for detailed roadmap.

## License

This project is licensed under the MIT License.

## Related Projects

- **[go-agents](https://github.com/JaimeStill/go-agents)**: Go library for building LLM-powered applications
