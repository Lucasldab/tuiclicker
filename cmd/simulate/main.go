// cmd/simulate/main.go — cost curve validation tool (dev only, not part of game binary).
//
// Usage: go run ./cmd/simulate/main.go
//
// Models a player clicking at 30 clicks/min on blood (and flesh/bones once unlocked),
// with 1 Hz harvester ticks, and prints the time-to-first-purchase for every mutation
// and harvester. Exits 1 if any ratio vs previous item exceeds 5.0x or if first flesh
// mutation takes more than 3 minutes.
//
// Zone unlock model:
//   - Blood zone: always unlocked (player can click for blood from start)
//   - Flesh Power mutation costs BLOOD (gateway purchase) and unlocks the flesh zone
//   - Bones Power mutation costs FLESH (gateway purchase) and unlocks the bones zone
//
// T-02-02 mitigation: simulation is capped at MaxSimSeconds (10000 s). If the cap
// is reached before all items are purchased the script prints an error and exits 1.
package main

import (
	"fmt"
	"math"
	"os"
)

// ---------------------------------------------------------------------------
// Constants — validated by running this script.
// These are the values that will be committed to balance.go after validation.
// ---------------------------------------------------------------------------

const (
	// Player input rate: 30 clicks/min = 0.5 clicks/sec
	clicksPerSec = 0.5

	// Cost scaling per purchase (D-13)
	costScale = 1.25

	// Harvester output rates (per owned, per second)
	harvTier1Rate = 0.5
	harvTier2Rate = 2.0

	// Simulation safety cap — mitigates T-02-02 infinite-loop denial-of-service risk
	maxSimSeconds = 10000

	// --- Mutation base costs ---
	// Blood branch (all costs in blood; defense adds flesh)
	mutBloodPowerBase = 10.0 // blood only
	mutBloodSpeedBase = 25.0 // blood only
	mutBloodDefBlood  = 15.0 // blood component of defense cost
	mutBloodDefFlesh  = 10.0 // flesh component of defense cost

	// Flesh branch:
	//   Power costs BLOOD — it is the gateway mutation that unlocks the flesh zone.
	//   Speed and defense cost flesh (available only after flesh zone unlocked).
	mutFleshPowerBlood = 15.0 // blood cost of gateway (unlocks flesh zone)
	mutFleshSpeedBase  = 20.0 // flesh only
	mutFleshDefFlesh   = 15.0 // flesh component of defense cost
	mutFleshDefBones   = 10.0 // bones component of defense cost

	// Bones branch:
	//   Power costs FLESH — it is the gateway mutation that unlocks the bones zone.
	//   Speed and defense cost bones (available only after bones zone unlocked).
	mutBonesPowerFlesh = 15.0 // flesh cost of gateway (unlocks bones zone)
	mutBonesSpeedBase  = 20.0 // bones only
	mutBonesDefBones   = 15.0 // bones component of defense cost
	mutBonesDefBlood   = 10.0 // blood component of defense cost

	// --- Harvester base costs ---
	harvBloodT1Cost   = 20.0 // blood
	harvBloodT2CostB  = 50.0 // blood component
	harvBloodT2CostF  = 30.0 // flesh component
	harvFleshT1Cost   = 20.0 // flesh
	harvFleshT2CostF  = 50.0 // flesh component
	harvFleshT2CostBo = 30.0 // bones component
	harvBoneT1Cost    = 20.0 // bones
	harvBoneT2CostBo  = 50.0 // bones component
	harvBoneT2CostBl  = 30.0 // blood component
)

// ResourceType mirrors internal/model/resources.go — simulation does not import model.
type ResourceType int

const (
	ResBlood ResourceType = iota
	ResFlesh
	ResBones
)

// ResourceCost is a single component of a multi-resource cost.
type ResourceCost struct {
	Resource ResourceType
	Amount   float64
}

// SimItem represents one purchasable item in the simulation.
type SimItem struct {
	Name         string
	Costs        []ResourceCost
	UnlocksFlesh bool        // set on the Flesh Power mutation
	UnlocksBones bool        // set on the Bones Power mutation
	HarvResource ResourceType // for harvesters: which resource is generated
	HarvRate     float64      // 0 = not a harvester
}

// SimState holds dynamic simulation state.
type SimState struct {
	Amounts       [3]float64
	FleshUnlocked bool
	BonesUnlocked bool
	HarvRates     [3]float64 // total passive income/s per resource
}

func (s *SimState) canAfford(costs []ResourceCost) bool {
	for _, c := range costs {
		if s.Amounts[c.Resource] < c.Amount {
			return false
		}
	}
	return true
}

func (s *SimState) deduct(costs []ResourceCost) {
	for _, c := range costs {
		s.Amounts[c.Resource] -= c.Amount
	}
}

// scaledCosts returns costs scaled by 1.25^n (n = prior purchase count; always 0 here
// since each item is purchased once, but kept for structural correctness).
func scaledCosts(base []ResourceCost, n int) []ResourceCost {
	factor := math.Pow(costScale, float64(n))
	out := make([]ResourceCost, len(base))
	for i, c := range base {
		out[i] = ResourceCost{c.Resource, c.Amount * factor}
	}
	return out
}

// resourceAccessible returns true if every resource in costs is currently harvestable.
func (s *SimState) resourceAccessible(costs []ResourceCost) bool {
	for _, c := range costs {
		switch c.Resource {
		case ResFlesh:
			if !s.FleshUnlocked {
				return false
			}
		case ResBones:
			if !s.BonesUnlocked {
				return false
			}
		}
	}
	return true
}

func formatTime(seconds int) string {
	m := seconds / 60
	s := seconds % 60
	return fmt.Sprintf("%d:%02d", m, s)
}

func main() {
	// Simulation item list — ordered from cheapest to most expensive.
	// Gateway mutations (Flesh Power, Bones Power) unlock zones so subsequent items become accessible.
	items := []SimItem{
		// Blood branch (blood zone always unlocked)
		{Name: "Blood Power (mut)", Costs: []ResourceCost{{ResBlood, mutBloodPowerBase}}},
		{Name: "Blood Speed (mut)", Costs: []ResourceCost{{ResBlood, mutBloodSpeedBase}}},

		// Blood tier 1 harvester (unlocks passive blood income)
		{Name: "Bleeding Vessel (harv-B1)", Costs: []ResourceCost{{ResBlood, harvBloodT1Cost}}, HarvResource: ResBlood, HarvRate: harvTier1Rate},

		// Flesh gateway — costs blood, unlocks flesh zone
		{Name: "Flesh Power (mut)", Costs: []ResourceCost{{ResBlood, mutFleshPowerBlood}}, UnlocksFlesh: true},

		// Blood defense — needs flesh (available after flesh zone unlocked)
		{Name: "Blood Defense (mut)", Costs: []ResourceCost{{ResBlood, mutBloodDefBlood}, {ResFlesh, mutBloodDefFlesh}}},

		// Flesh mutations (flesh zone now available)
		{Name: "Flesh Speed (mut)", Costs: []ResourceCost{{ResFlesh, mutFleshSpeedBase}}},

		// Flesh tier 1 harvester
		{Name: "Crawling Tendril (harv-F1)", Costs: []ResourceCost{{ResFlesh, harvFleshT1Cost}}, HarvResource: ResFlesh, HarvRate: harvTier1Rate},

		// Bones gateway — costs flesh, unlocks bones zone
		{Name: "Bones Power (mut)", Costs: []ResourceCost{{ResFlesh, mutBonesPowerFlesh}}, UnlocksBones: true},

		// Flesh defense — needs bones (available after bones zone unlocked)
		{Name: "Flesh Defense (mut)", Costs: []ResourceCost{{ResFlesh, mutFleshDefFlesh}, {ResBones, mutFleshDefBones}}},

		// Bones mutations (bones zone now available)
		{Name: "Bones Speed (mut)", Costs: []ResourceCost{{ResBones, mutBonesSpeedBase}}},

		// Bones tier 1 harvester
		{Name: "Grinding Shard (harv-Bo1)", Costs: []ResourceCost{{ResBones, harvBoneT1Cost}}, HarvResource: ResBones, HarvRate: harvTier1Rate},

		// Bones defense — needs blood + bones
		{Name: "Bones Defense (mut)", Costs: []ResourceCost{{ResBones, mutBonesDefBones}, {ResBlood, mutBonesDefBlood}}},

		// Tier 2 harvesters (multi-resource costs)
		{Name: "Throbbing Artery (harv-B2)", Costs: []ResourceCost{{ResBlood, harvBloodT2CostB}, {ResFlesh, harvBloodT2CostF}}, HarvResource: ResBlood, HarvRate: harvTier2Rate},
		{Name: "Writhing Mass (harv-F2)", Costs: []ResourceCost{{ResFlesh, harvFleshT2CostF}, {ResBones, harvFleshT2CostBo}}, HarvResource: ResFlesh, HarvRate: harvTier2Rate},
		{Name: "Calcified Spine (harv-Bo2)", Costs: []ResourceCost{{ResBones, harvBoneT2CostBo}, {ResBlood, harvBoneT2CostBl}}, HarvResource: ResBones, HarvRate: harvTier2Rate},
	}

	state := SimState{}
	purchased := make([]bool, len(items))
	firstFleshSecond := -1
	failed := false
	brickWallItem := ""
	cappedItem := ""

	type Result struct {
		Name   string
		Second int
	}
	results := []Result{}

	clickAccum := 0.0

outer:
	for tick := 0; tick < maxSimSeconds; tick++ {
		// Apply one second of manual clicks.
		// Blood is always clickable. Flesh/bones clickable once unlocked.
		clickAccum += clicksPerSec
		if clickAccum >= 1.0 {
			whole := math.Floor(clickAccum)
			clickAccum -= whole
			state.Amounts[ResBlood] += whole
			if state.FleshUnlocked {
				state.Amounts[ResFlesh] += whole
			}
			if state.BonesUnlocked {
				state.Amounts[ResBones] += whole
			}
		}

		// Apply 1 Hz harvester income
		for r := 0; r < 3; r++ {
			state.Amounts[r] += state.HarvRates[r]
		}

		// Greedy purchase: scan all items, buy first affordable accessible one
		for i, item := range items {
			if purchased[i] {
				continue
			}
			costs := scaledCosts(item.Costs, 0)

			// Skip if zone for this cost isn't accessible yet
			if !state.resourceAccessible(costs) {
				continue
			}
			// Skip if can't afford
			if !state.canAfford(costs) {
				continue
			}

			// Purchase
			state.deduct(costs)
			purchased[i] = true

			if item.UnlocksFlesh && !state.FleshUnlocked {
				state.FleshUnlocked = true
				firstFleshSecond = tick
			}
			if item.UnlocksBones && !state.BonesUnlocked {
				state.BonesUnlocked = true
			}
			if item.HarvRate > 0 {
				state.HarvRates[item.HarvResource] += item.HarvRate
			}

			results = append(results, Result{Name: item.Name, Second: tick})
		}

		// Check if all items purchased
		allDone := true
		for _, p := range purchased {
			if !p {
				allDone = false
				break
			}
		}
		if allDone {
			break outer
		}
	}

	// Check for uncompleted items (cap reached)
	for i, p := range purchased {
		if !p {
			cappedItem = items[i].Name
			failed = true
			break
		}
	}

	// Print results table
	fmt.Printf("%-42s  %-14s  %s\n", "Item", "Time-to-first", "Ratio-vs-prev")
	fmt.Printf("%-42s  %-14s  %s\n", "----", "-------------", "-------------")

	for idx, r := range results {
		ratioStr := "-"
		if idx > 0 {
			prev := results[idx-1].Second
			curr := r.Second
			if prev == 0 && curr == 0 {
				ratioStr = "1.0x"
			} else if prev == 0 {
				ratioStr = "n/a"
			} else {
				ratioVal := float64(curr) / float64(prev)
				ratioStr = fmt.Sprintf("%.1fx", ratioVal)
				if ratioVal > 5.0 {
					brickWallItem = r.Name
					failed = true
				}
			}
		}
		fmt.Printf("%-42s  %-14s  %s\n", r.Name, formatTime(r.Second), ratioStr)
	}

	fmt.Println()

	if firstFleshSecond >= 0 {
		fmt.Printf("First flesh unlock at: %s\n", formatTime(firstFleshSecond))
		if firstFleshSecond > 3*60 {
			fmt.Println("FLESH UNLOCK TOO SLOW")
			failed = true
		}
	} else {
		fmt.Println("WARNING: Flesh zone never unlocked")
		failed = true
	}

	fmt.Println()

	if brickWallItem != "" {
		fmt.Printf("BRICK WALL DETECTED at %s\n", brickWallItem)
	}
	if cappedItem != "" {
		fmt.Printf("SIMULATION CAP REACHED: %s and subsequent items not purchased within %d seconds\n", cappedItem, maxSimSeconds)
	}

	if failed {
		os.Exit(1)
	}

	fmt.Println("Cost curve OK: all ratios within range, flesh unlock <= 3 min.")
}
