package baseline

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/shivam95/bundlespy/internal/parser"
)

// Baseline is the saved snapshot of a previous build, stored as JSON.
type Baseline struct {
	Tool       string         `json:"tool"`
	Hash       string         `json:"hash"`
	RecordedAt time.Time      `json:"recorded_at"`
	Assets     []parser.Asset `json:"assets"`
}

// Save writes the current build stats to a baseline JSON file.
func Save(stats *parser.BuildStats, outPath string) error {
	b := Baseline{
		Tool:       stats.Tool,
		Hash:       stats.Hash,
		RecordedAt: time.Now().UTC(),
		Assets:     stats.Assets,
	}

	data, err := json.MarshalIndent(b, "", "  ")
	if err != nil {
		return fmt.Errorf("baseline: marshal: %w", err)
	}

	if err := os.WriteFile(outPath, data, 0644); err != nil {
		return fmt.Errorf("baseline: write %s: %w", outPath, err)
	}

	return nil
}

// Load reads a baseline JSON file from disk.
func Load(path string) (*Baseline, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("baseline: read %s: %w", path, err)
	}

	var b Baseline
	if err := json.Unmarshal(data, &b); err != nil {
		return nil, fmt.Errorf("baseline: parse: %w", err)
	}

	return &b, nil
}
