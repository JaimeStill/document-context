# Project Scope and Roadmap

## Overview

**document-context** is a Go library for converting documents into context-friendly formats suitable for LLM consumption and analysis. The library provides format-agnostic interfaces for document processing with extensible format support.

This project was created as a tooling extension for the [go-agents](https://github.com/JaimeStill/go-agents) project, but is not directly dependent on go-agents and can be used as a standalone library for document processing needs.

## Current Status

**Phase**: Pre-Release Development - Phase 2 Session 2 Complete

Phase 2 Session 2 (Cache and Logging Infrastructure) completed with comprehensive test coverage. The logger and cache packages establish interface-based abstractions for structured logging and persistent image storage. Both Type 1 (Logger) and Type 2 (Renderer Settings) configuration transformation patterns are now implemented, enabling cache-aware rendering operations.

The API is under active development and subject to change as Phase 2 features are added. The library is functional for its current capabilities but should be considered experimental until the first versioned release (v0.1.0).

**v0.1.0 Scope**: Core PDF processing with image caching, enhancement filters, and web service readiness features. Multi-format support (Office documents, OCR) is deferred to v1.0.0+.

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
├── config/            # Configuration data structures (Session 1 ✅)
│   ├── doc.go         # Package documentation
│   ├── image.go       # ImageConfig with filter fields
│   ├── cache.go       # CacheConfig structure
│   └── logger.go      # LoggerConfig structure (Session 2 ✅)
├── logger/            # Structured logging infrastructure (Session 2 ✅)
│   ├── doc.go         # Package documentation
│   ├── logger.go      # Logger interface
│   └── slogger.go     # log/slog implementation
├── cache/             # Persistent image caching infrastructure (Session 2 ✅)
│   ├── doc.go         # Package documentation
│   └── cache.go       # Cache interface, CacheEntry, key generation
├── image/             # Image rendering domain objects (Session 1 ✅, enhanced Session 2 ✅)
│   ├── image.go       # Renderer interface with Settings() method
│   └── imagemagick.go # ImageMagick implementation
├── document/          # Core document processing (enhanced Session 2 ✅)
│   ├── document.go    # Document and Page interfaces
│   └── pdf.go         # PDF implementation with cache-aware rendering
└── encoding/          # Output encoding utilities
    └── image.go       # Base64 data URI encoding
```

### What This Library Does NOT Provide

The library is intentionally scoped as a document processing utility. The following capabilities are outside the project scope:

- **LLM Integration**: Interaction with language models (use go-agents for this)
- **Document Classification**: Semantic analysis or categorization of documents
- **Document Generation**: Creating or modifying documents (read-only operations only)
- **Format Conversion Chains**: Multi-step conversion pipelines (single-step conversions only)

### Version Scope Clarification

**v0.1.0 Focus**: Production-ready PDF processing with web service capabilities (caching, enhancement filters, JSON configuration). This release completes the core infrastructure needed for agent-lab integration.

**v1.0.0+ Focus**: Multi-format document support (Office documents, OCR), advanced text extraction, and sophisticated processing pipelines. These features build upon the stable v0.1.0 foundation.

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

## v0.1.0 Completion Roadmap (Phase 2)

Phase 2 completes the v0.1.0 release by adding web service readiness capabilities: persistent image caching for evaluation and verification, enhancement filters for document clarity optimization, and JSON configuration support for web service integration.

### Architecture Decisions

The following architectural decisions were established during Phase 2 planning and guide all implementation sessions:

**Configuration Strategy**:
- Rename `ImageOptions` → `ImageConfig` and move to `pkg/config/` following go-agents patterns
- Integrate filter parameters directly into `ImageConfig` (brightness, contrast, saturation, rotation)
- Use pointer fields (`*int`) only for optional parameters without meaningful zero values
- Follow go-agents config patterns: JSON tags (snake_case), `Default*()` functions, `Merge()` methods

**Cache Purpose and Design**:
- Primary purpose: Persistent storage for generated images enabling visual evaluation and accuracy verification
- Not for performance optimization (no in-memory LRU layer)
- Interface-based design agnostic to storage mechanism (filesystem, blob storage, SQL, etc.)
- Registry pattern following go-agents provider registry for pluggable implementations
- Filesystem cache as initial reference implementation

**Cache Registry Pattern**:
- Registry structure with thread-safe registration and lookup (`sync.RWMutex`)
- Factory functions: `type Factory func(c *config.CacheConfig) (Cache, error)`
- Azure-style validation: extract options from `map[string]any`, validate immediately in factory
- Implementations register in `init()` functions
- Clear error messages for missing or invalid configurations

**Logging Strategy**:
- Define `Logger` interface in `pkg/cache/` with Debug/Info/Warn/Error methods
- Optional logging via `NoOpLogger` as default (no nil checks needed)
- Applications provide slog adapters or custom implementations
- Log levels: Debug (cache hits/misses), Info (initialization), Warn (issues), Error (failures)
- Follows "Interface-Based Layer Interconnection" principle from CLAUDE.md

**Filter Integration**:
- Filters are part of image conversion process, not separate operations
- Applied via ImageMagick parameters: rotation → brightness/saturation → contrast
- Parameter mappings: Brightness/Saturation (-100 to +100) → ImageMagick (0 to 200)
- Only add filter arguments when config fields are non-nil

**Examples Structure**:
- Follow go-agents-orchestration patterns: one directory per example
- Each example: `main.go` + comprehensive `README.md`
- Progressive complexity: basic conversion → filters → caching
- Use slog JSON handler for observability
- Numbered output sections with progress indicators (✓)

### Development Sessions

Phase 2 is broken down into eight focused development sessions, organized from lowest to highest level dependencies. Each session is designed to be executed as an isolated Claude Code session with clear deliverables.

#### Session 1: Configuration Foundation

**Goal**: Establish `pkg/config/` with `ImageConfig` and `CacheConfig` following go-agents patterns

**Tasks**:
1. Create `pkg/config/` package structure (doc.go, image.go, cache.go)
2. Migrate `ImageOptions` → `ImageConfig` from `pkg/document/document.go` to `pkg/config/image.go`
3. Add JSON tags following go-agents conventions (snake_case, `omitempty`)
4. Add filter fields: `Brightness *int`, `Contrast *int`, `Saturation *int`, `Rotation *int`
5. Implement `DefaultImageConfig()` function with sensible defaults
6. Implement `Merge(source *ImageConfig)` method for config composition
7. Define `CacheConfig` structure:
   ```go
   type CacheConfig struct {
       Name    string         `json:"name"`              // Registry lookup name
       Options map[string]any `json:"options,omitempty"` // Implementation-specific
   }
   ```
8. Document expected options for filesystem cache in code comments
9. Update all existing code to import and use `config.ImageConfig`
10. Update test files to use new config package
11. Create comprehensive tests in `tests/config/`:
    - JSON marshaling/unmarshaling
    - Default value creation
    - Merge operations
    - Round-trip serialization

**Deliverables**:
- ✅ `pkg/config/` package complete with doc.go, image.go, cache.go
- ✅ `ImageConfig` with filter fields and JSON support
- ✅ `CacheConfig` structure defined
- ✅ All existing code updated to use `config.ImageConfig`
- ✅ Tests with 80%+ coverage

**Estimated Effort**: 2-3 hours

---

#### Session 2: Cache and Logging Infrastructure ✅

**Goal**: Establish infrastructure for structured logging and persistent image caching

**Completed Tasks**:
1. ✅ Created `pkg/config/logger.go` with LogLevel type and LoggerConfig structure
2. ✅ Created `pkg/logger/` package with Logger interface and Slogger implementation
3. ✅ Created `pkg/cache/` package with Cache interface and CacheEntry structure
4. ✅ Implemented `GenerateKey()` function for deterministic SHA256-based cache keys
5. ✅ Added `Settings()` method to Renderer interface (Type 2 Configuration Pattern)
6. ✅ Enhanced `PDFPage.ToImage()` with optional cache parameter for transparent caching
7. ✅ Implemented `buildCacheKey()` and `prepareCache()` methods in PDFPage
8. ✅ Created comprehensive tests:
   - `tests/config/logger_test.go` (124 lines, 6 test functions)
   - `tests/logger/slogger_test.go` (203 lines, 10 test functions)
   - `tests/cache/cache_test.go` (86 lines, 7 test functions)
   - Updated `tests/image/imagemagick_test.go` with Settings() test
   - Fixed `tests/document/pdf_test.go` for cache parameter
9. ✅ Created comprehensive package documentation (doc.go files)
10. ✅ Updated all code with godoc comments

**Deliverables**:
- ✅ Logger package with Type 1 Configuration Pattern implementation
- ✅ Cache package with interface, CacheEntry, and key generation
- ✅ Type 2 Configuration Pattern (Settings() method) for Renderer
- ✅ Cache-aware rendering in PDFPage with transparent caching behavior
- ✅ 413 lines of new test code with 23+ test functions
- ✅ Complete godoc documentation across all Session 2 components
- ✅ All tests passing

**Architectural Contributions**:
- Established Type 1 vs Type 2 Configuration Pattern distinction
- Implemented Interface-Based Layer Interconnection for logging and caching
- Demonstrated optional dependency pattern (cache parameter can be nil)
- Created foundation for Session 3 (cache registry and implementations)

**Actual Effort**: Complete session

---

#### Session 3: Cache Registry Infrastructure

**Goal**: Implement registry system for pluggable cache implementations

**Tasks**:
1. Create `pkg/cache/registry.go` with registry structure:
   ```go
   type Factory func(c *config.CacheConfig) (Cache, error)

   type registry struct {
       factories map[string]Factory
       mu        sync.RWMutex
   }

   var register = &registry{
       factories: make(map[string]Factory),
   }
   ```
2. Implement `Register(name string, factory Factory)`:
   - Thread-safe write lock
   - Store factory in map
3. Implement `Create(c *config.CacheConfig) (Cache, error)`:
   - Thread-safe read lock
   - Lookup factory by name
   - Return error if not found: `"unknown cache type: %s"`
   - Call factory function with config
4. Implement `ListCaches() []string`:
   - Thread-safe read lock
   - Return sorted list of registered names
5. Create `tests/cache/registry_test.go`:
   - Test registration of mock cache factory
   - Test `Create()` with valid cache name
   - Test `Create()` with invalid cache name (error case)
   - Test `ListCaches()` returns all registered names
   - Test thread safety with concurrent registration and lookup

**Deliverables**:
- ✅ Registry infrastructure complete
- ✅ Thread-safe registration and lookup
- ✅ Clear error messages for unknown cache types
- ✅ Tests with 80%+ coverage including concurrency tests

**Estimated Effort**: 1-2 hours

---

#### Session 4: Filesystem Cache Implementation

**Goal**: Implement filesystem cache with validation, logging, and persistence

**Tasks**:
1. Create `pkg/cache/filesystem.go` with `FilesystemCache` struct:
   ```go
   type FilesystemCache struct {
       directory string
       logger    Logger
   }
   ```
2. Implement `NewFilesystem(c *config.CacheConfig) (Cache, error)` factory:
   - Extract `directory` from `c.Options["directory"]`
   - Validate required: return error if missing or empty
   - Normalize path with `filepath.Abs()`
   - Create directory if missing with 0755 permissions
   - Validate directory is writable (attempt to create temp file)
   - Extract optional `logger` from `c.Options["logger"]`, default to NoOpLogger
   - Return fully-initialized FilesystemCache
3. Implement `Get(key string) ([]byte, bool)`:
   - Detect file extension from existing cache files (`.png`, `.jpg`)
   - Read file with `os.ReadFile()`
   - Log debug: `"cache.get"`, key, found status
   - Return (data, true) on success, (nil, false) if not found
   - Distinguish "not found" from actual errors (only return false for ENOENT)
4. Implement `Set(key string, data []byte) error`:
   - Detect image format from data header (PNG: `0x89 'P'`, JPEG: `0xFF 0xD8`)
   - Determine file extension (`.png` or `.jpg`)
   - Write file atomically with `os.WriteFile(filepath.Join(directory, key+ext), data, 0644)`
   - Log debug: `"cache.set"`, key, size
   - Return error on failure with context
5. Implement `Invalidate(key string) error`:
   - Check for both `.png` and `.jpg` extensions
   - Remove file with `os.Remove()`
   - Log debug: `"cache.invalidate"`, key
   - Return nil if file doesn't exist (already invalidated)
6. Implement `Clear() error`:
   - Read directory with `os.ReadDir()`
   - Remove only `.cache`, `.png`, `.jpg` files (safety: don't remove unknown files)
   - Log info: `"cache.clear"`, file count
   - Return error on failure
7. Update `pkg/cache/registry.go` init():
   ```go
   func init() {
       Register("filesystem", NewFilesystem)
   }
   ```
8. Create `tests/cache/filesystem_test.go`:
   - Test factory validation (missing directory, invalid path)
   - Test directory auto-creation
   - Test Get/Set/Invalidate/Clear operations with temp directories
   - Test error conditions (permission denied simulations)
   - Test concurrent access (multiple goroutines)
   - Test logging (use mock logger to verify log calls)

**Deliverables**:
- ✅ `FilesystemCache` implementation complete
- ✅ Registered in cache registry
- ✅ Atomic file operations with proper error handling
- ✅ Logging integrated at appropriate levels
- ✅ Thread-safe concurrent access
- ✅ Tests with 80%+ coverage
- ✅ Clear error messages with context

**Estimated Effort**: 3-4 hours

---

#### Session 5: ImageMagick Filter Integration

**Goal**: Extend PDF image generation to support filter parameters via ImageMagick

**Tasks**:
1. Update `PDFPage.ToImage()` in `pkg/document/pdf.go` to use filter fields from `config.ImageConfig`
2. Create helper function `buildImageMagickArgs(page int, opts config.ImageConfig, outputPath string) []string`:
   - Base arguments: `"-density", strconv.Itoa(dpi), inputPath + "[" + page + "]", "-background", "white", "-flatten"`
   - Add rotation if `opts.Rotation != nil`: `"-rotate", strconv.Itoa(*opts.Rotation)`
   - Add brightness/saturation if either is non-nil:
     - Convert: Brightness (-100 to +100) → `brightness = *opts.Brightness + 100` (0 to 200)
     - Convert: Saturation (-100 to +100) → `saturation = *opts.Saturation + 100` (0 to 200)
     - Add: `"-modulate", fmt.Sprintf("%d,%d", brightness, saturation)`
   - Add contrast if `opts.Contrast != nil`: `"-brightness-contrast", fmt.Sprintf("%dx0", *opts.Contrast)`
   - Add output path
   - Return complete argument slice
3. Update `PDFPage.ToImage()` to call `buildImageMagickArgs()` and execute command
4. Add validation safety checks (optional, for robustness):
   - Clamp filter values to valid ranges if out of bounds
   - Log warnings for out-of-range values
5. Update `tests/document/pdf_test.go`:
   - Test image generation with each filter individually (brightness only, contrast only, etc.)
   - Test image generation with all filters combined
   - Test with nil filter values (default ImageMagick behavior)
   - Validate generated images (magic bytes correct, file size reasonable)
   - Test filter value ranges (ensure ImageMagick accepts converted values)
6. Add code comments documenting:
   - Filter parameter ranges and meanings
   - ImageMagick command structure and argument order
   - Parameter conversion formulas

**Deliverables**:
- ✅ Filter support integrated in `PDFPage.ToImage()`
- ✅ Helper function for command construction
- ✅ Tests validating filter application
- ✅ Documentation of filter parameters and ImageMagick mappings
- ✅ Code comments explaining conversion logic

**Estimated Effort**: 2-3 hours

---

#### Session 6: Cache Integration with Document Processing

**Goal**: Integrate cache into PDF document operations with transparent caching

**Tasks**:
1. Add `cache` field to `PDFDocument` struct in `pkg/document/pdf.go`:
   ```go
   type PDFDocument struct {
       path  string
       doc   *pdfcpu.PDFDocument
       cache cache.Cache  // Optional, may be nil
   }
   ```
2. Update `OpenPDF()` signature to accept optional cache:
   ```go
   func OpenPDF(path string, cache cache.Cache) (*PDFDocument, error)
   ```
   - Store cache in struct (may be nil for no caching)
3. Add parent document reference to `PDFPage`:
   ```go
   type PDFPage struct {
       doc    *PDFDocument  // Parent document for cache access
       number int
   }
   ```
4. Update `ExtractPage()` and `ExtractAllPages()` to set parent reference
5. Update `PDFPage.ToImage()` to integrate caching:
   - Import `pkg/cache` package
   - Before generation: check cache if available
     ```go
     if p.doc.cache != nil {
         key := cache.GenerateKey(p.doc.path, p.number, opts)
         if data, ok := p.doc.cache.Get(key); ok {
             return data, nil  // Cache hit
         }
     }
     ```
   - Generate image (existing logic)
   - After generation: store in cache if available
     ```go
     if p.doc.cache != nil {
         key := cache.GenerateKey(p.doc.path, p.number, opts)
         _ = p.doc.cache.Set(key, imageData)  // Ignore cache errors
     }
     ```
6. Error handling strategy:
   - Cache read failures (except ENOENT) should log warning but not fail generation
   - Cache write failures should log warning but return generated image
   - Always prioritize returning image data over cache operations
7. Update `tests/document/pdf_test.go`:
   - Test with cache enabled (verify cache hits after first generation)
   - Test with cache disabled (nil cache, existing behavior)
   - Test cache persistence across multiple `ToImage()` calls
   - Test with different `ImageConfig` values produce different cache keys
   - Test graceful degradation (cache failures don't break image generation)
8. Update integration example in code comments

**Deliverables**:
- ✅ Cache integration in `PDFDocument` and `PDFPage`
- ✅ Transparent caching with cache-miss fallback
- ✅ Graceful cache error handling (failures don't break generation)
- ✅ Tests validating cache behavior
- ✅ Updated code examples showing cache usage

**Estimated Effort**: 2-3 hours

---

#### Session 7: Examples and Documentation

**Goal**: Create example programs demonstrating library features and comprehensive documentation

**Tasks**:
1. Create `examples/` directory structure:
   ```
   examples/
   ├── README.md                          # Overview of all examples
   ├── 01-basic-conversion/
   │   ├── main.go
   │   ├── README.md
   │   └── sample.pdf                     # Small sample PDF
   ├── 02-filter-usage/
   │   ├── main.go
   │   ├── README.md
   │   └── sample.pdf
   └── 03-caching/
       ├── main.go
       ├── README.md
       ├── config.json                    # Example cache config
       └── sample.pdf
   ```
2. Implement `01-basic-conversion/main.go`:
   - Simple PDF to image conversion
   - Demonstrates basic `OpenPDF()` and `ToImage()` usage
   - Shows both PNG and JPEG output
   - Uses slog JSON handler for observability
   - Numbered output sections with progress indicators (✓)
3. Implement `02-filter-usage/main.go`:
   - Demonstrates filter parameters
   - Shows before/after with different filter combinations
   - Saves images to disk for visual comparison
   - Documents filter parameter effects
4. Implement `03-caching/main.go`:
   - Demonstrates cache configuration and usage
   - Loads `CacheConfig` from JSON file
   - Creates cache from registry
   - Processes same PDF twice, showing cache hit on second pass
   - Includes timing metrics to show cache benefits
5. Write comprehensive README.md for each example:
   - Purpose and scenario
   - Prerequisites (Go version, ImageMagick)
   - Running instructions (`go run main.go`)
   - Expected output (verbatim sample)
   - Configuration options and customization
   - Key concepts demonstrated
6. Create `examples/README.md`:
   - Overview of all examples
   - Progressive complexity explanation
   - Links to individual example READMEs
7. Update root `README.md`:
   - Add prerequisites section (Go 1.25.4+, ImageMagick 7.0+)
   - Add cache configuration examples
   - Add filter usage examples with code snippets
   - Update installation/usage sections
   - Link to examples directory
8. Update `ARCHITECTURE.md`:
   - Document cache architecture (interface, registry, implementations)
   - Document `ImageConfig` structure including filter parameters
   - Document cache key generation algorithm
   - Add logging integration details
   - Update package structure diagram
9. Update `PROJECT.md`:
   - Mark Phase 2 checklist items complete (✅)
   - Document architectural decisions made during implementation
   - Update version 0.1.0 goals progress

**Deliverables**:
- ✅ Three complete, runnable examples with READMEs
- ✅ Examples follow go-agents-orchestration patterns
- ✅ Updated README.md with cache and filter documentation
- ✅ Updated ARCHITECTURE.md with Phase 2 details
- ✅ Updated PROJECT.md with completion status

**Estimated Effort**: 3-4 hours

---

#### Session 8: Integration Testing and Validation

**Goal**: End-to-end validation, performance verification, and v0.1.0 readiness confirmation

**Tasks**:
1. Create `tests/integration/` directory for integration tests:
   - `document_cache_test.go` - End-to-end cache integration
   - `concurrent_test.go` - Concurrency and thread-safety validation
2. Implement `document_cache_test.go`:
   - Load `CacheConfig` from JSON
   - Create cache from registry
   - Process PDF with cache enabled
   - Verify cache hits on repeated operations
   - Test with different filter configurations
   - Verify cached images are identical to generated images (byte-for-byte comparison)
3. Implement `concurrent_test.go`:
   - Multiple goroutines processing same document concurrently
   - Verify thread-safe cache access (no race conditions)
   - Verify correct image generation under concurrent load
   - Test cache under concurrent writes to same keys
   - Use `testing.T.Parallel()` for parallel execution
4. Create `tests/document/pdf_benchmark_test.go`:
   - Benchmark `ToImage()` without cache (baseline)
   - Benchmark `ToImage()` with cache miss (includes cache write overhead)
   - Benchmark `ToImage()` with cache hit (should be significantly faster)
   - Measure cache overhead (miss vs no cache)
   - Document expected performance characteristics
5. Run comprehensive test validation:
   - `go test ./...` - all tests pass
   - `go test -race ./...` - no race conditions detected
   - `go test -cover ./...` - verify 80%+ coverage across all packages
   - `go test -bench=. ./...` - run benchmarks, document results
6. Validate Phase 2 success criteria:
   - ✅ Image caching reduces redundant conversions (measurable via benchmarks)
   - ✅ Enhancement filters enable document clarity optimization
   - ✅ JSON configuration enables web service integration
   - ✅ Thread-safe concurrent operations validated under load
   - ✅ 80%+ test coverage maintained across all packages
   - ✅ Ready for v0.1.0 pre-release
7. Create performance summary document (optional):
   - Benchmark results
   - Cache effectiveness metrics
   - Memory usage characteristics
   - Recommendations for production deployment
8. Final verification checklist:
   - All examples run successfully
   - All tests pass (unit, integration, concurrency)
   - Documentation is complete and accurate
   - Code follows CLAUDE.md design principles
   - Ready for version tag

**Deliverables**:
- ✅ Integration test suite complete
- ✅ Concurrency tests passing with `-race` flag
- ✅ Performance benchmarks showing cache effectiveness
- ✅ All Phase 2 success criteria validated
- ✅ 80%+ test coverage confirmed
- ✅ Project ready for v0.1.0 tag

**Estimated Effort**: 3-4 hours

---

### Success Criteria

Phase 2 is complete when all of the following criteria are met:

- ✅ **Image caching reduces redundant conversions**: Benchmarks demonstrate measurable performance improvement for cache hits vs generation
- ✅ **Enhancement filters enable document clarity optimization**: All filter parameters (brightness, contrast, saturation, rotation) work correctly via ImageMagick
- ✅ **JSON configuration enables web service integration**: `CacheConfig` and `ImageConfig` fully support JSON marshaling with validation
- ✅ **Thread-safe concurrent operations validated**: Concurrency tests pass with `-race` flag, demonstrating safe concurrent cache access
- ✅ **80%+ test coverage maintained**: All packages (config, cache, document) have comprehensive test coverage
- ✅ **Examples demonstrate library usage**: Three complete examples show basic conversion, filters, and caching
- ✅ **Documentation is comprehensive**: README, ARCHITECTURE, and PROJECT docs fully updated
- ✅ **Ready for v0.1.0 pre-release**: All features implemented, tested, and documented

### Estimated Total Effort

- **Session 1**: 2-3 hours (configuration foundation)
- **Session 2**: 2-3 hours (cache interface, logging, key generation)
- **Session 3**: 1-2 hours (registry infrastructure)
- **Session 4**: 3-4 hours (filesystem cache implementation)
- **Session 5**: 2-3 hours (ImageMagick filter integration)
- **Session 6**: 2-3 hours (cache integration with documents)
- **Session 7**: 3-4 hours (examples and documentation)
- **Session 8**: 3-4 hours (integration testing and validation)

**Total**: ~18-26 hours of focused development time

### Session Dependencies

```
Session 1 (Config Foundation)
    ↓
Session 2 (Cache Interface & Key Generation)
    ↓
Session 3 (Cache Registry)
    ↓
Session 4 (Filesystem Cache) ←──┐
    ↓                           │
Session 5 (Filter Integration) ─┤ (Sessions 5 & 6 could be parallel)
    ↓                           │
Session 6 (Cache Integration) ←─┘
    ↓
Session 7 (Examples & Documentation)
    ↓
Session 8 (Integration Testing)
```

Sessions 5 and 6 could potentially be executed in parallel since Session 5 extends image generation (filters) while Session 6 adds cache integration. However, executing them sequentially allows for easier testing and validation.

## v1.0.0+ Future Enhancements

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

**Streaming Support**:
- Streaming support for large file sets
- Memory-efficient processing for batch operations
- Progressive document loading

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

The library is currently in pre-release development and has not yet reached v0.1.0. The API is subject to change as Phase 2 features are added and refined.

**Development Focus (Phase 2)**:
- Image caching infrastructure (LRU + filesystem)
- Image enhancement filters (ImageMagick parameters)
- JSON configuration marshaling and validation
- Thread-safe concurrent operations
- Web service readiness for agent-lab integration

**Post-v0.1.0 Development**:
- Office document format support (v1.0.0+)
- Text extraction alternatives (v1.0.0+)
- Additional output formats (v1.0.0+)

### Version 0.1.0 Goals

The first versioned release will include:
- ✅ PDF support (complete)
- ✅ Image encoding (complete)
- ✅ Configuration infrastructure (pkg/config complete with logger and cache configs)
- ✅ Logger infrastructure (Session 2 complete - interface, implementation, configuration)
- ✅ Cache infrastructure - interfaces and abstractions (Session 2 complete)
- ⬜ Cache infrastructure - registry and implementations (Sessions 3-4 pending)
- ⬜ Image enhancement filters (config ready Session 1, application Session 5)
- ✅ JSON configuration marshaling and validation (config structures complete)
- ✅ Cache-aware rendering operations (Session 2 complete - ToImage with cache parameter)
- ⬜ Thread-safe concurrent request handling (Sessions 6-8)
- ⬜ Comprehensive documentation (ARCHITECTURE.md updated, examples pending Session 7)
- ✅ 80%+ test coverage (maintained with comprehensive Session 2 tests)
- ⬜ Agent-lab integration validation (pending later sessions)

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
