package document_test

import (
	"testing"

	"github.com/JaimeStill/document-context/pkg/document"
)

func TestParseImageFormat(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    document.ImageFormat
		wantErr bool
	}{
		{"png lowercase", "png", document.PNG, false},
		{"PNG uppercase", "PNG", document.PNG, false},
		{"Png mixed case", "Png", document.PNG, false},
		{"png with spaces", "  png  ", document.PNG, false},
		{"jpg lowercase", "jpg", document.JPEG, false},
		{"JPG uppercase", "JPG", document.JPEG, false},
		{"jpeg lowercase", "jpeg", document.JPEG, false},
		{"JPEG uppercase", "JPEG", document.JPEG, false},
		{"empty string defaults to png", "", document.PNG, false},
		{"whitespace only defaults to png", "   ", document.PNG, false},
		{"invalid format gif", "gif", "", true},
		{"invalid format bmp", "bmp", "", true},
		{"invalid format webp", "webp", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := document.ParseImageFormat(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Error("ParseImageFormat() error = nil, want error")
				}
			} else {
				if err != nil {
					t.Errorf("ParseImageFormat() error = %v, want nil", err)
				}
				if got != tt.want {
					t.Errorf("ParseImageFormat() = %q, want %q", got, tt.want)
				}
			}
		})
	}
}

func TestSupportedFormats(t *testing.T) {
	formats := document.SupportedFormats()

	if len(formats) == 0 {
		t.Fatal("SupportedFormats() returned empty slice")
	}

	found := false
	for _, f := range formats {
		if f == "application/pdf" {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("SupportedFormats() = %v, want to contain 'application/pdf'", formats)
	}
}

func TestIsSupported(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		want        bool
	}{
		{"pdf supported", "application/pdf", true},
		{"docx not supported", "application/vnd.openxmlformats-officedocument.wordprocessingml.document", false},
		{"image/png not supported", "image/png", false},
		{"empty string not supported", "", false},
		{"text/plain not supported", "text/plain", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := document.IsSupported(tt.contentType)
			if got != tt.want {
				t.Errorf("IsSupported(%q) = %v, want %v", tt.contentType, got, tt.want)
			}
		})
	}
}

func TestOpen(t *testing.T) {
	pdfPath := testPDFPath(t)

	t.Run("open pdf success", func(t *testing.T) {
		doc, err := document.Open(pdfPath, "application/pdf")
		if err != nil {
			t.Fatalf("Open() error = %v", err)
		}
		defer doc.Close()

		if doc.PageCount() == 0 {
			t.Error("Open() returned document with zero pages")
		}
	})

	t.Run("unsupported content type", func(t *testing.T) {
		_, err := document.Open(pdfPath, "application/msword")
		if err == nil {
			t.Error("Open() error = nil, want error for unsupported content type")
		}
	})

	t.Run("invalid path with valid content type", func(t *testing.T) {
		_, err := document.Open("/nonexistent/file.pdf", "application/pdf")
		if err == nil {
			t.Error("Open() error = nil, want error for invalid path")
		}
	})
}
