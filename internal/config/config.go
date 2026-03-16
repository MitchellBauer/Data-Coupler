package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mitchellbauer/data-coupler/internal/types"
)

// LoadProfile reads a JSON file and returns a Profile, automatically migrating old-format profiles.
func LoadProfile(path string) (types.Profile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return types.Profile{}, fmt.Errorf("could not read profile at %s: %w", path, err)
	}
	return migrateProfile(data)
}

// migrateProfile detects old-format profiles (Settings field present, no Input/Output blocks)
// and upgrades them so Phase 1A profiles load without errors.
func migrateProfile(data []byte) (types.Profile, error) {
	// Peek at the raw keys to detect old vs new format.
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return types.Profile{}, fmt.Errorf("invalid JSON format in profile: %w", err)
	}

	_, hasSettings := raw["settings"]
	_, hasInput := raw["input"]

	if hasSettings && !hasInput {
		return migrateV0Profile(raw)
	}

	// New format — unmarshal directly.
	var p types.Profile
	if err := json.Unmarshal(data, &p); err != nil {
		return p, fmt.Errorf("invalid JSON format in profile: %w", err)
	}
	return p, nil
}

// migrateV0Profile converts a Phase 1A flat profile into the new connector-based schema.
func migrateV0Profile(raw map[string]json.RawMessage) (types.Profile, error) {
	// Parse the common top-level fields using a temporary struct.
	var meta struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	// Re-marshal just the keys we need and unmarshal into the temporary struct.
	subset, _ := json.Marshal(raw)
	_ = json.Unmarshal(subset, &meta)

	var p types.Profile
	p.ID = meta.ID
	p.Name = meta.Name
	p.Description = meta.Description
	p.Version = 1

	// Default both connectors to csv; paths will be overridden at runtime by CLI flags.
	p.Input = types.IOConfig{Connector: "csv"}
	p.Output = types.IOConfig{Connector: "csv"}

	// Parse old-style mappings (Transform string → Transforms []Transform).
	if rawMappings, ok := raw["mappings"]; ok {
		var oldMappings []struct {
			InputCol  string `json:"inputCol"`
			OutputCol string `json:"outputCol"`
			Transform string `json:"transform"`
		}
		if err := json.Unmarshal(rawMappings, &oldMappings); err != nil {
			return p, fmt.Errorf("could not parse mappings: %w", err)
		}
		for _, m := range oldMappings {
			newM := types.Mapping{
				InputCol:   m.InputCol,
				OutputCol:  m.OutputCol,
				Transforms: []types.Transform{},
			}
			// Carry over a non-empty, non-"none" transform string as a best-effort migration.
			if m.Transform != "" && m.Transform != "none" {
				newM.Transforms = append(newM.Transforms, types.Transform{Type: m.Transform})
			}
			p.Mappings = append(p.Mappings, newM)
		}
	}

	return p, nil
}

// LoadSettings reads AppSettings from a JSON file.
// If the file does not exist, it returns a default AppSettings (LastConnector = "csv").
func LoadSettings(path string) (types.AppSettings, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return types.AppSettings{LastConnector: "csv"}, nil
	}
	if err != nil {
		return types.AppSettings{}, fmt.Errorf("could not read settings at %s: %w", path, err)
	}

	var s types.AppSettings
	if err := json.Unmarshal(data, &s); err != nil {
		return types.AppSettings{}, fmt.Errorf("invalid JSON format in settings: %w", err)
	}
	if s.LastConnector == "" {
		s.LastConnector = "csv"
	}
	return s, nil
}

// SaveProfile writes a Profile to a JSON file, creating parent directories as needed.
func SaveProfile(path string, p types.Profile) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("could not create profile directory: %w", err)
	}
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return fmt.Errorf("could not marshal profile: %w", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("could not write profile to %s: %w", path, err)
	}
	return nil
}

// SaveSettings writes AppSettings to a JSON file.
func SaveSettings(path string, s types.AppSettings) error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("could not marshal settings: %w", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("could not write settings to %s: %w", path, err)
	}
	return nil
}
