package test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"THW-JugendOlympiade/backend/io"
	"THW-JugendOlympiade/backend/models"

	"github.com/xuri/excelize/v2"
)

// TestValidateHeaders tests the header validation logic
func TestValidateHeaders(t *testing.T) {
	tests := []struct {
		name        string
		headers     []string
		expected    []string
		shouldError bool
		errorMsg    string
	}{
		{
			name:        "Valid headers - exact match",
			headers:     []string{"Name", "Ortsverband", "Alter", "Geschlecht", "PreGroup"},
			expected:    []string{"Name", "Ortsverband", "Alter", "Geschlecht", "PreGroup"},
			shouldError: false,
		},
		{
			name:        "Valid headers - case insensitive",
			headers:     []string{"name", "ORTSVERBAND", "Alter", "geschlecht", "pregroup"},
			expected:    []string{"Name", "Ortsverband", "Alter", "Geschlecht", "PreGroup"},
			shouldError: false,
		},
		{
			name:        "Valid headers - with whitespace",
			headers:     []string{"  Name  ", "Ortsverband", "Alter  ", "Geschlecht", "PreGroup"},
			expected:    []string{"Name", "Ortsverband", "Alter", "Geschlecht", "PreGroup"},
			shouldError: false,
		},
		{
			name:        "Invalid headers - insufficient columns",
			headers:     []string{"Name", "Ortsverband"},
			expected:    []string{"Name", "Ortsverband", "Alter", "Geschlecht", "PreGroup"},
			shouldError: true,
			errorMsg:    "insufficient columns",
		},
		{
			name:        "Invalid headers - wrong column name",
			headers:     []string{"Name", "Location", "Alter", "Geschlecht", "PreGroup"},
			expected:    []string{"Name", "Ortsverband", "Alter", "Geschlecht", "PreGroup"},
			shouldError: true,
			errorMsg:    "invalid header",
		},
		{
			name:        "Valid headers - extra columns",
			headers:     []string{"Name", "Ortsverband", "Alter", "Geschlecht", "PreGroup", "Extra"},
			expected:    []string{"Name", "Ortsverband", "Alter", "Geschlecht", "PreGroup"},
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := io.ValidateHeaders(tt.headers, tt.expected)
			if tt.shouldError {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("expected error containing %q, got: %v", tt.errorMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

// TestValidateParticipantRow tests the participant row validation logic
func TestValidateParticipantRow(t *testing.T) {
	tests := []struct {
		name        string
		row         []string
		rowNum      int
		shouldError bool
		errorMsg    string
	}{
		{
			name:        "Valid row - all fields",
			row:         []string{"Max Mustermann", "Berlin", "25", "M"},
			rowNum:      2,
			shouldError: false,
		},
		{
			name:        "Valid row - empty ortsverband (warning only)",
			row:         []string{"Anna Schmidt", "", "30", "W"},
			rowNum:      3,
			shouldError: false,
		},
		{
			name:        "Valid row - empty geschlecht (warning only)",
			row:         []string{"Tom Meyer", "Hamburg", "20", ""},
			rowNum:      4,
			shouldError: false,
		},
		{
			name:        "Valid row - empty age",
			row:         []string{"Lisa Weber", "München", "", "W"},
			rowNum:      5,
			shouldError: false,
		},
		{
			name:        "Valid row - empty name (skipped)",
			row:         []string{"", "Berlin", "25", "M"},
			rowNum:      6,
			shouldError: false,
		},
		{
			name:        "Invalid row - insufficient columns",
			row:         []string{"Max Mustermann", "Berlin"},
			rowNum:      7,
			shouldError: true,
			errorMsg:    "insufficient columns",
		},
		{
			name:        "Invalid row - non-numeric age",
			row:         []string{"Max Mustermann", "Berlin", "twenty", "M"},
			rowNum:      8,
			shouldError: true,
			errorMsg:    "invalid age",
		},
		{
			name:        "Invalid row - negative age",
			row:         []string{"Max Mustermann", "Berlin", "-5", "M"},
			rowNum:      9,
			shouldError: true,
			errorMsg:    "invalid age",
		},
		{
			name:        "Invalid row - age over 150",
			row:         []string{"Max Mustermann", "Berlin", "200", "M"},
			rowNum:      10,
			shouldError: true,
			errorMsg:    "invalid age",
		},
		{
			name:        "Valid row - unusual gender (warning only)",
			row:         []string{"Alex Smith", "Berlin", "25", "X"},
			rowNum:      11,
			shouldError: false,
		},
		{
			name:        "Valid row - gender variations",
			row:         []string{"Test User", "Berlin", "25", "männlich"},
			rowNum:      12,
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := io.ValidateParticipantRow(tt.row, tt.rowNum)
			if tt.shouldError {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("expected error containing %q, got: %v", tt.errorMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

// Helper function to create a test Excel file
func createTestExcelFile(t *testing.T, filename string, sheetName string, data [][]string) string {
	f := excelize.NewFile()
	defer f.Close()

	// Create or use existing sheet
	index, err := f.NewSheet(sheetName)
	if err != nil {
		t.Fatalf("Failed to create sheet: %v", err)
	}
	f.SetActiveSheet(index)

	// Write data
	for rowIdx, row := range data {
		for colIdx, cell := range row {
			cellName, err := excelize.CoordinatesToCellName(colIdx+1, rowIdx+1)
			if err != nil {
				t.Fatalf("Failed to get cell name: %v", err)
			}
			f.SetCellValue(sheetName, cellName, cell)
		}
	}

	// Save file in test directory
	testDir := t.TempDir()
	filepath := filepath.Join(testDir, filename)

	if err := f.SaveAs(filepath); err != nil {
		t.Fatalf("Failed to save Excel file: %v", err)
	}

	return filepath
}

// TestReadXLSXFile_ValidFile tests reading a valid Excel file
func TestReadXLSXFile_ValidFile(t *testing.T) {
	data := [][]string{
		{"Name", "Ortsverband", "Alter", "Geschlecht", "PreGroup"},
		{"Max Mustermann", "Berlin", "25", "M", ""},
		{"Anna Schmidt", "Hamburg", "30", "W", ""},
		{"Tom Meyer", "München", "22", "M", ""},
	}

	filepath := createTestExcelFile(t, "valid_test.xlsx", models.SheetName, data)
	defer os.Remove(filepath)

	rows, err := io.ReadXLSXFile(filepath)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(rows) != 4 {
		t.Errorf("Expected 4 rows, got %d", len(rows))
	}

	// Check header row
	if rows[0][0] != "Name" {
		t.Errorf("Expected header 'Name', got '%s'", rows[0][0])
	}

	// Check data row
	if rows[1][0] != "Max Mustermann" {
		t.Errorf("Expected 'Max Mustermann', got '%s'", rows[1][0])
	}
}

// TestReadXLSXFile_EmptySheet tests reading an empty Excel sheet
func TestReadXLSXFile_EmptySheet(t *testing.T) {
	data := [][]string{} // Empty sheet

	filepath := createTestExcelFile(t, "empty_test.xlsx", models.SheetName, data)
	defer os.Remove(filepath)

	_, err := io.ReadXLSXFile(filepath)
	if err == nil {
		t.Fatal("Expected error for empty sheet, got nil")
	}

	if !strings.Contains(err.Error(), "empty") {
		t.Errorf("Expected error message to contain 'empty', got: %v", err)
	}
}

// TestReadXLSXFile_OnlyHeader tests reading a file with only header row
func TestReadXLSXFile_OnlyHeader(t *testing.T) {
	data := [][]string{
		{"Name", "Ortsverband", "Alter", "Geschlecht", "PreGroup"},
	}

	filepath := createTestExcelFile(t, "header_only_test.xlsx", models.SheetName, data)
	defer os.Remove(filepath)

	_, err := io.ReadXLSXFile(filepath)
	if err == nil {
		t.Fatal("Expected error for header-only sheet, got nil")
	}

	if !strings.Contains(err.Error(), "at least a header row and one data row") {
		t.Errorf("Expected error about missing data rows, got: %v", err)
	}
}

// TestReadXLSXFile_InvalidHeaders tests reading a file with invalid headers
func TestReadXLSXFile_InvalidHeaders(t *testing.T) {
	data := [][]string{
		{"Name", "Location", "Age", "Gender", "PreGroup"}, // Wrong headers (Location instead of Ortsverband, Age/Gender instead of Alter/Geschlecht)
		{"Max Mustermann", "Berlin", "25", "M", ""},
	}

	filepath := createTestExcelFile(t, "invalid_headers_test.xlsx", models.SheetName, data)
	defer os.Remove(filepath)

	_, err := io.ReadXLSXFile(filepath)
	if err == nil {
		t.Fatal("Expected error for invalid headers, got nil")
	}

	if !strings.Contains(err.Error(), "invalid header") {
		t.Errorf("Expected error about invalid headers, got: %v", err)
	}
}

// TestReadXLSXFile_InvalidAge tests reading a file with invalid age values
func TestReadXLSXFile_InvalidAge(t *testing.T) {
	data := [][]string{
		{"Name", "Ortsverband", "Alter", "Geschlecht", "PreGroup"},
		{"Max Mustermann", "Berlin", "twenty", "M", ""}, // Invalid age
	}

	filepath := createTestExcelFile(t, "invalid_age_test.xlsx", models.SheetName, data)
	defer os.Remove(filepath)

	_, err := io.ReadXLSXFile(filepath)
	if err == nil {
		t.Fatal("Expected error for invalid age, got nil")
	}

	if !strings.Contains(err.Error(), "invalid age") {
		t.Errorf("Expected error about invalid age, got: %v", err)
	}
}

// TestReadXLSXFile_InsufficientColumns tests reading a file with insufficient columns
func TestReadXLSXFile_InsufficientColumns(t *testing.T) {
	data := [][]string{
		{"Name", "Ortsverband", "Alter", "Geschlecht", "PreGroup"},
		{"Max Mustermann", "Berlin"}, // Only 2 columns
	}

	filepath := createTestExcelFile(t, "insufficient_columns_test.xlsx", models.SheetName, data)
	defer os.Remove(filepath)

	_, err := io.ReadXLSXFile(filepath)
	if err == nil {
		t.Fatal("Expected error for insufficient columns, got nil")
	}

	if !strings.Contains(err.Error(), "insufficient columns") {
		t.Errorf("Expected error about insufficient columns, got: %v", err)
	}
}

// TestReadXLSXFile_EmptyRows tests reading a file with empty rows (should be skipped)
func TestReadXLSXFile_EmptyRows(t *testing.T) {
	data := [][]string{
		{"Name", "Ortsverband", "Alter", "Geschlecht", "PreGroup"},
		{"Max Mustermann", "Berlin", "25", "M", ""},
		{"  ", "  ", "  ", "  ", ""}, // Row with whitespace - should be skipped gracefully
		{"Anna Schmidt", "Hamburg", "30", "W", ""},
	}

	filepath := createTestExcelFile(t, "empty_rows_test.xlsx", models.SheetName, data)
	defer os.Remove(filepath)

	rows, err := io.ReadXLSXFile(filepath)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Should have 4 rows total (header + 2 valid + 1 whitespace)
	if len(rows) != 4 {
		t.Errorf("Expected 4 rows, got %d", len(rows))
	}

	// The whitespace row will be validated as having empty name after trim, so it passes validation
}

// TestReadXLSXFile_FileNotFound tests reading a non-existent file
func TestReadXLSXFile_FileNotFound(t *testing.T) {
	_, err := io.ReadXLSXFile("/nonexistent/path/test.xlsx")
	if err == nil {
		t.Fatal("Expected error for non-existent file, got nil")
	}

	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Expected error about file not found, got: %v", err)
	}
}

// TestReadStationsFromXLSX_ValidFile tests reading valid stations
func TestReadStationsFromXLSX_ValidFile(t *testing.T) {
	// Create a file with both Teilnehmende and Stationen sheets
	f := excelize.NewFile()
	defer f.Close()

	// Create Teilnehmende sheet (required)
	teilnehmerData := [][]string{
		{"Name", "Ortsverband", "Alter", "Geschlecht", "PreGroup"},
		{"Max Mustermann", "Berlin", "25", "M", ""},
	}
	index1, _ := f.NewSheet(models.SheetName)
	for rowIdx, row := range teilnehmerData {
		for colIdx, cell := range row {
			cellName, _ := excelize.CoordinatesToCellName(colIdx+1, rowIdx+1)
			f.SetCellValue(models.SheetName, cellName, cell)
		}
	}
	f.SetActiveSheet(index1)

	// Create Stationen sheet
	stationData := [][]string{
		{"Station"},
		{"Weitsprung"},
		{"Sprint"},
		{"Ballwurf"},
	}
	f.NewSheet(models.StationsSheetName)
	for rowIdx, row := range stationData {
		for colIdx, cell := range row {
			cellName, _ := excelize.CoordinatesToCellName(colIdx+1, rowIdx+1)
			f.SetCellValue(models.StationsSheetName, cellName, cell)
		}
	}

	testDir := t.TempDir()
	filepath := filepath.Join(testDir, "stations_test.xlsx")
	if err := f.SaveAs(filepath); err != nil {
		t.Fatalf("Failed to save Excel file: %v", err)
	}
	defer os.Remove(filepath)

	rows, err := io.ReadStationsFromXLSX(filepath)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(rows) != 4 {
		t.Errorf("Expected 4 rows (header + 3 stations), got %d", len(rows))
	}

	if rows[1][0] != "Weitsprung" {
		t.Errorf("Expected 'Weitsprung', got '%s'", rows[1][0])
	}
}

// TestReadStationsFromXLSX_NoStationsSheet tests handling missing stations sheet
func TestReadStationsFromXLSX_NoStationsSheet(t *testing.T) {
	data := [][]string{
		{"Name", "Ortsverband", "Alter", "Geschlecht", "PreGroup"},
		{"Max Mustermann", "Berlin", "25", "M", ""},
	}

	filepath := createTestExcelFile(t, "no_stations_test.xlsx", models.SheetName, data)
	defer os.Remove(filepath)

	rows, err := io.ReadStationsFromXLSX(filepath)
	if err != nil {
		t.Fatalf("Expected no error (stations are optional), got: %v", err)
	}

	// Should return empty slice when no stations sheet exists
	if len(rows) != 0 {
		t.Errorf("Expected 0 rows for missing stations sheet, got %d", len(rows))
	}
}

// TestReadXLSXFile_ValidPreGroup tests reading a file with valid PreGroup values
func TestReadXLSXFile_ValidPreGroup(t *testing.T) {
	data := [][]string{
		{"Name", "Ortsverband", "Alter", "Geschlecht", "PreGroup"},
		{"Alice", "Berlin", "20", "W", "TeamA"},
		{"Bob", "Hamburg", "22", "M", "Team123"},
		{"Carol", "München", "21", "W", "ABC123def456GHI789"},
	}

	filepath := createTestExcelFile(t, "valid_pregroup_test.xlsx", models.SheetName, data)
	defer os.Remove(filepath)

	rows, err := io.ReadXLSXFile(filepath)
	if err != nil {
		t.Fatalf("Expected no error for valid PreGroup, got: %v", err)
	}

	if len(rows) != 4 {
		t.Errorf("Expected 4 rows, got %d", len(rows))
	}
}

// TestReadXLSXFile_InvalidPreGroupSpecialChars tests reading a file with invalid PreGroup containing special characters
func TestReadXLSXFile_InvalidPreGroupSpecialChars(t *testing.T) {
	data := [][]string{
		{"Name", "Ortsverband", "Alter", "Geschlecht", "PreGroup"},
		{"Alice", "Berlin", "20", "W", "Team-A"}, // Invalid: contains hyphen
	}

	filepath := createTestExcelFile(t, "invalid_pregroup_chars_test.xlsx", models.SheetName, data)
	defer os.Remove(filepath)

	_, err := io.ReadXLSXFile(filepath)
	if err == nil {
		t.Fatal("Expected error for PreGroup with special characters, got nil")
	}

	if !strings.Contains(err.Error(), "invalid PreGroup") {
		t.Errorf("Expected error about invalid PreGroup, got: %v", err)
	}
}

// TestReadXLSXFile_InvalidPreGroupTooLong tests reading a file with PreGroup exceeding 20 characters
func TestReadXLSXFile_InvalidPreGroupTooLong(t *testing.T) {
	data := [][]string{
		{"Name", "Ortsverband", "Alter", "Geschlecht", "PreGroup"},
		{"Alice", "Berlin", "20", "W", "ThisPreGroupIsTooLongToBeValid"}, // 30 characters
	}

	filepath := createTestExcelFile(t, "invalid_pregroup_length_test.xlsx", models.SheetName, data)
	defer os.Remove(filepath)

	_, err := io.ReadXLSXFile(filepath)
	if err == nil {
		t.Fatal("Expected error for PreGroup too long, got nil")
	}

	if !strings.Contains(err.Error(), "too long") {
		t.Errorf("Expected error about PreGroup being too long, got: %v", err)
	}
}

// TestReadXLSXFile_EmptyPreGroup tests reading a file with empty PreGroup values (should be valid)
func TestReadXLSXFile_EmptyPreGroup(t *testing.T) {
	data := [][]string{
		{"Name", "Ortsverband", "Alter", "Geschlecht", "PreGroup"},
		{"Alice", "Berlin", "20", "W", ""},
		{"Bob", "Hamburg", "22", "M", "  "}, // Just whitespace
	}

	filepath := createTestExcelFile(t, "empty_pregroup_test.xlsx", models.SheetName, data)
	defer os.Remove(filepath)

	rows, err := io.ReadXLSXFile(filepath)
	if err != nil {
		t.Fatalf("Expected no error for empty PreGroup, got: %v", err)
	}

	if len(rows) != 3 {
		t.Errorf("Expected 3 rows, got %d", len(rows))
	}
}
