package model

import (
	"math"
	"testing"
)

// --- Task 1: HarvesterDef types, registry, and cost/rate logic ---

func TestAllHarvestersLength(t *testing.T) {
	if len(AllHarvesters) != 6 {
		t.Errorf("AllHarvesters: want 6 entries, got %d", len(AllHarvesters))
	}
}

func TestAllHarvestersNames(t *testing.T) {
	want := []string{
		"Bleeding Vessel",
		"Throbbing Artery",
		"Crawling Tendril",
		"Writhing Mass",
		"Grinding Shard",
		"Calcified Spine",
	}
	for i, name := range want {
		if AllHarvesters[i].Name != name {
			t.Errorf("AllHarvesters[%d].Name: want %q, got %q", i, name, AllHarvesters[i].Name)
		}
	}
}

func TestHarvesterResources(t *testing.T) {
	// Blood harvesters (IDs 0,1) generate blood
	if AllHarvesters[0].Resource != ResourceBlood {
		t.Errorf("AllHarvesters[0].Resource: want ResourceBlood, got %v", AllHarvesters[0].Resource)
	}
	if AllHarvesters[1].Resource != ResourceBlood {
		t.Errorf("AllHarvesters[1].Resource: want ResourceBlood, got %v", AllHarvesters[1].Resource)
	}
	// Flesh harvesters (IDs 2,3) generate flesh
	if AllHarvesters[2].Resource != ResourceFlesh {
		t.Errorf("AllHarvesters[2].Resource: want ResourceFlesh, got %v", AllHarvesters[2].Resource)
	}
	if AllHarvesters[3].Resource != ResourceFlesh {
		t.Errorf("AllHarvesters[3].Resource: want ResourceFlesh, got %v", AllHarvesters[3].Resource)
	}
	// Bones harvesters (IDs 4,5) generate bones
	if AllHarvesters[4].Resource != ResourceBones {
		t.Errorf("AllHarvesters[4].Resource: want ResourceBones, got %v", AllHarvesters[4].Resource)
	}
	if AllHarvesters[5].Resource != ResourceBones {
		t.Errorf("AllHarvesters[5].Resource: want ResourceBones, got %v", AllHarvesters[5].Resource)
	}
}

func TestHarvesterEffectiveRateZeroOwned(t *testing.T) {
	def := AllHarvesters[0] // Bleeding Vessel, BaseRate=0.5
	state := HarvesterState{Owned: 0}
	rate := def.EffectiveRate(state, 1.0)
	if rate != 0.0 {
		t.Errorf("EffectiveRate(Owned=0): want 0.0, got %f", rate)
	}
}

func TestHarvesterEffectiveRateTwoOwned(t *testing.T) {
	def := AllHarvesters[0] // Bleeding Vessel, BaseRate=0.5
	state := HarvesterState{Owned: 2}
	rate := def.EffectiveRate(state, 1.0)
	want := 1.0 // 2 * 0.5 * 1.0
	if math.Abs(rate-want) > 1e-9 {
		t.Errorf("EffectiveRate(Owned=2, buff=1.0): want %f, got %f", want, rate)
	}
}

func TestHarvesterEffectiveRateWithBuff(t *testing.T) {
	def := AllHarvesters[0] // Bleeding Vessel, BaseRate=0.5
	state := HarvesterState{Owned: 2}
	rate := def.EffectiveRate(state, 1.1)
	want := 2 * 0.5 * 1.1 // 1.1
	if math.Abs(rate-want) > 1e-9 {
		t.Errorf("EffectiveRate(Owned=2, buff=1.1): want %f, got %f", want, rate)
	}
}

func TestHarvesterCurrentCostBaseCase(t *testing.T) {
	def := AllHarvesters[0] // Bleeding Vessel: [{ResourceBlood, HarvBloodT1Cost=20}]
	state := HarvesterState{Owned: 0}
	costs := def.CurrentCost(state)
	if len(costs) != 1 {
		t.Fatalf("CurrentCost(Owned=0): want 1 cost entry, got %d", len(costs))
	}
	if costs[0].Resource != ResourceBlood {
		t.Errorf("CurrentCost(Owned=0): resource: want ResourceBlood, got %v", costs[0].Resource)
	}
	want := 20.0 // HarvBloodT1Cost
	if math.Abs(costs[0].Amount-want) > 1e-9 {
		t.Errorf("CurrentCost(Owned=0): amount: want %f, got %f", want, costs[0].Amount)
	}
}

func TestHarvesterCurrentCostScaling(t *testing.T) {
	def := AllHarvesters[0] // Bleeding Vessel: base cost 20 blood
	state := HarvesterState{Owned: 2}
	costs := def.CurrentCost(state)
	// 20 * 1.25^2 = 20 * 1.5625 = 31.25
	want := 20.0 * math.Pow(1.25, 2)
	if math.Abs(costs[0].Amount-want) > 1e-9 {
		t.Errorf("CurrentCost(Owned=2): want %f, got %f", want, costs[0].Amount)
	}
}

func TestHarvesterTier2MultiCost(t *testing.T) {
	def := AllHarvesters[1] // Throbbing Artery: [{blood,50},{flesh,30}]
	state := HarvesterState{Owned: 0}
	costs := def.CurrentCost(state)
	if len(costs) != 2 {
		t.Fatalf("Throbbing Artery CurrentCost: want 2 cost entries, got %d", len(costs))
	}
	if costs[0].Resource != ResourceBlood {
		t.Errorf("Throbbing Artery cost[0]: want ResourceBlood, got %v", costs[0].Resource)
	}
	if costs[1].Resource != ResourceFlesh {
		t.Errorf("Throbbing Artery cost[1]: want ResourceFlesh, got %v", costs[1].Resource)
	}
}

// --- Task 2: TryPurchaseHarvester and RecalcAllRates ---

func newTestGameModel() GameModel {
	m := GameModel{
		width:  80,
		height: 24,
	}
	m.ZoneUnlocked[ResourceBlood] = true
	m.HarvesterDefs = AllHarvesters
	m.HarvesterStates = make([]HarvesterState, len(AllHarvesters))
	m.MutationDefs = []MutationDef{}
	m.MutationStates = []MutationState{}
	return m
}

func TestTryPurchaseHarvesterAffordable(t *testing.T) {
	m := newTestGameModel()
	// Give enough blood to buy Bleeding Vessel (costs 20 blood)
	m.Ledger.Amounts[ResourceBlood] = 100.0

	m2, ok := TryPurchaseHarvester(m, 0)
	if !ok {
		t.Fatal("TryPurchaseHarvester: expected success, got false")
	}
	if m2.HarvesterStates[0].Owned != 1 {
		t.Errorf("TryPurchaseHarvester: Owned: want 1, got %d", m2.HarvesterStates[0].Owned)
	}
	// Blood should be deducted by 20
	wantBlood := 80.0
	if math.Abs(m2.Ledger.Amounts[ResourceBlood]-wantBlood) > 1e-9 {
		t.Errorf("TryPurchaseHarvester: blood after purchase: want %f, got %f", wantBlood, m2.Ledger.Amounts[ResourceBlood])
	}
	// Rates should be updated: Bleeding Vessel, Owned=1, BaseRate=0.5, buff=1.0 → 0.5
	wantRate := 0.5
	if math.Abs(m2.Ledger.Rates[ResourceBlood]-wantRate) > 1e-9 {
		t.Errorf("TryPurchaseHarvester: blood rate after purchase: want %f, got %f", wantRate, m2.Ledger.Rates[ResourceBlood])
	}
}

func TestTryPurchaseHarvesterUnaffordable(t *testing.T) {
	m := newTestGameModel()
	// Not enough blood (need 20, have 5)
	m.Ledger.Amounts[ResourceBlood] = 5.0

	m2, ok := TryPurchaseHarvester(m, 0)
	if ok {
		t.Fatal("TryPurchaseHarvester: expected failure, got true")
	}
	if m2.HarvesterStates[0].Owned != 0 {
		t.Errorf("TryPurchaseHarvester: Owned should remain 0, got %d", m2.HarvesterStates[0].Owned)
	}
	if m2.Ledger.Amounts[ResourceBlood] != 5.0 {
		t.Errorf("TryPurchaseHarvester: blood should be unchanged at 5.0, got %f", m2.Ledger.Amounts[ResourceBlood])
	}
}

func TestTryPurchaseHarvesterOutOfBounds(t *testing.T) {
	m := newTestGameModel()
	m.Ledger.Amounts[ResourceBlood] = 1000.0

	_, ok := TryPurchaseHarvester(m, -1)
	if ok {
		t.Error("TryPurchaseHarvester(-1): expected false for negative index")
	}

	_, ok = TryPurchaseHarvester(m, 100)
	if ok {
		t.Error("TryPurchaseHarvester(100): expected false for out-of-range index")
	}
}

func TestRecalcAllRatesWithTwoBleedingVessels(t *testing.T) {
	m := newTestGameModel()
	m.HarvesterStates[0].Owned = 2 // 2 Bleeding Vessels, BaseRate=0.5, buff=1.0

	ledger := RecalcAllRates(m)
	// 2 * 0.5 * 1.0 = 1.0
	want := 1.0
	if math.Abs(ledger.Rates[ResourceBlood]-want) > 1e-9 {
		t.Errorf("RecalcAllRates: blood rate: want %f, got %f", want, ledger.Rates[ResourceBlood])
	}
	// Other resources should be 0
	if ledger.Rates[ResourceFlesh] != 0.0 {
		t.Errorf("RecalcAllRates: flesh rate should be 0, got %f", ledger.Rates[ResourceFlesh])
	}
	if ledger.Rates[ResourceBones] != 0.0 {
		t.Errorf("RecalcAllRates: bones rate should be 0, got %f", ledger.Rates[ResourceBones])
	}
}

func TestRecalcAllRatesZeroesExistingRates(t *testing.T) {
	m := newTestGameModel()
	// Pre-populate rates with stale values
	m.Ledger.Rates[ResourceBlood] = 999.0
	m.Ledger.Rates[ResourceFlesh] = 888.0
	// No harvesters owned

	ledger := RecalcAllRates(m)
	// All rates should be 0 since no harvesters are owned
	for r := ResourceBlood; r <= ResourceBones; r++ {
		if ledger.Rates[r] != 0.0 {
			t.Errorf("RecalcAllRates: Rates[%d] should be 0 with no harvesters owned, got %f", r, ledger.Rates[r])
		}
	}
}

func TestRecalcAllRatesIdempotent(t *testing.T) {
	m := newTestGameModel()
	m.HarvesterStates[0].Owned = 1 // 1 Bleeding Vessel

	ledger1 := RecalcAllRates(m)
	// Apply to model and recalc again — should be identical
	m.Ledger = ledger1
	ledger2 := RecalcAllRates(m)

	if math.Abs(ledger1.Rates[ResourceBlood]-ledger2.Rates[ResourceBlood]) > 1e-9 {
		t.Errorf("RecalcAllRates is not idempotent: first=%f, second=%f",
			ledger1.Rates[ResourceBlood], ledger2.Rates[ResourceBlood])
	}
}
