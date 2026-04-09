package model

// ZoneType identifies a harvestable click zone.
type ZoneType int

const (
	ZoneNone  ZoneType = iota // sentinel — no zone / no flash
	ZoneBlood                 // always unlocked at start
	ZoneFlesh                 // locked until Phase 2 mutation
	ZoneBones                 // locked until Phase 2 mutation
)

// ZoneOffsets are the vertical offsets (in rows) from the content area start
// where each zone box begins. Both recalculateLayout and renderZonePanel must
// read these constants — single source of truth to prevent hit-test drift (Pitfall 2).
const (
	BloodZoneOffset = 0
	FleshZoneOffset = 7
	BonesZoneOffset = 14
	ZoneBoxHeight   = 5 // rows occupied by one zone box including border
)

// TabID identifies the active top-level navigation tab.
type TabID int

const (
	TabZones     TabID = iota // only functional tab in Phase 1
	TabMutations              // Phase 2 placeholder — no-op in Phase 1
	TabHarvesters             // Phase 2 placeholder — no-op in Phase 1
)
