package model

import (
	"time"

	"github.com/lucasldab/tuiclicker/internal/persistence"
)

// ToSaveData converts a GameModel to a persistence.SaveData snapshot.
// Sets Version and SavedAt automatically.
func ToSaveData(m GameModel) persistence.SaveData {
	sd := persistence.SaveData{
		Version:      persistence.CurrentVersion,
		SavedAt:      time.Now(),
		Ledger:       persistence.LedgerData{Amounts: m.Ledger.Amounts, Rates: m.Ledger.Rates},
		ZoneUnlocked: m.ZoneUnlocked,
	}

	sd.MutationStates = make([]persistence.MutationStateData, len(m.MutationStates))
	for i, s := range m.MutationStates {
		sd.MutationStates[i] = persistence.MutationStateData{PurchaseCount: s.PurchaseCount}
	}

	sd.HarvesterStates = make([]persistence.HarvesterStateData, len(m.HarvesterStates))
	for i, s := range m.HarvesterStates {
		sd.HarvesterStates[i] = persistence.HarvesterStateData{Owned: s.Owned}
	}

	return sd
}

// FromSaveData reconstructs a GameModel from a persistence.SaveData snapshot.
// Calls New() first to populate all defaults (including unexported fields),
// then overwrites the serializable fields from sd.
// Ends with RecalcAllRates to fix up rates from restored harvester/mutation state.
func FromSaveData(sd persistence.SaveData) GameModel {
	m := New() // establishes defaults for all unexported fields

	m.Ledger.Amounts = sd.Ledger.Amounts
	m.Ledger.Rates = sd.Ledger.Rates
	m.ZoneUnlocked = sd.ZoneUnlocked

	for i, s := range sd.MutationStates {
		if i < len(m.MutationStates) {
			m.MutationStates[i].PurchaseCount = s.PurchaseCount
		}
	}
	for i, s := range sd.HarvesterStates {
		if i < len(m.HarvesterStates) {
			m.HarvesterStates[i].Owned = s.Owned
		}
	}

	// Recompute rates from restored harvester/mutation state.
	m.Ledger = RecalcAllRates(m)
	return m
}
