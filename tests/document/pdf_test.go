package document_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

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
