package io

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"time"

	"THW-JugendOlympiade/backend/database"

	"github.com/go-pdf/fpdf"
)

// ovMarginLR is the left/right page margin (mm) for ortsverband certificates.
const ovMarginLR = 15.0

// GenerateOrtsverbandCertificates creates one PDF page per Ortsverband.
//
// Ranking (from GetOrtsverbandEvaluations, ordered by average score desc):
//   - Ortsverbände sharing the top score → Siegerurkunde (ov_winner layout)
//   - All others → Urkunde (ov_participant layout)
//
// If templates/background_urkunde_ovs.png exists it is rendered as a
// full-page background on every certificate page before the text content.
// Layout positions are loaded from certificate_layout.toml.
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
		return fmt.Errorf("certificate_layout.toml: %w", err)
	}

	theme := DefaultTheme
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(ovMarginLR, ovMarginLR, ovMarginLR)
	pdf.SetAutoPageBreak(false, 0) // absolute positioning throughout

	// Background images per variant: use layout settings, fall back to legacy filename
	bgWinner := certLayout.OVWinner.BackgroundImage
	if bgWinner == "" {
		bgWinner = "templates/background_urkunde_ovs.png"
	}
	bgParticipant := certLayout.OVParticipant.BackgroundImage
	if bgParticipant == "" {
		bgParticipant = "templates/background_urkunde_ovs.png"
	}
	bgWinner = resolveTemplateImagePath(bgWinner)
	bgParticipant = resolveTemplateImagePath(bgParticipant)

	currentYear := eventYear
	if currentYear == 0 {
		currentYear = time.Now().Year()
	}

	topScore := evaluations[0].AverageScore
	for _, eval := range evaluations {
		pdf.AddPage()
		bg := bgParticipant
		if eval.AverageScore == topScore {
			bg = bgWinner
		}
		if bg != "" {
			pdf.Image(bg, 0, 0, 210, 297, false, imageTypeFromFile(bg), 0, "")
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
