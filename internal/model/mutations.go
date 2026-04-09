package model

import (
	"fmt"
	"math"
	"strings"

	"github.com/lucasldab/tuiclicker/internal/balance"
)

// BranchID identifies one of the three mutation branches.
// Values match the ResourceType constants for numeric cast safety (ZoneUnlocked indexing).
type BranchID int

const (
	BranchBlood BranchID = iota // 0 — matches ResourceBlood
	BranchFlesh                  // 1 — matches ResourceFlesh
	BranchBones                  // 2 — matches ResourceBones
)

// ResourceCost represents a cost requirement for one resource type.
type ResourceCost struct {
	Resource ResourceType
	Amount   float64
}

// MutationDef is the static definition of one purchasable mutation.
// Instances are never modified — all runtime state lives in MutationState.
type MutationDef struct {
	ID          int
	Name        string        // body horror flavored name
	Description string        // 1-line flavor text
	Branch      BranchID      // organizes display order
	BaseCosts   []ResourceCost
	YieldMultiplier float64   // +N% to manual harvest yield per purchase
	HarvesterBuff   float64   // +N% to harvester output per purchase
}

// MutationState tracks runtime purchase counts for one mutation.
type MutationState struct {
	PurchaseCount int
}

// CurrentCost returns the scaled cost for the Nth purchase.
// Formula: baseCost * MutationCostScale^purchaseCount (D-13).
func (d MutationDef) CurrentCost(state MutationState) []ResourceCost {
	n := float64(state.PurchaseCount)
	scaled := make([]ResourceCost, len(d.BaseCosts))
	for i, c := range d.BaseCosts {
		scaled[i] = ResourceCost{
			Resource: c.Resource,
			Amount:   c.Amount * math.Pow(balance.MutationCostScale, n),
		}
	}
	return scaled
}

// CanAfford returns true if the ledger has enough of every required resource.
func CanAfford(costs []ResourceCost, ledger ResourceLedger) bool {
	for _, c := range costs {
		if ledger.Amounts[c.Resource] < c.Amount {
			return false
		}
	}
	return true
}

// resourceName returns the lowercase display name for a ResourceType.
func resourceName(r ResourceType) string {
	switch r {
	case ResourceBlood:
		return "blood"
	case ResourceFlesh:
		return "flesh"
	case ResourceBones:
		return "bones"
	default:
		return "unknown"
	}
}

// FormatCosts formats costs as "N resource" (single) or "N resource / M resource2" (two).
// Resource names are lowercase: blood, flesh, bones.
func FormatCosts(costs []ResourceCost) string {
	parts := make([]string, len(costs))
	for i, c := range costs {
		parts[i] = fmt.Sprintf("%d %s", int(math.Floor(c.Amount)), resourceName(c.Resource))
	}
	return strings.Join(parts, " / ")
}

// AllMutations is the static registry of all 9 purchasable mutations.
// Ordered: blood-power, blood-speed, blood-defense, flesh-power, flesh-speed,
// flesh-defense, bones-power, bones-speed, bones-defense.
// Do NOT modify — state lives in MutationState slices on GameModel.
var AllMutations = []MutationDef{
	// ── Blood branch ──────────────────────────────────────────────────────────
	{
		ID:          0,
		Name:        "Hemorrhagic Surge",
		Description: "Veins rupture and regrow thicker each time",
		Branch:      BranchBlood,
		BaseCosts:   []ResourceCost{{Resource: ResourceBlood, Amount: balance.MutBloodPowerBase}},
		YieldMultiplier: balance.MutYieldBonusPerPurchase,
		HarvesterBuff:   0,
	},
	{
		ID:          1,
		Name:        "Ichor Weeping",
		Description: "Black fluid seeps where blood should flow",
		Branch:      BranchBlood,
		BaseCosts:   []ResourceCost{{Resource: ResourceBlood, Amount: balance.MutBloodSpeedBase}},
		YieldMultiplier: balance.MutYieldBonusPerPurchase,
		HarvesterBuff:   balance.MutHarvesterBuffPerPurchase,
	},
	{
		ID:          2,
		Name:        "Clotted Carapace",
		Description: "Dried blood hardens into protective plating",
		Branch:      BranchBlood,
		BaseCosts: []ResourceCost{
			{Resource: ResourceBlood, Amount: balance.MutBloodDefBlood},
			{Resource: ResourceFlesh, Amount: balance.MutBloodDefFlesh},
		},
		YieldMultiplier: 0,
		HarvesterBuff:   balance.MutHarvesterBuffPerPurchase,
	},
	// ── Flesh branch ──────────────────────────────────────────────────────────
	// ID 3 is the gateway mutation: costs blood, unlocks flesh harvest zone (D-05).
	{
		ID:          3,
		Name:        "Proliferating Tissue",
		Description: "Meat grows faster than it can be consumed",
		Branch:      BranchFlesh,
		BaseCosts:   []ResourceCost{{Resource: ResourceBlood, Amount: balance.MutFleshPowerBlood}},
		YieldMultiplier: balance.MutYieldBonusPerPurchase,
		HarvesterBuff:   0,
	},
	{
		ID:          4,
		Name:        "Squirming Mass",
		Description: "Something moves beneath the surface",
		Branch:      BranchFlesh,
		BaseCosts:   []ResourceCost{{Resource: ResourceFlesh, Amount: balance.MutFleshSpeedBase}},
		YieldMultiplier: balance.MutYieldBonusPerPurchase,
		HarvesterBuff:   balance.MutHarvesterBuffPerPurchase,
	},
	{
		ID:          5,
		Name:        "Putrefactive Bloom",
		Description: "Rot that feeds itself, endlessly cycling",
		Branch:      BranchFlesh,
		BaseCosts: []ResourceCost{
			{Resource: ResourceFlesh, Amount: balance.MutFleshDefFlesh},
			{Resource: ResourceBones, Amount: balance.MutFleshDefBones},
		},
		YieldMultiplier: 0,
		HarvesterBuff:   balance.MutHarvesterBuffPerPurchase,
	},
	// ── Bones branch ──────────────────────────────────────────────────────────
	// ID 6 is the gateway mutation: costs flesh, unlocks bones harvest zone (D-05).
	{
		ID:          6,
		Name:        "Osseous Spike Growth",
		Description: "New bones erupt through skin without warning",
		Branch:      BranchBones,
		BaseCosts:   []ResourceCost{{Resource: ResourceFlesh, Amount: balance.MutBonesPowerFlesh}},
		YieldMultiplier: balance.MutYieldBonusPerPurchase,
		HarvesterBuff:   0,
	},
	{
		ID:          7,
		Name:        "Skeletal Haste",
		Description: "The skeleton loosens and re-knits at speed",
		Branch:      BranchBones,
		BaseCosts:   []ResourceCost{{Resource: ResourceBones, Amount: balance.MutBonesSpeedBase}},
		YieldMultiplier: balance.MutYieldBonusPerPurchase,
		HarvesterBuff:   balance.MutHarvesterBuffPerPurchase,
	},
	{
		ID:          8,
		Name:        "Calcium Fortress",
		Description: "Bone density beyond any natural limit",
		Branch:      BranchBones,
		BaseCosts: []ResourceCost{
			{Resource: ResourceBones, Amount: balance.MutBonesDefBones},
			{Resource: ResourceBlood, Amount: balance.MutBonesDefBlood},
		},
		YieldMultiplier: 0,
		HarvesterBuff:   balance.MutHarvesterBuffPerPurchase,
	},
}

// TryPurchaseMutation attempts to purchase the mutation at index idx.
// Returns the updated GameModel and true on success, or the unchanged model
// and false if the player cannot afford it or idx is out of range.
// Side effects: zone unlock when first purchase in flesh/bones branch (D-05).
func TryPurchaseMutation(m GameModel, idx int) (GameModel, bool) {
	// T-02-03: bounds-check idx before indexing
	if idx < 0 || idx >= len(m.MutationDefs) {
		return m, false
	}

	def := m.MutationDefs[idx]
	state := m.MutationStates[idx]
	costs := def.CurrentCost(state)

	if !CanAfford(costs, m.Ledger) {
		return m, false
	}

	// Deduct resources
	for _, c := range costs {
		m.Ledger.Amounts[c.Resource] -= c.Amount
	}

	// Increment purchase count
	m.MutationStates[idx].PurchaseCount++

	// Zone unlock: first purchase in flesh branch unlocks flesh zone (D-05)
	if def.Branch == BranchFlesh && m.MutationStates[idx].PurchaseCount == 1 {
		m.ZoneUnlocked[ResourceFlesh] = true
	}
	// Zone unlock: first purchase in bones branch unlocks bones zone (D-05)
	if def.Branch == BranchBones && m.MutationStates[idx].PurchaseCount == 1 {
		m.ZoneUnlocked[ResourceBones] = true
	}

	return m, true
}

// BranchYieldMultiplier returns the total yield multiplier for manual harvests
// in the given branch, accumulated across all purchases of all branch mutations.
// Returns 1.0 (base) + sum of (YieldMultiplier * PurchaseCount) for branch mutations.
func BranchYieldMultiplier(defs []MutationDef, states []MutationState, branch BranchID) float64 {
	total := 1.0
	for i, def := range defs {
		if def.Branch == branch {
			total += def.YieldMultiplier * float64(states[i].PurchaseCount)
		}
	}
	return total
}

// BranchHarvesterBuff returns 1.0 + sum of HarvesterBuff * PurchaseCount for
// all mutations in the given branch. Used by harvesters to compute effective rate.
func BranchHarvesterBuff(defs []MutationDef, states []MutationState, branch BranchID) float64 {
	total := 1.0
	for i, def := range defs {
		if def.Branch == branch {
			total += def.HarvesterBuff * float64(states[i].PurchaseCount)
		}
	}
	return total
}
