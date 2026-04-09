package model

import (
	"math"
	"testing"
)

// --- AllMutations registry ---

func TestMutationRegistryHasNineEntries(t *testing.T) {
	if len(AllMutations) != 9 {
		t.Errorf("AllMutations should have 9 entries, got %d", len(AllMutations))
	}
}

func TestMutationRegistryBranchCounts(t *testing.T) {
	counts := [3]int{}
	for _, m := range AllMutations {
		counts[int(m.Branch)]++
	}
	if counts[BranchBlood] != 3 {
		t.Errorf("BranchBlood should have 3 mutations, got %d", counts[BranchBlood])
	}
	if counts[BranchFlesh] != 3 {
		t.Errorf("BranchFlesh should have 3 mutations, got %d", counts[BranchFlesh])
	}
	if counts[BranchBones] != 3 {
		t.Errorf("BranchBones should have 3 mutations, got %d", counts[BranchBones])
	}
}

func TestMutationRegistryOrdering(t *testing.T) {
	expectedBranches := []BranchID{
		BranchBlood, BranchBlood, BranchBlood,
		BranchFlesh, BranchFlesh, BranchFlesh,
		BranchBones, BranchBones, BranchBones,
	}
	for i, m := range AllMutations {
		if m.Branch != expectedBranches[i] {
			t.Errorf("AllMutations[%d].Branch = %v, want %v", i, m.Branch, expectedBranches[i])
		}
		if m.ID != i {
			t.Errorf("AllMutations[%d].ID = %d, want %d", i, m.ID, i)
		}
	}
}

func TestMutationNonEmptyNamesAndDescriptions(t *testing.T) {
	for i, m := range AllMutations {
		if m.Name == "" {
			t.Errorf("AllMutations[%d].Name is empty", i)
		}
		if m.Description == "" {
			t.Errorf("AllMutations[%d].Description is empty", i)
		}
	}
}

// --- BranchID constants ---

func TestBranchIDValues(t *testing.T) {
	if BranchBlood != 0 {
		t.Errorf("BranchBlood should be 0, got %d", BranchBlood)
	}
	if BranchFlesh != 1 {
		t.Errorf("BranchFlesh should be 1, got %d", BranchFlesh)
	}
	if BranchBones != 2 {
		t.Errorf("BranchBones should be 2, got %d", BranchBones)
	}
}

// --- CurrentCost ---

func TestMutationCurrentCostZeroPurchases(t *testing.T) {
	// ID 0: blood-only cost (MutBloodPowerBase=10.0)
	def := AllMutations[0]
	costs := def.CurrentCost(MutationState{PurchaseCount: 0})
	if len(costs) != 1 {
		t.Fatalf("expected 1 cost entry, got %d", len(costs))
	}
	if costs[0].Resource != ResourceBlood {
		t.Errorf("expected ResourceBlood, got %v", costs[0].Resource)
	}
	if costs[0].Amount != 10.0 {
		t.Errorf("expected 10.0, got %v", costs[0].Amount)
	}
}

func TestMutationCurrentCostThreePurchases(t *testing.T) {
	// ID 0: base cost 10.0, after 3 purchases: 10.0 * 1.25^3 = 10.0 * 1.953125 = 19.53125
	def := AllMutations[0]
	costs := def.CurrentCost(MutationState{PurchaseCount: 3})
	if len(costs) != 1 {
		t.Fatalf("expected 1 cost entry, got %d", len(costs))
	}
	expected := 10.0 * math.Pow(1.25, 3)
	if math.Abs(costs[0].Amount-expected) > 1e-9 {
		t.Errorf("expected %v, got %v", expected, costs[0].Amount)
	}
}

func TestMutationCurrentCostMultiResource(t *testing.T) {
	// ID 2: blood+flesh defense — MutBloodDefBlood=15.0, MutBloodDefFlesh=10.0
	def := AllMutations[2]
	costs := def.CurrentCost(MutationState{PurchaseCount: 0})
	if len(costs) != 2 {
		t.Fatalf("expected 2 cost entries, got %d", len(costs))
	}
}

// --- CanAfford ---

func TestCanAffordAllSufficient(t *testing.T) {
	ledger := ResourceLedger{}
	ledger.Amounts[ResourceBlood] = 100.0
	costs := []ResourceCost{{Resource: ResourceBlood, Amount: 10.0}}
	if !CanAfford(costs, ledger) {
		t.Error("should afford when ledger has enough")
	}
}

func TestCanAffordExactlyEnough(t *testing.T) {
	ledger := ResourceLedger{}
	ledger.Amounts[ResourceBlood] = 10.0
	costs := []ResourceCost{{Resource: ResourceBlood, Amount: 10.0}}
	if !CanAfford(costs, ledger) {
		t.Error("should afford when ledger has exact amount")
	}
}

func TestCanAffordInsufficientOneSingle(t *testing.T) {
	ledger := ResourceLedger{}
	ledger.Amounts[ResourceBlood] = 5.0
	costs := []ResourceCost{{Resource: ResourceBlood, Amount: 10.0}}
	if CanAfford(costs, ledger) {
		t.Error("should not afford when ledger has less than cost")
	}
}

func TestCanAffordInsufficientOneOfTwo(t *testing.T) {
	ledger := ResourceLedger{}
	ledger.Amounts[ResourceBlood] = 100.0
	ledger.Amounts[ResourceFlesh] = 5.0
	costs := []ResourceCost{
		{Resource: ResourceBlood, Amount: 10.0},
		{Resource: ResourceFlesh, Amount: 10.0},
	}
	if CanAfford(costs, ledger) {
		t.Error("should not afford when one resource is insufficient")
	}
}

// --- FormatCosts ---

func TestFormatCostsSingleResource(t *testing.T) {
	costs := []ResourceCost{{Resource: ResourceBlood, Amount: 10.0}}
	got := FormatCosts(costs)
	want := "10 blood"
	if got != want {
		t.Errorf("FormatCosts single = %q, want %q", got, want)
	}
}

func TestFormatCostsTwoResources(t *testing.T) {
	costs := []ResourceCost{
		{Resource: ResourceBlood, Amount: 15.0},
		{Resource: ResourceFlesh, Amount: 10.0},
	}
	got := FormatCosts(costs)
	want := "15 blood / 10 flesh"
	if got != want {
		t.Errorf("FormatCosts two = %q, want %q", got, want)
	}
}

func TestFormatCostsBonesResource(t *testing.T) {
	costs := []ResourceCost{{Resource: ResourceBones, Amount: 20.0}}
	got := FormatCosts(costs)
	want := "20 bones"
	if got != want {
		t.Errorf("FormatCosts bones = %q, want %q", got, want)
	}
}

func TestFormatCostsFleshResource(t *testing.T) {
	costs := []ResourceCost{{Resource: ResourceFlesh, Amount: 25.0}}
	got := FormatCosts(costs)
	want := "25 flesh"
	if got != want {
		t.Errorf("FormatCosts flesh = %q, want %q", got, want)
	}
}

// --- TryPurchaseMutation ---

// makeTestModel returns a GameModel with AllMutations wired and zeroed states.
func makeTestModel() GameModel {
	m := GameModel{}
	m.MutationDefs = make([]MutationDef, len(AllMutations))
	copy(m.MutationDefs, AllMutations)
	m.MutationStates = make([]MutationState, len(AllMutations))
	m.ZoneUnlocked[ResourceBlood] = true
	return m
}

func TestTryPurchaseMutationInsufficientResourcesReturnsFalse(t *testing.T) {
	m := makeTestModel()
	// ID 0 costs 10 blood — ledger has 0
	_, ok := TryPurchaseMutation(m, 0)
	if ok {
		t.Error("should return false when ledger insufficient")
	}
}

func TestTryPurchaseMutationModelUnchangedOnFailure(t *testing.T) {
	m := makeTestModel()
	m2, _ := TryPurchaseMutation(m, 0)
	if m2.Ledger.Amounts[ResourceBlood] != 0 {
		t.Error("ledger should be unchanged on failed purchase")
	}
	if m2.MutationStates[0].PurchaseCount != 0 {
		t.Error("PurchaseCount should be unchanged on failed purchase")
	}
}

func TestTryPurchaseMutationDeductsResourcesOnSuccess(t *testing.T) {
	m := makeTestModel()
	m.Ledger.Amounts[ResourceBlood] = 100.0
	m2, ok := TryPurchaseMutation(m, 0)
	if !ok {
		t.Fatal("should succeed when ledger has enough")
	}
	// ID 0 costs MutBloodPowerBase=10.0 blood
	expected := 100.0 - 10.0
	if m2.Ledger.Amounts[ResourceBlood] != expected {
		t.Errorf("expected blood=%v after purchase, got %v", expected, m2.Ledger.Amounts[ResourceBlood])
	}
}

func TestTryPurchaseMutationIncrementsPurchaseCount(t *testing.T) {
	m := makeTestModel()
	m.Ledger.Amounts[ResourceBlood] = 100.0
	m2, _ := TryPurchaseMutation(m, 0)
	if m2.MutationStates[0].PurchaseCount != 1 {
		t.Errorf("PurchaseCount should be 1 after first purchase, got %d", m2.MutationStates[0].PurchaseCount)
	}
}

func TestTryPurchaseMutationMultiResourceDeductsAll(t *testing.T) {
	m := makeTestModel()
	// ID 2: costs MutBloodDefBlood=15 blood + MutBloodDefFlesh=10 flesh
	m.Ledger.Amounts[ResourceBlood] = 50.0
	m.Ledger.Amounts[ResourceFlesh] = 50.0
	m2, ok := TryPurchaseMutation(m, 2)
	if !ok {
		t.Fatal("should succeed with enough resources")
	}
	if m2.Ledger.Amounts[ResourceBlood] != 35.0 {
		t.Errorf("blood should be 35.0, got %v", m2.Ledger.Amounts[ResourceBlood])
	}
	if m2.Ledger.Amounts[ResourceFlesh] != 40.0 {
		t.Errorf("flesh should be 40.0, got %v", m2.Ledger.Amounts[ResourceFlesh])
	}
}

func TestTryPurchaseMutationOutOfBoundsReturnsFalse(t *testing.T) {
	m := makeTestModel()
	_, ok := TryPurchaseMutation(m, -1)
	if ok {
		t.Error("out-of-range idx=-1 should return false")
	}
	_, ok = TryPurchaseMutation(m, 100)
	if ok {
		t.Error("out-of-range idx=100 should return false")
	}
}

// --- Zone unlock side effects (D-05) ---

func TestFirstFleshMutationUnlocksFleshZone(t *testing.T) {
	m := makeTestModel()
	// ID 3 is the flesh gateway: costs MutFleshPowerBlood=15 blood
	m.Ledger.Amounts[ResourceBlood] = 100.0
	m2, ok := TryPurchaseMutation(m, 3)
	if !ok {
		t.Fatal("purchase should succeed")
	}
	if !m2.ZoneUnlocked[ResourceFlesh] {
		t.Error("ZoneUnlocked[ResourceFlesh] should be true after first flesh mutation")
	}
}

func TestFirstBonesMutationUnlocksBonesZone(t *testing.T) {
	m := makeTestModel()
	// ID 6 is the bones gateway: costs MutBonesPowerFlesh=15 flesh
	m.Ledger.Amounts[ResourceFlesh] = 100.0
	m2, ok := TryPurchaseMutation(m, 6)
	if !ok {
		t.Fatal("purchase should succeed")
	}
	if !m2.ZoneUnlocked[ResourceBones] {
		t.Error("ZoneUnlocked[ResourceBones] should be true after first bones mutation")
	}
}

func TestBloodMutationDoesNotUnlockFleshOrBones(t *testing.T) {
	m := makeTestModel()
	m.Ledger.Amounts[ResourceBlood] = 100.0
	m2, _ := TryPurchaseMutation(m, 0) // blood mutation
	if m2.ZoneUnlocked[ResourceFlesh] {
		t.Error("blood mutation should not unlock flesh zone")
	}
	if m2.ZoneUnlocked[ResourceBones] {
		t.Error("blood mutation should not unlock bones zone")
	}
}

func TestSubsequentFleshPurchaseDoesNotResetZone(t *testing.T) {
	m := makeTestModel()
	// First purchase of flesh mutation
	m.Ledger.Amounts[ResourceBlood] = 100.0
	m2, _ := TryPurchaseMutation(m, 3)
	// Second purchase — zone already true, must remain true
	m2.Ledger.Amounts[ResourceBlood] = 100.0
	m3, ok := TryPurchaseMutation(m2, 3)
	if !ok {
		t.Fatal("second purchase should succeed")
	}
	if !m3.ZoneUnlocked[ResourceFlesh] {
		t.Error("ZoneUnlocked[ResourceFlesh] should still be true after second purchase")
	}
	if m3.MutationStates[3].PurchaseCount != 2 {
		t.Errorf("PurchaseCount should be 2, got %d", m3.MutationStates[3].PurchaseCount)
	}
}

// --- BranchYieldMultiplier ---

func TestBranchYieldMultiplierNoPurchases(t *testing.T) {
	defs := AllMutations
	states := make([]MutationState, len(defs))
	got := BranchYieldMultiplier(defs, states, BranchBlood)
	if got != 1.0 {
		t.Errorf("expected 1.0 with no purchases, got %v", got)
	}
}

func TestBranchYieldMultiplierOnePurchase(t *testing.T) {
	defs := AllMutations
	states := make([]MutationState, len(defs))
	// ID 0 is blood branch with YieldMultiplier=0.10
	states[0].PurchaseCount = 1
	got := BranchYieldMultiplier(defs, states, BranchBlood)
	// 1.0 + 0.10*1 = 1.10
	if math.Abs(got-1.10) > 1e-9 {
		t.Errorf("expected 1.10 after 1 purchase, got %v", got)
	}
}

func TestBranchYieldMultiplierOnlyCountsBranch(t *testing.T) {
	defs := AllMutations
	states := make([]MutationState, len(defs))
	// Purchase flesh mutation (ID 3) — should not affect blood yield
	states[3].PurchaseCount = 5
	got := BranchYieldMultiplier(defs, states, BranchBlood)
	if got != 1.0 {
		t.Errorf("flesh purchases should not affect blood yield, got %v", got)
	}
}

// --- BranchHarvesterBuff ---

func TestBranchHarvesterBuffNoPurchases(t *testing.T) {
	defs := AllMutations
	states := make([]MutationState, len(defs))
	got := BranchHarvesterBuff(defs, states, BranchBlood)
	if got != 1.0 {
		t.Errorf("expected 1.0 with no purchases, got %v", got)
	}
}

func TestBranchHarvesterBuffOnePurchase(t *testing.T) {
	defs := AllMutations
	states := make([]MutationState, len(defs))
	// ID 1 is blood branch with HarvesterBuff=0.05
	states[1].PurchaseCount = 1
	got := BranchHarvesterBuff(defs, states, BranchBlood)
	// 1.0 + 0.05*1 = 1.05
	if math.Abs(got-1.05) > 1e-9 {
		t.Errorf("expected 1.05 after 1 purchase, got %v", got)
	}
}
