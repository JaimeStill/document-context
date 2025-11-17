package document

import (
	"fmt"

	"github.com/JaimeStill/document-context/pkg/cache"
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
	ToImage(renderer image.Renderer, c cache.Cache) ([]byte, error)
}
