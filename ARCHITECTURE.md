# Architecture

This document describes the current architecture and implementation of the document-context library.

## Package Structure

```
pkg/
├── document/           # Core document processing abstractions
│   ├── document.go     # Document and Page interfaces, ImageFormat types
│   └── pdf.go          # PDF implementation using pdfcpu and ImageMagick
└── encoding/           # Output encoding utilities
    └── image.go        # Base64 data URI encoding
```

## Core Abstractions

### Document and Page Interfaces

The library provides format-agnostic interfaces for document processing:

```go
type Document interface {
    PageCount() int
    ExtractPage(pageNum int) (Page, error)
    ExtractAllPages() ([]Page, error)
    Close() error
}

type Page interface {
    Number() int
    ToImage(opts ImageOptions) ([]byte, error)
}
```

**Design Rationale**:
- **Format Independence**: The same code can process PDFs, Office documents, or any future format without changes
- **Lazy Evaluation**: Pages are extracted on-demand rather than loading entire documents into memory
- **Resource Management**: Explicit `Close()` method for cleanup of document resources
- **Composability**: Page operations are independent, enabling parallel processing

**Interface Benefits**:
- New format support requires only interface implementation
- Testing through mocks without file I/O
- Clear contracts between document access and conversion
- Enables batch processing and pipeline composition

### Image Format System

The library defines supported output formats with associated metadata:

```go
type ImageFormat string

const (
    PNG  ImageFormat = "png"
    JPEG ImageFormat = "jpg"
)

func (f ImageFormat) MimeType() (string, error) {
    switch f {
    case PNG:
        return "image/png", nil
    case JPEG:
        return "image/jpeg", nil
    default:
        return "", fmt.Errorf("unsupported image format: %s", f)
    }
}
```

**Format Properties**:
- **PNG**: Lossless compression, transparency support, larger files, ideal for text
- **JPEG**: Lossy compression, configurable quality (1-100), smaller files, suitable for photos

**MimeType Method**: Provides format-specific MIME types for data URI encoding and HTTP content-type headers.

### Image Options

Configuration structure for image conversion:

```go
type ImageOptions struct {
    Format  ImageFormat  // Output format (PNG, JPEG)
    Quality int          // JPEG quality (1-100), ignored for PNG
    DPI     int          // Dots per inch (resolution)
}

func DefaultImageOptions() ImageOptions {
    return ImageOptions{
        Format:  PNG,
        Quality: 0,      // Not used for PNG
        DPI:     300,    // Standard high-resolution
    }
}
```

**Design Decisions**:
- **DPI Default (300)**: Standard for high-quality document rendering, balances quality and file size
- **PNG Default**: Lossless format ensures text clarity without compression artifacts
- **Quality Field**: Format-specific parameter (JPEG only), validated at conversion time
- **Zero Values**: Missing DPI triggers defaults to prevent user error

## PDF Implementation

### Document Loading

```go
func OpenPDF(path string) (*PDFDocument, error) {
    ctx, err := api.ReadContextFile(path)
    if err != nil {
        return nil, fmt.Errorf("failed to open PDF: %w", err)
    }

    pageCount := ctx.PageCount
    if pageCount == 0 {
        return nil, fmt.Errorf("PDF has no pages")
    }

    return &PDFDocument{
        path:      path,
        ctx:       ctx,
        pageCount: pageCount,
    }, nil
}
```

**pdfcpu Library**: Pure Go PDF processing library for metadata extraction and validation.

**Validation**: Empty PDFs rejected at load time to fail fast.

**Resource Retention**: File path and context stored for page extraction operations.

### Page Extraction

```go
func (d *PDFDocument) ExtractPage(pageNum int) (Page, error) {
    if pageNum < 1 || pageNum > d.pageCount {
        return nil, fmt.Errorf("page %d out of range [1-%d]", pageNum, d.pageCount)
    }

    return &PDFPage{
        doc:    d,
        number: pageNum,
    }, nil
}
```

**Design Pattern**: Lightweight page objects hold references to parent document, enabling lazy image conversion only when needed.

**Bounds Checking**: Page numbers validated at extraction time (1-indexed per PDF convention).

### Image Conversion

The critical implementation detail leveraging external binaries:

```go
func (p *PDFPage) ToImage(opts ImageOptions) ([]byte, error) {
    // Apply defaults for zero-value DPI
    if opts.DPI == 0 {
        opts = DefaultImageOptions()
    }

    // Validate JPEG quality
    if opts.Format == JPEG {
        if opts.Quality == 0 {
            opts.Quality = 85  // Standard quality default
        }
        if opts.Quality < 1 || opts.Quality > 100 {
            return nil, fmt.Errorf("JPEG quality must be 1-100, got %d", opts.Quality)
        }
    }

    // Determine file extension
    var ext string
    switch opts.Format {
    case PNG:
        ext = "png"
    case JPEG:
        ext = "jpg"
    default:
        return nil, fmt.Errorf("unsupported image format: %s", opts.Format)
    }

    // Create temporary file for output
    tmpFile, err := os.CreateTemp("", fmt.Sprintf("page-%d-*.%s", p.number, ext))
    if err != nil {
        return nil, fmt.Errorf("failed to create temp file: %w", err)
    }
    tmpPath := tmpFile.Name()
    tmpFile.Close()
    defer os.Remove(tmpPath)

    // Build ImageMagick command
    pageIndex := p.number - 1  // ImageMagick uses 0-based indexing
    inputSpec := fmt.Sprintf("%s[%d]", p.doc.path, pageIndex)

    args := []string{
        "-density", fmt.Sprintf("%d", opts.DPI),
        inputSpec,
        "-background", "white",
        "-flatten",
    }

    if opts.Format == JPEG {
        args = append(args, "-quality", fmt.Sprintf("%d", opts.Quality))
    }

    args = append(args, tmpPath)

    // Execute ImageMagick
    cmd := exec.Command("magick", args...)
    output, err := cmd.CombinedOutput()
    if err != nil {
        return nil, fmt.Errorf(
            "imagemagick failed for page %d: %w\nOutput: %s",
            p.number, err, string(output),
        )
    }

    // Read generated image
    imgData, err := os.ReadFile(tmpPath)
    if err != nil {
        return nil, fmt.Errorf("failed to read generated image: %w", err)
    }

    return imgData, nil
}
```

**Critical Design Decisions**:

1. **External Binary Strategy**: 
   - ImageMagick provides professional-quality PDF rendering
   - Avoids reimplementing complex PDF-to-image conversion
   - Leverages decades of development and testing
   - Trade-off: Requires binary availability in deployment

2. **Temporary File Management**:
   - `os.CreateTemp()` with unique naming prevents conflicts
   - File handle closed before ImageMagick writes to it
   - `defer os.Remove()` ensures cleanup even on errors
   - Pattern-based naming aids debugging

3. **Command Construction**:
   - String slice arguments prevent injection vulnerabilities
   - Page indexing: PDF (1-based) → ImageMagick (0-based)
   - Input specification: `path[pageIndex]` targets specific page
   - Background flattening: Ensures consistent output for transparent PDFs
   - Format-specific options: Quality only for JPEG

4. **Error Reporting**:
   - Includes ImageMagick output for troubleshooting
   - Clear indication of which page failed
   - Wrapped errors with context

## Image Encoding

### Data URI Generation

```go
func EncodeImageDataURI(data []byte, format document.ImageFormat) (string, error) {
    if len(data) == 0 {
        return "", fmt.Errorf("image data is empty")
    }

    mimeType, err := format.MimeType()
    if err != nil {
        return "", err
    }

    var builder strings.Builder
    builder.WriteString("data:")
    builder.WriteString(mimeType)
    builder.WriteString(";base64,")
    builder.WriteString(base64.StdEncoding.EncodeToString(data))

    return builder.String(), nil
}
```

**Data URI Format**: `data:<mimetype>;base64,<encoded-data>`

**Design Decisions**:
- **Empty Data Validation**: Fail fast for invalid input
- **String Builder**: Memory-efficient construction for large images
- **Standard Base64**: RFC 4648 standard encoding without padding variations
- **Format Coupling**: MIME type retrieved from ImageFormat for consistency

**Use Cases**:
- Direct embedding in HTML/JSON responses
- LLM vision API inputs (OpenAI, Anthropic, etc.)
- Elimination of external storage requirements
- Immediate availability for processing

## Testing Strategy

### Test Organization

Tests are separated from implementation in a parallel `tests/` directory:

```
tests/
├── document/
│   └── pdf_test.go
└── encoding/
    └── image_test.go
```

**Black-Box Testing**: All tests use `package <name>_test` to test only the public API.

### External Dependency Handling

Tests requiring ImageMagick use conditional execution:

```go
func hasImageMagick() bool {
    _, err := exec.LookPath("magick")
    return err == nil
}

func requireImageMagick(t *testing.T) {
    t.Helper()
    if !hasImageMagick() {
        t.Skip("ImageMagick not installed, skipping image conversion test")
    }
}

func TestPDFPage_ToImage_PNG(t *testing.T) {
    requireImageMagick(t)
    // Test implementation
}
```

**Benefits**:
- Local development without all binaries installed
- Clear indication of skipped tests
- CI/CD flexibility (can run subset without binary installation)
- Helper function marked with `t.Helper()` for clean stack traces

### Test Coverage

**Current Focus**:
- PDF document loading and validation
- Page extraction and bounds checking
- Image format validation
- Data URI encoding correctness
- Error handling for missing binaries
- Format-specific behavior (PNG vs JPEG)

**Test Patterns**:
- Table-driven tests for multiple scenarios
- Magic byte validation for image format verification
- Error case testing (missing files, invalid pages, out of range)
- Default value behavior testing

## Extension Points

### Adding New Document Formats

Implement the `Document` and `Page` interfaces:

```go
type WordDocument struct {
    path      string
    pageCount int
    // Format-specific fields
}

func OpenWord(path string) (*WordDocument, error) {
    // Format-specific loading
}

func (d *WordDocument) PageCount() int { ... }
func (d *WordDocument) ExtractPage(pageNum int) (Page, error) { ... }
func (d *WordDocument) ExtractAllPages() ([]Page, error) { ... }
func (d *WordDocument) Close() error { ... }

type WordPage struct {
    doc    *WordDocument
    number int
}

func (p *WordPage) Number() int { ... }
func (p *WordPage) ToImage(opts ImageOptions) ([]byte, error) { ... }
```

**Integration**: New format automatically works with existing encoding and pipeline code.

### Adding New Output Formats

Extend the `ImageFormat` type:

```go
const (
    PNG  ImageFormat = "png"
    JPEG ImageFormat = "jpg"
    WEBP ImageFormat = "webp"  // New format
)

func (f ImageFormat) MimeType() (string, error) {
    switch f {
    case PNG:
        return "image/png", nil
    case JPEG:
        return "image/jpeg", nil
    case WEBP:
        return "image/webp", nil
    default:
        return "", fmt.Errorf("unsupported image format: %s", f)
    }
}
```

Update image conversion logic to handle new format parameters.

### Adding Text Extraction

Create new output type and method:

```go
type TextOptions struct {
    PreserveFormatting bool
    IncludeMetadata    bool
}

type Page interface {
    Number() int
    ToImage(opts ImageOptions) ([]byte, error)
    ToText(opts TextOptions) (string, error)  // New method
}
```

Implement for each format with appropriate extraction strategy.

## Dependencies

### Pure Go Dependencies

**pdfcpu** (`github.com/pdfcpu/pdfcpu`):
- Purpose: PDF metadata extraction and validation
- Scope: Document loading, page counting
- Not used for: Image conversion (ImageMagick handles this)

### External Binary Dependencies

**ImageMagick** (Required):
- Binary: `magick` command
- Purpose: High-quality PDF page rendering to images
- Minimum Version: 7.0+ (using modern `magick` command, not deprecated `convert`)
- Installation: Platform-specific package managers

**Rationale for External Binary**: 
PDF rendering is complex (fonts, vector graphics, color spaces, transparency, compression). ImageMagick represents decades of development by experts in document rendering. Reimplementing would be error-prone, time-consuming, and unlikely to achieve comparable quality.

**Deployment Requirement**: 
Services using this library must ensure ImageMagick is available in their deployment environment. For containerized deployments, this means including ImageMagick in the Dockerfile.

## Design Patterns

### Interface-Based Abstraction

Document and Page interfaces decouple format-specific implementations from processing logic. Benefits:
- New formats added without modifying existing code
- Testing through mocks without file dependencies
- Clear contracts between components
- Enables batch processing and parallel execution

### Lazy Evaluation

Pages are not converted to images until `ToImage()` is called. Benefits:
- Memory efficiency for large documents
- Skip processing for unwanted pages
- Enable selective page extraction
- Parallel processing of independent pages

### Temporary File Pattern

Image conversion uses temporary files for intermediate storage:
```go
tmpFile, _ := os.CreateTemp("", "pattern-*.ext")
tmpPath := tmpFile.Name()
tmpFile.Close()
defer os.Remove(tmpPath)

// Use tmpPath with external command
cmd := exec.Command("tool", tmpPath)
cmd.Run()

// Read result
data, _ := os.ReadFile(tmpPath)
```

Benefits:
- Explicit cleanup with defer
- Unique naming prevents conflicts
- Works with external tools expecting file paths
- Handles large data that shouldn't be piped

### Error Context Wrapping

Errors wrapped with context using `%w` verb for error chain inspection:
```go
if err != nil {
    return nil, fmt.Errorf("operation failed: %w", err)
}
```

Enables error type checking and root cause identification.

## Future Architectural Considerations

As new features are added, these architectural elements will need consideration:

### Streaming Support

For very large documents, streaming pages on-demand:
```go
type StreamingDocument interface {
    Pages() <-chan Page
    Errors() <-chan error
}
```

### Parallel Processing

Concurrent page conversion with worker pools:
```go
type BatchOptions struct {
    MaxConcurrency int
    ProgressCallback func(pageNum int, total int)
}

func (d *Document) ConvertAll(opts BatchOptions) ([][]byte, error)
```

### Caching Layer

Cache converted pages to avoid redundant processing:
```go
type CachedDocument struct {
    doc   Document
    cache map[int][]byte
}
```

### Format Detection

Auto-detect format from file content:
```go
func Open(path string) (Document, error) {
    format := detectFormat(path)
    switch format {
    case FormatPDF:
        return OpenPDF(path)
    case FormatWord:
        return OpenWord(path)
    }
}
```

These enhancements will be documented as they are implemented.
