package models

const (
	DbFile            = "data.db"
	XlsxFile          = "data.xlsx"
	SheetName         = "Teilnehmer"
	StationsSheetName = "Stationen"
	TableName         = "teilnehmer"
	MaxGroupSize      = 8
)

// Teilnehmer represents a participant
type Teilnehmer struct {
	ID           int
	TeilnehmerID int
	Name         string
	Ortsverband  string
	Alter        int
	Geschlecht   string
}

// Group represents a group of participants
type Group struct {
	GroupID      int
	Teilnehmers  []Teilnehmer
	Ortsverbands map[string]int
	Geschlechts  map[string]int
	AlterSum     int
}

// GroupScore represents a group's score at a station
type GroupScore struct {
	GroupID int
	Score   int
}

// Station represents a station with groups that visited and their scores
type Station struct {
	StationID   int
	StationName string
	GroupScores []GroupScore
}

// GroupEvaluation represents a group's total score across all stations
type GroupEvaluation struct {
	GroupID      int
	TotalScore   int
	StationCount int
}

// OrtsverbandEvaluation represents an ortsverband's total score across all stations
type OrtsverbandEvaluation struct {
	Ortsverband     string
	TotalScore      int
	StationsVisited int
}
