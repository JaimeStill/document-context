package image

import (
	"fmt"
	"os/exec"
	"strconv"

	"github.com/JaimeStill/document-context/pkg/config"
)

type imagemagickRenderer struct {
	settings config.ImageConfig
}

// NewImageMagickRenderer creates a new Renderer using ImageMagick for rendering.
//
// This transformation function validates the provided configuration and creates
// an immutable renderer instance. Validation is performed at this boundary,
// ensuring that invalid configurations are rejected before creating domain objects.
//
// Configuration is finalized (defaults applied) and then validated:
//   - Format must be "png" or "jpg"
//   - Quality must be 1-100 for JPEG format
//   - Brightness, Contrast, Saturation must be -100 to +100 if set
//   - Rotation must be 0 to 360 degrees if set
//
// The returned Renderer is safe for concurrent use and will use ImageMagick's
// 'magick' command for rendering operations.
//
// Returns an error if configuration validation fails or if ImageMagick is not installed.
func NewImageMagickRenderer(cfg config.ImageConfig) (Renderer, error) {
	cfg.Finalize()

	if cfg.Format != "png" && cfg.Format != "jpg" {
		return nil, fmt.Errorf("unsupported image format: %s (must be a 'png' or 'jpg')", cfg.Format)
	}

	if cfg.Format == "jpg" {
		if cfg.Quality < 1 || cfg.Quality > 100 {
			return nil, fmt.Errorf("JPEG quality must be 1-100, got %d", cfg.Quality)
		}
	}

	if cfg.Brightness != nil {
		if *cfg.Brightness < -100 || *cfg.Brightness > 100 {
			return nil, fmt.Errorf("brightness must be -100 to +100, got %d", *cfg.Brightness)
		}
	}

	if cfg.Contrast != nil {
		if *cfg.Contrast < -100 || *cfg.Contrast > 100 {
			return nil, fmt.Errorf("contrast must be -100 to +100, got %d", *cfg.Contrast)
		}
	}

	if cfg.Saturation != nil {
		if *cfg.Saturation < -100 || *cfg.Saturation > 100 {
			return nil, fmt.Errorf("saturation must be -100 to +100, got %d", *cfg.Saturation)
		}
	}

	if cfg.Rotation != nil {
		if *cfg.Rotation < 0 || *cfg.Rotation > 360 {
			return nil, fmt.Errorf("rotation must be 0 to 360 degrees, got %d", *cfg.Rotation)
		}
	}

	return &imagemagickRenderer{
		settings: cfg,
	}, nil
}

func (r *imagemagickRenderer) Render(inputPath string, pageNum int, outputPath string) error {
	pageIndex := pageNum - 1
	inputSpec := fmt.Sprintf("%s[%d]", inputPath, pageIndex)

	args := []string{
		"-density", strconv.Itoa(r.settings.DPI),
		inputSpec,
		"-background", "white",
		"-flatten",
	}

	if r.settings.Format == "jpg" {
		args = append(args, "-quality", strconv.Itoa(r.settings.Quality))
	}

	args = append(args, outputPath)

	cmd := exec.Command("magick", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("imagemagick failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}

func (r *imagemagickRenderer) FileExtension() string {
	return r.settings.Format
}

func (r *imagemagickRenderer) Settings() config.ImageConfig {
	return r.settings
}
