package model

import (
	"math"

	"github.com/lucasldab/tuiclicker/internal/balance"
)

// HarvesterDef is the static definition of one purchasable auto-harvester.
// Instances are never modified — runtime state lives in HarvesterState.
type HarvesterDef struct {
	ID       int
	Name     string        // body horror name per UI-SPEC
	Branch   BranchID      // which mutation branch this belongs to (for buff lookup)
	Resource ResourceType  // which resource it generates each tick
	BaseRate float64       // units per second per owned (before mutation buffs)
	BaseCost []ResourceCost // base costs, scaled per purchase (see CurrentCost)
}

// HarvesterState tracks how many of a given harvester have been purchased.
type HarvesterState struct {
	Owned int
}

// CurrentCost returns the scaled cost for the next purchase.
// Formula: each BaseCost amount * HarvesterCostScale^Owned (D-13: 1.25x per purchase).
func (d HarvesterDef) CurrentCost(state HarvesterState) []ResourceCost {
	n := float64(state.Owned)
	scaled := make([]ResourceCost, len(d.BaseCost))
	for i, c := range d.BaseCost {
		scaled[i] = ResourceCost{
			Resource: c.Resource,
			Amount:   c.Amount * math.Pow(balance.HarvesterCostScale, n),
		}
	}
	return scaled
}

// EffectiveRate returns the total output per second for this harvester:
//
//	Owned * BaseRate * buffMultiplier
//
// buffMultiplier should be BranchHarvesterBuff() for this harvester's branch.
// Returns 0.0 when Owned == 0 (no output if none purchased).
func (d HarvesterDef) EffectiveRate(state HarvesterState, buffMultiplier float64) float64 {
	if state.Owned == 0 {
		return 0.0
	}
	return float64(state.Owned) * d.BaseRate * buffMultiplier
}

// AllHarvesters is the static registry of all 6 harvesters (2 per resource type).
// IDs must match slice indices. Never modify this slice at runtime.
var AllHarvesters = []HarvesterDef{
	// --- Blood harvesters ---
	{
		ID:       0,
		Name:     "Bleeding Vessel",
		Branch:   BranchBlood,
		Resource: ResourceBlood,
		BaseRate: balance.HarvTier1Rate,
		BaseCost: []ResourceCost{
			{Resource: ResourceBlood, Amount: balance.HarvBloodT1Cost},
		},
	},
	{
		ID:       1,
		Name:     "Throbbing Artery",
		Branch:   BranchBlood,
		Resource: ResourceBlood,
		BaseRate: balance.HarvTier2Rate,
		BaseCost: []ResourceCost{
			{Resource: ResourceBlood, Amount: balance.HarvBloodT2CostB},
			{Resource: ResourceFlesh, Amount: balance.HarvBloodT2CostF},
		},
	},
	// --- Flesh harvesters ---
	{
		ID:       2,
		Name:     "Crawling Tendril",
		Branch:   BranchFlesh,
		Resource: ResourceFlesh,
		BaseRate: balance.HarvTier1Rate,
		BaseCost: []ResourceCost{
			{Resource: ResourceFlesh, Amount: balance.HarvFleshT1Cost},
		},
	},
	{
		ID:       3,
		Name:     "Writhing Mass",
		Branch:   BranchFlesh,
		Resource: ResourceFlesh,
		BaseRate: balance.HarvTier2Rate,
		BaseCost: []ResourceCost{
			{Resource: ResourceFlesh, Amount: balance.HarvFleshT2CostF},
			{Resource: ResourceBones, Amount: balance.HarvFleshT2CostBo},
		},
	},
	// --- Bones harvesters ---
	{
		ID:       4,
		Name:     "Grinding Shard",
		Branch:   BranchBones,
		Resource: ResourceBones,
		BaseRate: balance.HarvTier1Rate,
		BaseCost: []ResourceCost{
			{Resource: ResourceBones, Amount: balance.HarvBoneT1Cost},
		},
	},
	{
		ID:       5,
		Name:     "Calcified Spine",
		Branch:   BranchBones,
		Resource: ResourceBones,
		BaseRate: balance.HarvTier2Rate,
		BaseCost: []ResourceCost{
			{Resource: ResourceBones, Amount: balance.HarvBoneT2CostBo},
			{Resource: ResourceBlood, Amount: balance.HarvBoneT2CostBl},
		},
	},
}

// TryPurchaseHarvester attempts to buy the harvester at index idx from the
// player's current ledger.
//
//   - Returns the updated GameModel and true on success (costs deducted, Owned++,
//     rates recomputed).
//   - Returns the unchanged GameModel and false if idx is out of range or the
//     ledger cannot afford the current cost (T-02-05: bounds-check guard).
func TryPurchaseHarvester(m GameModel, idx int) (GameModel, bool) {
	// Bounds-check guard (T-02-05)
	if idx < 0 || idx >= len(m.HarvesterDefs) {
		return m, false
	}

	def := m.HarvesterDefs[idx]
	state := m.HarvesterStates[idx]
	costs := def.CurrentCost(state)

	if !CanAfford(costs, m.Ledger) {
		return m, false
	}

	for _, c := range costs {
		m.Ledger.Amounts[c.Resource] -= c.Amount
	}
	m.HarvesterStates[idx].Owned++

	// Recompute rates from scratch after purchase (T-02-06: prevents rate drift)
	m.Ledger = RecalcAllRates(m)
	return m, true
}

// RecalcAllRates recomputes Ledger.Rates for all resources from scratch.
// Must be called after any harvester purchase or mutation purchase that
// affects harvester buff multipliers.
//
// IMPORTANT: Rates are zeroed before summing (T-02-06: no state accumulation).
// Caller must NOT add to Rates after this call — the result is authoritative.
func RecalcAllRates(m GameModel) ResourceLedger {
	ledger := m.Ledger
	// Zero rates first — prevent drift from prior state (Pitfall 1 / T-02-06)
	for r := range ledger.Rates {
		ledger.Rates[r] = 0.0
	}
	for i, def := range m.HarvesterDefs {
		buff := BranchHarvesterBuff(m.MutationDefs, m.MutationStates, def.Branch)
		rate := def.EffectiveRate(m.HarvesterStates[i], buff)
		ledger.Rates[def.Resource] += rate
	}
	return ledger
}
