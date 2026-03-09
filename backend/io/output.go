package io

import (
	"database/sql"
	"fmt"

	"experiment1/backend/database"

	"github.com/jung-kurt/gofpdf"
)

// GeneratePDFReport creates a PDF report with one group per page
func GeneratePDFReport(db *sql.DB) error {
	// Get all groups with their participants
	groups, err := database.GetGroupsForReport(db)
	if err != nil {
		return fmt.Errorf("failed to get groups: %w", err)
	}

	if len(groups) == 0 {
		return fmt.Errorf("no groups found to generate report")
	}

	// Initialize PDF with A4 portrait
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.SetAutoPageBreak(true, 15)

	// Add each group on a separate page
	for _, group := range groups {
		pdf.AddPage()

		// Title
		pdf.SetFont("Arial", "B", 24)
		pdf.CellFormat(0, 15, fmt.Sprintf("Gruppe %d", group.GroupID), "", 1, "C", false, 0, "")
		pdf.Ln(5)

		// Group statistics
		pdf.SetFont("Arial", "", 11)
		pdf.SetTextColor(100, 100, 100)
		pdf.CellFormat(0, 8, fmt.Sprintf("Anzahl Teilnehmer: %d", len(group.Teilnehmers)), "", 1, "L", false, 0, "")
		pdf.Ln(3)

		// Table header
		pdf.SetFont("Arial", "B", 11)
		pdf.SetFillColor(200, 200, 200)
		pdf.SetTextColor(0, 0, 0)

		colWidths := []float64{50, 50, 30, 40}
		headers := []string{"Name", "Ortsverband", "Alter", "Geschlecht"}

		for i, header := range headers {
			pdf.CellFormat(colWidths[i], 10, header, "1", 0, "C", true, 0, "")
		}
		pdf.Ln(-1)

		// Table rows
		pdf.SetFont("Arial", "", 10)
		pdf.SetFillColor(240, 240, 240)

		for i, teilnehmer := range group.Teilnehmers {
			fill := i%2 == 0

			pdf.CellFormat(colWidths[0], 9, teilnehmer.Name, "1", 0, "L", fill, 0, "")
			pdf.CellFormat(colWidths[1], 9, teilnehmer.Ortsverband, "1", 0, "L", fill, 0, "")
			pdf.CellFormat(colWidths[2], 9, fmt.Sprintf("%d", teilnehmer.Alter), "1", 0, "C", fill, 0, "")
			pdf.CellFormat(colWidths[3], 9, teilnehmer.Geschlecht, "1", 0, "C", fill, 0, "")
			pdf.Ln(-1)
		}

		// Group statistics at bottom
		pdf.Ln(8)
		pdf.SetFont("Arial", "I", 10)
		pdf.SetTextColor(80, 80, 80)

		// Calculate statistics
		ortsverbandStats := make(map[string]int)
		geschlechtStats := make(map[string]int)
		alterSum := 0
		alterCount := 0

		for _, t := range group.Teilnehmers {
			ortsverbandStats[t.Ortsverband]++
			geschlechtStats[t.Geschlecht]++
			if t.Alter > 0 {
				alterSum += t.Alter
				alterCount++
			}
		}

		pdf.CellFormat(0, 6, "Gruppenstatistik:", "", 1, "L", false, 0, "")

		// Ortsverband distribution
		pdf.SetFont("Arial", "", 9)
		ortsverbandStr := "Ortsverband: "
		first := true
		for ov, count := range ortsverbandStats {
			if !first {
				ortsverbandStr += ", "
			}
			ortsverbandStr += fmt.Sprintf("%s (%d)", ov, count)
			first = false
		}
		pdf.CellFormat(0, 5, ortsverbandStr, "", 1, "L", false, 0, "")

		// Geschlecht distribution
		geschlechtStr := "Geschlecht: "
		first = true
		for g, count := range geschlechtStats {
			if !first {
				geschlechtStr += ", "
			}
			geschlechtStr += fmt.Sprintf("%s (%d)", g, count)
			first = false
		}
		pdf.CellFormat(0, 5, geschlechtStr, "", 1, "L", false, 0, "")

		// Average age
		if alterCount > 0 {
			avgAlter := float64(alterSum) / float64(alterCount)
			pdf.CellFormat(0, 5, fmt.Sprintf("Durchschnittsalter: %.1f Jahre", avgAlter), "", 1, "L", false, 0, "")
		}
	}

	// Save PDF
	err = pdf.OutputFileAndClose("groups_report.pdf")
	if err != nil {
		return fmt.Errorf("failed to save PDF: %w", err)
	}

	return nil
}

// GenerateGroupEvaluationPDF creates a PDF report with group rankings and scores
func GenerateGroupEvaluationPDF(db *sql.DB) error {
	// Get group evaluations
	evaluations, err := database.GetGroupEvaluations(db)
	if err != nil {
		return fmt.Errorf("failed to get evaluations: %w", err)
	}

	if len(evaluations) == 0 {
		return fmt.Errorf("no group evaluations found to generate report")
	}

	// Initialize PDF with A4 portrait
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.SetAutoPageBreak(true, 15)
	pdf.AddPage()

	// Title
	pdf.SetFont("Arial", "B", 24)
	pdf.CellFormat(0, 15, "Gruppenauswertung", "", 1, "C", false, 0, "")
	pdf.SetFont("Arial", "", 12)
	pdf.SetTextColor(100, 100, 100)
	pdf.CellFormat(0, 8, "Ranking nach Gesamtpunktzahl", "", 1, "C", false, 0, "")
	pdf.Ln(8)

	// Table header
	pdf.SetFont("Arial", "B", 12)
	pdf.SetFillColor(102, 126, 234) // Purple color
	pdf.SetTextColor(255, 255, 255)

	colWidths := []float64{30, 70, 45, 45}
	headers := []string{"Platz", "Gruppe", "Stationen", "Gesamtscore"}

	for i, header := range headers {
		pdf.CellFormat(colWidths[i], 10, header, "1", 0, "C", true, 0, "")
	}
	pdf.Ln(-1)

	// Table rows
	pdf.SetFont("Arial", "", 11)
	pdf.SetTextColor(0, 0, 0)
	pdf.SetFillColor(255, 243, 205) // Light yellow for top 3

	for i, eval := range evaluations {
		rank := fmt.Sprintf("%d", i+1)
		if i == 0 {
			rank = "1"
		} else if i == 1 {
			rank = "2"
		} else if i == 2 {
			rank = "3"
		}

		fill := i < 3
		if fill {
			pdf.SetFont("Arial", "B", 11)
		} else {
			pdf.SetFont("Arial", "", 11)
		}

		pdf.CellFormat(colWidths[0], 9, rank, "1", 0, "C", fill, 0, "")
		pdf.CellFormat(colWidths[1], 9, fmt.Sprintf("Gruppe %d", eval.GroupID), "1", 0, "L", fill, 0, "")
		pdf.CellFormat(colWidths[2], 9, fmt.Sprintf("%d", eval.StationCount), "1", 0, "C", fill, 0, "")
		pdf.CellFormat(colWidths[3], 9, fmt.Sprintf("%d", eval.TotalScore), "1", 0, "C", fill, 0, "")
		pdf.Ln(-1)
	}

	// Statistics summary
	pdf.Ln(10)
	pdf.SetFont("Arial", "B", 14)
	pdf.SetTextColor(0, 0, 0)
	pdf.CellFormat(0, 8, "Zusammenfassung", "", 1, "L", false, 0, "")
	pdf.Ln(3)

	pdf.SetFont("Arial", "", 11)
	pdf.CellFormat(0, 6, fmt.Sprintf("Gesamtanzahl Gruppen: %d", len(evaluations)), "", 1, "L", false, 0, "")

	if len(evaluations) > 0 {
		pdf.CellFormat(0, 6, fmt.Sprintf("Hochster Score: %d (Gruppe %d)", evaluations[0].TotalScore, evaluations[0].GroupID), "", 1, "L", false, 0, "")

		lastEval := evaluations[len(evaluations)-1]
		pdf.CellFormat(0, 6, fmt.Sprintf("Niedrigster Score: %d (Gruppe %d)", lastEval.TotalScore, lastEval.GroupID), "", 1, "L", false, 0, "")

		totalScore := 0
		for _, e := range evaluations {
			totalScore += e.TotalScore
		}
		avgScore := float64(totalScore) / float64(len(evaluations))
		pdf.CellFormat(0, 6, fmt.Sprintf("Durchschnittlicher Score: %.1f", avgScore), "", 1, "L", false, 0, "")
	}

	// Save PDF
	err = pdf.OutputFileAndClose("group_evaluations.pdf")
	if err != nil {
		return fmt.Errorf("failed to save PDF: %w", err)
	}

	return nil
}

// GenerateOrtsverbandEvaluationPDF creates a PDF report with ortsverband rankings and average scores
func GenerateOrtsverbandEvaluationPDF(db *sql.DB) error {
	// Get ortsverband evaluations
	evaluations, err := database.GetOrtsverbandEvaluations(db)
	if err != nil {
		return fmt.Errorf("failed to get evaluations: %w", err)
	}

	if len(evaluations) == 0 {
		return fmt.Errorf("no ortsverband evaluations found to generate report")
	}

	// Initialize PDF with A4 portrait
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.SetAutoPageBreak(true, 15)
	pdf.AddPage()

	// Title
	pdf.SetFont("Arial", "B", 24)
	pdf.CellFormat(0, 15, "Ortsverband-Auswertung", "", 1, "C", false, 0, "")
	pdf.SetFont("Arial", "", 12)
	pdf.SetTextColor(100, 100, 100)
	pdf.CellFormat(0, 8, "Ranking nach Durchschnittspunktzahl", "", 1, "C", false, 0, "")
	pdf.Ln(8)

	// Table header
	pdf.SetFont("Arial", "B", 11)
	pdf.SetFillColor(250, 112, 154) // Pink color
	pdf.SetTextColor(255, 255, 255)

	colWidths := []float64{25, 65, 35, 35, 40}
	headers := []string{"Platz", "Ortsverband", "Teiln.", "Gesamt", "O Score"}

	for i, header := range headers {
		pdf.CellFormat(colWidths[i], 10, header, "1", 0, "C", true, 0, "")
	}
	pdf.Ln(-1)

	// Table rows
	pdf.SetFont("Arial", "", 10)
	pdf.SetTextColor(0, 0, 0)
	pdf.SetFillColor(255, 243, 205) // Light yellow for top 3

	for i, eval := range evaluations {
		rank := fmt.Sprintf("%d", i+1)

		fill := i < 3
		if fill {
			pdf.SetFont("Arial", "B", 10)
		} else {
			pdf.SetFont("Arial", "", 10)
		}

		pdf.CellFormat(colWidths[0], 9, rank, "1", 0, "C", fill, 0, "")
		pdf.CellFormat(colWidths[1], 9, eval.Ortsverband, "1", 0, "L", fill, 0, "")
		pdf.CellFormat(colWidths[2], 9, fmt.Sprintf("%d", eval.ParticipantCount), "1", 0, "C", fill, 0, "")
		pdf.CellFormat(colWidths[3], 9, fmt.Sprintf("%d", eval.TotalScore), "1", 0, "C", fill, 0, "")
		pdf.CellFormat(colWidths[4], 9, fmt.Sprintf("%.1f", eval.AverageScore), "1", 0, "C", fill, 0, "")
		pdf.Ln(-1)
	}

	// Statistics summary
	pdf.Ln(10)
	pdf.SetFont("Arial", "B", 14)
	pdf.SetTextColor(0, 0, 0)
	pdf.CellFormat(0, 8, "Zusammenfassung", "", 1, "L", false, 0, "")
	pdf.Ln(3)

	pdf.SetFont("Arial", "", 11)
	pdf.CellFormat(0, 6, fmt.Sprintf("Gesamtanzahl Ortsverbande: %d", len(evaluations)), "", 1, "L", false, 0, "")

	if len(evaluations) > 0 {
		pdf.CellFormat(0, 6, fmt.Sprintf("Hochster O-Score: %.1f (%s)", evaluations[0].AverageScore, evaluations[0].Ortsverband), "", 1, "L", false, 0, "")

		lastEval := evaluations[len(evaluations)-1]
		pdf.CellFormat(0, 6, fmt.Sprintf("Niedrigster O-Score: %.1f (%s)", lastEval.AverageScore, lastEval.Ortsverband), "", 1, "L", false, 0, "")

		totalAvg := 0.0
		for _, e := range evaluations {
			totalAvg += e.AverageScore
		}
		overallAvg := totalAvg / float64(len(evaluations))
		pdf.CellFormat(0, 6, fmt.Sprintf("Durchschnittlicher O-Score: %.1f", overallAvg), "", 1, "L", false, 0, "")
	}

	// Save PDF
	err = pdf.OutputFileAndClose("ortsverband_evaluations.pdf")
	if err != nil {
		return fmt.Errorf("failed to save PDF: %w", err)
	}

	return nil
}
