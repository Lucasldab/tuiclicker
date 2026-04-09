package persistence

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"time"

	"github.com/lucasldab/tuiclicker/internal/balance"
)

// CurrentVersion is the save file schema version. Increment when SaveData
// fields change in a backwards-incompatible way.
const CurrentVersion = 1

// SaveData holds the subset of GameModel that must survive a restart.
// All fields are exported so encoding/json can marshal them.
// Transient fields (width, height, layout cache, flash state, cursors, scroll)
// are intentionally omitted — they re-derive on first WindowSizeMsg / Init.
type SaveData struct {
	Version         int                  `json:"version"`
	SavedAt         time.Time            `json:"saved_at"`
	Ledger          LedgerData           `json:"ledger"`
	ZoneUnlocked    [3]bool              `json:"zone_unlocked"`
	MutationStates  []MutationStateData  `json:"mutation_states"`
	HarvesterStates []HarvesterStateData `json:"harvester_states"`
}

// LedgerData holds the serializable subset of ResourceLedger.
type LedgerData struct {
	Amounts [3]float64 `json:"amounts"`
	Rates   [3]float64 `json:"rates"` // rates at save time — used for offline calculation
}

// MutationStateData holds the durable state for a single mutation slot.
type MutationStateData struct {
	PurchaseCount int `json:"purchase_count"`
}

// HarvesterStateData holds the durable state for a single harvester slot.
type HarvesterStateData struct {
	Owned int `json:"owned"`
}

// SavePath returns the canonical save file path, respecting XDG_DATA_HOME.
// Falls back to ~/.local/share/tuiclicker/save.json if XDG_DATA_HOME is unset.
func SavePath() (string, error) {
	dataHome := os.Getenv("XDG_DATA_HOME")
	if dataHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("persistence: home dir: %w", err)
		}
		dataHome = filepath.Join(home, ".local", "share")
	}
	return filepath.Join(dataHome, "tuiclicker", "save.json"), nil
}

// Save atomically writes sd to savePath using a temp-file + rename pattern.
// The temp file is created in the same directory as savePath to guarantee
// a same-filesystem rename (avoids EXDEV on split /tmp mounts).
// The written content is validated by re-reading before the rename is committed.
func Save(sd SaveData, savePath string) error {
	data, err := json.Marshal(sd)
	if err != nil {
		return fmt.Errorf("persistence: marshal: %w", err)
	}

	dir := filepath.Dir(savePath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("persistence: mkdir: %w", err)
	}

	// Write to temp file in same directory (same filesystem — atomic rename guarantee).
	tmp, err := os.CreateTemp(dir, "save-*.json.tmp")
	if err != nil {
		return fmt.Errorf("persistence: create temp: %w", err)
	}
	tmpPath := tmp.Name()

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("persistence: write temp: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("persistence: close temp: %w", err)
	}

	// Validate written content (parse it back before committing).
	rawCheck, err := os.ReadFile(tmpPath)
	if err != nil || json.Unmarshal(rawCheck, &SaveData{}) != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("persistence: validation failed")
	}

	// Atomic rename — replaces existing save file if present.
	if err := os.Rename(tmpPath, savePath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("persistence: rename: %w", err)
	}
	return nil
}

// Load reads and unmarshals the save file at savePath.
// Returns (SaveData{}, false, nil) when the file does not exist (first launch).
// Returns (SaveData{}, false, err) when the file exists but cannot be parsed.
// Returns (sd, true, nil) on success.
func Load(savePath string) (SaveData, bool, error) {
	data, err := os.ReadFile(savePath)
	if err != nil {
		if os.IsNotExist(err) {
			return SaveData{}, false, nil
		}
		return SaveData{}, false, err
	}
	var sd SaveData
	if err := json.Unmarshal(data, &sd); err != nil {
		return SaveData{}, false, fmt.Errorf("persistence: corrupt save: %w", err)
	}
	return sd, true, nil
}

// ApplyOfflineProgress adds passive income earned since sd.SavedAt to sd.Ledger.Amounts.
// Uses rates stored in sd.Ledger.Rates (rates at save time — not current rates).
// Offline credit is capped at balance.OfflineCapSeconds (4 hours = 14400 s).
// No-op if elapsed <= 0 (clock skew or future-dated save).
func ApplyOfflineProgress(sd *SaveData, now time.Time) {
	elapsed := now.Sub(sd.SavedAt).Seconds()
	if elapsed <= 0 {
		return
	}
	capped := math.Min(elapsed, balance.OfflineCapSeconds)
	for i := range sd.Ledger.Amounts {
		sd.Ledger.Amounts[i] += sd.Ledger.Rates[i] * capped
	}
}

