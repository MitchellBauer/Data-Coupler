package config

import (
	"github.com/mitchellbauer/data-coupler/internal/types"
	"encoding/json"
	"fmt"
	"os"
)

// LoadProfile reads a JSON file and unmarshals it into a Profile struct.
func LoadProfile(path string) (types.Profile, error) {
	var p types.Profile

	// 1. Read the file
	data, err := os.ReadFile(path)
	if err != nil {
		return p, fmt.Errorf("could not read profile at %s: %w", path, err)
	}

	// 2. Parse JSON
	if err := json.Unmarshal(data, &p); err != nil {
		return p, fmt.Errorf("invalid JSON format in profile: %w", err)
	}

	return p, nil
}


