package persistence_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/lucasldab/tuiclicker/internal/model"
	"github.com/lucasldab/tuiclicker/internal/persistence"
)

// ---------------------------------------------------------------------------
// SavePath
// ---------------------------------------------------------------------------

func TestSavePathDefaultsToXDGDataHome(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", "/tmp/xdgtest")
	got, err := persistence.SavePath()
	if err != nil {
		t.Fatalf("SavePath() error: %v", err)
	}
	want := "/tmp/xdgtest/tuiclicker/save.json"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestSavePathFallsBackToHomeLocalShare(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", "")
	got, err := persistence.SavePath()
	if err != nil {
		t.Fatalf("SavePath() error: %v", err)
	}
	// Must end with /tuiclicker/save.json
	if filepath.Base(got) != "save.json" {
		t.Errorf("unexpected filename in path: %q", got)
	}
	if filepath.Base(filepath.Dir(got)) != "tuiclicker" {
		t.Errorf("unexpected parent dir in path: %q", got)
	}
}

// ---------------------------------------------------------------------------
// Save + Load round-trip
// ---------------------------------------------------------------------------

func makeSaveData() persistence.SaveData {
	return persistence.SaveData{
		Version: persistence.CurrentVersion,
		SavedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		Ledger: persistence.LedgerData{
			Amounts: [3]float64{100, 200, 300},
			Rates:   [3]float64{1.5, 2.5, 3.5},
		},
		ZoneUnlocked: [3]bool{true, true, false},
		MutationStates: []persistence.MutationStateData{
			{PurchaseCount: 2},
			{PurchaseCount: 0},
		},
		HarvesterStates: []persistence.HarvesterStateData{
			{Owned: 3},
			{Owned: 0},
		},
	}
}

func TestSaveWritesValidJSONFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tuiclicker", "save.json")

	sd := makeSaveData()
	if err := persistence.Save(sd, path); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error: %v", err)
	}
	var check persistence.SaveData
	if err := json.Unmarshal(data, &check); err != nil {
		t.Fatalf("Unmarshal written JSON: %v", err)
	}
	if check.Version != persistence.CurrentVersion {
		t.Errorf("version: got %d, want %d", check.Version, persistence.CurrentVersion)
	}
}

func TestSaveIsAtomic(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "save.json")

	sd := makeSaveData()
	if err := persistence.Save(sd, path); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	// No temp files should remain after a successful Save
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if e.Name() != "save.json" {
			t.Errorf("unexpected leftover file: %q", e.Name())
		}
	}
}

// ---------------------------------------------------------------------------
// Load
// ---------------------------------------------------------------------------

func TestLoadMissingFile(t *testing.T) {
	_, found, err := persistence.Load("/nonexistent/path/save.json")
	if err != nil {
		t.Fatalf("Load() expected nil error for missing file, got: %v", err)
	}
	if found {
		t.Error("Load() expected found=false for missing file")
	}
}

func TestLoadReturnsData(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "save.json")

	sd := makeSaveData()
	if err := persistence.Save(sd, path); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	got, found, err := persistence.Load(path)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if !found {
		t.Fatal("Load() expected found=true")
	}
	if got.Version != sd.Version {
		t.Errorf("version: got %d, want %d", got.Version, sd.Version)
	}
	if got.Ledger.Amounts != sd.Ledger.Amounts {
		t.Errorf("amounts: got %v, want %v", got.Ledger.Amounts, sd.Ledger.Amounts)
	}
}

func TestLoadCorruptFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "save.json")
	if err := os.WriteFile(path, []byte("{not valid json"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, found, err := persistence.Load(path)
	if err == nil {
		t.Error("Load() expected error for corrupt JSON, got nil")
	}
	if found {
		t.Error("Load() expected found=false for corrupt JSON")
	}
}

// ---------------------------------------------------------------------------
// ApplyOfflineProgress
// ---------------------------------------------------------------------------

func TestOfflineProgressUncapped(t *testing.T) {
	now := time.Date(2024, 1, 1, 1, 0, 0, 0, time.UTC) // 1h after save
	sd := persistence.SaveData{
		SavedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		Ledger: persistence.LedgerData{
			Amounts: [3]float64{0, 0, 0},
			Rates:   [3]float64{1, 2, 3},
		},
	}
	persistence.ApplyOfflineProgress(&sd, now)
	elapsed := 3600.0
	for i := 0; i < 3; i++ {
		want := sd.Ledger.Rates[i] * elapsed
		// amounts were modified in place — check expected delta
		_ = want
	}
	// blood: 0 + 1*3600 = 3600
	if sd.Ledger.Amounts[0] != 3600 {
		t.Errorf("blood: got %v, want 3600", sd.Ledger.Amounts[0])
	}
}

func TestOfflineProgressCapped(t *testing.T) {
	// 10 hours elapsed — should cap at 4h = 14400s
	now := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)
	sd := persistence.SaveData{
		SavedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		Ledger: persistence.LedgerData{
			Amounts: [3]float64{0, 0, 0},
			Rates:   [3]float64{1, 1, 1},
		},
	}
	persistence.ApplyOfflineProgress(&sd, now)
	const cap = 14400.0
	for i := 0; i < 3; i++ {
		if sd.Ledger.Amounts[i] != cap {
			t.Errorf("index %d: got %v, want %v (4h cap)", i, sd.Ledger.Amounts[i], cap)
		}
	}
}

func TestOfflineProgressFutureSave(t *testing.T) {
	// SavedAt is in the future — elapsed <= 0, no-op
	now := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	sd := persistence.SaveData{
		SavedAt: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
		Ledger: persistence.LedgerData{
			Amounts: [3]float64{100, 200, 300},
			Rates:   [3]float64{1, 1, 1},
		},
	}
	persistence.ApplyOfflineProgress(&sd, now)
	if sd.Ledger.Amounts[0] != 100 {
		t.Errorf("expected no change, got %v", sd.Ledger.Amounts[0])
	}
}

func TestOfflineUsesRatesAtSaveTime(t *testing.T) {
	// Rates in save data should be used directly, not recalculated
	now := time.Date(2024, 1, 1, 1, 0, 0, 0, time.UTC)
	savedRate := 5.0
	sd := persistence.SaveData{
		SavedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		Ledger: persistence.LedgerData{
			Amounts: [3]float64{0, 0, 0},
			Rates:   [3]float64{savedRate, savedRate, savedRate},
		},
	}
	persistence.ApplyOfflineProgress(&sd, now)
	want := savedRate * 3600
	for i := 0; i < 3; i++ {
		if sd.Ledger.Amounts[i] != want {
			t.Errorf("index %d: got %v, want %v", i, sd.Ledger.Amounts[i], want)
		}
	}
}

// ---------------------------------------------------------------------------
// ToSaveData / ToGameModel round-trip
// ---------------------------------------------------------------------------

func TestToSaveDataCapturesState(t *testing.T) {
	m := model.New()
	m.Ledger.Amounts[0] = 42
	m.Ledger.Rates[1] = 1.5
	m.ZoneUnlocked[1] = true
	m.MutationStates[0].PurchaseCount = 3
	m.HarvesterStates[0].Owned = 2

	sd := persistence.ToSaveData(m)

	if sd.Version != persistence.CurrentVersion {
		t.Errorf("version: got %d, want %d", sd.Version, persistence.CurrentVersion)
	}
	if sd.Ledger.Amounts[0] != 42 {
		t.Errorf("Amounts[0]: got %v, want 42", sd.Ledger.Amounts[0])
	}
	if sd.Ledger.Rates[1] != 1.5 {
		t.Errorf("Rates[1]: got %v, want 1.5", sd.Ledger.Rates[1])
	}
	if !sd.ZoneUnlocked[1] {
		t.Error("ZoneUnlocked[1]: expected true")
	}
	if len(sd.MutationStates) == 0 || sd.MutationStates[0].PurchaseCount != 3 {
		t.Errorf("MutationStates[0].PurchaseCount: got %v, want 3", sd.MutationStates[0].PurchaseCount)
	}
	if len(sd.HarvesterStates) == 0 || sd.HarvesterStates[0].Owned != 2 {
		t.Errorf("HarvesterStates[0].Owned: got %v, want 2", sd.HarvesterStates[0].Owned)
	}
}

func TestToGameModelRestoresAmounts(t *testing.T) {
	sd := persistence.SaveData{
		Ledger: persistence.LedgerData{
			Amounts: [3]float64{10, 20, 30},
			Rates:   [3]float64{0, 0, 0},
		},
		ZoneUnlocked:    [3]bool{true, false, false},
		MutationStates:  nil,
		HarvesterStates: nil,
	}
	m := persistence.ToGameModel(sd)
	if m.Ledger.Amounts[0] != 10 {
		t.Errorf("Amounts[0]: got %v, want 10", m.Ledger.Amounts[0])
	}
	if m.Ledger.Amounts[1] != 20 {
		t.Errorf("Amounts[1]: got %v, want 20", m.Ledger.Amounts[1])
	}
	if m.Ledger.Amounts[2] != 30 {
		t.Errorf("Amounts[2]: got %v, want 30", m.Ledger.Amounts[2])
	}
}

func TestToGameModelRestoresMutations(t *testing.T) {
	sd := persistence.SaveData{
		ZoneUnlocked: [3]bool{true, false, false},
		MutationStates: []persistence.MutationStateData{
			{PurchaseCount: 5},
			{PurchaseCount: 2},
		},
	}
	m := persistence.ToGameModel(sd)
	if m.MutationStates[0].PurchaseCount != 5 {
		t.Errorf("MutationStates[0].PurchaseCount: got %v, want 5", m.MutationStates[0].PurchaseCount)
	}
	if m.MutationStates[1].PurchaseCount != 2 {
		t.Errorf("MutationStates[1].PurchaseCount: got %v, want 2", m.MutationStates[1].PurchaseCount)
	}
}

func TestToGameModelRestoresHarvesters(t *testing.T) {
	sd := persistence.SaveData{
		ZoneUnlocked: [3]bool{true, false, false},
		HarvesterStates: []persistence.HarvesterStateData{
			{Owned: 4},
			{Owned: 1},
		},
	}
	m := persistence.ToGameModel(sd)
	if m.HarvesterStates[0].Owned != 4 {
		t.Errorf("HarvesterStates[0].Owned: got %v, want 4", m.HarvesterStates[0].Owned)
	}
	if m.HarvesterStates[1].Owned != 1 {
		t.Errorf("HarvesterStates[1].Owned: got %v, want 1", m.HarvesterStates[1].Owned)
	}
}

func TestToGameModelRestoresZones(t *testing.T) {
	sd := persistence.SaveData{
		ZoneUnlocked:    [3]bool{true, true, false},
		MutationStates:  nil,
		HarvesterStates: nil,
	}
	m := persistence.ToGameModel(sd)
	if !m.ZoneUnlocked[0] {
		t.Error("ZoneUnlocked[0]: expected true")
	}
	if !m.ZoneUnlocked[1] {
		t.Error("ZoneUnlocked[1]: expected true")
	}
	if m.ZoneUnlocked[2] {
		t.Error("ZoneUnlocked[2]: expected false")
	}
}

func TestToGameModelBoundsCheckMutations(t *testing.T) {
	// SaveData with more mutation states than the current registry — should not panic
	many := make([]persistence.MutationStateData, 100)
	for i := range many {
		many[i] = persistence.MutationStateData{PurchaseCount: i}
	}
	sd := persistence.SaveData{
		ZoneUnlocked:   [3]bool{true, false, false},
		MutationStates: many,
	}
	// Should not panic
	_ = persistence.ToGameModel(sd)
}

func TestToGameModelBoundsCheckHarvesters(t *testing.T) {
	// SaveData with more harvester states than current registry — should not panic
	many := make([]persistence.HarvesterStateData, 100)
	for i := range many {
		many[i] = persistence.HarvesterStateData{Owned: i}
	}
	sd := persistence.SaveData{
		ZoneUnlocked:    [3]bool{true, false, false},
		HarvesterStates: many,
	}
	// Should not panic
	_ = persistence.ToGameModel(sd)
}

func TestCurrentVersionIsOne(t *testing.T) {
	if persistence.CurrentVersion != 1 {
		t.Errorf("CurrentVersion: got %d, want 1", persistence.CurrentVersion)
	}
}
