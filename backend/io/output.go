package io

import (
	"fmt"
	"os"

	"golang.org/x/text/encoding/charmap"
)

var pdfOutputDir = "pdfdocs"

// SetPDFOutputDir overrides the directory where generated PDFs are saved.
func SetPDFOutputDir(dir string) {
	if dir != "" {
		pdfOutputDir = dir
	}
}

// ensurePDFDirectory creates the pdfdocs directory if it doesn't exist.
func ensurePDFDirectory() error {
	if err := os.MkdirAll(pdfOutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create PDF output directory: %w", err)
	}
	return nil
}

// enc converts a UTF-8 string to ISO-8859-1 so that gofpdf's built-in fonts
// (Arial, Helvetica, etc.) render German umlauts (ä ö ü ß Ä Ö Ü) correctly.
// Characters outside Latin-1 are replaced with '?'.
func enc(s string) string {
	encoded, err := charmap.ISO8859_1.NewEncoder().String(s)
	if err != nil {
		return s // fall back to original on unexpected errors
	}
	return encoded
}
