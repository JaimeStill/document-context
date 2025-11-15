# Session 1: Configuration Foundation + Image Renderer Pattern

## Overview

Implement the Configuration Transformation Pattern by creating the configuration package, image renderer domain package, and updating document processing to use interface-based rendering. This establishes the correct architectural foundation following the principles in safe-object-architecture.md.

**Architecture**:
```
pkg/document (high-level: uses image.Renderer interface)
    ↓
pkg/image (mid-level: transforms config → Renderer)
    ↓
pkg/config (low-level: data structures only)
```

---

## Task 1: Create pkg/config Package

### Step 1.1: Create pkg/config directory

```bash
mkdir -p pkg/config
```

### Step 1.2: Create pkg/config/doc.go

**File**: `pkg/config/doc.go`

```go
package config
```

### Step 1.3: Create pkg/config/image.go

**File**: `pkg/config/image.go`

```go
package config

type ImageConfig struct {
	Format     string `json:"format,omitempty"`
	Quality    int    `json:"quality,omitempty"`
	DPI        int    `json:"dpi,omitempty"`
	Brightness *int   `json:"brightness,omitempty"`
	Contrast   *int   `json:"contrast,omitempty"`
	Saturation *int   `json:"saturation,omitempty"`
	Rotation   *int   `json:"rotation,omitempty"`
}

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

func (c *ImageConfig) Finalize() {
	defaults := DefaultImageConfig()
	defaults.Merge(c)
	*c = defaults
}
```

### Step 1.4: Create pkg/config/cache.go

**File**: `pkg/config/cache.go`

```go
package config

type CacheConfig struct {
	Name    string         `json:"name"`
	Options map[string]any `json:"options,omitempty"`
}

func DefaultCacheConfig() CacheConfig {
	return CacheConfig{
		Name:    "",
		Options: make(map[string]any),
	}
}

func (c *CacheConfig) Merge(source *CacheConfig) {
	if source.Name != "" {
		c.Name = source.Name
	}

	if source.Options != nil {
		if c.Options == nil {
			c.Options = make(map[string]any)
		}
		for k, v := range source.Options {
			c.Options[k] = v
		}
	}
}
```

---

## Task 2: Create pkg/image Package

### Step 2.1: Create pkg/image directory

```bash
mkdir -p pkg/image
```

### Step 2.2: Create pkg/image/image.go

**File**: `pkg/image/image.go`

```go
package image

type Renderer interface {
	Render(inputPath string, pageNum int, outputPath string) error
	FileExtension() string
}
```

### Step 2.3: Create pkg/image/imagemagick.go

**File**: `pkg/image/imagemagick.go`

```go
package image

import (
	"fmt"
	"os/exec"
	"strconv"

	"github.com/JaimeStill/document-context/pkg/config"
)

type imagemagickRenderer struct {
	format     string
	quality    int
	dpi        int
	brightness int
	contrast   int
	saturation int
	rotation   int
}

func NewImageMagickRenderer(cfg config.ImageConfig) (Renderer, error) {
	cfg.Finalize()

	if cfg.Format != "png" && cfg.Format != "jpg" {
		return nil, fmt.Errorf("unsupported image format: %s (must be 'png' or 'jpg')", cfg.Format)
	}

	if cfg.Format == "jpg" {
		if cfg.Quality < 1 || cfg.Quality > 100 {
			return nil, fmt.Errorf("JPEG quality must be 1-100, got %d", cfg.Quality)
		}
	}

	brightness := 0
	if cfg.Brightness != nil {
		if *cfg.Brightness < -100 || *cfg.Brightness > 100 {
			return nil, fmt.Errorf("brightness must be -100 to +100, got %d", *cfg.Brightness)
		}
		brightness = *cfg.Brightness
	}

	contrast := 0
	if cfg.Contrast != nil {
		if *cfg.Contrast < -100 || *cfg.Contrast > 100 {
			return nil, fmt.Errorf("contrast must be -100 to +100, got %d", *cfg.Contrast)
		}
		contrast = *cfg.Contrast
	}

	saturation := 0
	if cfg.Saturation != nil {
		if *cfg.Saturation < -100 || *cfg.Saturation > 100 {
			return nil, fmt.Errorf("saturation must be -100 to +100, got %d", *cfg.Saturation)
		}
		saturation = *cfg.Saturation
	}

	rotation := 0
	if cfg.Rotation != nil {
		if *cfg.Rotation < 0 || *cfg.Rotation > 360 {
			return nil, fmt.Errorf("rotation must be 0 to 360 degrees, got %d", *cfg.Rotation)
		}
		rotation = *cfg.Rotation
	}

	return &imagemagickRenderer{
		format:     cfg.Format,
		quality:    cfg.Quality,
		dpi:        cfg.DPI,
		brightness: brightness,
		contrast:   contrast,
		saturation: saturation,
		rotation:   rotation,
	}, nil
}

func (r *imagemagickRenderer) Render(inputPath string, pageNum int, outputPath string) error {
	pageIndex := pageNum - 1
	inputSpec := fmt.Sprintf("%s[%d]", inputPath, pageIndex)

	args := []string{
		"-density", strconv.Itoa(r.dpi),
		inputSpec,
		"-background", "white",
		"-flatten",
	}

	if r.format == "jpg" {
		args = append(args, "-quality", strconv.Itoa(r.quality))
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
	return r.format
}
```

---

## Task 3: Update pkg/document/document.go

### Step 3.1: Add imports

Update imports section:

```go
import (
	"fmt"

	"github.com/JaimeStill/document-context/pkg/image"
)
```

### Step 3.2: Keep ImageFormat (domain type)

ImageFormat stays in document package - it's a domain type, not configuration.

**No changes to ImageFormat, constants, or MimeType() method.**

### Step 3.3: Remove ImageOptions

Delete the `ImageOptions` struct and `DefaultImageOptions()` function.

### Step 3.4: Update Page interface

Change the `ToImage` method signature:

```go
type Page interface {
	Number() int
	ToImage(renderer image.Renderer) ([]byte, error)
}
```

**Complete pkg/document/document.go**:

```go
package document

import (
	"fmt"

	"github.com/JaimeStill/document-context/pkg/image"
)

type ImageFormat string

const (
	PNG  ImageFormat = "png"
	JPEG ImageFormat = "jpg"
)

func (f ImageFormat) MimeType() (string, error) {
	switch f {
	case PNG:
		return "image/png", nil
	case JPEG:
		return "image/jpeg", nil
	default:
		return "", fmt.Errorf("unsupported image format: %s", f)
	}
}

type Document interface {
	PageCount() int
	ExtractPage(pageNum int) (Page, error)
	ExtractAllPages() ([]Page, error)
	Close() error
}

type Page interface {
	Number() int
	ToImage(renderer image.Renderer) ([]byte, error)
}
```

---

## Task 4: Update pkg/document/pdf.go

### Step 4.1: Add image import

Add to imports section:

```go
"github.com/JaimeStill/document-context/pkg/image"
```

### Step 4.2: Update ToImage method signature

Change from:

```go
func (p *PDFPage) ToImage(opts ImageOptions) ([]byte, error) {
```

to:

```go
func (p *PDFPage) ToImage(renderer image.Renderer) ([]byte, error) {
```

### Step 4.3: Replace ToImage implementation

Replace the entire `ToImage` method body:

```go
func (p *PDFPage) ToImage(renderer image.Renderer) ([]byte, error) {
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

	return imgData, nil
}
```

**Key changes**:
- No validation (renderer already validated during construction)
- No ImageMagick command building (renderer encapsulates this)
- Clean delegation to renderer.Render()
- Simple file operations and error handling

---

## Completion Checklist

After completing all tasks:

- [ ] pkg/config/ created with doc.go, image.go, cache.go
- [ ] ImageConfig has Finalize() and Merge() methods
- [ ] CacheConfig has Merge() method
- [ ] pkg/image/ created with image.go, imagemagick.go
- [ ] Renderer interface defined
- [ ] ImageMagickRenderer implements Renderer
- [ ] NewImageMagickRenderer() validates and transforms config
- [ ] pkg/document/document.go updated (Page uses image.Renderer)
- [ ] pkg/document/pdf.go updated (ToImage uses renderer)
- [ ] Code compiles: `go build ./pkg/...`

---

## Expected Outcomes

**Architectural Improvements**:
- Configuration → Domain Object transformation implemented
- Interface-based public API (Renderer interface)
- Clear layer separation (document → image → config)
- Validation at transformation boundary
- All ImageMagick details encapsulated in image package

**Usage Pattern** (after implementation):
```go
// Create configuration
cfg := config.ImageConfig{Format: "png", DPI: 150}

// Transform to domain object (validates)
renderer, err := image.NewImageMagickRenderer(cfg)
if err != nil {
    return err  // Configuration invalid
}

// Use domain object
doc, _ := document.OpenPDF("file.pdf")
page, _ := doc.ExtractPage(1)
img, err := page.ToImage(renderer)  // Clean delegation
```

**Filter Note**: Filters are validated but not yet applied to ImageMagick commands. Session 5 will add filter application logic to imagemagickRenderer.Render().
