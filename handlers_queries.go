package main

import (
	"fmt"

	"THW-JugendOlympiade/backend/database"
)

// ShowGroups retrieves and returns groups from the database.
func (a *App) ShowGroups() map[string]interface{} {
	if a.db == nil {
		return map[string]interface{}{
			"status":  "error",
			"message": "Bitte zuerst eine Excel-Datei laden.",
		}
	}

	groups, err := database.GetGroupsForReport(a.db)
	if err != nil {
		return map[string]interface{}{
			"status":  "error",
			"message": fmt.Sprintf("Gruppen konnten nicht abgerufen werden: %v", err),
		}
	}

	if len(groups) == 0 {
		return map[string]interface{}{
			"status":  "error",
			"message": "Keine Gruppen gefunden. Bitte zuerst eine Datei laden.",
		}
	}

	return map[string]interface{}{
		"status": "success",
		"count":  len(groups),
		"groups": groups,
	}
}

// ShowStations retrieves and returns stations with group scores from the database.
func (a *App) ShowStations() map[string]interface{} {
	if a.db == nil {
		return map[string]interface{}{
			"status":  "error",
			"message": "Bitte zuerst eine Excel-Datei laden.",
		}
	}

	stations, err := database.GetStationsForReport(a.db)
	if err != nil {
		return map[string]interface{}{
			"status":  "error",
			"message": fmt.Sprintf("Stationen konnten nicht abgerufen werden: %v", err),
		}
	}

	if len(stations) == 0 {
		return map[string]interface{}{
			"status":  "error",
			"message": "Keine Stationen gefunden. Bitte sicherstellen, dass die Excel-Datei ein 'Stationen'-Blatt enthält.",
		}
	}

	return map[string]interface{}{
		"status":   "success",
		"count":    len(stations),
		"stations": stations,
	}
}

// GetAllGroups retrieves all group IDs from the database.
func (a *App) GetAllGroups() map[string]interface{} {
	if a.db == nil {
		return map[string]interface{}{
			"status":  "error",
			"message": "Bitte zuerst eine Excel-Datei laden.",
		}
	}

	groupIDs, err := database.GetAllGroupIDs(a.db)
	if err != nil {
		return map[string]interface{}{
			"status":  "error",
			"message": fmt.Sprintf("Gruppen konnten nicht abgerufen werden: %v", err),
		}
	}

	return map[string]interface{}{
		"status": "success",
		"groups": groupIDs,
	}
}

// AssignScore assigns a score to a group at a station.
func (a *App) AssignScore(groupID int, stationID int, score int) map[string]interface{} {
	if a.db == nil {
		return map[string]interface{}{
			"status":  "error",
			"message": "Bitte zuerst eine Excel-Datei laden.",
		}
	}

	if score < a.cfg.Ergebnisse.MinPunkte || score > a.cfg.Ergebnisse.MaxPunkte {
		return map[string]interface{}{
			"status":  "error",
			"message": fmt.Sprintf("Ungültiges Ergebnis %d: muss zwischen %d und %d liegen", score, a.cfg.Ergebnisse.MinPunkte, a.cfg.Ergebnisse.MaxPunkte),
		}
	}

	if err := database.AssignGroupStationScore(a.db, groupID, stationID, score); err != nil {
		return map[string]interface{}{
			"status":  "error",
			"message": fmt.Sprintf("Ergebnis konnte nicht gespeichert werden: %v", err),
		}
	}

	return map[string]interface{}{
		"status":  "success",
		"message": fmt.Sprintf("Ergebnis %d für Gruppe %d gespeichert", score, groupID),
	}
}

// GetGroupEvaluations retrieves all groups with their total scores ranked from high to low.
func (a *App) GetGroupEvaluations() map[string]interface{} {
	if a.db == nil {
		return map[string]interface{}{
			"status":  "error",
			"message": "Bitte zuerst eine Excel-Datei laden.",
		}
	}

	evaluations, err := database.GetGroupEvaluations(a.db)
	if err != nil {
		return map[string]interface{}{
			"status":  "error",
			"message": fmt.Sprintf("Gruppenauswertungen konnten nicht abgerufen werden: %v", err),
		}
	}

	if len(evaluations) == 0 {
		return map[string]interface{}{
			"status":  "error",
			"message": "Keine Gruppen mit Ergebnissen gefunden.",
		}
	}

	return map[string]interface{}{
		"status":      "success",
		"evaluations": evaluations,
	}
}

// GetOrtsverbandEvaluations retrieves all ortsverbands with their average scores ranked from high to low.
func (a *App) GetOrtsverbandEvaluations() map[string]interface{} {
	if a.db == nil {
		return map[string]interface{}{
			"status":  "error",
			"message": "Bitte zuerst eine Excel-Datei laden.",
		}
	}

	evaluations, err := database.GetOrtsverbandEvaluations(a.db)
	if err != nil {
		return map[string]interface{}{
			"status":  "error",
			"message": fmt.Sprintf("Ortsverband-Auswertungen konnten nicht abgerufen werden: %v", err),
		}
	}

	if len(evaluations) == 0 {
		return map[string]interface{}{
			"status":  "error",
			"message": "Keine Ortsverbände mit Ergebnissen gefunden.",
		}
	}

	return map[string]interface{}{
		"status":      "success",
		"evaluations": evaluations,
	}
}
