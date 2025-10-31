# Project Scope and Roadmap

## Overview

**document-context** is a Go library for converting documents into context-friendly formats suitable for LLM consumption and analysis. The library provides format-agnostic interfaces for document processing with extensible format support.

This project was created as a tooling extension for the [go-agents](https://github.com/JaimeStill/go-agents) project, but is not directly dependent on go-agents and can be used as a standalone library for document processing needs.

## Current Status

**Phase**: Pre-Release Development (not yet versioned)

The API is under active development and subject to change as additional features and format support are added. The library is functional for its current capabilities but should be considered experimental until the first versioned release (v0.1.0).

### What This Library Provides

**Document Processing Abstractions**:
- Format-agnostic `Document` and `Page` interfaces
- Extensible format support through interface implementation
- Clean separation between document access and format conversion

**PDF Support**:
- PDF document loading and page extraction using pdfcpu
- Individual page rendering to images (PNG, JPEG)
- Configurable image quality and DPI settings
- Integration with ImageMagick for high-quality rendering

**Image Encoding**:
- Base64 data URI encoding for direct LLM consumption
- MIME type handling for different image formats
- Memory-efficient encoding of image data

**Current Implementation**:
```
pkg/
├── document/
│   ├── document.go    # Core interfaces (Document, Page)
│   └── pdf.go         # PDF implementation
└── encoding/
    └── image.go       # Data URI encoding
```

### What This Library Does NOT Provide

The library is intentionally scoped as a document processing utility. The following capabilities are outside the project scope:

- **LLM Integration**: Interaction with language models (use go-agents for this)
- **Document Classification**: Semantic analysis or categorization of documents
- **Document Generation**: Creating or modifying documents (read-only operations only)
- **Format Conversion Chains**: Multi-step conversion pipelines (single-step conversions only)

## Design Philosophy

### Core Principles

1. **Format Agnostic**: Provide common interfaces that work across document formats
2. **External Tool Leverage**: Use proven external tools (ImageMagick, Tesseract) rather than reimplementing
3. **Flexible Output**: Support multiple output formats for different use cases
4. **Clean Abstractions**: Separate document access from format conversion
5. **Minimal Dependencies**: Keep pure Go dependencies minimal, document external binary requirements

### External Binary Strategy

A critical design decision is leveraging mature external tools rather than reimplementing complex functionality:

- **ImageMagick**: PDF page rendering with professional-quality output
- **Future OCR**: Tesseract for text extraction from images
- **Future Office**: OpenXML for Office document processing

**Rationale**: These tools are battle-tested, cross-platform, feature-rich, and maintained by dedicated communities. Reimplementing would be error-prone and time-consuming.

**Trade-offs**:
- ✅ High-quality results with minimal code
- ✅ Leverage decades of development effort
- ✅ Cross-platform compatibility
- ❌ Deployment requires binary availability
- ❌ External process execution overhead
- ❌ Version compatibility considerations

**Deployment**: External binaries must be available in deployment environments. For containerized deployments, include binary installation in Dockerfile.

## Current Implementation

### Supported Formats

**PDF** (Complete):
- Format detection and validation
- Page count and metadata extraction
- Individual page rendering to images
- Configurable output format (PNG, JPEG)
- Quality and DPI control

### Image Output Formats

**PNG** (Lossless):
- High-quality rendering
- Transparency support
- Larger file sizes
- Ideal for text-heavy documents

**JPEG** (Lossy):
- Configurable quality (1-100)
- Smaller file sizes
- Suitable for photo-heavy documents
- Quality/size trade-off control

### Data URI Encoding

Converts image bytes to base64 data URIs for direct embedding:
- Format: `data:image/png;base64,<encoded-data>`
- Suitable for LLM vision APIs
- No external storage required
- Immediate availability for processing

## Future Development

### Additional Document Formats

The library will expand to support common document formats using appropriate processing strategies:

**Office Documents** (Planned):
- **Format**: .docx, .xlsx, .pptx (OpenXML)
- **Approach**: Direct XML parsing of unzipped archives
- **Output**: Structured text extraction, optional image rendering
- **Rationale**: Pure Go implementation possible via OpenXML standard

**Images** (Planned):
- **Format**: .png, .jpg, .tiff
- **Approach**: OCR via Tesseract integration
- **Output**: Extracted text with confidence scores
- **Rationale**: Native image format support for text extraction

### Output Format Flexibility

**Current**: Documents → Images (PNG/JPEG) → Data URIs

**Planned**:
- Documents → Raw Text (plain text extraction)
- Documents → Structured Text (preserve formatting, hierarchy)

**Use Cases**:
- **Raw Text**: Token efficiency, searchability, pure text analysis
- **Images**: Visual layout preservation, handwriting, complex formatting
- **Structured**: Hierarchical processing, section-aware analysis

### Processing Pipeline Enhancements

**Chunking Strategies**:
- Split large documents into processable chunks
- Intelligent boundary detection (page breaks, sections)
- Overlapping contexts for continuity
- Token budget management

**Optimization**:
- Parallel page processing for multi-page documents
- Streaming support for large files
- Memory-efficient processing
- Caching for repeated operations

**Error Handling**:
- Graceful degradation for partially corrupt documents
- Page-level error isolation
- Detailed error reporting with recovery suggestions

## Extension Points

The library is designed for extensibility through well-defined interfaces:

### Adding New Document Formats

Implement the `Document` interface:
```go
type Document interface {
    PageCount() int
    ExtractPage(pageNum int) (Page, error)
    ExtractAllPages() ([]Page, error)
    Close() error
}
```

And the `Page` interface:
```go
type Page interface {
    Number() int
    ToImage(opts ImageOptions) ([]byte, error)
}
```

### Adding New Output Formats

Extend `ImageOptions` or create new option structures:
```go
type ImageOptions struct {
    Format  ImageFormat  // PNG, JPEG, future: WEBP, TIFF
    Quality int          // Format-specific quality
    DPI     int          // Resolution control
}
```

Add new encoding functions in `pkg/encoding/`:
```go
func EncodeText(data []byte) (string, error)
func EncodeStructured(data []byte, format StructureFormat) (string, error)
```

## Testing Strategy

### Current Coverage

The library includes comprehensive unit tests:
- PDF document operations
- Image format validation
- Data URI encoding
- Error handling scenarios

### External Dependency Testing

Tests requiring ImageMagick use conditional execution:
```go
func requireImageMagick(t *testing.T) {
    t.Helper()
    if !hasImageMagick() {
        t.Skip("ImageMagick not installed, skipping test")
    }
}
```

This approach:
- Allows local development without all binaries
- Enables CI/CD flexibility
- Provides clear feedback about missing dependencies
- Separates unit tests from integration tests

### Testing Priorities

**High Priority** (Current):
- Interface compliance for all format implementations
- Error handling for missing binaries
- Image encoding correctness
- Format validation

**Future Priorities**:
- Multi-format processing chains
- Performance benchmarks for large documents
- Memory usage profiling
- Concurrent processing safety

## Publishing and Versioning

### Pre-Release Development

The library is currently in pre-release development and has not yet reached v0.1.0. The API is subject to change as features are added and refined.

**Development Focus**:
- Stabilize core abstractions (Document, Page interfaces)
- Add support for Office document formats
- Implement text extraction alternatives
- Validate API design through usage patterns

### Version 0.1.0 Goals

The first versioned release will include:
- ✅ PDF support (complete)
- ✅ Image encoding (complete)
- ⬜ Office document support (OpenXML)
- ⬜ Text extraction capabilities
- ⬜ Comprehensive documentation
- ⬜ 80%+ test coverage
- ⬜ Real-world usage validation

### Semantic Versioning

Once v0.1.0 is released, the library will follow [Semantic Versioning 2.0.0](https://semver.org/):

**Pre-Release (v0.x.x)**:
- API may change between minor versions
- Breaking changes documented in CHANGELOG
- Community feedback actively incorporated

**Stable Release (v1.0.0+)**:
- API stability commitment
- Breaking changes only in major versions
- Backward compatibility within major version

## Integration Patterns

### Standalone Usage

Process documents without LLM integration:
```go
// Load document
doc, err := document.OpenPDF("report.pdf")
defer doc.Close()

// Convert page to image
page, _ := doc.ExtractPage(1)
imageData, _ := page.ToImage(document.ImageOptions{
    Format: document.PNG,
    DPI:    300,
})

// Encode for storage or transmission
dataURI, _ := encoding.EncodeImageDataURI(imageData, document.PNG)
```

### LLM Integration (with go-agents)

Process documents for LLM analysis:
```go
// Extract and encode document pages
doc, _ := document.OpenPDF("contract.pdf")
pages, _ := doc.ExtractAllPages()

var images []string
for _, page := range pages {
    imageData, _ := page.ToImage(document.DefaultImageOptions())
    dataURI, _ := encoding.EncodeImageDataURI(imageData, document.PNG)
    images = append(images, dataURI)
}

// Send to LLM via go-agents
response, err := agent.Vision(ctx, 
    "Analyze this contract for key terms", 
    images,
)
```

### Batch Processing

Process multiple documents efficiently:
```go
// Process directory of PDFs
files, _ := os.ReadDir("documents/")
for _, file := range files {
    if filepath.Ext(file.Name()) == ".pdf" {
        // Process each document
        // Convert to desired format
        // Store or transmit results
    }
}
```

## Deployment Considerations

### Container Requirements

When deploying services that use document-context:

**Dockerfile Example**:
```dockerfile
FROM golang:1.25-alpine

# Install ImageMagick
RUN apk add --no-cache imagemagick

# Future: Add other binaries as needed
# RUN apk add --no-cache tesseract-ocr libreoffice

COPY . .
RUN go build -o /app

CMD ["/app"]
```

### Binary Verification

Check for required binaries at startup:
```go
func checkDependencies() error {
    required := []string{"magick"}
    for _, binary := range required {
        if _, err := exec.LookPath(binary); err != nil {
            return fmt.Errorf("required binary not found: %s", binary)
        }
    }
    return nil
}
```

### Environment Configuration

Configure binary paths for non-standard installations:
```go
// Allow custom binary paths via environment
magickPath := os.Getenv("IMAGEMAGICK_PATH")
if magickPath == "" {
    magickPath = "magick"
}
```
