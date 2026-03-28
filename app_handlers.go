package main

import (
	"THW-JugendOlympiade/backend/handlers"
	"THW-JugendOlympiade/backend/io"
)

// --- Config ---

func (a *App) GetConfig() map[string]interface{} {
	return handlers.GetConfig(a.cfg)
}

func (a *App) GetConfigRaw() map[string]interface{} {
	return handlers.GetConfigRaw()
}

func (a *App) SaveConfigRaw(content string) map[string]interface{} {
	cfg, result := handlers.SaveConfigRaw(content)
	if result["status"] == "ok" {
		a.cfg = cfg
		io.SetPDFOutputDir(cfg.Ausgabe.PDFOrdner)
	}
	return result
}

func (a *App) GetCertLayoutRaw() map[string]interface{} {
	return handlers.GetCertLayoutRaw()
}

func (a *App) SaveCertLayoutRaw(content string) map[string]interface{} {
	return handlers.SaveCertLayoutRaw(content)
}

func (a *App) GetCertLayoutJSON() map[string]interface{} {
	return handlers.GetCertLayoutJSON()
}

func (a *App) SaveCertLayoutJSON(jsonData string) map[string]interface{} {
	return handlers.SaveCertLayoutJSON(jsonData)
}

func (a *App) ListBackgroundImages() map[string]interface{} {
	return handlers.ListBackgroundImages()
}

func (a *App) GetImageAsBase64(filename string) map[string]interface{} {
	return handlers.GetImageAsBase64(filename)
}

func (a *App) ListGroupPictures() map[string]interface{} {
	return handlers.ListGroupPictures(a.cfg.Ausgabe.BilderOrdner)
}

func (a *App) GetGroupPictureAsBase64(filename string) map[string]interface{} {
	return handlers.GetGroupPictureAsBase64(a.cfg.Ausgabe.BilderOrdner, filename)
}

// --- Files / Startup ---

func (a *App) CheckStartup() map[string]interface{} {
	return handlers.CheckStartup()
}

func (a *App) UseExistingDB() map[string]interface{} {
	return handlers.UseExistingDB(&a.db)
}

func (a *App) ResetToFreshDB() map[string]interface{} {
	return handlers.ResetToFreshDB(&a.db)
}

func (a *App) CheckDB() map[string]interface{} {
	return handlers.CheckDB(a.db)
}

func (a *App) LoadFile() map[string]interface{} {
	return handlers.LoadFile(a.ctx, &a.db)
}

func (a *App) HasScores() (bool, error) {
	return handlers.HasScores(a.db)
}

func (a *App) DistributeGroups() map[string]interface{} {
	return handlers.DistributeGroups(a.db, a.cfg.Gruppen.MaxGroesse)
}

// --- Queries ---

func (a *App) ShowGroups() map[string]interface{} {
	return handlers.ShowGroups(a.db)
}

func (a *App) ShowStations() map[string]interface{} {
	return handlers.ShowStations(a.db)
}

func (a *App) GetAllGroups() map[string]interface{} {
	return handlers.GetAllGroups(a.db)
}

func (a *App) AssignScore(groupID int, stationID int, score int) map[string]interface{} {
	return handlers.AssignScore(a.db, groupID, stationID, score, a.cfg.Ergebnisse.MinPunkte, a.cfg.Ergebnisse.MaxPunkte)
}

func (a *App) GetGroupEvaluations() map[string]interface{} {
	return handlers.GetGroupEvaluations(a.db)
}

func (a *App) GetOrtsverbandEvaluations() map[string]interface{} {
	return handlers.GetOrtsverbandEvaluations(a.db)
}

// --- Reports ---

func (a *App) GeneratePDF() map[string]interface{} {
	return handlers.GeneratePDF(a.db, a.cfg.Veranstaltung.Name, a.cfg.Veranstaltung.Jahr)
}

func (a *App) GenerateGroupEvaluationPDF() map[string]interface{} {
	return handlers.GenerateGroupEvaluationPDF(a.db)
}

func (a *App) GenerateOrtsverbandEvaluationPDF() map[string]interface{} {
	return handlers.GenerateOrtsverbandEvaluationPDF(a.db)
}

// --- Certificates ---

func (a *App) GenerateParticipantCertificates() map[string]interface{} {
	return handlers.GenerateParticipantCertificates(a.db, a.cfg.Veranstaltung.Jahr, a.cfg.Ausgabe.UrkunderStil, a.cfg.Ausgabe.BilderOrdner)
}

func (a *App) GenerateOrtsverbandCertificates() map[string]interface{} {
	return handlers.GenerateOrtsverbandCertificates(a.db, a.cfg.Veranstaltung.Jahr, a.cfg.Veranstaltung.Name)
}

// --- Backup ---

func (a *App) BackupDatabase() map[string]interface{} {
	return handlers.BackupDatabase(a.db)
}

func (a *App) ListBackups() map[string]interface{} {
	return handlers.ListBackups()
}

func (a *App) RestoreDatabase(backupFilename string) map[string]interface{} {
	return handlers.RestoreDatabase(&a.db, backupFilename)
}
