package io

import (
	"database/sql"
	"fmt"
	"path/filepath"

	"experiment1/backend/database"

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
		pdf.CellFormat(0, 6, fmt.Sprintf("Höchster Score: %d (Gruppe %d)", evaluations[0].TotalScore, evaluations[0].GroupID), "", 1, "L", false, 0, "")
		last := evaluations[len(evaluations)-1]
		pdf.CellFormat(0, 6, fmt.Sprintf("Niedrigster Score: %d (Gruppe %d)", last.TotalScore, last.GroupID), "", 1, "L", false, 0, "")
		total := 0
		for _, e := range evaluations {
			total += e.TotalScore
		}
		pdf.CellFormat(0, 6, fmt.Sprintf("Durchschnittlicher Score: %.1f", float64(total)/float64(len(evaluations))), "", 1, "L", false, 0, "")
	}

	if err = pdf.OutputFileAndClose(filepath.Join(pdfOutputDir, "group_evaluations.pdf")); err != nil {
		return fmt.Errorf("failed to save PDF: %w", err)
	}
	return nil
}

// GenerateOrtsverbandEvaluationPDF creates a PDF report with ortsverband rankings and average scores.
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
	pdf.CellFormat(0, 8, "Ranking nach Durchschnittspunktzahl", "", 1, "C", false, 0, "")
	pdf.Ln(8)

	// Table header
	theme.Font(pdf, "B", theme.SizeTableHeader)
	theme.FillColor(pdf, theme.ColorSecondary)
	theme.TextColor(pdf, theme.ColorOnHeader)

	colWidths := []float64{20, 60, 30, 35, 35}
	for i, header := range []string{"Platz", "Ortsverband", "Teiln.", "Gesamt", "Ø Score"} {
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
		pdf.CellFormat(colWidths[1], 9, eval.Ortsverband, "1", 0, "L", fill, 0, "")
		pdf.CellFormat(colWidths[2], 9, fmt.Sprintf("%d", eval.ParticipantCount), "1", 0, "C", fill, 0, "")
		pdf.CellFormat(colWidths[3], 9, fmt.Sprintf("%d", eval.TotalScore), "1", 0, "C", fill, 0, "")
		pdf.CellFormat(colWidths[4], 9, fmt.Sprintf("%.1f", eval.AverageScore), "1", 0, "C", fill, 0, "")
		pdf.Ln(-1)
	}

	// Summary
	pdf.Ln(10)
	theme.Font(pdf, "B", theme.SizeHeading)
	theme.TextColor(pdf, theme.ColorText)
	pdf.CellFormat(0, 8, "Zusammenfassung", "", 1, "L", false, 0, "")
	pdf.Ln(3)
	theme.Font(pdf, "", theme.SizeBody)
	pdf.CellFormat(0, 6, fmt.Sprintf("Gesamtanzahl Ortsverbände: %d", len(evaluations)), "", 1, "L", false, 0, "")
	if len(evaluations) > 0 {
		pdf.CellFormat(0, 6, fmt.Sprintf("Höchster Ø-Score: %.1f (%s)", evaluations[0].AverageScore, evaluations[0].Ortsverband), "", 1, "L", false, 0, "")
		last := evaluations[len(evaluations)-1]
		pdf.CellFormat(0, 6, fmt.Sprintf("Niedrigster Ø-Score: %.1f (%s)", last.AverageScore, last.Ortsverband), "", 1, "L", false, 0, "")
		totalAvg := 0.0
		for _, e := range evaluations {
			totalAvg += e.AverageScore
		}
		pdf.CellFormat(0, 6, fmt.Sprintf("Durchschnittlicher Ø-Score: %.1f", totalAvg/float64(len(evaluations))), "", 1, "L", false, 0, "")
	}

	if err = pdf.OutputFileAndClose(filepath.Join(pdfOutputDir, "ortsverband_evaluations.pdf")); err != nil {
		return fmt.Errorf("failed to save PDF: %w", err)
	}
	return nil
}
