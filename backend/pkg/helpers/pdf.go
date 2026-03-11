package helpers

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/ledongthuc/pdf"
)

// PDFPage represents a single page extracted from a PDF document.
type PDFPage struct {
	// Number is the 1-based page number.
	Number int
	// Content is the extracted text content of the page.
	Content string
}

// ExtractPDFPages reads a PDF from raw bytes and returns text content per page.
func ExtractPDFPages(data []byte) ([]PDFPage, error) {
	tmpFile, err := os.CreateTemp("", "pdf-extract-*.pdf")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	if _, err := tmpFile.Write(data); err != nil {
		return nil, fmt.Errorf("failed to write temp file: %w", err)
	}
	tmpFile.Close()

	f, r, err := pdf.Open(tmpFile.Name())
	if err != nil {
		return nil, fmt.Errorf("failed to open PDF: %w", err)
	}
	defer f.Close()

	totalPages := r.NumPage()
	if totalPages == 0 {
		return nil, fmt.Errorf("PDF has no pages")
	}

	var pages []PDFPage
	for i := 1; i <= totalPages; i++ {
		p := r.Page(i)
		if p.V.IsNull() {
			continue
		}

		var buf bytes.Buffer
		rows, err := p.GetTextByRow()
		if err != nil {
			pages = append(pages, PDFPage{Number: i, Content: ""})
			continue
		}

		for _, row := range rows {
			for _, word := range row.Content {
				buf.WriteString(word.S)
			}
			buf.WriteString("\n")
		}

		content := strings.TrimSpace(buf.String())
		pages = append(pages, PDFPage{Number: i, Content: content})
	}

	return pages, nil
}

// ExtractPDFText reads a PDF from raw bytes and returns the full text content.
func ExtractPDFText(data []byte) (string, error) {
	pages, err := ExtractPDFPages(data)
	if err != nil {
		return "", err
	}

	var parts []string
	for _, page := range pages {
		if page.Content != "" {
			parts = append(parts, page.Content)
		}
	}

	return strings.Join(parts, "\n\n"), nil
}
