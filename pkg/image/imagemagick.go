package image

import (
	"fmt"
	"os/exec"
	"strconv"

	"github.com/JaimeStill/document-context/pkg/config"
)

// renderState encapsulates the parameters needed for a single render operation.
//
// This internal type groups render parameters to simplify the buildImageMagickArgs
// method signature and improve code organization.
type renderState struct {
	inputPath  string // Path to the input PDF file
	pageNum    int    // Page number to render (1-indexed)
	outputPath string // Path where the rendered image will be written
}

// parseImageMagickConfig transforms generic ImageConfig.Options into typed ImageMagickConfig.
//
// This function performs the configuration composition pattern's transformation step,
// extracting ImageMagick-specific options from the generic Options map and validating
// their types and ranges.
//
// Parsing process:
//  1. Extract "background" string (default: "white")
//  2. Extract optional filter values from Options map
//  3. Validate each filter value is within valid range
//  4. Return typed ImageMagickConfig with parsed values
//
// Filter validation ranges:
//   - brightness: 0-200 (100 is neutral)
//   - contrast: -100 to +100 (0 is neutral)
//   - saturation: 0-200 (100 is neutral)
//   - rotation: 0-360 degrees
//
// Returns an error if any option value is invalid (wrong type or out of range).
// The returned ImageMagickConfig embeds the base ImageConfig for unified access.
func parseImageMagickConfig(cfg config.ImageConfig) (*config.ImageMagickConfig, error) {
	background, err := config.ParseString(cfg.Options, "background", "white")
	if err != nil {
		return nil, err
	}

	brightness, err := config.ParseNilIntRanged(cfg.Options, "brightness", 0, 200)
	if err != nil {
		return nil, err
	}

	contrast, err := config.ParseNilIntRanged(cfg.Options, "contrast", -100, 100)
	if err != nil {
		return nil, err
	}

	saturation, err := config.ParseNilIntRanged(cfg.Options, "saturation", 0, 200)
	if err != nil {
		return nil, err
	}

	rotation, err := config.ParseNilIntRanged(cfg.Options, "rotation", 0, 360)
	if err != nil {
		return nil, err
	}

	return &config.ImageMagickConfig{
		Config:     cfg,
		Background: background,
		Brightness: brightness,
		Contrast:   contrast,
		Saturation: saturation,
		Rotation:   rotation,
	}, nil
}

type imagemagickRenderer struct {
	settings config.ImageMagickConfig
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

	if cfg.Format != "png" && cfg.Format != "jpg" && cfg.Format != "jpeg" {
		return nil, fmt.Errorf("unsupported image format: %s (must be 'png' or 'jpg')", cfg.Format)
	}

	if cfg.Format == "jpg" {
		if cfg.Quality < 1 || cfg.Quality > 100 {
			return nil, fmt.Errorf("JPEG quality must be 1-100, got %d", cfg.Quality)
		}
	}

	imCfg, err := parseImageMagickConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("invalid ImageMagick options: %w", err)
	}

	return &imagemagickRenderer{
		settings: *imCfg,
	}, nil
}

func (r *imagemagickRenderer) Render(inputPath string, pageNum int, outputPath string) error {
	state := renderState{
		inputPath:  inputPath,
		pageNum:    pageNum,
		outputPath: outputPath,
	}

	args := r.buildImageMagickArgs(state)

	cmd := exec.Command("magick", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("imagemagick failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}

func (r *imagemagickRenderer) FileExtension() string {
	return r.settings.Config.Format
}

func (r *imagemagickRenderer) Settings() config.ImageConfig {
	return r.settings.Config
}

// Parameters returns ImageMagick-specific rendering parameters for cache key generation.
//
// This method implements the Renderer.Parameters interface, providing a deterministic
// string representation of all ImageMagick-specific settings that affect render output.
// These parameters are used in cache key generation to ensure different filter
// configurations produce unique cache entries.
//
// Parameters are returned in alphabetical order for consistency:
//  1. background (always included)
//  2. brightness (if set)
//  3. contrast (if set)
//  4. rotation (if set)
//  5. saturation (if set)
//
// Format: Each parameter is formatted as "key=value"
//
// Example output: ["background=white", "brightness=10", "contrast=-5", "saturation=15"]
func (r *imagemagickRenderer) Parameters() []string {
	params := []string{
		fmt.Sprintf("background=%s", r.settings.Background),
	}

	if r.settings.Brightness != nil {
		params = append(params, fmt.Sprintf("brightness=%d", *r.settings.Brightness))
	}

	if r.settings.Contrast != nil {
		params = append(params, fmt.Sprintf("contrast=%d", *r.settings.Contrast))
	}

	if r.settings.Rotation != nil {
		params = append(params, fmt.Sprintf("rotation=%d", *r.settings.Rotation))
	}

	if r.settings.Saturation != nil {
		params = append(params, fmt.Sprintf("saturation=%d", *r.settings.Saturation))
	}

	return params
}

// buildImageMagickArgs constructs the ImageMagick command-line arguments for rendering.
//
// This method builds the complete argument list for the ImageMagick 'magick' command,
// applying settings and filters in the correct order according to ImageMagick's
// processing pipeline.
//
// Argument order (critical for correct rendering):
//  1. Settings before input: -density (affects input interpretation)
//  2. Input specification: path[pageIndex]
//  3. Operations after input: -background, -flatten
//  4. Filters (applied sequentially): -rotate, -modulate, -brightness-contrast
//  5. Output settings: -quality (for JPEG only)
//  6. Output path
//
// Filter optimization:
//   - Rotation: Applied only if set and non-zero
//   - Modulate: Applied only if brightness or saturation differ from neutral (100)
//   - Brightness-contrast: Applied only if contrast is set and non-zero
//   - Quality: Applied only for JPEG format
//
// Returns a string slice ready for exec.Command("magick", args...).
func (r *imagemagickRenderer) buildImageMagickArgs(state renderState) []string {
	pageIndex := state.pageNum - 1
	inputSpec := fmt.Sprintf("%s[%d]", state.inputPath, pageIndex)

	args := []string{
		"-density", strconv.Itoa(r.settings.Config.DPI),
		inputSpec,
		"-background", r.settings.Background,
		"-flatten",
	}

	if r.settings.Rotation != nil && *r.settings.Rotation != 0 {
		args = append(args, "-rotate", strconv.Itoa(*r.settings.Rotation))
	}

	needsModulate := false
	brightness := 100
	saturation := 100

	if r.settings.Brightness != nil && *r.settings.Brightness != 100 {
		brightness = *r.settings.Brightness
		needsModulate = true
	}

	if r.settings.Saturation != nil && *r.settings.Saturation != 100 {
		saturation = *r.settings.Saturation
		needsModulate = true
	}

	if needsModulate {
		modulate := fmt.Sprintf("%d,%d", brightness, saturation)
		args = append(args, "-modulate", modulate)
	}

	if r.settings.Contrast != nil && *r.settings.Contrast != 0 {
		contrast := fmt.Sprintf("0,%d", *r.settings.Contrast)
		args = append(args, "-brightness-contrast", contrast)
	}

	if r.settings.Config.Format == "jpg" || r.settings.Config.Format == "jpeg" {
		args = append(args, "-quality", strconv.Itoa(r.settings.Config.Quality))
	}

	args = append(args, state.outputPath)

	return args
}
