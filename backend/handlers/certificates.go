package handlers

import (
	"database/sql"
	"fmt"
	"os"

	"THW-JugendOlympiade/backend/io"
)

// GenerateParticipantCertificates generates the participant certificates PDF.
func GenerateParticipantCertificates(db *sql.DB, eventYear int, certStyle string, pictureDir string, eventLocation string) map[string]interface{} {
	if db == nil {
		return map[string]interface{}{
			"status":  "error",
			"message": "Bitte zuerst eine Excel-Datei laden.",
		}
	}
	if err := io.GenerateParticipantCertificates(db, eventYear, certStyle, pictureDir, eventLocation); err != nil {
		return map[string]interface{}{
			"status":  "error",
			"message": fmt.Sprintf("Urkunden konnten nicht erstellt werden: %v", err),
		}
	}
	absPath, _ := os.Getwd()
	return map[string]interface{}{
		"status":  "success",
		"message": "Urkunden Teilnehmende erfolgreich erstellt",
		"file":    "Urkunden_Teilnehmende.pdf",
		"path":    absPath + string(os.PathSeparator) + "pdfdocs" + string(os.PathSeparator) + "Urkunden_Teilnehmende.pdf",
	}
}

// GenerateOrtsverbandCertificates generates ortsverband certificates PDF.
func GenerateOrtsverbandCertificates(db *sql.DB, eventYear int, eventName string) map[string]interface{} {
	if db == nil {
		return map[string]interface{}{
			"status":  "error",
			"message": "Bitte zuerst eine Excel-Datei laden.",
		}
	}
	if err := io.GenerateOrtsverbandCertificates(db, eventYear, eventName); err != nil {
		return map[string]interface{}{
			"status":  "error",
			"message": fmt.Sprintf("Ortsverband-Urkunden konnten nicht erstellt werden: %v", err),
		}
	}
	absPath, _ := os.Getwd()
	return map[string]interface{}{
		"status":  "success",
		"message": "Urkunden Ortsverbände erfolgreich erstellt",
		"file":    "Urkunden_Ortsverbaende.pdf",
		"path":    absPath + string(os.PathSeparator) + "pdfdocs" + string(os.PathSeparator) + "Urkunden_Ortsverbaende.pdf",
	}
}
