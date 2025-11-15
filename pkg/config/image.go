package config

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
	Format     string `json:"format,omitempty"`      // Image format: "png" or "jpg"
	Quality    int    `json:"quality,omitempty"`     // JPEG quality: 1-100 (ignored for PNG)
	DPI        int    `json:"dpi,omitempty"`         // Render density in dots per inch
	Brightness *int   `json:"brightness,omitempty"`  // Brightness adjustment: -100 to +100
	Contrast   *int   `json:"contrast,omitempty"`    // Contrast adjustment: -100 to +100
	Saturation *int   `json:"saturation,omitempty"`  // Saturation adjustment: -100 to +100
	Rotation   *int   `json:"rotation,omitempty"`    // Rotation in degrees: 0 to 360
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
		Format:     "png",
		Quality:    0,
		DPI:        300,
		Brightness: nil,
		Contrast:   nil,
		Saturation: nil,
		Rotation:   nil,
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

	if source.Brightness != nil {
		c.Brightness = source.Brightness
	}

	if source.Contrast != nil {
		c.Contrast = source.Contrast
	}

	if source.Saturation != nil {
		c.Saturation = source.Saturation
	}

	if source.Rotation != nil {
		c.Rotation = source.Rotation
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
