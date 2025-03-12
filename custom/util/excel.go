package util

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/xuri/excelize/v2"
)

// IsExcel checks if the file has a valid Excel extension.
func IsExcel(fileName string) bool {
	fileName = strings.ToLower(fileName)
	return strings.HasSuffix(fileName, ".xlsx")
}

// ReadExcel reads data from one or more sheets in an Excel file.
// If no sheet names are provided, all sheets in the workbook are processed.
func ReadExcel(file io.Reader, sheetNames ...string) (map[string][][]string, error) {
	// Map to store data by sheet name
	sheetData := make(map[string][][]string)

	// Create a buffer to store file contents and open the Excel file
	teeReader := io.TeeReader(file, new(bytes.Buffer))
	workbook, err := excelize.OpenReader(teeReader)
	if err != nil {
		return nil, fmt.Errorf("failed to open Excel file, err: %w", err)
	}
	defer workbook.Close()

	// If no sheet names are provided, process all sheets in the workbook
	if len(sheetNames) == 0 {
		sheetNames = workbook.GetSheetList()
	}

	// Process each sheet
	for _, sheetName := range sheetNames {
		data, err := readSheetData(workbook, sheetName)
		if err != nil {
			return nil, err
		}
		sheetData[sheetName] = data
	}

	// Return the collected data by sheet name
	return sheetData, nil
}

// ReadExcelSheet reads data from a specific sheet in an Excel file.
func ReadExcelSheet(file io.Reader, sheetName string) ([][]string, error) {
	// Create a buffer to store file contents and open the Excel file
	teeReader := io.TeeReader(file, new(bytes.Buffer))
	workbook, err := excelize.OpenReader(teeReader)
	if err != nil {
		return nil, fmt.Errorf("failed to open Excel file, err: %w", err)
	}
	defer workbook.Close()

	// Read data from the specific sheet
	return readSheetData(workbook, sheetName)
}

// readSheetData reads rows from a single sheet and returns valid rows, skipping empty rows.
func readSheetData(workbook *excelize.File, sheetName string) ([][]string, error) {
	// Get all rows from the sheet
	rows, err := workbook.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("failed to read rows from sheet %s, err: %w", sheetName, err)
	}

	// Process rows, skipping empty rows
	var validRows [][]string
	for _, row := range rows {
		// Skip rows that are completely empty
		if isRowEmpty(row) {
			continue
		}
		validRows = append(validRows, row)
	}

	// Return valid rows for the sheet
	return validRows, nil
}

// isRowEmpty checks if a row is completely empty (i.e., all cells are empty).
func isRowEmpty(row []string) bool {
	for _, cell := range row {
		if cell != "" {
			return false
		}
	}
	return true
}
