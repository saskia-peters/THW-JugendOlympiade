package io

import (
	"database/sql"
	"fmt"
	"path/filepath"

	"experiment1/backend/database"

	"github.com/jung-kurt/gofpdf"
)

// GeneratePDFReport creates a PDF report with one group per page.
func GeneratePDFReport(db *sql.DB) error {
	if err := ensurePDFDirectory(); err != nil {
		return err
	}

	groups, err := database.GetGroupsForReport(db)
	if err != nil {
		return fmt.Errorf("failed to get groups: %w", err)
	}
	if len(groups) == 0 {
		return fmt.Errorf("no groups found to generate report")
	}

	theme := DefaultTheme
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.SetAutoPageBreak(true, 15)

	for _, group := range groups {
		pdf.AddPage()

		// Title
		theme.Font(pdf, "B", theme.SizeTitle)
		theme.TextColor(pdf, theme.ColorText)
		pdf.CellFormat(0, 15, fmt.Sprintf("Gruppe %d", group.GroupID), "", 1, "C", false, 0, "")
		pdf.Ln(5)

		// Participant count
		theme.Font(pdf, "", theme.SizeBody)
		theme.TextColor(pdf, theme.ColorSubtext)
		pdf.CellFormat(0, 8, fmt.Sprintf("Anzahl Teilnehmer: %d", len(group.Teilnehmers)), "", 1, "L", false, 0, "")
		pdf.Ln(3)

		// Table header
		theme.Font(pdf, "B", theme.SizeTableHeader)
		theme.FillColor(pdf, theme.ColorTableHeader)
		theme.TextColor(pdf, theme.ColorText)

		colWidths := []float64{50, 50, 30, 40}
		for i, header := range []string{"Name", "Ortsverband", "Alter", "Geschlecht"} {
			pdf.CellFormat(colWidths[i], 10, header, "1", 0, "C", true, 0, "")
		}
		pdf.Ln(-1)

		// Table rows
		theme.Font(pdf, "", theme.SizeBody)
		theme.FillColor(pdf, theme.ColorTableRowAlt)

		for i, t := range group.Teilnehmers {
			fill := i%2 == 0
			pdf.CellFormat(colWidths[0], 9, t.Name, "1", 0, "L", fill, 0, "")
			pdf.CellFormat(colWidths[1], 9, t.Ortsverband, "1", 0, "L", fill, 0, "")
			pdf.CellFormat(colWidths[2], 9, fmt.Sprintf("%d", t.Alter), "1", 0, "C", fill, 0, "")
			pdf.CellFormat(colWidths[3], 9, t.Geschlecht, "1", 0, "C", fill, 0, "")
			pdf.Ln(-1)
		}

		// Group statistics footer
		pdf.Ln(8)
		theme.Font(pdf, "I", theme.SizeBody)
		theme.TextColor(pdf, theme.ColorSubtext)

		ortsverbandStats := make(map[string]int)
		geschlechtStats := make(map[string]int)
		alterSum, alterCount := 0, 0
		for _, t := range group.Teilnehmers {
			ortsverbandStats[t.Ortsverband]++
			geschlechtStats[t.Geschlecht]++
			if t.Alter > 0 {
				alterSum += t.Alter
				alterCount++
			}
		}

		pdf.CellFormat(0, 6, "Gruppenstatistik:", "", 1, "L", false, 0, "")
		theme.Font(pdf, "", theme.SizeSmall)

		ovStr := "Ortsverband: "
		first := true
		for ov, count := range ortsverbandStats {
			if !first {
				ovStr += ", "
			}
			ovStr += fmt.Sprintf("%s (%d)", ov, count)
			first = false
		}
		pdf.CellFormat(0, 5, ovStr, "", 1, "L", false, 0, "")

		gStr := "Geschlecht: "
		first = true
		for g, count := range geschlechtStats {
			if !first {
				gStr += ", "
			}
			gStr += fmt.Sprintf("%s (%d)", g, count)
			first = false
		}
		pdf.CellFormat(0, 5, gStr, "", 1, "L", false, 0, "")

		if alterCount > 0 {
			pdf.CellFormat(0, 5, fmt.Sprintf("Durchschnittsalter: %.1f Jahre", float64(alterSum)/float64(alterCount)), "", 1, "L", false, 0, "")
		}
	}

	if err = pdf.OutputFileAndClose(filepath.Join(pdfOutputDir, "groups_report.pdf")); err != nil {
		return fmt.Errorf("failed to save PDF: %w", err)
	}
	return nil
}
