package io

import (
	"fmt"
	"log"
	"os"

	"experiment1/backend/models"

	"github.com/xuri/excelize/v2"
)

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

	// Read all rows from the "Teilnehmer" sheet
	rows, err := f.GetRows(models.SheetName)
	if err != nil {
		return nil, fmt.Errorf("failed to read sheet '%s': %w", models.SheetName, err)
	}

	if len(rows) == 0 {
		return nil, fmt.Errorf("sheet '%s' is empty", models.SheetName)
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
		return [][]string{}, nil
	}

	return rows, nil
}
