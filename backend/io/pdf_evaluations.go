package io

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"strings"

	"THW-JugendOlympiade/backend/database"

	"github.com/jung-kurt/gofpdf"
)

// GenerateGroupEvaluationPDF creates a PDF report with group rankings and scores.
func GenerateGroupEvaluationPDF(db *sql.DB) error {
	if err := ensurePDFDirectory(); err != nil {
		return err
	}

	evaluations, err := database.GetGroupEvaluations(db)
	if err != nil {
		return fmt.Errorf("failed to get evaluations: %w", err)
	}
	if len(evaluations) == 0 {
		return fmt.Errorf("no group evaluations found to generate report")
	}

	theme := DefaultTheme
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.SetAutoPageBreak(true, 15)
	pdf.AddPage()

	// Title
	theme.Font(pdf, "B", theme.SizeTitle)
	theme.TextColor(pdf, theme.ColorText)
	pdf.CellFormat(0, 15, "Gruppenauswertung", "", 1, "C", false, 0, "")
	theme.Font(pdf, "", theme.SizeSubtitle)
	theme.TextColor(pdf, theme.ColorSubtext)
	pdf.CellFormat(0, 8, "Ranking nach Gesamtpunktzahl", "", 1, "C", false, 0, "")
	pdf.Ln(8)

	// Table header
	theme.Font(pdf, "B", theme.SizeTableHeader)
	theme.FillColor(pdf, theme.ColorPrimary)
	theme.TextColor(pdf, theme.ColorOnHeader)

	colWidths := []float64{30, 70, 45, 45}
	for i, header := range []string{"Platz", "Gruppe", "Stationen", "Gesamtscore"} {
		pdf.CellFormat(colWidths[i], 10, header, "1", 0, "C", true, 0, "")
	}
	pdf.Ln(-1)

	// Table rows
	theme.TextColor(pdf, theme.ColorText)
	theme.FillColor(pdf, theme.ColorTableHighlight)

	for i, eval := range evaluations {
		fill := i < 3
		if fill {
			theme.Font(pdf, "B", theme.SizeBody)
		} else {
			theme.Font(pdf, "", theme.SizeBody)
		}
		pdf.CellFormat(colWidths[0], 9, fmt.Sprintf("%d", i+1), "1", 0, "C", fill, 0, "")
		pdf.CellFormat(colWidths[1], 9, fmt.Sprintf("Gruppe %d", eval.GroupID), "1", 0, "L", fill, 0, "")
		pdf.CellFormat(colWidths[2], 9, fmt.Sprintf("%d", eval.StationCount), "1", 0, "C", fill, 0, "")
		pdf.CellFormat(colWidths[3], 9, fmt.Sprintf("%d", eval.TotalScore), "1", 0, "C", fill, 0, "")
		pdf.Ln(-1)
	}

	// Summary
	pdf.Ln(10)
	theme.Font(pdf, "B", theme.SizeHeading)
	theme.TextColor(pdf, theme.ColorText)
	pdf.CellFormat(0, 8, "Zusammenfassung", "", 1, "L", false, 0, "")
	pdf.Ln(3)
	theme.Font(pdf, "", theme.SizeBody)
	pdf.CellFormat(0, 6, fmt.Sprintf("Gesamtanzahl Gruppen: %d", len(evaluations)), "", 1, "L", false, 0, "")
	if len(evaluations) > 0 {
		pdf.CellFormat(0, 6, enc(fmt.Sprintf("Höchster Score: %d (Gruppe %d)", evaluations[0].TotalScore, evaluations[0].GroupID)), "", 1, "L", false, 0, "")
		last := evaluations[len(evaluations)-1]
		pdf.CellFormat(0, 6, enc(fmt.Sprintf("Niedrigster Score: %d (Gruppe %d)", last.TotalScore, last.GroupID)), "", 1, "L", false, 0, "")
		total := 0
		for _, e := range evaluations {
			total += e.TotalScore
		}
		pdf.CellFormat(0, 6, fmt.Sprintf("Durchschnittlicher Score: %.1f", float64(total)/float64(len(evaluations))), "", 1, "L", false, 0, "")
	}

	if err = pdf.OutputFileAndClose(filepath.Join(pdfOutputDir, "Auswertung_nach_Gruppe.pdf")); err != nil {
		return fmt.Errorf("failed to save PDF: %w", err)
	}
	return nil
}

// GenerateOrtsverbandEvaluationPDF creates a PDF report listing all Ortsverbände.
// The best Ortsverband is highlighted; all others are listed without ranking.
func GenerateOrtsverbandEvaluationPDF(db *sql.DB) error {
	if err := ensurePDFDirectory(); err != nil {
		return err
	}

	evaluations, err := database.GetOrtsverbandEvaluations(db)
	if err != nil {
		return fmt.Errorf("failed to get evaluations: %w", err)
	}
	if len(evaluations) == 0 {
		return fmt.Errorf("no ortsverband evaluations found to generate report")
	}

	theme := DefaultTheme
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.SetAutoPageBreak(true, 15)
	pdf.AddPage()

	// Title
	theme.Font(pdf, "B", theme.SizeTitle)
	theme.TextColor(pdf, theme.ColorText)
	pdf.CellFormat(0, 15, "Ortsverband-Auswertung", "", 1, "C", false, 0, "")
	theme.Font(pdf, "", theme.SizeSubtitle)
	theme.TextColor(pdf, theme.ColorSubtext)
	pdf.CellFormat(0, 8, "Teilnehmende Ortsverbände", "", 1, "C", false, 0, "")
	pdf.Ln(8)

	// Table header — no Platz column
	theme.Font(pdf, "B", theme.SizeTableHeader)
	theme.FillColor(pdf, theme.ColorSecondary)
	theme.TextColor(pdf, theme.ColorOnHeader)

	colWidths := []float64{110, 40}
	for i, header := range []string{"Ortsverband", "Teilnehmende"} {
		pdf.CellFormat(colWidths[i], 10, enc(header), "1", 0, "C", true, 0, "")
	}
	pdf.Ln(-1)

	// Table rows — winner row highlighted, rest uniform (no rank numbers)
	theme.TextColor(pdf, theme.ColorText)

	topScore := evaluations[0].AverageScore
	for _, eval := range evaluations {
		isWinner := eval.AverageScore == topScore
		if isWinner {
			theme.Font(pdf, "B", theme.SizeBody)
			theme.FillColor(pdf, theme.ColorTableHighlight)
		} else {
			theme.Font(pdf, "", theme.SizeBody)
			theme.FillColor(pdf, [3]int{255, 255, 255})
		}
		name := enc(eval.Ortsverband)
		if isWinner {
			name = enc("★ " + eval.Ortsverband + " – Bester Ortsverband")
		}

		pdf.CellFormat(colWidths[0], 9, name, "1", 0, "L", isWinner, 0, "")
		pdf.CellFormat(colWidths[1], 9, fmt.Sprintf("%d", eval.ParticipantCount), "1", 0, "C", isWinner, 0, "")
		pdf.Ln(-1)
	}

	// Summary
	pdf.Ln(10)
	theme.Font(pdf, "B", theme.SizeHeading)
	theme.TextColor(pdf, theme.ColorText)
	pdf.CellFormat(0, 8, "Zusammenfassung", "", 1, "L", false, 0, "")
	pdf.Ln(3)
	theme.Font(pdf, "", theme.SizeBody)
	pdf.CellFormat(0, 6, enc(fmt.Sprintf("Ortsverbände gesamt: %d", len(evaluations))), "", 1, "L", false, 0, "")
	if len(evaluations) > 0 {
		var winners []string
		for _, e := range evaluations {
			if e.AverageScore == topScore {
				winners = append(winners, e.Ortsverband)
			}
		}
		pdf.CellFormat(0, 6, enc(fmt.Sprintf("Bester Ortsverband: %s", strings.Join(winners, ", "))), "", 1, "L", false, 0, "")
	}

	if err = pdf.OutputFileAndClose(filepath.Join(pdfOutputDir, "Auswertung_nach_Ortsverband.pdf")); err != nil {
		return fmt.Errorf("failed to save PDF: %w", err)
	}
	return nil
}
