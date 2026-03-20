package main

import (
	"fmt"
	"os"

	"THW-JugendOlympiade/backend/io"
)

// GenerateParticipantCertificates generates the participant certificates PDF.
func (a *App) GenerateParticipantCertificates() map[string]interface{} {
	if a.db == nil {
		return map[string]interface{}{
			"status":  "error",
			"message": "Bitte zuerst eine Excel-Datei laden.",
		}
	}

	if err := io.GenerateParticipantCertificates(a.db); err != nil {
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
// The best-ranked Ortsverband receives a special Siegerurkunde; all others get
// an identical participation certificate with no ranking mentioned.
func (a *App) GenerateOrtsverbandCertificates() map[string]interface{} {
	if a.db == nil {
		return map[string]interface{}{
			"status":  "error",
			"message": "Bitte zuerst eine Excel-Datei laden.",
		}
	}

	if err := io.GenerateOrtsverbandCertificates(a.db); err != nil {
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
