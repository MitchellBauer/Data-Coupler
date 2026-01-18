package engine

import (
	"bufio"
	"bytes"
	"github.com/mitchellbauer/data-coupler/internal/types" // <--- Import your types package
	"encoding/csv"
	"fmt"
	"os"
)

// ReadCSV opens a file, strips any UTF-8 BOM, and returns all records.
func ReadCSV(path string) ([][]string, error) {
	// ... (Previous code remains unchanged) ...
	// (Re-paste the ReadCSV code here if you are overwriting the whole file,
	// otherwise just append the function below)

	// For brevity in this snippet, I am focusing on the new function:
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("could not open file: %w", err)
	}
	defer file.Close()

	br := bufio.NewReader(file)
	bom := []byte{0xEF, 0xBB, 0xBF}
	peekBytes, err := br.Peek(3)
	if err == nil && bytes.Equal(peekBytes, bom) {
		_, _ = br.Discard(3)
	}

	reader := csv.NewReader(br)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to parse CSV: %w", err)
	}
	return records, nil
}

// ValidateHeaders checks if the input CSV contains all columns defined in the Profile.
// It returns a lookup map of [ColumnName] -> ColumnIndex.
func ValidateHeaders(headers []string, p types.Profile) (map[string]int, error) {
	// 1. Create a "Lookup Map" of the CSV headers
	// e.g. map["EmployeeID"] = 0, map["Name"] = 1
	headerMap := make(map[string]int)
	for i, h := range headers {
		headerMap[h] = i
	}

	// 2. Verify every requirement in the Profile exists in the CSV
	for _, mapping := range p.Mappings {
		if _, exists := headerMap[mapping.InputCol]; !exists {
			// 🛑 STOP! The CSV is wrong.
			return nil, fmt.Errorf("validation failed: Input CSV is missing required column '%s'", mapping.InputCol)
		}
	}

	return headerMap, nil
}

// MapRow transforms a single input row into the desired output row based on the Profile.
func MapRow(inputRow []string, headerMap map[string]int, p types.Profile) []string {
	// 1. Initialize the bucket for the new data.
	// The output will always have exactly the same number of columns as the "mappings" list.
	outputRow := make([]string, len(p.Mappings))

	// 2. Iterate through the mapping instructions
	for i, mapping := range p.Mappings {
		// A. Find where the data lives in the source CSV
		sourceIndex, found := headerMap[mapping.InputCol]

		// B. Safety Check: Does the column exist and is the row long enough?
		if found && sourceIndex < len(inputRow) {
			val := inputRow[sourceIndex]

			// --- Phase 2 Placeholder: Transformations ---
			// if mapping.Transform == "uppercase" { val = strings.ToUpper(val) }
			// --------------------------------------------

			// C. Place the value in the specific output position
			outputRow[i] = val
		}
	}

	return outputRow
}

// Run coordinates the entire conversion process.
func Run(inputPath string, outputPath string, p types.Profile) error {
	// 1. Read the Input File
	records, err := ReadCSV(inputPath)
	if err != nil {
		return err
	}
	if len(records) < 1 {
		return fmt.Errorf("input CSV is empty")
	}

	// 2. Validate the Input Headers (Row 0)
	inputHeaders := records[0]
	headerMap, err := ValidateHeaders(inputHeaders, p)
	if err != nil {
		return err
	}

	// 3. Create the Output File
	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	writer := csv.NewWriter(outFile)
	defer writer.Flush()

	// 4. Write the Output Headers (from Profile)
	// We build the header row based on what the Profile *wants* the columns to be named.
	outputHeaders := make([]string, len(p.Mappings))
	for i, m := range p.Mappings {
		outputHeaders[i] = m.OutputCol
	}
	if err := writer.Write(outputHeaders); err != nil {
		return fmt.Errorf("failed to write output headers: %w", err)
	}

	// 5. Process and Write Data Rows
	// Start at index 1 to skip the header row we already processed.
	for i, row := range records {
		if i == 0 {
			continue
		}

		newRow := MapRow(row, headerMap, p)
		if err := writer.Write(newRow); err != nil {
			return fmt.Errorf("error writing row %d: %w", i+1, err)
		}
	}

	return nil
}


