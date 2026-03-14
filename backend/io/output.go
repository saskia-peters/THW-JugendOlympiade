package io

import (
	"fmt"
	"os"
)

const pdfOutputDir = "pdfdocs"

// ensurePDFDirectory creates the pdfdocs directory if it doesn't exist.
func ensurePDFDirectory() error {
	if err := os.MkdirAll(pdfOutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create PDF output directory: %w", err)
	}
	return nil
}
