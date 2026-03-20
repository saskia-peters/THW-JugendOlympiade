package main

import (
	"fmt"
	"os"

	"THW-JugendOlympiade/backend/io"
)

// GeneratePDF generates a groups PDF report.
func (a *App) GeneratePDF() map[string]interface{} {
	if a.db == nil {
		return map[string]interface{}{
			"status":  "error",
			"message": "Bitte zuerst eine Excel-Datei laden.",
		}
	}

	if err := io.GeneratePDFReport(a.db); err != nil {
		return map[string]interface{}{
			"status":  "error",
			"message": fmt.Sprintf("Gruppen-PDF konnte nicht erstellt werden: %v", err),
		}
	}

	absPath, _ := os.Getwd()
	return map[string]interface{}{
		"status":  "success",
		"message": "Gruppen-PDF erfolgreich erstellt",
		"file":    "Gruppeneinteilung.pdf",
		"path":    absPath + string(os.PathSeparator) + "pdfdocs" + string(os.PathSeparator) + "Gruppeneinteilung.pdf",
	}
}

// GenerateGroupEvaluationPDF generates a PDF report for group rankings.
func (a *App) GenerateGroupEvaluationPDF() map[string]interface{} {
	if a.db == nil {
		return map[string]interface{}{
			"status":  "error",
			"message": "Bitte zuerst eine Excel-Datei laden.",
		}
	}

	if err := io.GenerateGroupEvaluationPDF(a.db); err != nil {
		return map[string]interface{}{
			"status":  "error",
			"message": fmt.Sprintf("Auswertungs-PDF konnte nicht erstellt werden: %v", err),
		}
	}

	absPath, _ := os.Getwd()
	return map[string]interface{}{
		"status":  "success",
		"message": "Auswertungs-PDF erfolgreich erstellt",
		"file":    "Auswertung_nach_Gruppe.pdf",
		"path":    absPath + string(os.PathSeparator) + "pdfdocs" + string(os.PathSeparator) + "Auswertung_nach_Gruppe.pdf",
	}
}

// GenerateOrtsverbandEvaluationPDF generates a PDF report for ortsverband rankings.
func (a *App) GenerateOrtsverbandEvaluationPDF() map[string]interface{} {
	if a.db == nil {
		return map[string]interface{}{
			"status":  "error",
			"message": "Bitte zuerst eine Excel-Datei laden.",
		}
	}

	if err := io.GenerateOrtsverbandEvaluationPDF(a.db); err != nil {
		return map[string]interface{}{
			"status":  "error",
			"message": fmt.Sprintf("Ortsverband-Auswertungs-PDF konnte nicht erstellt werden: %v", err),
		}
	}

	absPath, _ := os.Getwd()
	return map[string]interface{}{
		"status":  "success",
		"message": "Ortsverband-Auswertungs-PDF erfolgreich erstellt",
		"file":    "Auswertung_nach_Ortsverband.pdf",
		"path":    absPath + string(os.PathSeparator) + "pdfdocs" + string(os.PathSeparator) + "Auswertung_nach_Ortsverband.pdf",
	}
}
