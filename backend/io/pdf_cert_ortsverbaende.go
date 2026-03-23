package io

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"time"

	"THW-JugendOlympiade/backend/database"

	"github.com/jung-kurt/gofpdf"
)

// A4 page geometry constants (mm) used throughout ortsverband certificates.
const (
	ovPageW    = 210.0
	ovMarginLR = 15.0
	ovContentW = ovPageW - 2*ovMarginLR // 180 mm
)

// GenerateOrtsverbandCertificates creates one PDF page per Ortsverband.
//
// Ranking (from GetOrtsverbandEvaluations, ordered by average score desc):
//   - Index 0 → Siegerurkunde: shows ov_winner_image.png above "Bester Ortsverband"
//   - All others → Urkunde: identical layout, no ranking mentioned
//
// The layout is always fully programmatic and centered on A4.
// No background image is used.
func GenerateOrtsverbandCertificates(db *sql.DB, eventYear int) error {
	if err := ensurePDFDirectory(); err != nil {
		return err
	}

	evaluations, err := database.GetOrtsverbandEvaluations(db)
	if err != nil {
		return fmt.Errorf("failed to get ortsverband evaluations: %w", err)
	}
	if len(evaluations) == 0 {
		return fmt.Errorf("keine Ortsverbände mit Bewertungen gefunden")
	}

	// Build ortsverband → participant names map
	teilnehmende, err := database.GetAllTeilnehmende(db)
	if err != nil {
		return fmt.Errorf("failed to get participants: %w", err)
	}
	ovParticipants := make(map[string][]string)
	for _, t := range teilnehmende {
		if t.Ortsverband != "" {
			ovParticipants[t.Ortsverband] = append(ovParticipants[t.Ortsverband], t.Name)
		}
	}

	theme := DefaultTheme
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(ovMarginLR, ovMarginLR, ovMarginLR)
	pdf.SetAutoPageBreak(false, 0) // absolute positioning throughout

	currentYear := eventYear
	if currentYear == 0 {
		currentYear = time.Now().Year()
	}

	for i, eval := range evaluations {
		pdf.AddPage()
		if i == 0 {
			ovRenderWinner(pdf, theme, eval.Ortsverband, ovParticipants[eval.Ortsverband], currentYear)
		} else {
			ovRenderParticipant(pdf, theme, eval.Ortsverband, ovParticipants[eval.Ortsverband], currentYear)
		}
	}

	if err := pdf.OutputFileAndClose(filepath.Join(pdfOutputDir, "Urkunden_Ortsverbaende.pdf")); err != nil {
		return fmt.Errorf("failed to save PDF: %w", err)
	}
	return nil
}

// ovRenderWinner renders the Siegerurkunde for the best Ortsverband.
// ov_winner_image.png (90 mm wide, centred; height auto from aspect ratio)
// is placed above the "Bester Ortsverband" text.
//
// Vertical layout (y in mm):
//
//	 25  "Jugendolympiade" title
//	 44  year
//	 62  "Siegerurkunde"
//	 78  ortsverband name
//	105  ov_winner_image.png  (90 mm wide, centred)
//	185  "Bester Ortsverband" (gold)
//	200  "Teilnehmende" section label
//	210  participant list
func ovRenderWinner(pdf *gofpdf.Fpdf, theme PDFTheme, ortsverband string, participants []string, year int) {
	const left = ovMarginLR
	const w = ovContentW

	pdf.SetXY(left, 25)
	theme.Font(pdf, "B", theme.SizeCertTitle)
	theme.TextColor(pdf, theme.ColorPrimary)
	pdf.CellFormat(w, 14, "Jugendolympiade", "", 0, "C", false, 0, "")

	pdf.SetXY(left, 44)
	theme.Font(pdf, "B", theme.SizeCertYear)
	theme.TextColor(pdf, theme.ColorPrimary)
	pdf.CellFormat(w, 12, fmt.Sprintf("%d", year), "", 0, "C", false, 0, "")

	pdf.SetXY(left, 62)
	theme.Font(pdf, "B", theme.SizeCertGroup)
	theme.TextColor(pdf, theme.ColorText)
	pdf.CellFormat(w, 12, "Siegerurkunde", "", 0, "C", false, 0, "")

	pdf.SetXY(left, 78)
	theme.Font(pdf, "B", theme.SizeCertName)
	theme.TextColor(pdf, theme.ColorText)
	pdf.CellFormat(w, 14, enc(ortsverband), "", 0, "C", false, 0, "")

	// Winner image: centred horizontally, 90 mm wide, height from aspect ratio.
	const imgW = 90.0
	pdf.Image("ov_winner_image.png", (ovPageW-imgW)/2, 105, imgW, 0, false, "", 0, "")

	pdf.SetXY(left, 185)
	theme.Font(pdf, "B", theme.SizeCertRank)
	theme.TextColor(pdf, theme.ColorAccent)
	pdf.CellFormat(w, 14, "Bester Ortsverband", "", 0, "C", false, 0, "")

	pdf.SetXY(left, 200)
	theme.Font(pdf, "B", theme.SizeCertLabel)
	theme.TextColor(pdf, theme.ColorText)
	pdf.CellFormat(w, 8, "Teilnehmende", "", 0, "C", false, 0, "")

	ovParticipantsList(pdf, theme, participants, left, w, 210)
}

// ovRenderParticipant renders the standard participation Urkunde.
// No image and no ranking are shown.
//
// Vertical layout (y in mm):
//
//	 40  "Jugendolympiade" title
//	 60  year
//	 80  "Urkunde"
//	100  ortsverband name
//	125  "hat erfolgreich teilgenommen"
//	150  "Teilnehmende" section label
//	160  participant list
func ovRenderParticipant(pdf *gofpdf.Fpdf, theme PDFTheme, ortsverband string, participants []string, year int) {
	const left = ovMarginLR
	const w = ovContentW

	pdf.SetXY(left, 40)
	theme.Font(pdf, "B", theme.SizeCertTitle)
	theme.TextColor(pdf, theme.ColorPrimary)
	pdf.CellFormat(w, 14, "Jugendolympiade", "", 0, "C", false, 0, "")

	pdf.SetXY(left, 60)
	theme.Font(pdf, "B", theme.SizeCertYear)
	theme.TextColor(pdf, theme.ColorPrimary)
	pdf.CellFormat(w, 12, fmt.Sprintf("%d", year), "", 0, "C", false, 0, "")

	pdf.SetXY(left, 80)
	theme.Font(pdf, "B", theme.SizeCertGroup)
	theme.TextColor(pdf, theme.ColorText)
	pdf.CellFormat(w, 12, "Urkunde", "", 0, "C", false, 0, "")

	pdf.SetXY(left, 100)
	theme.Font(pdf, "B", theme.SizeCertName)
	theme.TextColor(pdf, theme.ColorText)
	pdf.CellFormat(w, 14, enc(ortsverband), "", 0, "C", false, 0, "")

	pdf.SetXY(left, 125)
	theme.Font(pdf, "", theme.SizeCertOrtsverband)
	theme.TextColor(pdf, theme.ColorSubtext)
	pdf.CellFormat(w, 10, "hat erfolgreich teilgenommen", "", 0, "C", false, 0, "")

	pdf.SetXY(left, 150)
	theme.Font(pdf, "B", theme.SizeCertLabel)
	theme.TextColor(pdf, theme.ColorText)
	pdf.CellFormat(w, 8, "Teilnehmende", "", 0, "C", false, 0, "")

	ovParticipantsList(pdf, theme, participants, left, w, 160)
}

// ovParticipantsList renders a single-column list of participant names
// starting at the given absolute y position (mm).
func ovParticipantsList(pdf *gofpdf.Fpdf, theme PDFTheme, names []string, left, width, startY float64) {
	pdf.SetXY(left, startY)
	theme.Font(pdf, "", theme.SizeCertTableBody)
	theme.FillColor(pdf, theme.ColorTableRowAlt)
	for i, name := range names {
		pdf.SetX(left)
		pdf.CellFormat(width, 7, enc(name), "1", 1, "L", i%2 == 0, 0, "")
	}
}
