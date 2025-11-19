package document_test

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"

	"github.com/JaimeStill/document-context/pkg/cache"
	"github.com/JaimeStill/document-context/pkg/config"
	"github.com/JaimeStill/document-context/pkg/document"
	"github.com/JaimeStill/document-context/pkg/image"
)

func testPDFPath(t *testing.T) string {
	t.Helper()

	path := filepath.Join("vim-cheatsheet.pdf")

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("Test PDF not found: %s", path)
	}

	return path
}

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

func TestOpenPDF(t *testing.T) {
	path := testPDFPath(t)

	doc, err := document.OpenPDF(path)
	if err != nil {
		t.Fatalf("OpenPDF failed: %v", err)
	}
	defer doc.Close()

	if doc.PageCount() == 0 {
		t.Error("Expected non-zero page count")
	}

	t.Logf("Successfully opened PDF with %d pages", doc.PageCount())
}

func TestOpenPDF_InvalidPath(t *testing.T) {
	_, err := document.OpenPDF("/nonexistent/file.pdf")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestPDFDocument_ExtractPage(t *testing.T) {
	path := testPDFPath(t)
	doc, err := document.OpenPDF(path)
	if err != nil {
		t.Fatalf("OpenPDF failed: %v", err)
	}
	defer doc.Close()

	page, err := doc.ExtractPage(1)
	if err != nil {
		t.Fatalf("ExtractPage failed: %v", err)
	}

	if page.Number() != 1 {
		t.Errorf("Expected page number 1, got %d", page.Number())
	}

	_, err = doc.ExtractPage(0)
	if err == nil {
		t.Error("Expected error for page 0")
	}

	_, err = doc.ExtractPage(doc.PageCount() + 1)
	if err == nil {
		t.Error("Expected error for page beyond document")
	}
}

func TestPDFDocument_ExtractAllPages(t *testing.T) {
	path := testPDFPath(t)
	doc, err := document.OpenPDF(path)
	if err != nil {
		t.Fatalf("OpenPDF failed: %v", err)
	}
	defer doc.Close()

	pages, err := doc.ExtractAllPages()
	if err != nil {
		t.Fatalf("ExtractAllPages failed: %v", err)
	}

	if len(pages) != doc.PageCount() {
		t.Errorf("Expected %d pages, got %d", doc.PageCount(), len(pages))
	}

	// Verify page numbers are sequential
	for i, page := range pages {
		expectedNum := i + 1
		if page.Number() != expectedNum {
			t.Errorf("Page %d has wrong number: %d", i, page.Number())
		}
	}
}

func TestPDFPage_ToImage_PNG(t *testing.T) {
	requireImageMagick(t)

	path := testPDFPath(t)
	doc, err := document.OpenPDF(path)
	if err != nil {
		t.Fatalf("OpenPDF failed: %v", err)
	}
	defer doc.Close()

	page, err := doc.ExtractPage(1)
	if err != nil {
		t.Fatalf("ExtractPage failed: %v", err)
	}

	cfg := config.ImageConfig{
		Format: "png",
		DPI:    150,
	}

	renderer, err := image.NewImageMagickRenderer(cfg)
	if err != nil {
		t.Fatalf("NewImageMagickRenderer failed: %v", err)
	}

	imgData, err := page.ToImage(renderer, nil)
	if err != nil {
		t.Fatalf("ToImage failed: %v", err)
	}

	if len(imgData) == 0 {
		t.Error("Expected non-empty image data")
	}

	// PNG files start with specific magic bytes
	if len(imgData) < 8 || imgData[0] != 0x89 || imgData[1] != 'P' {
		t.Error("Image data does not appear to be PNG format")
	}

	t.Logf("Generated PNG: %d bytes", len(imgData))
}

func TestPDFPage_ToImage_JPEG(t *testing.T) {
	requireImageMagick(t)

	path := testPDFPath(t)
	doc, err := document.OpenPDF(path)
	if err != nil {
		t.Fatalf("OpenPDF failed: %v", err)
	}
	defer doc.Close()

	page, err := doc.ExtractPage(1)
	if err != nil {
		t.Fatalf("ExtractPage failed: %v", err)
	}

	cfg := config.ImageConfig{
		Format:  "jpg",
		Quality: 85,
		DPI:     150,
	}

	renderer, err := image.NewImageMagickRenderer(cfg)
	if err != nil {
		t.Fatalf("NewImageMagickRenderer failed: %v", err)
	}

	imgData, err := page.ToImage(renderer, nil)
	if err != nil {
		t.Fatalf("ToImage failed: %v", err)
	}

	if len(imgData) == 0 {
		t.Error("Expected non-empty image data")
	}

	// JPEG files start with 0xFF 0xD8
	if len(imgData) < 2 || imgData[0] != 0xFF || imgData[1] != 0xD8 {
		t.Error("Image data does not appear to be JPEG format")
	}

	t.Logf("Generated JPEG: %d bytes", len(imgData))
}

func TestPDFPage_ToImage_DefaultOptions(t *testing.T) {
	requireImageMagick(t)

	path := testPDFPath(t)
	doc, err := document.OpenPDF(path)
	if err != nil {
		t.Fatalf("OpenPDF failed: %v", err)
	}
	defer doc.Close()

	page, err := doc.ExtractPage(1)
	if err != nil {
		t.Fatalf("ExtractPage failed: %v", err)
	}

	// Pass zero-value config to test Finalize() applying defaults
	cfg := config.ImageConfig{}

	renderer, err := image.NewImageMagickRenderer(cfg)
	if err != nil {
		t.Fatalf("NewImageMagickRenderer failed: %v", err)
	}

	imgData, err := page.ToImage(renderer, nil)
	if err != nil {
		t.Fatalf("ToImage with defaults failed: %v", err)
	}

	if len(imgData) == 0 {
		t.Error("Expected non-empty image data")
	}
}

// mockCache implements cache.Cache interface for testing cache integration.
type mockCache struct {
	mu      sync.RWMutex
	entries map[string]*cache.CacheEntry
	getErr  error
	setErr  error
}

func newMockCache() *mockCache {
	return &mockCache{
		entries: make(map[string]*cache.CacheEntry),
	}
}

func (m *mockCache) Get(key string) (*cache.CacheEntry, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.getErr != nil {
		return nil, m.getErr
	}

	entry, exists := m.entries[key]
	if !exists {
		return nil, cache.ErrCacheEntryNotFound
	}

	return entry, nil
}

func (m *mockCache) Set(entry *cache.CacheEntry) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.setErr != nil {
		return m.setErr
	}

	m.entries[entry.Key] = entry
	return nil
}

func (m *mockCache) Invalidate(key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.entries, key)
	return nil
}

func (m *mockCache) Clear() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.entries = make(map[string]*cache.CacheEntry)
	return nil
}

func (m *mockCache) setGetError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.getErr = err
}

func (m *mockCache) setSetError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.setErr = err
}

func (m *mockCache) hasKey(key string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, exists := m.entries[key]
	return exists
}

func (m *mockCache) entryCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.entries)
}

func TestPDFPage_ToImage_CacheMiss(t *testing.T) {
	requireImageMagick(t)

	path := testPDFPath(t)
	doc, err := document.OpenPDF(path)
	if err != nil {
		t.Fatalf("OpenPDF failed: %v", err)
	}
	defer doc.Close()

	page, err := doc.ExtractPage(1)
	if err != nil {
		t.Fatalf("ExtractPage failed: %v", err)
	}

	cfg := config.ImageConfig{
		Format: "png",
		DPI:    150,
	}

	renderer, err := image.NewImageMagickRenderer(cfg)
	if err != nil {
		t.Fatalf("NewImageMagickRenderer failed: %v", err)
	}

	mockCache := newMockCache()

	imgData, err := page.ToImage(renderer, mockCache)
	if err != nil {
		t.Fatalf("ToImage failed: %v", err)
	}

	if len(imgData) == 0 {
		t.Error("Expected non-empty image data")
	}

	if mockCache.entryCount() != 1 {
		t.Errorf("Expected 1 cache entry, got %d", mockCache.entryCount())
	}

	if len(imgData) < 8 || imgData[0] != 0x89 || imgData[1] != 'P' {
		t.Error("Image data does not appear to be PNG format")
	}

	t.Logf("Cache miss: rendered and cached %d bytes", len(imgData))
}

func TestPDFPage_ToImage_CacheHit(t *testing.T) {
	requireImageMagick(t)

	path := testPDFPath(t)
	doc, err := document.OpenPDF(path)
	if err != nil {
		t.Fatalf("OpenPDF failed: %v", err)
	}
	defer doc.Close()

	page, err := doc.ExtractPage(1)
	if err != nil {
		t.Fatalf("ExtractPage failed: %v", err)
	}

	cfg := config.ImageConfig{
		Format: "png",
		DPI:    150,
	}

	renderer, err := image.NewImageMagickRenderer(cfg)
	if err != nil {
		t.Fatalf("NewImageMagickRenderer failed: %v", err)
	}

	mockCache := newMockCache()

	imgData1, err := page.ToImage(renderer, mockCache)
	if err != nil {
		t.Fatalf("First ToImage failed: %v", err)
	}

	imgData2, err := page.ToImage(renderer, mockCache)
	if err != nil {
		t.Fatalf("Second ToImage failed: %v", err)
	}

	if len(imgData1) != len(imgData2) {
		t.Errorf("Expected same data length, got %d and %d", len(imgData1), len(imgData2))
	}

	for i := range imgData1 {
		if imgData1[i] != imgData2[i] {
			t.Errorf("Byte mismatch at position %d", i)
			break
		}
	}

	if mockCache.entryCount() != 1 {
		t.Errorf("Expected 1 cache entry, got %d", mockCache.entryCount())
	}

	t.Logf("Cache hit: returned %d bytes from cache", len(imgData2))
}

func TestPDFPage_ToImage_CacheKey_Deterministic(t *testing.T) {
	requireImageMagick(t)

	path := testPDFPath(t)
	doc, err := document.OpenPDF(path)
	if err != nil {
		t.Fatalf("OpenPDF failed: %v", err)
	}
	defer doc.Close()

	page, err := doc.ExtractPage(1)
	if err != nil {
		t.Fatalf("ExtractPage failed: %v", err)
	}

	cfg := config.ImageConfig{
		Format: "png",
		DPI:    150,
	}

	renderer1, err := image.NewImageMagickRenderer(cfg)
	if err != nil {
		t.Fatalf("NewImageMagickRenderer failed: %v", err)
	}

	renderer2, err := image.NewImageMagickRenderer(cfg)
	if err != nil {
		t.Fatalf("NewImageMagickRenderer failed: %v", err)
	}

	mockCache1 := newMockCache()
	mockCache2 := newMockCache()

	_, err = page.ToImage(renderer1, mockCache1)
	if err != nil {
		t.Fatalf("First ToImage failed: %v", err)
	}

	_, err = page.ToImage(renderer2, mockCache2)
	if err != nil {
		t.Fatalf("Second ToImage failed: %v", err)
	}

	if mockCache1.entryCount() != 1 || mockCache2.entryCount() != 1 {
		t.Fatal("Expected 1 entry in each cache")
	}

	var key1, key2 string
	for k := range mockCache1.entries {
		key1 = k
	}
	for k := range mockCache2.entries {
		key2 = k
	}

	if key1 != key2 {
		t.Errorf("Expected same cache key for same configuration\nKey1: %s\nKey2: %s", key1, key2)
	}

	t.Logf("Deterministic cache key: %s", key1)
}

func TestPDFPage_ToImage_CacheKey_DifferentPages(t *testing.T) {
	requireImageMagick(t)

	path := testPDFPath(t)
	doc, err := document.OpenPDF(path)
	if err != nil {
		t.Fatalf("OpenPDF failed: %v", err)
	}
	defer doc.Close()

	if doc.PageCount() < 2 {
		t.Skip("Test requires PDF with at least 2 pages")
	}

	page1, err := doc.ExtractPage(1)
	if err != nil {
		t.Fatalf("ExtractPage 1 failed: %v", err)
	}

	page2, err := doc.ExtractPage(2)
	if err != nil {
		t.Fatalf("ExtractPage 2 failed: %v", err)
	}

	cfg := config.ImageConfig{
		Format: "png",
		DPI:    150,
	}

	renderer, err := image.NewImageMagickRenderer(cfg)
	if err != nil {
		t.Fatalf("NewImageMagickRenderer failed: %v", err)
	}

	mockCache := newMockCache()

	_, err = page1.ToImage(renderer, mockCache)
	if err != nil {
		t.Fatalf("ToImage page 1 failed: %v", err)
	}

	_, err = page2.ToImage(renderer, mockCache)
	if err != nil {
		t.Fatalf("ToImage page 2 failed: %v", err)
	}

	if mockCache.entryCount() != 2 {
		t.Errorf("Expected 2 cache entries for different pages, got %d", mockCache.entryCount())
	}

	t.Log("Different pages produce different cache keys")
}

func TestPDFPage_ToImage_CacheKey_DifferentFormats(t *testing.T) {
	requireImageMagick(t)

	path := testPDFPath(t)
	doc, err := document.OpenPDF(path)
	if err != nil {
		t.Fatalf("OpenPDF failed: %v", err)
	}
	defer doc.Close()

	page, err := doc.ExtractPage(1)
	if err != nil {
		t.Fatalf("ExtractPage failed: %v", err)
	}

	pngCfg := config.ImageConfig{
		Format: "png",
		DPI:    150,
	}

	jpgCfg := config.ImageConfig{
		Format:  "jpg",
		Quality: 85,
		DPI:     150,
	}

	pngRenderer, err := image.NewImageMagickRenderer(pngCfg)
	if err != nil {
		t.Fatalf("PNG renderer creation failed: %v", err)
	}

	jpgRenderer, err := image.NewImageMagickRenderer(jpgCfg)
	if err != nil {
		t.Fatalf("JPEG renderer creation failed: %v", err)
	}

	mockCache := newMockCache()

	_, err = page.ToImage(pngRenderer, mockCache)
	if err != nil {
		t.Fatalf("PNG ToImage failed: %v", err)
	}

	_, err = page.ToImage(jpgRenderer, mockCache)
	if err != nil {
		t.Fatalf("JPEG ToImage failed: %v", err)
	}

	if mockCache.entryCount() != 2 {
		t.Errorf("Expected 2 cache entries for different formats, got %d", mockCache.entryCount())
	}

	t.Log("Different formats produce different cache keys")
}

func TestPDFPage_ToImage_CacheKey_DifferentDPI(t *testing.T) {
	requireImageMagick(t)

	path := testPDFPath(t)
	doc, err := document.OpenPDF(path)
	if err != nil {
		t.Fatalf("OpenPDF failed: %v", err)
	}
	defer doc.Close()

	page, err := doc.ExtractPage(1)
	if err != nil {
		t.Fatalf("ExtractPage failed: %v", err)
	}

	cfg150 := config.ImageConfig{
		Format: "png",
		DPI:    150,
	}

	cfg300 := config.ImageConfig{
		Format: "png",
		DPI:    300,
	}

	renderer150, err := image.NewImageMagickRenderer(cfg150)
	if err != nil {
		t.Fatalf("150 DPI renderer creation failed: %v", err)
	}

	renderer300, err := image.NewImageMagickRenderer(cfg300)
	if err != nil {
		t.Fatalf("300 DPI renderer creation failed: %v", err)
	}

	mockCache := newMockCache()

	_, err = page.ToImage(renderer150, mockCache)
	if err != nil {
		t.Fatalf("150 DPI ToImage failed: %v", err)
	}

	_, err = page.ToImage(renderer300, mockCache)
	if err != nil {
		t.Fatalf("300 DPI ToImage failed: %v", err)
	}

	if mockCache.entryCount() != 2 {
		t.Errorf("Expected 2 cache entries for different DPI, got %d", mockCache.entryCount())
	}

	t.Log("Different DPI values produce different cache keys")
}

func TestPDFPage_ToImage_CacheKey_WithFilters(t *testing.T) {
	requireImageMagick(t)

	path := testPDFPath(t)
	doc, err := document.OpenPDF(path)
	if err != nil {
		t.Fatalf("OpenPDF failed: %v", err)
	}
	defer doc.Close()

	page, err := doc.ExtractPage(1)
	if err != nil {
		t.Fatalf("ExtractPage failed: %v", err)
	}

	baseCfg := config.ImageConfig{
		Format: "png",
		DPI:    150,
	}

	filteredCfg := config.ImageConfig{
		Format: "png",
		DPI:    150,
		Options: map[string]any{
			"brightness": 110,
		},
	}

	baseRenderer, err := image.NewImageMagickRenderer(baseCfg)
	if err != nil {
		t.Fatalf("Base renderer creation failed: %v", err)
	}

	filteredRenderer, err := image.NewImageMagickRenderer(filteredCfg)
	if err != nil {
		t.Fatalf("Filtered renderer creation failed: %v", err)
	}

	mockCache := newMockCache()

	_, err = page.ToImage(baseRenderer, mockCache)
	if err != nil {
		t.Fatalf("Base ToImage failed: %v", err)
	}

	_, err = page.ToImage(filteredRenderer, mockCache)
	if err != nil {
		t.Fatalf("Filtered ToImage failed: %v", err)
	}

	if mockCache.entryCount() != 2 {
		t.Errorf("Expected 2 cache entries (with and without filters), got %d", mockCache.entryCount())
	}

	t.Log("Filters affect cache keys")
}

func TestPDFPage_ToImage_CacheKey_DifferentFilters(t *testing.T) {
	requireImageMagick(t)

	path := testPDFPath(t)
	doc, err := document.OpenPDF(path)
	if err != nil {
		t.Fatalf("OpenPDF failed: %v", err)
	}
	defer doc.Close()

	page, err := doc.ExtractPage(1)
	if err != nil {
		t.Fatalf("ExtractPage failed: %v", err)
	}

	cfg1 := config.ImageConfig{
		Format: "png",
		DPI:    150,
		Options: map[string]any{
			"brightness": 110,
		},
	}

	cfg2 := config.ImageConfig{
		Format: "png",
		DPI:    150,
		Options: map[string]any{
			"brightness": 120,
		},
	}

	renderer1, err := image.NewImageMagickRenderer(cfg1)
	if err != nil {
		t.Fatalf("Renderer 1 creation failed: %v", err)
	}

	renderer2, err := image.NewImageMagickRenderer(cfg2)
	if err != nil {
		t.Fatalf("Renderer 2 creation failed: %v", err)
	}

	mockCache := newMockCache()

	_, err = page.ToImage(renderer1, mockCache)
	if err != nil {
		t.Fatalf("ToImage 1 failed: %v", err)
	}

	_, err = page.ToImage(renderer2, mockCache)
	if err != nil {
		t.Fatalf("ToImage 2 failed: %v", err)
	}

	if mockCache.entryCount() != 2 {
		t.Errorf("Expected 2 cache entries for different filter values, got %d", mockCache.entryCount())
	}

	t.Log("Different filter values produce different cache keys")
}

func TestPDFPage_ToImage_CacheGetError(t *testing.T) {
	requireImageMagick(t)

	path := testPDFPath(t)
	doc, err := document.OpenPDF(path)
	if err != nil {
		t.Fatalf("OpenPDF failed: %v", err)
	}
	defer doc.Close()

	page, err := doc.ExtractPage(1)
	if err != nil {
		t.Fatalf("ExtractPage failed: %v", err)
	}

	cfg := config.ImageConfig{
		Format: "png",
		DPI:    150,
	}

	renderer, err := image.NewImageMagickRenderer(cfg)
	if err != nil {
		t.Fatalf("NewImageMagickRenderer failed: %v", err)
	}

	mockCache := newMockCache()
	testErr := errors.New("cache storage unavailable")
	mockCache.setGetError(testErr)

	_, err = page.ToImage(renderer, mockCache)
	if err == nil {
		t.Fatal("Expected error from cache.Get() to be propagated")
	}

	if !errors.Is(err, testErr) {
		t.Errorf("Expected error to wrap cache error, got: %v", err)
	}

	t.Log("Cache Get errors properly propagated")
}

func TestPDFPage_ToImage_CacheSetError(t *testing.T) {
	requireImageMagick(t)

	path := testPDFPath(t)
	doc, err := document.OpenPDF(path)
	if err != nil {
		t.Fatalf("OpenPDF failed: %v", err)
	}
	defer doc.Close()

	page, err := doc.ExtractPage(1)
	if err != nil {
		t.Fatalf("ExtractPage failed: %v", err)
	}

	cfg := config.ImageConfig{
		Format: "png",
		DPI:    150,
	}

	renderer, err := image.NewImageMagickRenderer(cfg)
	if err != nil {
		t.Fatalf("NewImageMagickRenderer failed: %v", err)
	}

	mockCache := newMockCache()
	testErr := errors.New("cache write failed")
	mockCache.setSetError(testErr)

	_, err = page.ToImage(renderer, mockCache)
	if err == nil {
		t.Fatal("Expected error from cache.Set() to be propagated")
	}

	if !errors.Is(err, testErr) {
		t.Errorf("Expected error to wrap cache error, got: %v", err)
	}

	t.Log("Cache Set errors properly propagated")
}

func TestPDFPage_ToImage_ConcurrentAccess(t *testing.T) {
	requireImageMagick(t)

	path := testPDFPath(t)
	doc, err := document.OpenPDF(path)
	if err != nil {
		t.Fatalf("OpenPDF failed: %v", err)
	}
	defer doc.Close()

	page, err := doc.ExtractPage(1)
	if err != nil {
		t.Fatalf("ExtractPage failed: %v", err)
	}

	cfg := config.ImageConfig{
		Format: "png",
		DPI:    150,
	}

	renderer, err := image.NewImageMagickRenderer(cfg)
	if err != nil {
		t.Fatalf("NewImageMagickRenderer failed: %v", err)
	}

	mockCache := newMockCache()

	const numGoroutines = 10
	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := page.ToImage(renderer, mockCache)
			if err != nil {
				errors <- err
			}
		}()
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Errorf("Concurrent ToImage failed: %v", err)
	}

	if mockCache.entryCount() != 1 {
		t.Errorf("Expected 1 cache entry, got %d", mockCache.entryCount())
	}

	t.Logf("Concurrent access test completed: %d goroutines, 1 cache entry", numGoroutines)
}

func TestPDFPage_ToImage_CacheEntry_Filename(t *testing.T) {
	requireImageMagick(t)

	path := testPDFPath(t)
	doc, err := document.OpenPDF(path)
	if err != nil {
		t.Fatalf("OpenPDF failed: %v", err)
	}
	defer doc.Close()

	page, err := doc.ExtractPage(1)
	if err != nil {
		t.Fatalf("ExtractPage failed: %v", err)
	}

	cfg := config.ImageConfig{
		Format: "png",
		DPI:    150,
	}

	renderer, err := image.NewImageMagickRenderer(cfg)
	if err != nil {
		t.Fatalf("NewImageMagickRenderer failed: %v", err)
	}

	mockCache := newMockCache()

	_, err = page.ToImage(renderer, mockCache)
	if err != nil {
		t.Fatalf("ToImage failed: %v", err)
	}

	if mockCache.entryCount() != 1 {
		t.Fatal("Expected 1 cache entry")
	}

	for _, entry := range mockCache.entries {
		expectedFilename := "vim-cheatsheet.1.png"
		if entry.Filename != expectedFilename {
			t.Errorf("Expected filename %s, got %s", expectedFilename, entry.Filename)
		}

		if entry.Key == "" {
			t.Error("Expected non-empty cache key")
		}

		if len(entry.Data) == 0 {
			t.Error("Expected non-empty image data")
		}

		t.Logf("Cache entry: key=%s, filename=%s, size=%d bytes", entry.Key, entry.Filename, len(entry.Data))
	}
}
