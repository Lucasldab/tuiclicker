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
//
// Layout within zone panel:
//   row 0: heading ("HARVEST ZONES")
//   row 1: blank (md spacing)
//   row 2-6: blood zone box (border top + 3 content lines + border bottom = 5)
//   row 7-8: gap (2 blank rows)
//   row 9-13: flesh zone box
//   row 14-15: gap
//   row 16-20: bones zone box
const (
	ZonePanelHeadingRows = 2  // heading + 1 blank row before first box
	BloodZoneOffset      = 2  // heading(1) + blank(1)
	FleshZoneOffset      = 9  // blood(5) + gap(2) + heading offset(2)
	BonesZoneOffset      = 16 // flesh(5) + gap(2) + blood offset(9)
	ZoneBoxHeight        = 7  // rows occupied by one zone box including border + gap below
)

// TabID identifies the active top-level navigation tab.
type TabID int

const (
	TabZones     TabID = iota // only functional tab in Phase 1
	TabMutations              // Phase 2 placeholder — no-op in Phase 1
	TabHarvesters             // Phase 2 placeholder — no-op in Phase 1
)
