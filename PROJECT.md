# Project Scope and Roadmap

## Overview

**document-context** is a Go library for converting documents into context-friendly formats suitable for LLM consumption and analysis. The library provides format-agnostic interfaces for document processing with extensible format support.

This project was created as a tooling extension for the [go-agents](https://github.com/JaimeStill/go-agents) project, but is not directly dependent on go-agents and can be used as a standalone library for document processing needs.

## Current Status

**Phase**: Pre-Release Development - Phase 2 Session 5 Complete

Phase 2 Session 5 (ImageMagick Filter Integration) completed with comprehensive test coverage. The library now supports configuration foundation, cache/logging infrastructure, cache registry, filesystem cache implementation, and ImageMagick filter integration (brightness, contrast, saturation, rotation, background). All configuration transformation patterns (Type 1, Type 2, and Enhanced Composition) are implemented and operational.

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

Phase 2 extends library functionality with persistent caching, structured logging, and image enhancement filters.

#### Session 1: Configuration Foundation ✅

**Goal**: Establish `pkg/config/` with ImageConfig and CacheConfig structures

**Deliverables**:
- ✅ Configuration package with ImageConfig (filter fields, JSON support) and CacheConfig
- ✅ DefaultImageConfig() and Merge() methods
- ✅ Tests with 94.3% coverage

---

#### Session 2: Cache and Logging Infrastructure ✅

**Goal**: Establish structured logging and persistent image caching interfaces

**Deliverables**:
- ✅ Logger package with Type 1 Configuration Pattern (LoggerConfig → Logger interface)
- ✅ Cache package with Cache interface, CacheEntry, and SHA256 key generation
- ✅ Type 2 Configuration Pattern (Settings() method) added to Renderer interface
- ✅ Cache-aware PDFPage.ToImage() with optional cache parameter
- ✅ Comprehensive tests with 23+ test functions

---

#### Session 3: Cache Registry Infrastructure ✅

**Goal**: Implement thread-safe registry for pluggable cache backends

**Deliverables**:
- ✅ Registry with Factory pattern, Register(), Create(), ListCaches()
- ✅ Thread-safe operations with sync.RWMutex
- ✅ Tests with 87.1% coverage including concurrency validation

---

#### Session 4: Filesystem Cache Implementation ✅

**Goal**: Implement filesystem cache with directory-per-key storage and logging

**Deliverables**:
- ✅ FilesystemCache with Configuration Composition Pattern (parseFilesystemConfig)
- ✅ CRUD operations (Get/Set/Invalidate/Clear) with corruption detection
- ✅ Auto-registration in cache registry as "filesystem"
- ✅ Logger integration with configurable output (LoggerOutput enum)
- ✅ Tests with 87.1% coverage including concurrency validation

---

#### Session 5: ImageMagick Filter Integration ✅

**Goal**: Extend PDF image rendering to support enhancement filters via ImageMagick

**Deliverables**:
- ✅ Enhanced Composition Pattern with ImageMagickConfig embedding ImageConfig
- ✅ Filter application (brightness, contrast, saturation, rotation, background)
- ✅ Parameters() method for filter-specific cache keys
- ✅ Tests validating filter application and parameter conversion

---

#### Session 6: Cache Integration with Document Processing

**Goal**: Integrate cache into PDF operations with transparent caching behavior

**Deliverables**:
- Cache-aware OpenPDF() and PDFPage.ToImage() with optional cache parameter
- Transparent cache checking and storage (graceful degradation on cache errors)
- Tests validating cache hits, misses, and error handling

---

#### Session 7: Examples and Documentation

**Goal**: Create example programs demonstrating library features and comprehensive documentation

**Deliverables**:
- Three progressive examples (basic conversion, filter usage, caching) with comprehensive READMEs
- Updated README.md with cache and filter documentation
- Updated ARCHITECTURE.md documenting Phase 2 architecture
- Updated PROJECT.md with completion status

---

#### Session 8: Integration Testing and Validation

**Goal**: End-to-end validation and v0.1.0 readiness confirmation

**Deliverables**:
- Integration tests (end-to-end cache integration, concurrency validation)
- Performance benchmarks demonstrating cache effectiveness
- Concurrency tests passing with `-race` flag
- 80%+ test coverage confirmed across all packages

---

## v1.0.0+ Future Enhancements

### Additional Document Formats

**Office Documents**: .docx, .xlsx, .pptx via OpenXML parsing with structured text extraction and optional rendering

**Images**: .png, .jpg, .tiff via Tesseract OCR for text extraction with confidence scores

### Output Format Flexibility

**Planned Outputs**:
- Raw text extraction for token efficiency and searchability
- Structured text preserving formatting and hierarchy for section-aware analysis
- Image rendering (current) for visual layout preservation

### Processing Pipeline Enhancements

**Chunking**: Intelligent document splitting with boundary detection and overlapping contexts for token budget management

**Streaming**: Memory-efficient progressive loading for large file sets and batch operations

**Error Handling**: Graceful degradation for corrupt documents with page-level isolation and detailed recovery suggestions

## Extension Points

The library provides extensibility through interface implementation:

**New Document Formats**: Implement `Document` and `Page` interfaces (see ARCHITECTURE.md)

**New Output Formats**: Extend configuration structures and add encoding functions in `pkg/encoding/`

## Publishing and Versioning

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
