package document

import (
	"fmt"
	"os"

	"github.com/JaimeStill/document-context/pkg/image"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

type PDFDocument struct {
	path      string
	ctx       *model.Context
	pageCount int
}

func OpenPDF(path string) (*PDFDocument, error) {
	ctx, err := api.ReadContextFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open PDF: %w", err)
	}

	pageCount := ctx.PageCount
	if pageCount == 0 {
		return nil, fmt.Errorf("PDF has no pages")
	}

	return &PDFDocument{
		path:      path,
		ctx:       ctx,
		pageCount: pageCount,
	}, nil
}

func (d *PDFDocument) PageCount() int {
	return d.pageCount
}

func (d *PDFDocument) ExtractPage(pageNum int) (Page, error) {
	if pageNum < 1 || pageNum > d.pageCount {
		return nil, fmt.Errorf("page %d out of range [1-%d]", pageNum, d.pageCount)
	}

	return &PDFPage{
		doc:    d,
		number: pageNum,
	}, nil
}

func (d *PDFDocument) ExtractAllPages() ([]Page, error) {
	pages := make([]Page, 0, d.pageCount)

	for i := 1; i <= d.pageCount; i++ {
		page, err := d.ExtractPage(i)
		if err != nil {
			return nil, fmt.Errorf("failed to extract page %d: %w", i, err)
		}
		pages = append(pages, page)
	}

	return pages, nil
}

func (d *PDFDocument) Close() error {
	d.ctx = nil
	return nil
}

type PDFPage struct {
	doc    *PDFDocument
	number int
}

func (p *PDFPage) Number() int {
	return p.number
}

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
