package document

import (
	"fmt"
	"strings"

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

func ParseImageFormat(s string) (ImageFormat, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "", "png":
		return PNG, nil
	case "jpg", "jpeg":
		return JPEG, nil
	default:
		return "", fmt.Errorf("unsupported image format: %s", s)
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

var formatRegistry = map[string]func(string) (Document, error){
	"application/pdf": func(path string) (Document, error) {
		return OpenPDF(path)
	},
}

func SupportedFormats() []string {
	formats := make([]string, 0, len(formatRegistry))
	for contentType := range formatRegistry {
		formats = append(formats, contentType)
	}
	return formats
}

func IsSupported(contentType string) bool {
	_, ok := formatRegistry[contentType]
	return ok
}

func Open(path string, contentType string) (Document, error) {
	opener, ok := formatRegistry[contentType]
	if !ok {
		return nil, fmt.Errorf("unsupported content type: %s", contentType)
	}
	return opener(path)
}
