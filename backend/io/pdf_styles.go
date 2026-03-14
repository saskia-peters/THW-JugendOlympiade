package io

import "github.com/jung-kurt/gofpdf"

// PDFTheme defines all visual properties for PDF generation.
// It is the single place to change fonts, sizes, and colors across all
// generated PDFs — analogous to a CSS stylesheet.
type PDFTheme struct {
	// Font family used throughout all PDFs.
	FontFamily string

	// Font sizes (pt) — general documents
	SizeTitle       float64 // Main page title
	SizeSubtitle    float64 // Subtitle / ranking description line
	SizeHeading     float64 // Section heading
	SizeBody        float64 // Normal body / table body text
	SizeSmall       float64 // Footnote / statistics text
	SizeTableHeader float64 // Table header row

	// Font sizes (pt) — participant certificates
	SizeCertTitle       float64 // "Jugendolympiade" heading
	SizeCertYear        float64 // Year below title
	SizeCertName        float64 // Participant name
	SizeCertOrtsverband float64 // Ortsverband line
	SizeCertGroup       float64 // Group number
	SizeCertRank        float64 // Rank text (highlighted)
	SizeCertLabel       float64 // "Gruppenmitglieder" section label
	SizeCertTableHeader float64 // Certificate members table header
	SizeCertTableBody   float64 // Certificate members table rows

	// Colors [R, G, B]
	ColorPrimary   [3]int // Titles, primary accents, group eval header
	ColorSecondary [3]int // Ortsverband eval header
	ColorAccent    [3]int // Rank highlight (gold)
	ColorOnHeader  [3]int // Text on colored table headers (white)
	ColorText      [3]int // Main body text
	ColorSubtext   [3]int // Secondary / muted text

	// Table background colors
	ColorTableHeader    [3]int // Plain grey header background
	ColorTableRowAlt    [3]int // Alternating row background
	ColorTableHighlight [3]int // Top-3 row highlight
}

// DefaultTheme is the active theme used by all PDF generators.
// Modify the values here to restyle every generated PDF at once.
var DefaultTheme = PDFTheme{
	FontFamily: "Arial",

	// General
	SizeTitle:       24,
	SizeSubtitle:    12,
	SizeHeading:     14,
	SizeBody:        11,
	SizeSmall:       9,
	SizeTableHeader: 11,

	// Certificates
	SizeCertTitle:       28,
	SizeCertYear:        24,
	SizeCertName:        28,
	SizeCertOrtsverband: 14,
	SizeCertGroup:       16,
	SizeCertRank:        22,
	SizeCertLabel:       12,
	SizeCertTableHeader: 10,
	SizeCertTableBody:   9,

	// Colors
	ColorPrimary:   [3]int{102, 126, 234},
	ColorSecondary: [3]int{250, 112, 154},
	ColorAccent:    [3]int{180, 140, 10},
	ColorOnHeader:  [3]int{255, 255, 255},
	ColorText:      [3]int{0, 0, 0},
	ColorSubtext:   [3]int{100, 100, 100},

	ColorTableHeader:    [3]int{200, 200, 200},
	ColorTableRowAlt:    [3]int{240, 240, 240},
	ColorTableHighlight: [3]int{255, 243, 205},
}

// Font sets the font on pdf using this theme's font family.
func (t PDFTheme) Font(pdf *gofpdf.Fpdf, style string, size float64) {
	pdf.SetFont(t.FontFamily, style, size)
}

// TextColor sets the active text color on pdf.
func (t PDFTheme) TextColor(pdf *gofpdf.Fpdf, c [3]int) {
	pdf.SetTextColor(c[0], c[1], c[2])
}

// FillColor sets the active fill color on pdf.
func (t PDFTheme) FillColor(pdf *gofpdf.Fpdf, c [3]int) {
	pdf.SetFillColor(c[0], c[1], c[2])
}
