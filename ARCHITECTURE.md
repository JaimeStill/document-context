# Architecture

This document describes the current architecture and implementation of the document-context library.

## Package Structure

```
pkg/
├── config/             # Configuration data structures
│   ├── doc.go          # Package documentation
│   ├── image.go        # ImageConfig with filter fields
│   └── cache.go        # CacheConfig structure
├── image/              # Image rendering domain objects
│   ├── image.go        # Renderer interface
│   └── imagemagick.go  # ImageMagick implementation
├── document/           # Core document processing abstractions
│   ├── document.go     # Document and Page interfaces, ImageFormat types
│   └── pdf.go          # PDF implementation using pdfcpu
└── encoding/           # Output encoding utilities
    └── image.go        # Base64 data URI encoding
```

**Package Dependencies** (higher layers depend on lower layers):
```
pkg/document → pkg/image → pkg/config
pkg/encoding (independent utility)
```

## Core Abstractions

### Configuration Package (pkg/config/)

The configuration package provides ephemeral data structures for initializing domain objects. Configuration structures support JSON serialization, default values, and merge semantics for layered configuration.

**Key Principle**: Configuration is data, not behavior. Validation happens in domain packages during transformation.

#### ImageConfig

Configuration for image rendering operations:

```go
type ImageConfig struct {
    Format     string `json:"format,omitempty"`      // "png" or "jpg"
    Quality    int    `json:"quality,omitempty"`     // JPEG quality: 1-100
    DPI        int    `json:"dpi,omitempty"`         // Render density
    Brightness *int   `json:"brightness,omitempty"`  // -100 to +100 (Session 5)
    Contrast   *int   `json:"contrast,omitempty"`    // -100 to +100 (Session 5)
    Saturation *int   `json:"saturation,omitempty"`  // -100 to +100 (Session 5)
    Rotation   *int   `json:"rotation,omitempty"`    // 0 to 360 degrees (Session 5)
}
```

**Filter Fields**: Use pointer types (*int) to distinguish "not set" (nil) from "explicitly set to zero" (pointer to 0).

**Configuration Methods**:
- `DefaultImageConfig()`: Returns PNG, 300 DPI, no filters
- `Merge(source *ImageConfig)`: Overlays non-zero values from source onto receiver
- `Finalize()`: Applies default values for any unset fields by merging onto defaults

**Merge Semantics**:
- String fields: only merge if non-empty
- Integer fields: only merge if greater than zero
- Pointer fields: only merge if non-nil (allows explicit zero via pointer to 0)

#### CacheConfig

Configuration for cache implementations:

```go
type CacheConfig struct {
    Name    string         `json:"name"`             // Implementation identifier
    Options map[string]any `json:"options,omitempty"` // Implementation-specific settings
}
```

**Design**: Name-based approach where Name identifies cache type and Options provides implementation-specific parameters.

**Configuration Methods**:
- `DefaultCacheConfig()`: Returns empty name and initialized options map
- `Merge(source *CacheConfig)`: Merges name and options using `maps.Copy()`

### Image Rendering Package (pkg/image/)

The image package defines the interface for rendering documents to images and provides implementations. This package transforms configuration (data) into renderers (behavior) with validation at the transformation boundary.

#### Renderer Interface

```go
type Renderer interface {
    Render(inputPath string, pageNum int, outputPath string) error
    FileExtension() string
}
```

**Design**: Interface-based public API hides implementation details. Consumers depend on the contract, not concrete types.

**Methods**:
- `Render(inputPath, pageNum, outputPath)`: Converts specified page to image file
- `FileExtension()`: Returns appropriate file extension for output format (no leading dot)

**Thread Safety**: Renderer instances are immutable once created and safe for concurrent use.

#### ImageMagickRenderer

Implementation using ImageMagick for PDF rendering:

```go
func NewImageMagickRenderer(cfg config.ImageConfig) (Renderer, error) {
    cfg.Finalize()  // Apply defaults

    // Validate configuration
    // - Format must be "png" or "jpg"
    // - JPEG quality 1-100
    // - Filter ranges: -100 to +100
    // - Rotation: 0 to 360 degrees

    return &imagemagickRenderer{
        format:     cfg.Format,
        quality:    cfg.Quality,
        dpi:        cfg.DPI,
        brightness: brightness,
        contrast:   contrast,
        saturation: saturation,
        rotation:   rotation,
    }, nil
}
```

**Transformation Pattern**: Configuration (data) → Validation → Domain Object (behavior)

**Validation Boundary**: All validation occurs in `NewImageMagickRenderer()`. Invalid configurations are rejected before creating the renderer.

**Implementation Details**:
- Struct `imagemagickRenderer` is unexported (private)
- Constructor returns `Renderer` interface (public API)
- Stores validated values as concrete int fields (no pointers in domain object)
- `Render()` method builds ImageMagick command with validated parameters
- Filter application implementation pending (Session 5)

**Benefits**:
- Consumers cannot access implementation-specific methods
- Easy to add new renderer implementations (e.g., LibreOffice, Ghostscript)
- Testing through interface mocks
- Validation centralized at transformation point

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
    ToImage(renderer image.Renderer) ([]byte, error)
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

The document layer delegates to the image rendering layer through the Renderer interface:

```go
func (p *PDFPage) ToImage(renderer image.Renderer) ([]byte, error) {
    // Get file extension from renderer
    ext := renderer.FileExtension()

    // Create temporary file for output
    tmpFile, err := os.CreateTemp("", fmt.Sprintf("page-%d-*.%s", p.number, ext))
    if err != nil {
        return nil, fmt.Errorf("failed to create temp file: %w", err)
    }
    tmpPath := tmpFile.Name()
    tmpFile.Close()
    defer os.Remove(tmpPath)

    // Delegate rendering to renderer
    err = renderer.Render(p.doc.path, p.number, tmpPath)
    if err != nil {
        return nil, fmt.Errorf("failed to render page %d: %w", p.number, err)
    }

    // Read generated image
    imgData, err := os.ReadFile(tmpPath)
    if err != nil {
        return nil, fmt.Errorf("failed to read generated image: %w", err)
    }

    return imgData, nil
}
```

**Design Pattern**: Clean delegation with clear separation of concerns.

**Responsibilities**:
- **Document Layer (PDFPage)**: Temporary file management, error context
- **Image Layer (Renderer)**: Format validation, ImageMagick integration, rendering logic

**Benefits**:
1. **No ImageMagick Knowledge**: Document layer doesn't know about ImageMagick, densities, formats, or quality settings
2. **No Validation**: Renderer is already validated at creation time (by `NewImageMagickRenderer()`)
3. **Simple Logic**: Create temp file → delegate to renderer → read result
4. **Clear Errors**: Contextual error messages with page numbers
5. **Interface-Based**: Can swap renderer implementations without changing document code

**Temporary File Management**:
- File extension obtained from renderer (encapsulates format knowledge)
- Unique naming prevents conflicts in concurrent operations
- `defer os.Remove()` ensures cleanup even on errors
- File handle closed before renderer writes to it

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

### Configuration Transformation Pattern

Configuration structures are ephemeral data containers that transform into domain objects at package boundaries through finalization, validation, and initialization functions. This pattern creates a clear separation between data (configuration) and behavior (domain objects).

**Lifecycle**:
```
1. Create/Load Configuration (JSON, code, defaults)
    ↓
2. Finalize Configuration (merge defaults)
    ↓
3. Transform to Domain Object via New*() (validate)
    ↓
4. Use Domain Object (configuration discarded)
```

**Example**:
```go
// 1. Create configuration
cfg := config.ImageConfig{
    Format: "png",
    DPI:    150,
}

// 2. Transform to domain object (Finalize + Validate)
renderer, err := image.NewImageMagickRenderer(cfg)
if err != nil {
    return fmt.Errorf("invalid configuration: %w", err)
}

// 3. Use domain object (config is now discarded)
imageData, err := page.ToImage(renderer)
```

**Configuration Responsibilities**:
- Structure definitions with JSON serialization
- Default value creation via `Default*()` functions
- Configuration merging via `Merge()` methods
- Finalization via `Finalize()` method (merges defaults)
- **NO** validation of domain-specific values
- **NO** imports of domain packages
- **NO** business logic

**Domain Object Responsibilities**:
- Validate configuration values during transformation
- Transform configuration into domain objects via `New*()` functions
- Encapsulate runtime behavior and business logic
- Provide interface-based public APIs
- Store validated values as concrete types (no pointers)

**Transformation Function Pattern**:
```go
func NewDomainObject(cfg config.DomainConfig) (Interface, error) {
    // Finalize configuration (merge defaults)
    cfg.Finalize()

    // Validate configuration values
    if cfg.Field < minValue || cfg.Field > maxValue {
        return nil, fmt.Errorf("field must be %d-%d, got %d",
            minValue, maxValue, cfg.Field)
    }

    // Transform to domain object
    return &domainObjectImpl{
        field: cfg.Field,
        // Extract and store validated values
    }, nil
}
```

**Benefits**:
- Clear separation: data (config) vs behavior (domain objects)
- Configuration is ephemeral and doesn't leak into runtime
- Domain objects are always constructed in a valid state
- Interface-based APIs prevent exposure of implementation details
- Enables clean testing through interface mocks
- Validation centralized at transformation boundary

**Layered Dependency Hierarchy Implementation**:

The library implements a strict layered dependency hierarchy where higher-level packages depend on lower-level interfaces:

```
pkg/document (high-level)
    ↓ uses image.Renderer interface
pkg/image (mid-level)
    ↓ transforms config.ImageConfig
pkg/config (low-level)
```

**Benefits**:
- Maximizes library reusability (image.Renderer usable beyond PDFs)
- Prevents tight coupling between layers
- Enables independent testing of each layer
- Facilitates parallel development across layers
- Clear architectural boundaries prevent responsibility creep

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
