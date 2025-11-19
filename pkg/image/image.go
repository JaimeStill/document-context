// Package image provides interfaces and implementations for rendering documents to image formats.
//
// This package defines the Renderer interface which abstracts image rendering operations,
// allowing different rendering implementations (ImageMagick, etc.) to be used interchangeably.
//
// Renderers are created through transformation functions (e.g., NewImageMagickRenderer) that
// accept configuration and return interface types, hiding implementation details.
package image

import "github.com/JaimeStill/document-context/pkg/config"

// Renderer defines the interface for rendering document pages to image files.
//
// Implementations handle the conversion of document pages to various image formats
// with configurable rendering options. The interface provides two key operations:
//   - Render: converts a specific page to an image file
//   - FileExtension: returns the appropriate file extension for the output format
//
// Renderer instances are immutable once created and safe for concurrent use.
type Renderer interface {
	// Render converts the specified page of a document to an image file.
	//
	// Parameters:
	//   - inputPath: path to the source document
	//   - pageNum: page number to render (1-indexed)
	//   - outputPath: path where the rendered image should be written
	//
	// Returns an error if rendering fails or if the external rendering tool
	// is not available.
	Render(inputPath string, pageNum int, outputPath string) error

	// FileExtension returns the file extension for the rendered image format.
	//
	// The extension does not include a leading dot (e.g., "png" not ".png").
	FileExtension() string

	// Settings returns the renderer's immutable configuration.
	//
	// This method exposes the ImageConfig used to create the renderer, enabling
	// access to rendering parameters throughout the renderer's lifetime. This
	// follows the Type 2 Configuration Pattern (Immutable Runtime Settings).
	//
	// The returned configuration is used for operations like cache key generation
	// where the complete rendering parameters must be available to ensure cache
	// correctness.
	//
	// The configuration is immutable after renderer creation - it cannot be
	// changed through this interface.
	Settings() config.ImageConfig

	// Parameters returns implementation-specific rendering parameters for cache key generation.
	//
	// This method provides a deterministic string representation of all implementation-specific
	// settings that affect render output but are not part of the base ImageConfig. These
	// parameters complement Settings() in cache key generation to ensure different
	// configurations produce unique cache entries.
	//
	// Format: Returns a slice of "key=value" strings in consistent (typically alphabetical) order.
	//
	// Example (ImageMagick): ["background=white", "brightness=10", "contrast=-5"]
	//
	// The returned parameters must be deterministic - the same configuration must always
	// produce the same parameter list in the same order to ensure cache key consistency.
	Parameters() []string
}
