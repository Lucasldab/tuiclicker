package balance

// HarvestYield is the base amount of a resource gained per manual click.
// All harvest amounts flow through this constant — never hardcode 1 elsewhere.
const HarvestYield float64 = 1.0

// ---------------------------------------------------------------------------
// Phase 2 constants — validated by cmd/simulate/main.go.
// Do NOT change these values without re-running the simulation and verifying
// the cost curve exits 0 (no brick walls, flesh unlock <= 3 min).
// ---------------------------------------------------------------------------

// Cost scaling factors (D-13: exponential scaling at 1.25x per purchase).
const MutationCostScale  = 1.25
const HarvesterCostScale = 1.25

// --- Mutation base costs ---

// Blood branch
// Power and Speed cost blood only; Defense requires blood + flesh (D-12/D-14).
const MutBloodPowerBase = 10.0 // blood only
const MutBloodSpeedBase = 25.0 // blood only
const MutBloodDefBlood  = 15.0 // blood component of defense cost
const MutBloodDefFlesh  = 10.0 // flesh component of defense cost

// Flesh branch
// Power is a GATEWAY mutation: costs blood, unlocks the flesh harvest zone (D-05).
// Speed and Defense cost flesh (available after flesh zone is unlocked).
const MutFleshPowerBlood = 15.0 // blood cost — gateway that unlocks flesh zone
const MutFleshSpeedBase  = 20.0 // flesh only
const MutFleshDefFlesh   = 15.0 // flesh component of defense cost
const MutFleshDefBones   = 10.0 // bones component of defense cost

// Bones branch
// Power is a GATEWAY mutation: costs flesh, unlocks the bones harvest zone (D-05).
// Speed and Defense cost bones (available after bones zone is unlocked).
const MutBonesPowerFlesh = 15.0 // flesh cost — gateway that unlocks bones zone
const MutBonesSpeedBase  = 20.0 // bones only
const MutBonesDefBones   = 15.0 // bones component of defense cost
const MutBonesDefBlood   = 10.0 // blood component of defense cost

// Mutation yield and harvester buff bonuses per purchase.
const MutYieldBonusPerPurchase    = 0.10 // +10% manual harvest yield per purchase in branch
const MutHarvesterBuffPerPurchase = 0.05 // +5% harvester output per purchase in branch

// --- Harvester base costs ---
// Tier 1: single-resource cost (branch resource only).
// Tier 2: multi-resource cost (D-12/D-14).
const HarvBloodT1Cost   = 20.0 // blood
const HarvBloodT2CostB  = 50.0 // blood component
const HarvBloodT2CostF  = 30.0 // flesh component
const HarvFleshT1Cost   = 20.0 // flesh
const HarvFleshT2CostF  = 50.0 // flesh component
const HarvFleshT2CostBo = 30.0 // bones component
const HarvBoneT1Cost    = 20.0 // bones
const HarvBoneT2CostBo  = 50.0 // bones component
const HarvBoneT2CostBl  = 30.0 // blood component

// Harvester output rates (units per owned, per second).
const HarvTier1Rate = 0.5
const HarvTier2Rate = 2.0

// Creature visual tier thresholds (total mutations purchased).
// Tier 0 = 0 mutations (pre-mutant), Tier 1 = 1-3, Tier 2 = 4-7, Tier 3 = 8+.
const CreatureTier1Threshold = 1 // 1-3 mutations: NASCENT FORM
const CreatureTier2Threshold = 4 // 4-7 mutations: GROTESQUE
const CreatureTier3Threshold = 8 // 8+  mutations: ABOMINATION
