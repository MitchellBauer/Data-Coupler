package engine

import (
	"data-coupler/internal/types"
	"reflect"
	"testing"
)

// Test 1: Does the Mapper actually move data to the right slots?
func TestMapRow(t *testing.T) {
	// Setup: A fake profile that wants "ID" -> "TargetID" and "Name" -> "TargetName"
	profile := types.Profile{
		Mappings: []types.Mapping{
			{InputCol: "ID", OutputCol: "TargetID"},
			{InputCol: "Name", OutputCol: "TargetName"},
		},
	}

	// Setup: A fake header map telling us where "ID" and "Name" are in the source
	// source_csv: [ "Name", "Irrelevant", "ID" ]
	// indices:       0         1          2
	headerMap := map[string]int{
		"Name": 0,
		"ID":   2,
	}

	// Input Row corresponding to indices above: Name="Alice", Irrelevant="X", ID="101"
	inputRow := []string{"Alice", "X", "101"}

	// Execution
	result := MapRow(inputRow, headerMap, profile)

	// Expectation: [ "101", "Alice" ] (Order based on Mappings list)
	expected := []string{"101", "Alice"}

	// Assertion
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("MapRow() = %v, want %v", result, expected)
	}
}

// Test 2: Does the Validator catch missing columns?
func TestValidateHeaders_MissingColumn(t *testing.T) {
	// Profile asks for "Email", but CSV headers won't have it
	profile := types.Profile{
		Mappings: []types.Mapping{
			{InputCol: "Email", OutputCol: "Contact"},
		},
	}

	csvHeaders := []string{"ID", "Name"} // No Email here!

	_, err := ValidateHeaders(csvHeaders, profile)

	if err == nil {
		t.Error("Expected an error for missing column 'Email', but got nil")
	}
}

// Test 3: Does it safely ignore irrelevant columns?
func TestValidateHeaders_Success(t *testing.T) {
	profile := types.Profile{
		Mappings: []types.Mapping{
			{InputCol: "ID", OutputCol: "OutputID"},
		},
	}

	csvHeaders := []string{"ID", "ExtraColumn1", "ExtraColumn2"}

	_, err := ValidateHeaders(csvHeaders, profile)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}
