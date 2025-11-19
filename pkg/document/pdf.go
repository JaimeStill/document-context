package document

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/JaimeStill/document-context/pkg/cache"
	"github.com/JaimeStill/document-context/pkg/image"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

type PDFDocument struct {
	path      string
	ctx       *model.Context
	pageCount int
}

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

func (d *PDFDocument) PageCount() int {
	return d.pageCount
}

func (d *PDFDocument) ExtractPage(pageNum int) (Page, error) {
	if pageNum < 1 || pageNum > d.pageCount {
		return nil, fmt.Errorf("page %d out of range [1-%d]", pageNum, d.pageCount)
	}

	return &PDFPage{
		doc:    d,
		number: pageNum,
	}, nil
}

func (d *PDFDocument) ExtractAllPages() ([]Page, error) {
	pages := make([]Page, 0, d.pageCount)

	for i := 1; i <= d.pageCount; i++ {
		page, err := d.ExtractPage(i)
		if err != nil {
			return nil, fmt.Errorf("failed to extract page %d: %w", i, err)
		}
		pages = append(pages, page)
	}

	return pages, nil
}

func (d *PDFDocument) Close() error {
	d.ctx = nil
	return nil
}

type PDFPage struct {
	doc    *PDFDocument
	number int
}

func (p *PDFPage) Number() int {
	return p.number
}

// ToImage converts the PDF page to an image using the specified renderer.
//
// This method supports optional caching to avoid redundant conversions. If a cache
// is provided and contains the rendered image, it returns the cached data immediately.
// Otherwise, it renders the page and stores the result in the cache for future use.
//
// Parameters:
//   - renderer: Image renderer implementation (e.g., ImageMagickRenderer)
//   - c: Optional cache for storing rendered images. Pass nil to disable caching.
//
// The cache key is deterministically generated from the document path, page number,
// and all rendering settings (format, DPI, quality, filters). The same inputs always
// produce the same cache key, enabling reliable cache lookups.
//
// Caching behavior:
//   - Cache hit: Returns cached image data immediately (no rendering)
//   - Cache miss: Renders page, stores in cache, returns image data
//   - No cache (nil): Always renders page (original behavior)
//   - Cache errors: Non-ErrCacheEntryNotFound errors are propagated
//
// Returns the rendered image data as bytes, or an error if rendering fails.
func (p *PDFPage) ToImage(renderer image.Renderer, c cache.Cache) ([]byte, error) {
	if c != nil {
		key, err := p.buildCacheKey(renderer)
		if err != nil {
			return nil, err
		}

		entry, err := c.Get(key)
		if err == nil {
			return entry.Data, nil
		}
		if !errors.Is(err, cache.ErrCacheEntryNotFound) {
			return nil, err
		}
	}

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

	if c != nil {
		entry, err := p.prepareCache(imgData, renderer)
		if err != nil {
			return nil, err
		}

		if err := c.Set(entry); err != nil {
			return nil, err
		}
	}

	return imgData, nil
}

// buildCacheKey generates a deterministic cache key from page and rendering settings.
//
// The cache key uniquely identifies a rendered page based on:
//   - Document path (normalized to absolute path)
//   - Page number
//   - Image format (png, jpg)
//   - All rendering parameters (DPI, quality, brightness, contrast, saturation, rotation)
//
// Key format (before hashing):
//
//	/absolute/path/to/document.pdf/1.png?dpi=300&quality=90&brightness=10
//
// Parameters are included in deterministic order:
//  1. Mandatory fields (alphabetically): dpi, quality
//  2. Optional fields present (alphabetically): brightness, contrast, rotation, saturation
//
// The formatted string is then hashed with SHA256 to produce a 64-character
// hexadecimal key. The same inputs always produce the same key.
//
// Returns an error if the document path cannot be normalized to an absolute path.
func (p *PDFPage) buildCacheKey(renderer image.Renderer) (string, error) {
	absPath, err := filepath.Abs(p.doc.path)
	if err != nil {
		return "", fmt.Errorf("failed to normalize path: %w", err)
	}

	settings := renderer.Settings()

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("%s/%d.%s", absPath, p.number, settings.Format))

	params := []string{
		fmt.Sprintf("dpi=%d", settings.DPI),
		fmt.Sprintf("quality=%d", settings.Quality),
	}

	params = append(params, renderer.Parameters()...)

	builder.WriteString(fmt.Sprintf("?%s", strings.Join(params, "&")))

	key := cache.GenerateKey(builder.String())
	return key, nil
}

// prepareCache constructs a cache entry from rendered image data and settings.
//
// This method creates a complete cache entry with:
//   - Key: Generated from document path, page number, and rendering settings
//   - Data: The rendered image bytes
//   - Filename: Suggested filename in format "basename.pagenum.ext"
//
// Filename construction:
//   - Extracts base filename from document path (without extension)
//   - Appends page number and output format extension
//   - Example: "document.pdf" page 1 as PNG → "document.1.png"
//
// Returns a complete cache entry ready for storage, or an error if the cache
// key cannot be generated.
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
