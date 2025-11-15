package image_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/JaimeStill/document-context/pkg/config"
	"github.com/JaimeStill/document-context/pkg/image"
)

func intPtr(i int) *int {
	return &i
}

func TestNewImageMagickRenderer_ValidConfig(t *testing.T) {
	tests := []struct {
		name   string
		config config.ImageConfig
	}{
		{
			name: "valid PNG config",
			config: config.ImageConfig{
				Format: "png",
				DPI:    150,
			},
		},
		{
			name: "valid JPEG config",
			config: config.ImageConfig{
				Format:  "jpg",
				Quality: 85,
				DPI:     200,
			},
		},
		{
			name: "empty config gets defaults",
			config: config.ImageConfig{},
		},
		{
			name: "config with valid filters",
			config: config.ImageConfig{
				Format:     "png",
				Brightness: intPtr(10),
				Contrast:   intPtr(-5),
				Saturation: intPtr(15),
				Rotation:   intPtr(90),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			renderer, err := image.NewImageMagickRenderer(tt.config)
			if err != nil {
				t.Fatalf("NewImageMagickRenderer failed: %v", err)
			}

			if renderer == nil {
				t.Error("expected non-nil renderer")
			}
		})
	}
}

func TestNewImageMagickRenderer_InvalidFormat(t *testing.T) {
	tests := []struct {
		name   string
		config config.ImageConfig
	}{
		{
			name: "invalid format webp",
			config: config.ImageConfig{
				Format: "webp",
			},
		},
		{
			name: "invalid format tiff",
			config: config.ImageConfig{
				Format: "tiff",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := image.NewImageMagickRenderer(tt.config)
			if err == nil {
				t.Error("expected error for invalid format")
			}
		})
	}
}

func TestNewImageMagickRenderer_InvalidQuality(t *testing.T) {
	tests := []struct {
		name    string
		quality int
	}{
		{
			name:    "quality too low",
			quality: 0,
		},
		{
			name:    "quality negative",
			quality: -1,
		},
		{
			name:    "quality too high",
			quality: 101,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.ImageConfig{
				Format:  "jpg",
				Quality: tt.quality,
			}

			_, err := image.NewImageMagickRenderer(cfg)
			if err == nil {
				t.Error("expected error for invalid quality")
			}
		})
	}
}

func TestNewImageMagickRenderer_InvalidFilters(t *testing.T) {
	tests := []struct {
		name   string
		config config.ImageConfig
		errMsg string
	}{
		{
			name: "brightness too low",
			config: config.ImageConfig{
				Format:     "png",
				Brightness: intPtr(-101),
			},
			errMsg: "brightness",
		},
		{
			name: "brightness too high",
			config: config.ImageConfig{
				Format:     "png",
				Brightness: intPtr(101),
			},
			errMsg: "brightness",
		},
		{
			name: "contrast too low",
			config: config.ImageConfig{
				Format:   "png",
				Contrast: intPtr(-101),
			},
			errMsg: "contrast",
		},
		{
			name: "contrast too high",
			config: config.ImageConfig{
				Format:   "png",
				Contrast: intPtr(101),
			},
			errMsg: "contrast",
		},
		{
			name: "saturation too low",
			config: config.ImageConfig{
				Format:     "png",
				Saturation: intPtr(-101),
			},
			errMsg: "saturation",
		},
		{
			name: "saturation too high",
			config: config.ImageConfig{
				Format:     "png",
				Saturation: intPtr(101),
			},
			errMsg: "saturation",
		},
		{
			name: "rotation negative",
			config: config.ImageConfig{
				Format:   "png",
				Rotation: intPtr(-1),
			},
			errMsg: "rotation",
		},
		{
			name: "rotation too high",
			config: config.ImageConfig{
				Format:   "png",
				Rotation: intPtr(361),
			},
			errMsg: "rotation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := image.NewImageMagickRenderer(tt.config)
			if err == nil {
				t.Error("expected error for invalid filter value")
			}
		})
	}
}

func TestNewImageMagickRenderer_Finalize(t *testing.T) {
	// Empty config should get defaults via Finalize
	cfg := config.ImageConfig{}

	renderer, err := image.NewImageMagickRenderer(cfg)
	if err != nil {
		t.Fatalf("NewImageMagickRenderer failed: %v", err)
	}

	// Should have default format
	ext := renderer.FileExtension()
	if ext != "png" {
		t.Errorf("expected default format 'png', got %q", ext)
	}
}

func TestRenderer_FileExtension(t *testing.T) {
	tests := []struct {
		name     string
		config   config.ImageConfig
		expected string
	}{
		{
			name: "PNG extension",
			config: config.ImageConfig{
				Format: "png",
			},
			expected: "png",
		},
		{
			name: "JPEG extension",
			config: config.ImageConfig{
				Format:  "jpg",
				Quality: 85,
			},
			expected: "jpg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			renderer, err := image.NewImageMagickRenderer(tt.config)
			if err != nil {
				t.Fatalf("NewImageMagickRenderer failed: %v", err)
			}

			ext := renderer.FileExtension()
			if ext != tt.expected {
				t.Errorf("expected extension %q, got %q", tt.expected, ext)
			}
		})
	}
}

func TestRenderer_Interface(t *testing.T) {
	cfg := config.ImageConfig{Format: "png"}

	// This should compile - renderer satisfies image.Renderer interface
	var renderer image.Renderer
	renderer, err := image.NewImageMagickRenderer(cfg)
	if err != nil {
		t.Fatalf("NewImageMagickRenderer failed: %v", err)
	}

	if renderer == nil {
		t.Error("expected non-nil renderer")
	}

	// Should be able to call interface methods
	ext := renderer.FileExtension()
	if ext != "png" {
		t.Errorf("expected extension 'png', got %q", ext)
	}
}

func TestRenderer_BoundaryValues(t *testing.T) {
	tests := []struct {
		name   string
		config config.ImageConfig
	}{
		{
			name: "minimum valid brightness",
			config: config.ImageConfig{
				Format:     "png",
				Brightness: intPtr(-100),
			},
		},
		{
			name: "maximum valid brightness",
			config: config.ImageConfig{
				Format:     "png",
				Brightness: intPtr(100),
			},
		},
		{
			name: "minimum valid contrast",
			config: config.ImageConfig{
				Format:   "png",
				Contrast: intPtr(-100),
			},
		},
		{
			name: "maximum valid contrast",
			config: config.ImageConfig{
				Format:   "png",
				Contrast: intPtr(100),
			},
		},
		{
			name: "minimum valid rotation",
			config: config.ImageConfig{
				Format:   "png",
				Rotation: intPtr(0),
			},
		},
		{
			name: "maximum valid rotation",
			config: config.ImageConfig{
				Format:   "png",
				Rotation: intPtr(360),
			},
		},
		{
			name: "JPEG minimum quality",
			config: config.ImageConfig{
				Format:  "jpg",
				Quality: 1,
			},
		},
		{
			name: "JPEG maximum quality",
			config: config.ImageConfig{
				Format:  "jpg",
				Quality: 100,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			renderer, err := image.NewImageMagickRenderer(tt.config)
			if err != nil {
				t.Fatalf("NewImageMagickRenderer failed for valid boundary value: %v", err)
			}

			if renderer == nil {
				t.Error("expected non-nil renderer for valid boundary value")
			}
		})
	}
}

func TestRenderer_Render_Integration(t *testing.T) {
	// This test requires a real PDF file and ImageMagick
	pdfPath := filepath.Join("..", "document", "vim-cheatsheet.pdf")
	if _, err := os.Stat(pdfPath); os.IsNotExist(err) {
		t.Skip("Test PDF not found, skipping integration test")
	}

	cfg := config.ImageConfig{
		Format: "png",
		DPI:    150,
	}

	renderer, err := image.NewImageMagickRenderer(cfg)
	if err != nil {
		t.Fatalf("NewImageMagickRenderer failed: %v", err)
	}

	tmpFile, err := os.CreateTemp("", "test-render-*.png")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	err = renderer.Render(pdfPath, 1, tmpPath)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(tmpPath); os.IsNotExist(err) {
		t.Error("Expected output file to be created")
	}

	// Verify file has content
	info, err := os.Stat(tmpPath)
	if err != nil {
		t.Fatalf("Failed to stat output file: %v", err)
	}

	if info.Size() == 0 {
		t.Error("Expected non-empty output file")
	}
}
