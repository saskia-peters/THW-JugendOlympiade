package io

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"THW-JugendOlympiade/backend/models"

	"github.com/xuri/excelize/v2"
)

// ValidateHeaders checks if the header row matches expected columns
func ValidateHeaders(headers []string, expected []string) error {
	if len(headers) < len(expected) {
		return fmt.Errorf("insufficient columns: expected at least %d columns, got %d", len(expected), len(headers))
	}

	for i, expectedHeader := range expected {
		actual := strings.TrimSpace(headers[i])
		if !strings.EqualFold(actual, expectedHeader) {
			return fmt.Errorf("invalid header in column %d: expected '%s', got '%s'", i+1, expectedHeader, actual)
		}
	}

	return nil
}

// ValidateParticipantRow validates a single participant data row
func ValidateParticipantRow(row []string, rowNum int) error {
	// Accept rows with at least 4 columns (PreGroup is optional and may be missing)
	if len(row) < 4 {
		return fmt.Errorf("row %d: insufficient columns (expected at least 4: Name, Ortsverband, Alter, Geschlecht; PreGroup is optional)", rowNum)
	}

	name := strings.TrimSpace(row[0])
	ortsverband := strings.TrimSpace(row[1])
	alterStr := strings.TrimSpace(row[2])
	geschlecht := strings.TrimSpace(row[3])

	// Validate PreGroup if present (alphanumeric, max 20 chars)
	var pregroup string
	if len(row) > 4 {
		pregroup = strings.TrimSpace(row[4])
		if pregroup != "" {
			// Check if alphanumeric only
			for _, r := range pregroup {
				if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')) {
					return fmt.Errorf("row %d (%s): invalid PreGroup '%s' - must contain only letters and numbers", rowNum, name, pregroup)
				}
			}
			// Check length
			if len(pregroup) > 20 {
				return fmt.Errorf("row %d (%s): PreGroup '%s' too long - maximum 20 characters", rowNum, name, pregroup)
			}
		}
	}

	// Validate name (required)
	if name == "" {
		// Empty name is acceptable - row will be skipped
		return nil
	}

	// Validate age (must be a valid positive integer if provided)
	if alterStr != "" {
		age, err := strconv.Atoi(alterStr)
		if err != nil {
			return fmt.Errorf("row %d (%s): invalid age '%s' - must be a number", rowNum, name, alterStr)
		}
		if age < 0 || age > 150 {
			return fmt.Errorf("row %d (%s): invalid age %d - must be between 0 and 150", rowNum, name, age)
		}
	}

	// Validate geschlecht (should be M/W/D or similar, but we'll be lenient)
	if geschlecht != "" {
		validGenders := []string{"M", "W", "D", "m", "w", "d", "männlich", "weiblich", "divers", "male", "female"}
		isValid := false
		for _, valid := range validGenders {
			if strings.EqualFold(geschlecht, valid) {
				isValid = true
				break
			}
		}
		if !isValid {
			log.Printf("Warning: row %d (%s): unusual gender value '%s' - proceeding anyway", rowNum, name, geschlecht)
		}
	}

	// Ortsverband is optional but log if missing
	if ortsverband == "" {
		log.Printf("Warning: row %d (%s): missing Ortsverband", rowNum, name)
	}

	return nil
}

// ReadXLSXFile reads the XLSX file and returns the rows
func ReadXLSXFile(filePath string) ([][]string, error) {
	// Check if XLSX file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("XLSX file '%s' not found", filePath)
	}

	// Open XLSX file
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open XLSX file: %w", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Printf("Failed to close XLSX file: %v", err)
		}
	}()

	// Read all rows from the "Teilnehmende" sheet
	rows, err := f.GetRows(models.SheetName)
	if err != nil {
		return nil, fmt.Errorf("failed to read sheet '%s': %w", models.SheetName, err)
	}

	if len(rows) == 0 {
		return nil, fmt.Errorf("sheet '%s' is empty", models.SheetName)
	}

	if len(rows) < 2 {
		return nil, fmt.Errorf("sheet '%s' must contain at least a header row and one data row", models.SheetName)
	}

	// Validate header row
	expectedHeaders := []string{"Name", "Ortsverband", "Alter", "Geschlecht", "PreGroup"}
	if err := ValidateHeaders(rows[0], expectedHeaders); err != nil {
		return nil, fmt.Errorf("invalid sheet structure: %w", err)
	}

	// Validate each data row (skip header)
	validRowCount := 0
	for i := 1; i < len(rows); i++ {
		if err := ValidateParticipantRow(rows[i], i+1); err != nil {
			return nil, err
		}
		// Count non-empty rows
		if len(rows[i]) > 0 && strings.TrimSpace(rows[i][0]) != "" {
			validRowCount++
		}
	}

	if validRowCount == 0 {
		return nil, fmt.Errorf("sheet '%s' contains no valid participant data", models.SheetName)
	}

	log.Printf("Successfully validated %d participant rows", validRowCount)

	return rows, nil
}

// ReadBetreuendeFromXLSX reads caretakers/drivers from the "Betreuende" sheet.
// Returns an empty slice (no error) if the sheet does not exist.
func ReadBetreuendeFromXLSX(filePath string) ([][]string, error) {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("XLSX file '%s' not found", filePath)
	}

	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open XLSX file: %w", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Printf("Failed to close XLSX file: %v", err)
		}
	}()

	rows, err := f.GetRows(models.BetreuendeSheetName)
	if err != nil {
		log.Printf("Sheet '%s' not found or error reading: %v - betreuende are optional", models.BetreuendeSheetName, err)
		return [][]string{}, nil
	}

	if len(rows) < 2 {
		log.Printf("Warning: sheet '%s' has no data rows", models.BetreuendeSheetName)
		return [][]string{}, nil
	}

	return rows, nil
}

// ReadStationsFromXLSX reads the stations from the Stationen sheet
func ReadStationsFromXLSX(filePath string) ([][]string, error) {
	// Check if XLSX file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("XLSX file '%s' not found", filePath)
	}

	// Open XLSX file
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open XLSX file: %w", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Printf("Failed to close XLSX file: %v", err)
		}
	}()

	// Read all rows from the "Stationen" sheet
	rows, err := f.GetRows(models.StationsSheetName)
	if err != nil {
		// If sheet doesn't exist, return empty slice (stations are optional)
		log.Printf("Sheet '%s' not found or error reading: %v - stations are optional", models.StationsSheetName, err)
		return [][]string{}, nil
	}

	// Validate stations if present
	if len(rows) > 0 {
		// Check if there's at least a header
		if len(rows) < 2 {
			log.Printf("Warning: sheet '%s' has only header row, no station data", models.StationsSheetName)
			return [][]string{}, nil
		}

		// Validate that each row has at least one column (station name)
		validStationCount := 0
		for i := 1; i < len(rows); i++ {
			if len(rows[i]) > 0 && strings.TrimSpace(rows[i][0]) != "" {
				validStationCount++
			}
		}

		log.Printf("Successfully validated %d station rows", validStationCount)
	}

	return rows, nil
}
