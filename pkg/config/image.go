package config

import "maps"

// ImageConfig defines configuration for image rendering operations.
//
// This structure is used to initialize image renderers and supports
// JSON serialization for configuration files. Filter fields (Brightness,
// Contrast, Saturation, Rotation) use pointers to distinguish between
// "not set" (nil) and "explicitly set to zero" (pointer to 0).
//
// Validation of field values is performed by the consuming package
// (e.g., pkg/image) during transformation to domain objects.
type ImageConfig struct {
	Format  string         `json:"format,omitempty"`  // Image format: "png" or "jpg"
	Quality int            `json:"quality,omitempty"` // JPEG quality: 1-100 (ignored for PNG)
	DPI     int            `json:"dpi,omitempty"`     // Render density in dots per inch
	Options map[string]any `json:"options,omitempty"`
}

// DefaultImageConfig returns an ImageConfig with recommended default values.
//
// Defaults:
//   - Format: "png"
//   - Quality: 0 (not applicable for PNG)
//   - DPI: 300
//   - All filter fields: nil (no filters applied)
func DefaultImageConfig() ImageConfig {
	return ImageConfig{
		Format:  "png",
		Quality: 0,
		DPI:     300,
		Options: make(map[string]any),
	}
}

// Merge overlays non-zero values from source onto the receiver.
//
// Merge semantics:
//   - String fields: only merge if source is non-empty
//   - Integer fields: only merge if source is greater than zero
//   - Pointer fields: only merge if source is non-nil (allows explicit zero via pointer to 0)
//
// This enables layered configuration where higher-priority sources
// can override lower-priority sources without affecting unset values.
func (c *ImageConfig) Merge(source *ImageConfig) {
	if source.Format != "" {
		c.Format = source.Format
	}

	if source.Quality > 0 {
		c.Quality = source.Quality
	}

	if source.DPI > 0 {
		c.DPI = source.DPI
	}

	if source.Options != nil {
		if c.Options == nil {
			c.Options = make(map[string]any)
		}
		maps.Copy(c.Options, source.Options)
	}
}

// Finalize applies default values for any unset fields.
//
// This method merges the receiver's values onto a fresh default configuration,
// ensuring all fields have valid values. It modifies the receiver in place.
//
// Finalize should be called before transforming configuration into domain objects
// to ensure all required fields are populated.
func (c *ImageConfig) Finalize() {
	defaults := DefaultImageConfig()
	defaults.Merge(c)
	*c = defaults
}

// ImageMagickConfig extends ImageConfig with ImageMagick-specific rendering options.
//
// This configuration is parsed from ImageConfig.Options during renderer initialization.
// Filter fields use pointers to distinguish between "not set" (nil) and "explicitly
// set" (non-nil pointer), enabling optional filter application.
//
// Filter value ranges:
//   - Background: Color name (default: "white")
//   - Brightness: 0-200, where 100 is neutral (no change)
//   - Contrast: -100 to +100, where 0 is neutral (no change)
//   - Saturation: 0-200, where 100 is neutral (no change)
//   - Rotation: 0-360 degrees clockwise
//
// Validation of field values is performed during transformation to renderer objects.
type ImageMagickConfig struct {
	Config     ImageConfig // Base configuration (format, DPI, quality)
	Background string      // Background color for alpha channel flattening
	Brightness *int        // Brightness adjustment (0-200, 100=neutral)
	Contrast   *int        // Contrast adjustment (-100 to +100, 0=neutral)
	Saturation *int        // Saturation adjustment (0-200, 100=neutral)
	Rotation   *int        // Rotation in degrees (0-360)
}

// DefaultImageMagickConfig returns an ImageMagickConfig with recommended defaults.
//
// Defaults:
//   - Config: DefaultImageConfig()
//   - Background: "white"
//   - All filter fields: nil (no filters applied)
func DefaultImageMagickConfig() ImageMagickConfig {
	return ImageMagickConfig{
		Config:     DefaultImageConfig(),
		Background: "white",
		Brightness: nil,
		Contrast:   nil,
		Saturation: nil,
		Rotation:   nil,
	}
}
