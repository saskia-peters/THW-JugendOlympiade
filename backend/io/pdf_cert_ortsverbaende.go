package io

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"THW-JugendOlympiade/backend/database"

	"github.com/jung-kurt/gofpdf"
)

// ovMarginLR is the left/right page margin (mm) for ortsverband certificates.
const ovMarginLR = 15.0

// GenerateOrtsverbandCertificates creates one PDF page per Ortsverband.
//
// Ranking (from GetOrtsverbandEvaluations, ordered by average score desc):
//   - Ortsverbände sharing the top score → Siegerurkunde (ov_winner layout)
//   - All others → Urkunde (ov_participant layout)
//
// If cert_background_ov.png exists in the working directory it is rendered as a
// full-page background on every certificate page before the text content.
// Layout positions are loaded from certificate_layout.json.
func GenerateOrtsverbandCertificates(db *sql.DB, eventYear int, eventName string) error {
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

	certLayout, err := LoadCertLayout()
	if err != nil {
		return fmt.Errorf("certificate_layout.json: %w", err)
	}

	theme := DefaultTheme
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(ovMarginLR, ovMarginLR, ovMarginLR)
	pdf.SetAutoPageBreak(false, 0) // absolute positioning throughout

	const bgFile = "cert_background_ov.png"
	_, statErr := os.Stat(bgFile)
	useBg := statErr == nil

	currentYear := eventYear
	if currentYear == 0 {
		currentYear = time.Now().Year()
	}

	topScore := evaluations[0].AverageScore
	for _, eval := range evaluations {
		pdf.AddPage()
		if useBg {
			pdf.Image(bgFile, 0, 0, 210, 297, false, imageTypeFromFile(bgFile), 0, "")
		}
		ctx := CertContext{
			EventName:   eventName,
			Year:        currentYear,
			Ortsverband: eval.Ortsverband,
			OVNames:     ovParticipants[eval.Ortsverband],
		}
		if eval.AverageScore == topScore {
			RenderCertPage(pdf, theme, certLayout.OVWinner, ctx)
		} else {
			RenderCertPage(pdf, theme, certLayout.OVParticipant, ctx)
		}
	}

	if err := pdf.OutputFileAndClose(filepath.Join(pdfOutputDir, "Urkunden_Ortsverbaende.pdf")); err != nil {
		return fmt.Errorf("failed to save PDF: %w", err)
	}
	return nil
}
