package engine

import (
	"encoding/csv"
	"fmt"
	"os"

	"github.com/mitchellbauer/data-coupler/internal/connector"
	"github.com/mitchellbauer/data-coupler/internal/credentials"
	"github.com/mitchellbauer/data-coupler/internal/transform"
	"github.com/mitchellbauer/data-coupler/internal/types"
)

// ValidateHeaders checks if the input connector's columns contain all columns defined in the Profile.
// It returns a lookup map of [ColumnName] -> ColumnIndex.
func ValidateHeaders(headers []string, p types.Profile) (map[string]int, error) {
	headerMap := make(map[string]int)
	for i, h := range headers {
		headerMap[h] = i
	}

	for _, mapping := range p.Mappings {
		if _, exists := headerMap[mapping.InputCol]; !exists {
			return nil, fmt.Errorf("validation failed: Input is missing required column '%s'", mapping.InputCol)
		}
	}

	return headerMap, nil
}

// MapRow transforms a single input row into the desired output row based on the Profile.
// Transforms defined on each mapping are applied in order.
func MapRow(inputRow []string, headerMap map[string]int, p types.Profile) ([]string, error) {
	outputRow := make([]string, len(p.Mappings))

	for i, mapping := range p.Mappings {
		sourceIndex, found := headerMap[mapping.InputCol]
		if found && sourceIndex < len(inputRow) {
			val := inputRow[sourceIndex]
			var err error
			val, err = applyTransforms(val, mapping.Transforms, inputRow, headerMap)
			if err != nil {
				return nil, fmt.Errorf("column %q: %w", mapping.OutputCol, err)
			}
			outputRow[i] = val
		}
	}

	return outputRow, nil
}

// applyTransforms runs a value through an ordered list of transforms.
// It passes the full input row and header map so RowTransformer implementations
// can read other columns (e.g. Concatenate).
// Unknown transform types return an error.
func applyTransforms(value string, transforms []types.Transform, inputRow []string, headerMap map[string]int) (string, error) {
	for _, t := range transforms {
		tr, ok := transform.Get(t.Type)
		if !ok {
			return "", fmt.Errorf("unknown transform %q", t.Type)
		}

		var err error
		if rt, ok := tr.(transform.RowTransformer); ok {
			value, err = rt.ApplyRow(inputRow, headerMap, t.Params)
		} else {
			value, err = tr.Apply(value, t.Params)
		}
		if err != nil {
			return value, fmt.Errorf("transform %q: %w", t.Type, err)
		}
	}
	return value, nil
}

// Run coordinates the entire conversion process using the connector registry.
// creds may be nil for CSV-only profiles that have no credentialRef.
// It returns the number of data rows written (excluding the header row).
func Run(p types.Profile, creds credentials.Store) (int, error) {
	// 1. Resolve input connector.
	inputConn, ok := connector.Get(p.Input.Connector)
	if !ok {
		return 0, fmt.Errorf("unknown input connector: %q", p.Input.Connector)
	}

	// 2. Build the input ConnectionConfig, resolving credentials if needed.
	inputCfg := connector.ConnectionConfig{
		FilePath: p.Input.Path,
		Host:     p.Input.Host,
		Port:     p.Input.Port,
		Database: p.Input.Database,
		Username: p.Input.Username,
	}
	if p.Input.CredentialRef != "" && creds != nil {
		pw, err := creds.Load(p.Input.CredentialRef)
		if err == nil {
			inputCfg.Password = pw
		}
	}

	// 3. Connect to input.
	if err := inputConn.Connect(inputCfg); err != nil {
		return 0, err
	}
	defer inputConn.Disconnect()

	// 4. Read headers and validate.
	headers, err := inputConn.Columns(p.Input.Query)
	if err != nil {
		return 0, err
	}
	headerMap, err := ValidateHeaders(headers, p)
	if err != nil {
		return 0, err
	}

	// 5. Stream rows.
	rowCh, err := inputConn.Rows(p.Input.Query)
	if err != nil {
		return 0, err
	}

	// 6. Create output file and writer.
	outFile, err := os.Create(p.Output.Path)
	if err != nil {
		return 0, fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	writer := csv.NewWriter(outFile)
	defer writer.Flush()

	// 7. Write output headers.
	outputHeaders := make([]string, len(p.Mappings))
	for i, m := range p.Mappings {
		outputHeaders[i] = m.OutputCol
	}
	if err := writer.Write(outputHeaders); err != nil {
		return 0, fmt.Errorf("failed to write output headers: %w", err)
	}

	// 8. Process and write data rows.
	rowsWritten := 0
	for row := range rowCh {
		newRow, err := MapRow(row, headerMap, p)
		if err != nil {
			return rowsWritten, fmt.Errorf("row %d: %w", rowsWritten+1, err)
		}
		if err := writer.Write(newRow); err != nil {
			return rowsWritten, fmt.Errorf("error writing row %d: %w", rowsWritten+1, err)
		}
		rowsWritten++
	}

	return rowsWritten, nil
}
