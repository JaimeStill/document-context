# Architecture

This document describes the current architecture and implementation of the document-context library.

## Package Structure

```
pkg/
├── config/             # Configuration data structures
│   ├── doc.go          # Package documentation
│   ├── image.go        # ImageConfig and ImageMagickConfig
│   ├── parse.go        # Options map parsing helpers
│   ├── cache.go        # CacheConfig structure
│   └── logger.go       # LoggerConfig structure
├── logger/             # Structured logging infrastructure
│   ├── doc.go          # Package documentation
│   ├── logger.go       # Logger interface
│   └── slogger.go      # log/slog implementation
├── cache/              # Persistent image caching infrastructure
│   ├── doc.go          # Package documentation
│   ├── cache.go        # Cache interface, CacheEntry, key generation
│   ├── registry.go     # Factory registration and cache creation
│   └── filesystem.go   # Filesystem-based cache implementation
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
pkg/document → pkg/cache
pkg/logger → pkg/config
pkg/encoding (independent utility)
```

## Architectural Framework

This library implements Layer 1 (Package) and Layer 2 (Library/Module) patterns from **Layered Composition Architecture**. Package-level transformations (configuration → domain objects) compose upward into library-level capabilities (module public API).

**Configuration flows downward**: Applications → Libraries → Packages → Configuration structs
**Interfaces flow upward**: Package interfaces → Library API → Application usage

See [LCA Framework](./_context/lca/layered-composition-architecture.md) for the complete philosophy spanning package through platform layers, or [LCA Synopsis](./_context/lca/lca-synopsis.md) for a quick overview.

## Core Abstractions

### Configuration Package (pkg/config/)

The configuration package provides ephemeral data structures for initializing domain objects. Configuration structures support JSON serialization, default values, and merge semantics for layered configuration.

**Key Principle**: Configuration is data, not behavior. Validation happens in domain packages during transformation.

#### ImageConfig

Base configuration for image rendering operations using the Configuration Composition Pattern:

```go
type ImageConfig struct {
    Format  string         `json:"format,omitempty"`  // "png" or "jpg"
    Quality int            `json:"quality,omitempty"` // JPEG quality: 1-100
    DPI     int            `json:"dpi,omitempty"`     // Render density
    Options map[string]any `json:"options,omitempty"` // Implementation-specific options
}
```

**Design**: ImageConfig provides universal rendering settings (format, quality, DPI) while Options map contains implementation-specific configuration that is parsed during renderer initialization.

**Configuration Methods**:
- `DefaultImageConfig()`: Returns PNG format, 300 DPI, 0 quality, empty Options map
- `Merge(source *ImageConfig)`: Overlays non-zero values and copies Options using `maps.Copy()`
- `Finalize()`: Applies default values for any unset fields by merging onto defaults

**Merge Semantics**:
- String fields: only merge if non-empty
- Integer fields: only merge if greater than zero
- Options map: merged using `maps.Copy()` (source values override base)

#### ImageMagickConfig

Implementation-specific configuration parsed from ImageConfig.Options:

```go
type ImageMagickConfig struct {
    Config     ImageConfig // Embedded base configuration
    Background string      // Background color for alpha flattening
    Brightness *int        // 0-200, where 100 is neutral
    Contrast   *int        // -100 to +100, where 0 is neutral
    Saturation *int        // 0-200, where 100 is neutral
    Rotation   *int        // 0-360 degrees clockwise
}
```

**Design**: Embeds ImageConfig for unified access to both universal and implementation-specific settings. Filter fields use pointers to distinguish "not set" (nil) from "explicitly set" (non-nil).

**Parsing**: Created via `parseImageMagickConfig()` which extracts and validates Options map entries using parsing helpers (`ParseString`, `ParseNilIntRanged`).

**Configuration Methods**:
- `DefaultImageMagickConfig()`: Returns default ImageConfig with "white" background and nil filters

#### Configuration Parsing Helpers

Type-safe extraction functions for Options map values:

**ParseString(options, key, fallback)**: Extracts string with fallback support
- Returns fallback if key absent
- Validates value is non-empty string
- Returns error for wrong type or empty string

**ParseNilIntRanged(options, key, low, high)**: Extracts optional integer with range validation
- Returns nil if key absent (not configured)
- Handles JSON float64 → int conversion
- Validates value is within [low, high] range
- Returns pointer to value if present and valid

#### CacheConfig

Configuration for cache implementations:

```go
type CacheConfig struct {
    Name    string         `json:"name"`              // Implementation identifier
    Logger  LoggerConfig   `json:"logger"`            // Logger configuration
    Options map[string]any `json:"options,omitempty"` // Implementation-specific settings
}
```

**Design**: Name-based approach where Name identifies cache type, Logger provides common logging configuration, and Options provides implementation-specific parameters.

**Configuration Methods**:
- `DefaultCacheConfig()`: Returns empty name, default logger config, and initialized options map
- `Merge(source *CacheConfig)`: Merges name, delegates to Logger.Merge(), and merges options using `maps.Copy()`

**Common Dependencies**: Logger is a common dependency for all cache implementations, so it's included in the base CacheConfig rather than being implementation-specific in Options.

#### LoggerConfig

Configuration for logger initialization following the Configuration Transformation Pattern (Type 1):

```go
type LogLevel string

const (
    LogLevelDebug    LogLevel = "debug"
    LogLevelInfo     LogLevel = "info"
    LogLevelWarn     LogLevel = "warn"
    LogLevelError    LogLevel = "error"
    LogLevelDisabled LogLevel = "disabled"
)

type LoggerOutput string

const (
    LoggerOutputDiscard LoggerOutput = "discard"
    LoggerOutputStdout  LoggerOutput = "stdout"
    LoggerOutputStderr  LoggerOutput = "stderr"
)

type LoggerConfig struct {
    Level  LogLevel      `json:"level,omitempty"`
    Format string        `json:"format,omitempty"`
    Output LoggerOutput  `json:"output"`
}
```

**Design**: Ephemeral configuration that transforms into Logger interface via `logger.NewSlogger()` and is discarded after initialization.

**Configuration Methods**:
- `DefaultLoggerConfig()`: Returns info level with text format
- `DisabledLoggerConfig()`: Returns disabled level for zero-overhead no-op logging
- `Finalize()`: Merges configuration with defaults for any unset fields
- `Merge(source *LoggerConfig)`: Overlays non-empty values from source for layered configuration

**Log Levels**:
- **Debug**: Detailed diagnostic information for development
- **Info**: General operational events and status messages
- **Warn**: Unexpected conditions that don't prevent operation
- **Error**: Failures and exceptional conditions
- **Disabled**: Suppresses all output with minimal overhead

**Log Output Destinations**:
- **discard**: Discards all output (io.Discard)
- **stdout**: Writes to standard output (os.Stdout)
- **stderr**: Writes to standard error (os.Stderr)

**Output Formats**:
- **text**: Human-readable text output
- **json**: Structured JSON output for log aggregation

### Logger Package (pkg/logger/)

The logger package provides structured logging infrastructure through interface-based abstraction. This package transforms configuration (data) into loggers (behavior) with validation at the transformation boundary.

#### Logger Interface

```go
type Logger interface {
    Debug(msg string, args ...any)
    Info(msg string, args ...any)
    Warn(msg string, args ...any)
    Error(msg string, args ...any)
}
```

**Design**: Interface-based public API hides implementation details and enables dependency injection throughout the application.

**Structured Logging**: All methods accept variadic key-value pairs for structured context:
```go
logger.Info("processing document", "path", docPath, "pages", pageCount)
logger.Error("conversion failed", "error", err, "document", docID)
```

**Thread Safety**: All Logger implementations must be safe for concurrent use by multiple goroutines.

#### Slogger Implementation

Implementation using Go's standard `log/slog` package:

```go
func NewSlogger(cfg config.LoggerConfig) (Logger, error) {
    cfg.Finalize()  // Apply defaults

    // Validate configuration
    // - Level must be valid LogLevel constant
    // - Format must be "text" or "json"
    // - Output must be valid LoggerOutput constant

    // Derive output writer from LoggerOutput
    var output io.Writer
    switch cfg.Output {
    case LoggerOutputDiscard, "":
        output = io.Discard
    case LoggerOutputStdout:
        output = os.Stdout
    case LoggerOutputStderr:
        output = os.Stderr
    }

    return &Slogger{
        logger: slog.New(handler),
    }, nil
}
```

**Transformation Pattern**: Configuration (data) → Validation → Logger Interface (behavior)

**Validation Boundary**: All validation occurs in `NewSlogger()`. Invalid configurations are rejected before creating the logger.

**Output Derivation**: Output writer is determined by LoggerOutput enum in configuration rather than passed as parameter, making logger creation fully configuration-driven.

**Implementation Details**:
- Struct `Slogger` is unexported (private)
- Constructor returns `Logger` interface (public API)
- Wraps `slog.Logger` for structured logging
- Thread-safe concurrent access via underlying slog implementation
- Disabled mode automatically uses `io.Discard` for zero overhead
- Output destination controlled by configuration

**Benefits**:
- Consumers cannot access implementation-specific methods
- Easy to add new logger implementations (e.g., zap, zerolog)
- Testing through interface mocks
- Validation centralized at transformation point
- Zero overhead for disabled logging
- Fully configuration-driven (no runtime parameters)

### Cache Package (pkg/cache/)

The cache package provides persistent storage infrastructure for rendered document images through interface-based abstraction. This package enables multiple storage backends (filesystem, blob storage, databases, in-memory) to be used interchangeably.

#### Cache Interface

```go
type Cache interface {
    Get(key string) (*CacheEntry, error)
    Set(entry *CacheEntry) error
    Invalidate(key string) error
    Clear() error
}
```

**Design**: Implementation-agnostic interface supporting multiple storage backends. Consumers depend on the contract, not concrete types.

**Methods**:
- `Get(key)`: Retrieves cache entry by key, returns `ErrCacheEntryNotFound` for cache misses
- `Set(entry)`: Stores cache entry, replacing existing entry with same key
- `Invalidate(key)`: Removes cache entry by key, idempotent (no error if key doesn't exist)
- `Clear()`: Removes all cache entries

**Thread Safety**: Cache implementations must be safe for concurrent use by multiple goroutines.

**Error Handling**: Distinguishes between cache misses (expected, `ErrCacheEntryNotFound`) and storage failures (unexpected, other errors).

#### CacheEntry

Structure encapsulating cached image data with metadata:

```go
type CacheEntry struct {
    Key      string  // SHA256 hash in hexadecimal format (64 characters)
    Data     []byte  // Raw image bytes (PNG, JPEG, etc.)
    Filename string  // Suggested filename: "basename.pagenum.ext"
}
```

**Metadata Purpose**:
- **Filename**: Meaningful names for HTTP Content-Disposition headers or file storage. File extension can be used to derive MIME type when needed (web service concern, not library concern)

#### Cache Key Generation

Deterministic SHA256-based key generation:

```go
func GenerateKey(input string) string {
    hash := sha256.Sum256([]byte(input))
    return hex.EncodeToString(hash[:])
}
```

**Input Format**: Normalized string representing all factors uniquely identifying a cached image:
```
/absolute/path/to/document.pdf/1.png?dpi=300&quality=90&brightness=10
```

**Key Properties**:
- Same input always produces same key (deterministic)
- Different inputs produce different keys with high probability (cryptographic hash)
- 64-character hexadecimal string (SHA256)

**Parameter Normalization**: Caller must normalize parameters (e.g., alphabetical ordering) to ensure consistent key generation.

**Sentinel Error**:
```go
var ErrCacheEntryNotFound = errors.New("cache entry not found")
```

Used to distinguish cache misses from storage failures. Callers should use `errors.Is(err, cache.ErrCacheEntryNotFound)` to detect cache misses.

#### Cache Registry Pattern

The cache package uses a registry pattern to support multiple cache implementations that can be selected via configuration.

**Factory Type**:
```go
type Factory func(c *config.CacheConfig) (Cache, error)
```

Cache implementations register factory functions that create configured Cache instances from CacheConfig.

**Registry Functions**:
```go
func Register(name string, factory Factory)
func Create(c *config.CacheConfig) (Cache, error)
func ListCaches() []string
```

**Registration Pattern**:
```go
func init() {
    cache.Register("filesystem", NewFilesystem)
}
```

Cache implementations register themselves in init() functions, making them available for creation via configuration.

**Creation Pattern**:
```go
cfg := &config.CacheConfig{
    Name: "filesystem",
    Options: map[string]any{"directory": "/var/cache"},
}
cache, err := cache.Create(cfg)
```

**Registry Behavior**:
- Thread-safe using sync.RWMutex
- Silent overwriting: registering the same name multiple times replaces previous factory
- Panics on invalid registration: empty name or nil factory
- Returns error for unknown cache types

**Benefits**:
- Pluggable cache backends without modifying core code
- Configuration-driven cache selection
- Discoverable implementations via `ListCaches()`
- Centralized creation logic with validation

#### Configuration Composition Pattern

Cache implementations use the Configuration Composition Pattern to handle implementation-specific configuration:

**Pattern Flow**:
```
BaseConfig (CacheConfig)
    ↓ Options map[string]any
    ↓ Parse Function
TypedImplConfig (FilesystemCacheConfig)
    ↓ Validation
Domain Object (FilesystemCache)
```

**Example** (FilesystemCache):
```go
// Type-specific configuration
type FilesystemCacheConfig struct {
    Directory string
}

// Parse from generic Options map
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
    return &FilesystemCacheConfig{Directory: directory}, nil
}

// Factory transforms config into domain object
func NewFilesystem(c *config.CacheConfig) (Cache, error) {
    fsConfig, err := parseFilesystemConfig(c.Options)
    if err != nil {
        return nil, err
    }
    // Create cache with validated typed config
    return &FilesystemCache{directory: fsConfig.Directory}, nil
}
```

**Pattern Benefits**:
- Type-safe implementation configuration
- Centralized validation in parse functions
- Extensible without modifying base CacheConfig
- Clear transformation from data to behavior

#### FilesystemCache Implementation

Filesystem-based cache using directory-per-key storage structure:

**Storage Structure**:
```
<cache_root>/
    <key>/
        <filename>
```

Example:
```
/var/cache/
    a3b5c7.../
        document.1.png
    d4e6f8.../
        report.2.jpg
```

**Configuration** (via Options map):
```go
cfg := &config.CacheConfig{
    Name: "filesystem",
    Logger: config.LoggerConfig{Level: config.LogLevelInfo},
    Options: map[string]any{
        "directory": "/var/cache",  // Required: cache root directory
    },
}
```

**Initialization**:
- Normalizes directory path to absolute path
- Creates cache root directory if it doesn't exist (0755 permissions)
- Creates logger from embedded LoggerConfig
- Validates directory option is non-empty string

**Cache Operations**:

- **Get(key)**: Reads key directory, validates exactly 1 file exists (detects corruption), returns CacheEntry
- **Set(entry)**: Creates key directory (0755), writes data to file with entry's filename (0644)
- **Invalidate(key)**: Removes entire key directory, idempotent (no error if key doesn't exist)
- **Clear()**: Iterates through cache root, removes all subdirectories, logs warnings for failures

**Error Handling**:
- **Cache miss**: Returns `ErrCacheEntryNotFound` only when key directory doesn't exist
- **Corruption detection**: Multiple files or directory instead of file returns descriptive error
- **Filesystem errors**: Permission denied, disk full, etc. return wrapped errors

**Thread Safety**: Safe for concurrent operations on different keys via OS filesystem atomicity. Operations on the same key are not atomic across multiple goroutines.

**Logging**: Uses structured logging for debugging:
```go
logger.Debug("cache.get", "key", key, "found", true)
logger.Debug("cache.set", "key", key, "size", len(data))
logger.Info("cache.clear", "file_count", removedCount)
```

**Benefits**:
- Simple, portable storage (no external dependencies)
- Easy inspection and debugging (files visible on filesystem)
- Corruption detection through validation
- Graceful degradation (Clear() continues on individual failures)

### Image Rendering Package (pkg/image/)

The image package defines the interface for rendering documents to images and provides implementations. This package transforms configuration (data) into renderers (behavior) with validation at the transformation boundary.

#### Renderer Interface

```go
type Renderer interface {
    Render(inputPath string, pageNum int, outputPath string) error
    FileExtension() string
    Settings() config.ImageConfig
    Parameters() []string
}
```

**Design**: Interface-based public API hides implementation details. Consumers depend on the contract, not concrete types.

**Methods**:
- `Render(inputPath, pageNum, outputPath)`: Converts specified page to image file
- `FileExtension()`: Returns appropriate file extension for output format (no leading dot)
- `Settings()`: Returns renderer's immutable base configuration (universal settings)
- `Parameters()`: Returns implementation-specific parameters for cache key generation

**Thread Safety**: Renderer instances are immutable once created and safe for concurrent use.

**Settings and Parameters**: The `Settings()` method exposes the universal configuration (format, DPI, quality) while `Parameters()` provides implementation-specific settings (filters, background, etc.). This separation follows the Configuration Composition Pattern (Enhanced: Embedded Base), enabling complete cache key generation that includes both universal and implementation-specific parameters.

#### ImageMagickRenderer

Implementation using ImageMagick for PDF rendering following the Configuration Composition Pattern (Enhanced: Embedded Base):

```go
// ImageMagickConfig embeds base config + adds implementation-specific fields
type ImageMagickConfig struct {
    Config     config.ImageConfig  // Embedded base (universal settings)
    Background string              // ImageMagick-specific
    Brightness *int                // 0-200, 100=neutral (nil=omit)
    Contrast   *int                // -100 to +100 (nil=omit)
    Saturation *int                // 0-200, 100=neutral (nil=omit)
    Rotation   *int                // 0-360 degrees (nil=omit)
}

func NewImageMagickRenderer(cfg config.ImageConfig) (Renderer, error) {
    cfg.Finalize()  // Apply defaults

    // Validate universal settings
    // - Format must be "png" or "jpg"
    // - JPEG quality 1-100

    // Parse Options map into ImageMagickConfig (embeds base + adds specific)
    imCfg, err := parseImageMagickConfig(cfg)
    if err != nil {
        return nil, err
    }

    return &imagemagickRenderer{
        settings: *imCfg,  // Single field contains base + specific
    }, nil
}

func (r *imagemagickRenderer) Settings() config.ImageConfig {
    return r.settings.Config  // Return embedded base config
}

func (r *imagemagickRenderer) Parameters() []string {
    // Return implementation-specific params for cache keys
    params := []string{fmt.Sprintf("background=%s", r.settings.Background)}
    if r.settings.Brightness != nil {
        params = append(params, fmt.Sprintf("brightness=%d", *r.settings.Brightness))
    }
    // ... other optional filters
    return params
}
```

**Transformation Pattern**:
```
BaseConfig → Parse Options → ImageMagickConfig (embeds base) → Validate → Renderer
```

**Configuration Composition Pattern (Enhanced)**:
- Base `ImageConfig` contains universal settings (Format, DPI, Quality) + Options map
- `parseImageMagickConfig()` creates `ImageMagickConfig` that embeds base + adds typed specific fields
- Renderer stores single `settings ImageMagickConfig` field (contains both base and specific)
- `Settings()` returns embedded base config for universal access
- `Parameters()` returns implementation-specific params for cache key generation

**Validation Boundary**:
- Universal settings validated in `NewImageMagickRenderer()`
- Implementation-specific options parsed and validated in `parseImageMagickConfig()`
- Invalid configurations rejected before creating renderer

**Implementation Details**:
- Struct `imagemagickRenderer` is unexported (private)
- Constructor returns `Renderer` interface (public API)
- Stores `ImageMagickConfig` as single `settings` field (immutable after creation)
- Access base via `r.settings.Config`, specific via `r.settings.Background`, etc.
- `Render()` method builds ImageMagick command with filters applied conditionally
- Filter ranges use ImageMagick-native values (Brightness/Saturation: 0-200, Contrast: -100 to +100)

**Benefits**:
- Consumers cannot access implementation-specific methods
- Easy to add new renderer implementations (e.g., LibreOffice, Ghostscript)
- Testing through interface mocks
- Validation centralized at transformation point
- Cache key generation can access complete rendering configuration
- Configuration remains immutable and accessible for introspection

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
    ToImage(renderer image.Renderer, c cache.Cache) ([]byte, error)
}
```

**Design Rationale**:
- **Format Independence**: The same code can process PDFs, Office documents, or any future format without changes
- **Lazy Evaluation**: Pages are extracted on-demand rather than loading entire documents into memory
- **Resource Management**: Explicit `Close()` method for cleanup of document resources
- **Composability**: Page operations are independent, enabling parallel processing
- **Optional Caching**: Cache parameter enables transparent caching without affecting non-cached workflows

**Interface Benefits**:
- New format support requires only interface implementation
- Testing through mocks without file I/O
- Clear contracts between document access and conversion
- Enables batch processing and pipeline composition
- Caching opt-in via parameter (pass `nil` to disable)

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

### Image Conversion with Optional Caching

The document layer delegates to the image rendering layer through the Renderer interface with transparent caching support:

```go
func (p *PDFPage) ToImage(renderer image.Renderer, c cache.Cache) ([]byte, error) {
    // Check cache if provided
    if c != nil {
        key, err := p.buildCacheKey(renderer.Settings())
        if err != nil {
            return nil, err
        }

        entry, err := c.Get(key)
        if err == nil {
            return entry.Data, nil  // Cache hit
        }
        if !errors.Is(err, cache.ErrCacheEntryNotFound) {
            return nil, err  // Real cache error
        }
    }

    // Cache miss or no cache - render page
    ext := renderer.FileExtension()
    tmpFile, err := os.CreateTemp("", fmt.Sprintf("page-%d-*.%s", p.number, ext))
    if err != nil {
        return nil, fmt.Errorf("failed to create temp file: %w", err)
    }
    tmpPath := tmpFile.Name()
    tmpFile.Close()
    defer os.Remove(tmpPath)

    err = renderer.Render(p.doc.path, p.number, tmpPath)
    if err != nil {
        return nil, fmt.Errorf("failed to render page %d: %w", p.number, err)
    }

    imgData, err := os.ReadFile(tmpPath)
    if err != nil {
        return nil, fmt.Errorf("failed to read rendered image: %w", err)
    }

    // Store in cache if provided
    if c != nil {
        entry, err := p.prepareCache(imgData, renderer.Settings())
        if err != nil {
            return nil, err
        }
        if err := c.Set(entry); err != nil {
            return nil, err
        }
    }

    return imgData, nil
}
```

**Cache-Aware Rendering Pattern**: Transparent caching layer that checks cache before rendering and stores results after rendering. Caching is completely optional via the cache parameter.

**Caching Behavior**:
- **Cache provided + cache hit**: Returns cached image immediately (no rendering)
- **Cache provided + cache miss**: Renders page, stores in cache, returns image
- **No cache (nil)**: Always renders page (original behavior, no caching overhead)
- **Cache errors**: Non-`ErrCacheEntryNotFound` errors are propagated as failures

**Design Pattern**: Clean delegation with optional optimization through caching.

**Responsibilities**:
- **Document Layer (PDFPage)**: Cache checking, temporary file management, cache storage, error context
- **Cache Layer**: Persistent storage of rendered images with metadata
- **Image Layer (Renderer)**: Format validation, ImageMagick integration, rendering logic

**Benefits**:
1. **Optional Performance Optimization**: Caching available when needed without affecting non-cached workflows
2. **No ImageMagick Knowledge**: Document layer doesn't know about ImageMagick internals
3. **Transparent Caching**: Callers don't need to manage cache checking/storing logic
4. **Clear Errors**: Contextual error messages with page numbers
5. **Interface-Based**: Can swap renderer and cache implementations independently
6. **Backward Compatible**: Pass `nil` for cache parameter to disable caching

**Temporary File Management**:
- File extension obtained from renderer (encapsulates format knowledge)
- Unique naming prevents conflicts in concurrent operations
- `defer os.Remove()` ensures cleanup even on errors
- File handle closed before renderer writes to it

#### Cache Key Generation

Deterministic cache key generation from page and rendering settings:

```go
func (p *PDFPage) buildCacheKey(settings config.ImageConfig) (string, error) {
    absPath, err := filepath.Abs(p.doc.path)
    if err != nil {
        return "", fmt.Errorf("failed to normalize path: %w", err)
    }

    var builder strings.Builder
    builder.WriteString(fmt.Sprintf("%s/%d.%s", absPath, p.number, settings.Format))

    params := make([]string, 0)
    params = append(params, fmt.Sprintf("dpi=%d", settings.DPI))
    params = append(params, fmt.Sprintf("quality=%d", settings.Quality))

    // Add optional filter parameters if present
    if settings.Brightness != nil {
        params = append(params, fmt.Sprintf("brightness=%d", *settings.Brightness))
    }
    if settings.Contrast != nil {
        params = append(params, fmt.Sprintf("contrast=%d", *settings.Contrast))
    }
    if settings.Rotation != nil {
        params = append(params, fmt.Sprintf("rotation=%d", *settings.Rotation))
    }
    if settings.Saturation != nil {
        params = append(params, fmt.Sprintf("saturation=%d", *settings.Saturation))
    }

    builder.WriteString(fmt.Sprintf("?%s", strings.Join(params, "&")))

    key := cache.GenerateKey(builder.String())
    return key, nil
}
```

**Cache Key Format** (before hashing):
```
/absolute/path/to/document.pdf/1.png?dpi=300&quality=90&brightness=10
```

**Key Components**:
- Document path (normalized to absolute path)
- Page number
- Image format
- All rendering parameters in deterministic alphabetical order

**Deterministic Ordering**: Parameters included alphabetically (brightness, contrast, rotation, saturation) to ensure the same inputs always produce the same key.

**Why Type 2 Pattern Enables This**: The `Settings()` method provides access to the complete rendering configuration needed for cache key generation. Without persistent settings access, cache key generation would be impossible.

#### Cache Entry Preparation

Constructs complete cache entry with metadata:

```go
func (p *PDFPage) prepareCache(data []byte, renderer image.Renderer) (*cache.CacheEntry, error) {
    key, err := p.buildCacheKey(renderer)
    if err != nil {
        return nil, err
    }

    settings := renderer.Settings()
    baseName := filepath.Base(p.doc.path)
    ext := filepath.Ext(baseName)
    nameWithoutExt := strings.TrimSuffix(baseName, ext)

    filename := fmt.Sprintf("%s.%d.%s", nameWithoutExt, p.number, settings.Format)

    return &cache.CacheEntry{
        Key:      key,
        Data:     data,
        Filename: filename,
    }, nil
}
```

**Filename Construction**: Formatted as `basename.pagenum.ext` (e.g., "document.1.png")

**Complete Entry**: Provides all metadata needed for cache storage and retrieval

### Cache Integration Specification

The document layer integrates persistent caching to avoid redundant PDF page rendering. This section specifies the cache key format, integration flow, and behavior matrix.

#### Cache Key Format

Cache keys uniquely identify rendered images based on all factors affecting the output. The key generation process creates a deterministic string representation of these factors, then hashes it with SHA256 for consistent key length and collision avoidance.

**Pre-Hash Format**:
```
/absolute/path/to/document.pdf/1.png?dpi=300&quality=90&background=white&brightness=110
```

**Components** (in order):
1. **Document Path**: Absolute path to source PDF (`/absolute/path/to/document.pdf`)
2. **Page Number**: 1-indexed page number (`/1`)
3. **Image Format**: Output format extension (`.png` or `.jpg`)
4. **Query Parameters**: Rendering settings in deterministic alphabetical order

**Mandatory Parameters** (always included):
- `dpi={value}` - Rendering density (e.g., `dpi=300`)
- `quality={value}` - JPEG quality or 0 for PNG (e.g., `quality=90`)

**Optional Parameters** (included when non-nil, alphabetically):
- `background={color}` - Background color for alpha flattening (e.g., `background=white`)
- `brightness={value}` - Brightness adjustment 0-200 (e.g., `brightness=110`)
- `contrast={value}` - Contrast adjustment -100 to +100 (e.g., `contrast=-10`)
- `rotation={degrees}` - Rotation in degrees 0-360 (e.g., `rotation=90`)
- `saturation={value}` - Saturation adjustment 0-200 (e.g., `saturation=120`)

**Post-Hash Key**: SHA256 hash in hexadecimal format (64 characters):
```
a3a6788c43b16d73b83cc01f34ea39e416bf1fcbff5cbaccceb818b1118f06ed
```

**Key Properties**:
- **Deterministic**: Same inputs always produce same key
- **Unique**: Different configurations produce different keys with high probability
- **Portable**: Absolute paths ensure different machines can share cache if paths match
- **Complete**: All rendering parameters included to prevent incorrect cache hits

**Parameter Ordering**: Alphabetical ordering of all parameters ensures deterministic key generation regardless of configuration source.

#### Cache Integration Flow

```
┌─────────────────────────────────────────────────────────────┐
│ PDFPage.ToImage(renderer, cache)                            │
└─────────────────────────────────────────────────────────────┘
                         │
                         ├─ cache != nil?
                         │
        ┌────────────────┴────────────────┐
        │ NO                              │ YES
        │                                 │
        ▼                                 ▼
   ┌─────────────┐              ┌──────────────────┐
   │ Skip Cache  │              │ Generate Cache   │
   │ Logic       │              │ Key from Renderer│
   └─────────────┘              └──────────────────┘
        │                                 │
        │                                 ▼
        │                        ┌──────────────────┐
        │                        │ cache.Get(key)   │
        │                        └──────────────────┘
        │                                 │
        │                ┌────────────────┴───────────────┐
        │                │ Found?                         │
        │                │                                │
        │         ┌──────┴─────┐                    ┌─────┴─────┐
        │         │ YES        │                    │ NO (Miss) │
        │         │            │                    │           │
        │         ▼            ▼                    ▼           │
        │    ┌─────────┐  ┌────────┐          ┌─────────┐       │
        │    │ Error?  │  │ Cache  │          │ Error?  │       │
        │    │         │  │ Hit    │          │         │       │
        │    └─────────┘  └────────┘          └─────────┘       │
        │         │            │                    │           │
        │         │ YES        │ NO                 │ YES       │ NO
        │         │            │                    │           │
        │         ▼            ▼                    ▼           ▼
        │    ┌─────────┐  ┌─────────┐         ┌─────────┐ ┌─────────┐
        │    │ Return  │  │ Return  │         │ Return  │ │ Continue│
        │    │ Error   │  │ entry.  │         │ Error   │ │ to      │
        │    │         │  │ Data    │         │         │ │ Render  │
        │    └─────────┘  └─────────┘         └─────────┘ └─────────┘
        │                                                       │
        └───────────────────────────────────────────────────────┘
                                 │
                                 ▼
                    ┌──────────────────────────┐
                    │ Create Temp File         │
                    │ renderer.Render(...)     │
                    │ Read Image Data          │
                    │ Delete Temp File         │
                    └──────────────────────────┘
                                 │
                                 ├─ cache != nil?
                                 │
                    ┌────────────┴────────────┐
                    │ NO                      │ YES
                    │                         │
                    ▼                         ▼
              ┌───────────┐         ┌──────────────────┐
              │ Return    │         │ Prepare Cache    │
              │ Image     │         │ Entry (key,data, │
              │ Data      │         │ filename)        │
              └───────────┘         └──────────────────┘
                                             │
                                             ▼
                                    ┌──────────────────┐
                                    │ cache.Set(entry) │
                                    └──────────────────┘
                                             │
                                ┌────────────┴───────────┐
                                │ Success?               │
                                │                        │
                         ┌──────┴─────┐          ┌───────┴──────┐
                         │ YES        │          │ NO           │
                         │            │          │              │
                         ▼            ▼          ▼              │
                    ┌─────────┐  ┌─────────┐ ┌─────────┐        │
                    │ Return  │  │ Return  │ │ Return  │        │
                    │ Image   │  │ Image   │ │ Error   │        │
                    │ Data    │  │ Data    │ │         │        │
                    └─────────┘  └─────────┘ └─────────┘        │
                                                                │
                    ┌───────────────────────────────────────────┘
                    │
                    ▼
              ┌───────────┐
              │   Done    │
              └───────────┘
```

**Flow Description**:
1. **Cache Check Phase** (if cache provided):
   - Generate cache key from renderer settings and filter parameters
   - Call `cache.Get(key)` to check for cached image
   - **Cache Hit**: Return cached data immediately (no rendering)
   - **Cache Miss**: Continue to rendering phase
   - **Cache Error**: Propagate error (fail fast on storage failures)

2. **Rendering Phase**:
   - Create temporary file for rendered output
   - Call `renderer.Render()` to convert PDF page to image
   - Read rendered image data from temporary file
   - Delete temporary file (via defer)

3. **Cache Store Phase** (if cache provided):
   - Prepare `CacheEntry` with key, data, and filename
   - Call `cache.Set(entry)` to store rendered image
   - **Store Success**: Return image data
   - **Store Error**: Propagate error (fail fast on storage failures)

4. **No Cache Phase** (if cache nil):
   - Skip all cache operations
   - Perform rendering only
   - Return image data directly

#### Caching Behavior Matrix

| Scenario | Cache Param | Cache.Get() | Rendering | Cache.Set() | Result | Notes |
|----------|-------------|-------------|-----------|-------------|--------|-------|
| **No Cache** | `nil` | Skipped | Always | Skipped | Image data | Original behavior, no caching overhead |
| **Cache Hit** | Provided | Success | Never | Skipped | Cached data | Fastest path, no ImageMagick execution |
| **Cache Miss** | Provided | `ErrCacheEntryNotFound` | Yes | Success | Image data | Normal cache population |
| **Cache Miss + Set Fail** | Provided | `ErrCacheEntryNotFound` | Yes | Error | Error | Storage failure prevents caching |
| **Cache Get Error** | Provided | Storage Error | Never | Skipped | Error | Fail fast on cache infrastructure issues |
| **Cache Set Error** | Provided | `ErrCacheEntryNotFound` | Yes | Error | Error | Fail fast on cache infrastructure issues |

**Error Handling Philosophy**:
- **Cache Misses**: Expected behavior (`ErrCacheEntryNotFound`), continue to rendering
- **Storage Failures**: Unexpected errors, propagated immediately (fail fast)
- **No Graceful Degradation**: Cache errors indicate infrastructure problems that should be surfaced, not hidden

**Performance Implications**:
- **Cache Hit**: ~1ms (cache retrieval only, no rendering)
- **Cache Miss**: ~500-1000ms (rendering + cache storage)
- **No Cache**: ~500ms (rendering only, no cache overhead)

**Concurrency Safety**:
- **Multiple Renders**: Safe to render same page concurrently in different goroutines
- **Cache Implementation**: Responsible for thread-safe operations
- **Filesystem Cache**: Directory-per-key structure enables concurrent access to different keys
- **Same Key Concurrent Access**: Last write wins (filesystem atomic replace)

**Cache Key Determinism Validation**:
Tests verify that:
1. Same configuration produces same cache key (deterministic)
2. Different document paths produce different keys
3. Different page numbers produce different keys
4. Different formats (PNG vs JPEG) produce different keys
5. Different DPI values produce different keys
6. Different filter values (brightness, contrast, etc.) produce different keys
7. Parameters included in alphabetical order regardless of configuration source

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
├── config/
│   ├── cache_test.go         # CacheConfig and LoggerConfig tests
│   └── logger_test.go        # LoggerConfig tests
├── logger/
│   └── slogger_test.go       # Slogger implementation tests
├── cache/
│   ├── cache_test.go         # CacheEntry and key generation tests
│   ├── registry_test.go      # Registry pattern tests (Session 3)
│   └── filesystem_test.go    # FilesystemCache implementation tests (Session 4)
├── image/
│   └── imagemagick_test.go   # ImageMagick renderer tests
├── document/
│   └── pdf_test.go           # PDF document and page tests
└── encoding/
    └── image_test.go         # Data URI encoding tests
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
- **Logger Configuration**: Default values, finalization, merging behavior, output destinations
- **Logger Implementation**: Output formats, log levels, argument handling, disabled mode
- **Cache Registry**: Registration, factory creation, listing, thread safety, concurrent access
- **Cache Implementation**: Filesystem CRUD operations, corruption detection, directory-per-key structure
- **Cache Infrastructure**: Key generation (deterministic, format, ordering), entry structure
- **Cache Configuration**: Options parsing, validation, Configuration Composition Pattern
- **Cache Integration** (Session 6): Cache hits, misses, error handling, key determinism, filter integration, concurrency
- **Image Rendering**: Renderer configuration, Settings() access, Parameters() method, ImageMagick integration with filters
- **PDF Processing**: Document loading, page extraction, bounds checking, cache-aware rendering with ToImage()
- **Image Encoding**: Data URI encoding, format validation
- **Error Handling**: Missing binaries, invalid configurations, cache misses vs errors, cache corruption, cache storage failures

**Test Patterns**:
- Table-driven tests for multiple scenarios
- Black-box testing using `package <name>_test`
- Interface compliance verification
- Magic byte validation for image format verification
- Error case testing (missing files, invalid pages, out of range, invalid configs)
- Default value and finalization behavior testing
- Conditional execution for external binary dependencies
- Concurrency testing using sync.WaitGroup and error channels
- Temporary directory isolation using t.TempDir()
- Mock implementations for controlled cache behavior testing

**Test Statistics** (Session 6):
- **Cache Package Coverage**: 87.1%
- **Config Package Coverage**: 95.3%
- **Document Package**: 13 cache integration tests + 7 basic operation tests
- **Registry Tests**: 10 test functions covering registration, creation, listing, concurrency
- **Filesystem Tests**: 13 test functions covering factory, CRUD, corruption, concurrency
- **Cache Integration Tests**: 13 test functions covering hits, misses, determinism, filters, errors, concurrency
- **All tests pass with -race flag** (no race conditions detected)

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

Configuration structures are data containers that transform into domain objects at package boundaries through finalization, validation, and initialization functions. This pattern creates a clear separation between data (configuration) and behavior (domain objects).

The library uses three distinct configuration pattern types based on how configuration persists after transformation:

#### Type 1: Configuration Transformation (Initialization-Only)

Configuration exists only during initialization and is discarded after transformation into domain objects.

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

**Example** (Logger):
```go
// 1. Create configuration
cfg := config.LoggerConfig{
    Level:  config.LogLevelInfo,
    Format: "json",
}

// 2. Transform to domain object (Finalize + Validate)
logger, err := logger.NewSlogger(cfg, os.Stderr)
if err != nil {
    return fmt.Errorf("invalid configuration: %w", err)
}

// 3. Use domain object (config is now discarded)
logger.Info("application started")
```

**Characteristics**:
- Configuration ephemeral, discarded after transformation
- Domain object stores only extracted primitive values
- No method to retrieve original configuration
- Examples: Logger (configuration → Logger interface)

#### Type 2: Immutable Runtime Settings

Configuration is stored directly in the domain object and accessible throughout its lifetime via `Settings()` method. Configuration remains immutable after creation.

**Lifecycle**:
```
1. Create/Load Configuration (JSON, code, defaults)
    ↓
2. Finalize Configuration (merge defaults)
    ↓
3. Transform to Domain Object via New*() (validate, store config)
    ↓
4. Use Domain Object (settings accessible via Settings() method)
```

**Example** (Image Renderer):
```go
// 1. Create configuration
cfg := config.ImageConfig{
    Format:  "png",
    DPI:     300,
    Quality: 90,
}

// 2. Transform to domain object (stores complete config)
renderer, err := image.NewImageMagickRenderer(cfg)
if err != nil {
    return fmt.Errorf("invalid configuration: %w", err)
}

// 3. Use domain object (settings remain accessible)
imageData, err := page.ToImage(renderer, cache)

// 4. Access settings when needed (e.g., cache key generation)
settings := renderer.Settings()
key := buildCacheKey(settings)
```

**Implementation Pattern**:
```go
type imagemagickRenderer struct {
    settings config.ImageConfig  // Store entire config
}

func NewImageMagickRenderer(cfg config.ImageConfig) (Renderer, error) {
    cfg.Finalize()
    // validate...
    return &imagemagickRenderer{
        settings: cfg,  // Store complete configuration
    }, nil
}

func (r *imagemagickRenderer) Settings() config.ImageConfig {
    return r.settings  // Expose immutable settings
}
```

**Characteristics**:
- Configuration stored as `settings` field in domain object
- Accessible via `Settings()` method in interface
- Configuration immutable after creation
- Enables runtime introspection (e.g., cache key generation)
- Examples: Renderer (ImageConfig accessible throughout lifetime)

**Why This Matters**: Cache key generation requires access to complete rendering parameters (DPI, quality, filters). Without the `Settings()` method, this would be impossible.

#### Type 3: Mutable Runtime Settings (Future)

Configuration can be modified after creation via setter methods. Enables runtime adjustment of operational parameters.

**Not yet implemented** - placeholder for future enhancement (e.g., adjustable log levels, dynamic rendering parameters).

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
