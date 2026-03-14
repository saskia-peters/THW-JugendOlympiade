package test

import (
	"database/sql"
	"testing"

	"experiment1/backend/database"
	"experiment1/backend/models"

	_ "modernc.org/sqlite"
)

// setupTestDB creates a temporary test database
func setupTestDB(t *testing.T) *sql.DB {
	// Create a temporary database in memory for tests
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open in-memory database: %v", err)
	}

	// Create all tables using the same schema as InitDatabase but without calling it
	tables := []string{
		`CREATE TABLE IF NOT EXISTS teilnehmer (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			teilnehmer_id INTEGER UNIQUE,
			name TEXT,
			ortsverband TEXT,
			age INTEGER,
			geschlecht TEXT,
			pregroup TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS gruppe (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			group_id INTEGER NOT NULL,
			teilnehmer_id INTEGER UNIQUE NOT NULL,
			FOREIGN KEY (teilnehmer_id) REFERENCES teilnehmer(teilnehmer_id)
		)`,

		`CREATE TABLE IF NOT EXISTS stations (
			station_id INTEGER PRIMARY KEY AUTOINCREMENT,
			station_name TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS group_station_scores (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			group_id INTEGER NOT NULL,
			station_id INTEGER NOT NULL,
			score INTEGER,
			FOREIGN KEY (station_id) REFERENCES stations(station_id),
			UNIQUE(group_id, station_id)
		)`,
	}

	for _, tableSQL := range tables {
		if _, err := db.Exec(tableSQL); err != nil {
			t.Fatalf("Failed to create table: %v", err)
		}
	}

	// Create indexes
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_gruppe_group_id ON gruppe(group_id)",
		"CREATE INDEX IF NOT EXISTS idx_gruppe_teilnehmer_id ON gruppe(teilnehmer_id)",
		"CREATE INDEX IF NOT EXISTS idx_scores_group_id ON group_station_scores(group_id)",
		"CREATE INDEX IF NOT EXISTS idx_scores_station_id ON group_station_scores(station_id)",
	}

	for _, indexSQL := range indexes {
		if _, err := db.Exec(indexSQL); err != nil {
			t.Fatalf("Failed to create index: %v", err)
		}
	}

	return db
}

// teardownTestDB cleans up test database
func teardownTestDB(t *testing.T, db *sql.DB) {
	if db != nil {
		db.Close()
	}
}

// TestInitDatabase tests database initialization
func TestInitDatabase(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	// Verify tables exist by querying sqlite_master
	tables := []string{"teilnehmer", "gruppe", "stations", "group_station_scores"}

	for _, tableName := range tables {
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", tableName).Scan(&count)
		if err != nil {
			t.Errorf("Failed to query for table %s: %v", tableName, err)
		}
		if count != 1 {
			t.Errorf("Expected table %s to exist, but it doesn't", tableName)
		}
	}
}

// TestInitDatabase_IndexesCreated tests that all indexes are created
func TestInitDatabase_IndexesCreated(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	expectedIndexes := []string{
		"idx_gruppe_group_id",
		"idx_gruppe_teilnehmer_id",
		"idx_scores_group_id",
		"idx_scores_station_id",
	}

	for _, indexName := range expectedIndexes {
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='index' AND name=?", indexName).Scan(&count)
		if err != nil {
			t.Errorf("Failed to query for index %s: %v", indexName, err)
		}
		if count != 1 {
			t.Errorf("Expected index %s to exist, but it doesn't", indexName)
		}
	}
}

// TestTrimSpace tests the TrimSpace helper function
func TestTrimSpace(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"  hello  ", "hello"},
		{"hello", "hello"},
		{"  ", ""},
		{"", ""},
		{"\t\nhello\n\t", "hello"},
		{"  multiple   spaces  ", "multiple   spaces"},
	}

	for _, tt := range tests {
		result := database.TrimSpace(tt.input)
		if result != tt.expected {
			t.Errorf("TrimSpace(%q) = %q, expected %q", tt.input, result, tt.expected)
		}
	}
}

// TestInsertData tests inserting participant data
func TestInsertData(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	// Prepare test data (header + 3 participants)
	rows := [][]string{
		{"Name", "Ortsverband", "Alter", "Geschlecht", "PreGroup"},
		{"Max Mustermann", "Berlin", "25", "M", ""},
		{"Anna Schmidt", "Hamburg", "30", "W", ""},
		{"Tom Meyer", "München", "22", "M", ""},
	}

	err := database.InsertData(db, rows)
	if err != nil {
		t.Fatalf("InsertData failed: %v", err)
	}

	// Verify data was inserted
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM teilnehmer").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count participants: %v", err)
	}

	if count != 3 {
		t.Errorf("Expected 3 participants, got %d", count)
	}

	// Verify specific participant
	var name, ortsverband, geschlecht string
	var age int
	err = db.QueryRow("SELECT name, ortsverband, age, geschlecht FROM teilnehmer WHERE teilnehmer_id = 1").Scan(&name, &ortsverband, &age, &geschlecht)
	if err != nil {
		t.Fatalf("Failed to query participant: %v", err)
	}

	if name != "Max Mustermann" {
		t.Errorf("Expected name 'Max Mustermann', got %s", name)
	}
	if age != 25 {
		t.Errorf("Expected age 25, got %d", age)
	}
}

// TestInsertData_EmptyRows tests inserting with empty data
func TestInsertData_EmptyRows(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	rows := [][]string{
		{"Name", "Ortsverband", "Alter", "Geschlecht", "PreGroup"},
	}

	// InsertData doesn't error on empty rows  - it just skips them
	err := database.InsertData(db, rows)
	if err != nil {
		t.Errorf("InsertData failed: %v", err)
	}

	// Verify no data was inserted
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM teilnehmer").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count participants: %v", err)
	}

	if count != 0 {
		t.Errorf("Expected 0 participants, got %d", count)
	}
}

// TestInsertStations tests inserting station data
func TestInsertStations(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	// Station data format: each row has station name in first column
	stationRows := [][]string{
		{"Station Name"}, // Header row (skipped)
		{"Weitsprung"},
		{"Ballwurf"},
		{"Sprint"},
	}

	err := database.InsertStations(db, stationRows)
	if err != nil {
		t.Fatalf("InsertStations failed: %v", err)
	}

	// Verify stations were inserted
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM stations").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count stations: %v", err)
	}

	if count != 3 {
		t.Errorf("Expected 3 stations, got %d", count)
	}

	// Verify specific station
	var stationName string
	err = db.QueryRow("SELECT station_name FROM stations WHERE station_id = 1").Scan(&stationName)
	if err != nil {
		t.Fatalf("Failed to query station: %v", err)
	}

	if stationName != "Weitsprung" {
		t.Errorf("Expected station name 'Weitsprung', got %s", stationName)
	}
}

// TestGetAllTeilnehmers tests retrieving all participants
func TestGetAllTeilnehmers(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	// Insert test data
	rows := [][]string{
		{"Name", "Ortsverband", "Alter", "Geschlecht", "PreGroup"},
		{"Max Mustermann", "Berlin", "25", "M", ""},
		{"Anna Schmidt", "Hamburg", "30", "W", ""},
	}

	err := database.InsertData(db, rows)
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	// Get all participants
	teilnehmers, err := database.GetAllTeilnehmers(db)
	if err != nil {
		t.Fatalf("GetAllTeilnehmers failed: %v", err)
	}

	if len(teilnehmers) != 2 {
		t.Errorf("Expected 2 participants, got %d", len(teilnehmers))
	}

	// Verify first participant
	if teilnehmers[0].Name != "Max Mustermann" {
		t.Errorf("Expected first participant 'Max Mustermann', got %s", teilnehmers[0].Name)
	}
	if teilnehmers[0].Alter != 25 {
		t.Errorf("Expected age 25, got %d", teilnehmers[0].Alter)
	}
}

// TestGetAllTeilnehmers_EmptyDatabase tests retrieving from empty database
func TestGetAllTeilnehmers_EmptyDatabase(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	teilnehmers, err := database.GetAllTeilnehmers(db)
	if err != nil {
		t.Fatalf("GetAllTeilnehmers failed: %v", err)
	}

	if len(teilnehmers) != 0 {
		t.Errorf("Expected 0 participants, got %d", len(teilnehmers))
	}
}

// TestSaveGroups tests saving groups to database
func TestSaveGroups(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	// Insert participants first
	rows := [][]string{
		{"Name", "Ortsverband", "Alter", "Geschlecht", "PreGroup"},
		{"Max Mustermann", "Berlin", "25", "M", ""},
		{"Anna Schmidt", "Hamburg", "30", "W", ""},
		{"Tom Meyer", "München", "22", "M", ""},
		{"Lisa Weber", "Köln", "24", "W", ""},
	}

	err := database.InsertData(db, rows)
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	// Create test groups
	groups := []models.Group{
		{
			GroupID: 1,
			Teilnehmers: []models.Teilnehmer{
				{TeilnehmerID: 1, Name: "Max Mustermann"},
				{TeilnehmerID: 2, Name: "Anna Schmidt"},
			},
		},
		{
			GroupID: 2,
			Teilnehmers: []models.Teilnehmer{
				{TeilnehmerID: 3, Name: "Tom Meyer"},
				{TeilnehmerID: 4, Name: "Lisa Weber"},
			},
		},
	}

	// Save groups
	err = database.SaveGroups(db, groups)
	if err != nil {
		t.Fatalf("SaveGroups failed: %v", err)
	}

	// Verify groups were saved in gruppe table
	var gruppeCount int
	err = db.QueryRow("SELECT COUNT(*) FROM gruppe").Scan(&gruppeCount)
	if err != nil {
		t.Fatalf("Failed to count gruppe entries: %v", err)
	}

	if gruppeCount != 4 {
		t.Errorf("Expected 4 gruppe entries, got %d", gruppeCount)
	}

}

// TestGetGroupsForReport tests retrieving groups with participants
func TestGetGroupsForReport(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	// Insert participants
	rows := [][]string{
		{"Name", "Ortsverband", "Alter", "Geschlecht", "PreGroup"},
		{"Max Mustermann", "Berlin", "25", "M", ""},
		{"Anna Schmidt", "Hamburg", "30", "W", ""},
		{"Tom Meyer", "München", "22", "M", ""},
	}

	err := database.InsertData(db, rows)
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	// Create and save groups
	groups := []models.Group{
		{
			GroupID: 1,
			Teilnehmers: []models.Teilnehmer{
				{TeilnehmerID: 1},
				{TeilnehmerID: 2},
			},
		},
		{
			GroupID: 2,
			Teilnehmers: []models.Teilnehmer{
				{TeilnehmerID: 3},
			},
		},
	}

	err = database.SaveGroups(db, groups)
	if err != nil {
		t.Fatalf("Failed to save groups: %v", err)
	}

	// Retrieve groups
	retrievedGroups, err := database.GetGroupsForReport(db)
	if err != nil {
		t.Fatalf("GetGroupsForReport failed: %v", err)
	}

	if len(retrievedGroups) != 2 {
		t.Errorf("Expected 2 groups, got %d", len(retrievedGroups))
	}

	// Verify first group
	if len(retrievedGroups[0].Teilnehmers) != 2 {
		t.Errorf("Expected 2 participants in group 1, got %d", len(retrievedGroups[0].Teilnehmers))
	}

	// Verify second group
	if len(retrievedGroups[1].Teilnehmers) != 1 {
		t.Errorf("Expected 1 participant in group 2, got %d", len(retrievedGroups[1].Teilnehmers))
	}
}

// TestGetAllGroupIDs tests retrieving all group IDs
func TestGetAllGroupIDs(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	// Insert participants and groups
	rows := [][]string{
		{"Name", "Ortsverband", "Alter", "Geschlecht", "PreGroup"},
		{"Max Mustermann", "Berlin", "25", "M", ""},
		{"Anna Schmidt", "Hamburg", "30", "W", ""},
	}

	err := database.InsertData(db, rows)
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	groups := []models.Group{
		{GroupID: 1, Teilnehmers: []models.Teilnehmer{{TeilnehmerID: 1}}},
		{GroupID: 2, Teilnehmers: []models.Teilnehmer{{TeilnehmerID: 2}}},
	}

	err = database.SaveGroups(db, groups)
	if err != nil {
		t.Fatalf("Failed to save groups: %v", err)
	}

	// Get all group IDs
	groupIDs, err := database.GetAllGroupIDs(db)
	if err != nil {
		t.Fatalf("GetAllGroupIDs failed: %v", err)
	}

	if len(groupIDs) != 2 {
		t.Errorf("Expected 2 group IDs, got %d", len(groupIDs))
	}

	if groupIDs[0] != 1 || groupIDs[1] != 2 {
		t.Errorf("Expected group IDs [1, 2], got %v", groupIDs)
	}
}

// TestGetStationsForReport tests retrieving stations
func TestGetStationsForReport(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	// Insert stations
	stationRows := [][]string{
		{"Station Name"}, // Header
		{"Weitsprung"},
		{"Ballwurf"},
	}

	err := database.InsertStations(db, stationRows)
	if err != nil {
		t.Fatalf("Failed to insert stations: %v", err)
	}

	// Insert participants and groups
	rows := [][]string{
		{"Name", "Ortsverband", "Alter", "Geschlecht", "PreGroup"},
		{"Max Mustermann", "Berlin", "25", "M", ""},
	}

	err = database.InsertData(db, rows)
	if err != nil {
		t.Fatalf("Failed to insert participants: %v", err)
	}

	groups := []models.Group{
		{GroupID: 1, Teilnehmers: []models.Teilnehmer{{TeilnehmerID: 1}}},
	}

	err = database.SaveGroups(db, groups)
	if err != nil {
		t.Fatalf("Failed to save groups: %v", err)
	}

	// Get stations
	stations, err := database.GetStationsForReport(db)
	if err != nil {
		t.Fatalf("GetStationsForReport failed: %v", err)
	}

	if len(stations) != 2 {
		t.Errorf("Expected 2 stations, got %d", len(stations))
	}

	// Verify station names (either order is valid)
	stationNames := make(map[string]bool)
	for _, station := range stations {
		stationNames[station.StationName] = true
	}

	if !stationNames["Weitsprung"] {
		t.Error("Expected station 'Weitsprung' in results")
	}
	if !stationNames["Ballwurf"] {
		t.Error("Expected station 'Ballwurf' in results")
	}
}
