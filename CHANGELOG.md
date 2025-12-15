# Changelog

## [v0.1.1] - 2025-12-15

**Added**:

- `pkg/document` - Image format parsing and document format registry

  Adds `ParseImageFormat()` function for parsing string input to `ImageFormat` type with case-insensitive matching ("png", "jpg", "jpeg") and empty string defaulting to PNG. Adds format registry with `Open()` for content-type-based document opening, `IsSupported()` for format validation, and `SupportedFormats()` for listing available content types. Currently supports "application/pdf" with extensible registry pattern.

**Changed**:

- Go version updated from 1.25.4 to 1.25.5

## [v0.1.0] - 2025-11-19

Initial pre-release.

**Added**:

- `pkg/config` - Configuration data structures for all library components

  Provides ImageConfig for rendering settings (format, DPI, quality, filter options), CacheConfig for cache backend configuration with pluggable Options map, and LoggerConfig for structured logging setup with levels and output destinations. All configuration types include default factory functions, Merge methods for layered configuration, and Finalize methods for applying defaults.

- `pkg/logger` - Structured logging infrastructure with interface-based abstraction

  Defines Logger interface for capturing operational events across all library components. Includes Slogger implementation using Go's standard log/slog package with support for text/JSON output formats, configurable log levels (debug, info, warn, error, disabled), and output destinations (stdout, stderr, discard). Zero-overhead disabled mode for production deployments.

- `pkg/cache` - Persistent image caching infrastructure with pluggable backends

  Provides Cache interface for storing rendered document images with deterministic SHA256-based key generation. Includes thread-safe registry pattern for pluggable cache implementations with factory-based creation. FilesystemCache implementation provides directory-per-key storage with corruption detection, CRUD operations, and structured logging integration. Supports concurrent access with filesystem atomicity guarantees.

- `pkg/image` - Image rendering domain objects with external binary integration

  Defines Renderer interface for converting document pages to images with Settings method exposing base configuration and Parameters method for cache key generation. ImageMagickRenderer implementation provides high-quality PDF page rendering via ImageMagick with support for PNG/JPEG output formats, configurable DPI and quality settings, and enhancement filters (brightness, contrast, saturation, rotation, background color). Follows Configuration Composition Pattern with ImageMagickConfig embedding base ImageConfig.

- `pkg/document` - Core document processing abstractions with format-agnostic interfaces

  Provides Document and Page interfaces enabling format-independent document processing with lazy page extraction and optional caching support. PDF implementation using pdfcpu for document loading and metadata extraction, with cache-aware ToImage method integrating persistent caching for rendered pages. Supports page-level operations with bounds checking, temporary file management, and deterministic cache key generation from document path, page number, and all rendering parameters.

- `pkg/encoding` - Output encoding utilities for LLM integration

  Provides base64 data URI encoding for image data with format-specific MIME type handling. Supports PNG and JPEG formats with efficient string builder construction for large images. Enables direct embedding in LLM vision API requests without external storage requirements.
