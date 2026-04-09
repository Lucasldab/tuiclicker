package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ZoneID mirrors model.ZoneType without importing internal/model.
type ZoneID int

const (
	ZoneNone  ZoneID = iota
	ZoneBlood        // ANSI 1 — red
	ZoneFlesh        // ANSI 9 — bright red
	ZoneBones        // ANSI 3 — yellow
)

// TabID mirrors model.TabID without importing internal/model.
type TabID int

const (
	TabZones     TabID = iota
	TabMutations
	TabHarvesters
)

// ResourceView carries display values for one resource.
type ResourceView struct {
	Label  string // "BLOOD", "FLESH", "BONES"
	Amount string // formatted: "42", "1,234", "1.23M"
	Rate   string // formatted: "+0.0/s"
	ZoneID ZoneID // which zone this resource maps to (for color selection)
}

// MutationView carries display data for one mutation in the mutations panel.
// All computation (affordability, cost string) done in toGameView() — View() is pure.
type MutationView struct {
	Name        string // body horror mutation name
	Description string // 1-line flavor text
	CostString  string // formatted: "100 blood" or "50 blood / 25 flesh"
	OwnedCount  int    // times purchased
	CanAfford   bool   // true if player has enough of all required resources
	BranchColor ZoneID // ZoneBlood / ZoneFlesh / ZoneBones for color selection
}

// HarvesterView carries display data for one harvester in the harvesters panel.
type HarvesterView struct {
	Name        string
	RateString  string // formatted: "+0.5/s each"
	CostString  string // formatted cost string
	OwnedCount  int
	CanAfford   bool
	BranchColor ZoneID
}

// GameView is the data-transfer type panels receive. No model dependency.
type GameView struct {
	Width        int
	Height       int
	ActiveTab    TabID
	FlashZone    ZoneID
	ZoneUnlocked [3]bool // index 0=blood, 1=flesh, 2=bones
	Resources    [3]ResourceView

	// Phase 2 additions:
	Mutations       []MutationView
	Harvesters      []HarvesterView
	CreatureTier    int // 0–3
	DominantBranch  int // 0=blood, 1=flesh, 2=bones
	MutationScroll  int // top visible item index
	HarvesterScroll int // top visible item index
	MutationCursor  int // highlighted item index (-1 = none)
	HarvesterCursor int // highlighted item index (-1 = none)
	MutationFlash   int // index of mutation to flash (-1 = none)
	HarvesterFlash  int // index of harvester to flash (-1 = none)
}

// --- Tab Bar ---

// renderTabBar renders the top navigation row per UI-SPEC Tab Navigation Contract.
// Active tab: bold + bright white. Inactive/locked tabs: dim (8).
// Tab widths: ZONES=7, MUTATIONS=11, HARVESTERS=12. Gap between tabs: 2 chars.
func renderTabBar(v GameView, width int) string {
	tabs := []struct {
		label string
		id    TabID
	}{
		{"ZONES", TabZones},
		{"MUTATIONS", TabMutations},
		{"HARVESTERS", TabHarvesters},
	}

	var parts []string
	for _, tab := range tabs {
		label := fmt.Sprintf("[%s]", tab.label)
		if v.ActiveTab == tab.id {
			parts = append(parts, StyleTabActive.Render(label))
		} else {
			parts = append(parts, StyleTabInactive.Render(label))
		}
	}

	bar := strings.Join(parts, "  ") // 2-char gap between tabs
	return lipgloss.NewStyle().Width(width).Render(bar)
}

// --- Resource Panel (left, ~30%) ---

func renderResourcePanel(v GameView, w, h int) string {
	heading := StyleHeading.Render("RESOURCES")

	var blocks []string
	for _, r := range v.Resources {
		colorStyle := resourceStyle(r.ZoneID)
		name := colorStyle.Render(r.Label)
		amount := colorStyle.Render(r.Amount)
		rate := StyleDim.Render(r.Rate)
		blocks = append(blocks, lipgloss.JoinVertical(lipgloss.Left, name, amount, rate))
	}

	content := lipgloss.JoinVertical(lipgloss.Left,
		heading,
		"", // md spacing = 1 row
		blocks[0],
		"", // 2 blank rows between resource blocks
		"",
		blocks[1],
		"",
		"",
		blocks[2],
	)

	return lipgloss.NewStyle().
		Width(w).
		Height(h).
		PaddingLeft(2).
		PaddingRight(2).
		Render(content)
}

// resourceStyle returns the Lipgloss style for a resource's color.
func resourceStyle(z ZoneID) lipgloss.Style {
	switch z {
	case ZoneBlood:
		return StyleBlood
	case ZoneFlesh:
		return StyleFlesh
	case ZoneBones:
		return StyleBones
	default:
		return StylePrimary
	}
}

// --- Creature Panel (center, ~40%) ---

func renderCreaturePanel(v GameView, w, h int) string {
	heading := StyleHeading.Render("THE ABERRATION")

	artStr := GetCreatureArt(v.CreatureTier, v.DominantBranch)

	var artStyle lipgloss.Style
	if v.CreatureTier == 0 {
		artStyle = StyleDim
	} else {
		switch v.DominantBranch {
		case 0:
			artStyle = StyleBlood
		case 1:
			artStyle = StyleFlesh
		case 2:
			artStyle = StyleBones
		default:
			artStyle = StyleDim
		}
	}

	art := artStyle.Render(artStr)

	tierLabel := tierName(v.CreatureTier)
	label := StyleHeading.Render(tierLabel)

	content := lipgloss.JoinVertical(lipgloss.Center,
		heading,
		"",
		art,
		"",
		label,
	)

	return lipgloss.NewStyle().
		Width(w).
		Height(h).
		AlignHorizontal(lipgloss.Center).
		Render(content)
}

func tierName(tier int) string {
	switch tier {
	case 1:
		return "NASCENT FORM"
	case 2:
		return "GROTESQUE"
	case 3:
		return "ABOMINATION"
	default:
		return "ABERRANT SEED"
	}
}

// --- Zone Panel (right, ~30%) ---

// renderZonePanel renders the three click zones with correct locked/unlocked/flash states.
// Zone box height = ZoneBoxHeight (5 rows). Gaps = 2 rows between boxes.
// Zone layout constants from model/zones.go: offsets 0, 7, 14 correspond to row spacing here.
func renderZonePanel(v GameView, w, h int) string {
	heading := StyleHeading.Render("HARVEST ZONES")

	bloodBox := renderZoneBox(v, ZoneBlood, "BLOOD", w-4)
	fleshBox := renderZoneBox(v, ZoneFlesh, "FLESH", w-4)
	bonesBox := renderZoneBox(v, ZoneBones, "BONES", w-4)

	content := lipgloss.JoinVertical(lipgloss.Left,
		heading,
		"",
		bloodBox,
		"",
		"",
		fleshBox,
		"",
		"",
		bonesBox,
	)

	return lipgloss.NewStyle().
		Width(w).
		Height(h).
		PaddingLeft(2).
		PaddingRight(2).
		Render(content)
}

// renderZoneBox renders a single 5-row click zone box.
// States: unlocked (resource color), locked (dim), flash (bright white).
func renderZoneBox(v GameView, z ZoneID, label string, innerW int) string {
	unlocked := v.ZoneUnlocked[zoneIndex(z)]
	flashing := v.FlashZone == z

	var borderColor lipgloss.Color
	var textStyle lipgloss.Style

	switch {
	case flashing:
		borderColor = lipgloss.Color("15") // bright white
		textStyle = StyleFlash
	case !unlocked:
		borderColor = lipgloss.Color("8") // dim
		textStyle = StyleDim
	default:
		borderColor = zoneColor(z)
		textStyle = resourceStyle(z)
	}

	var name, prompt string
	if unlocked {
		name = textStyle.Render(label)
		prompt = StyleDim.Render("Click to harvest")
	} else {
		name = textStyle.Render(fmt.Sprintf("%s  (locked)", label))
		prompt = StyleDim.Render("Unlock via mutation")
	}

	inner := lipgloss.JoinVertical(lipgloss.Left,
		name,
		prompt,
		textStyle.Render("[ HARVEST ]"),
	)

	return lipgloss.NewStyle().
		Border(lipgloss.ASCIIBorder()).
		BorderForeground(borderColor).
		Width(innerW).
		Render(inner)
}

func zoneIndex(z ZoneID) int {
	switch z {
	case ZoneBlood:
		return 0
	case ZoneFlesh:
		return 1
	case ZoneBones:
		return 2
	}
	return 0
}

func zoneColor(z ZoneID) lipgloss.Color {
	switch z {
	case ZoneBlood:
		return lipgloss.Color("1")
	case ZoneFlesh:
		return lipgloss.Color("9")
	case ZoneBones:
		return lipgloss.Color("3")
	}
	return lipgloss.Color("7")
}

// --- Mutation Panel (right, when TabMutations active) ---

// renderMutationPanel renders 9 mutations in 3 branch sections with names,
// descriptions, costs, owned counts, cursor indicator, flash, and scroll indicators.
func renderMutationPanel(v GameView, w, h int) string {
	heading := StyleHeading.Render("MUTATIONS")
	separator := strings.Repeat("-", w-4)

	headerRows := 2
	scrollIndicatorRows := 2
	contentH := h - headerRows - scrollIndicatorRows
	if contentH < 1 {
		contentH = 1
	}

	// Build flat item list: branch headers + mutation items
	type itemKind int
	const (
		kindHeader itemKind = iota
		kindMutation
	)
	type renderItem struct {
		kind        itemKind
		headerText  string
		headerStyle lipgloss.Style
		mutIdx      int // flat index into v.Mutations (for cursor/flash)
	}

	var items []renderItem
	branches := []struct {
		label string
		style lipgloss.Style
		start int
		end   int
	}{
		{"BLOOD BRANCH", StyleBlood, 0, 3},
		{"FLESH BRANCH", StyleFlesh, 3, 6},
		{"BONES BRANCH", StyleBones, 6, 9},
	}

	for bi, br := range branches {
		if bi > 0 {
			// lg gap (2 blank rows) between branches
			items = append(items, renderItem{kind: kindHeader, headerText: "", headerStyle: StyleDim})
			items = append(items, renderItem{kind: kindHeader, headerText: "", headerStyle: StyleDim})
		}
		items = append(items, renderItem{kind: kindHeader, headerText: br.label, headerStyle: br.style})
		end := br.end
		if end > len(v.Mutations) {
			end = len(v.Mutations)
		}
		for mi := br.start; mi < end; mi++ {
			items = append(items, renderItem{kind: kindMutation, mutIdx: mi})
		}
	}

	// Each mutation item is 4 rows; each header is 1 row.
	// Compute total rows needed.
	totalRows := 0
	for _, it := range items {
		if it.kind == kindMutation {
			totalRows += 4
		} else {
			totalRows += 1
		}
	}

	// Convert flat item list to a row-indexed list for scrolling.
	type rowItem struct {
		itemIdx int // index in items[]
		rowInItem int // which row within item (0-3 for mutations, 0 for headers)
	}
	var rows []rowItem
	for i, it := range items {
		if it.kind == kindMutation {
			for r := 0; r < 4; r++ {
				rows = append(rows, rowItem{i, r})
			}
		} else {
			rows = append(rows, rowItem{i, 0})
		}
	}

	// Clamp scroll
	scroll := v.MutationScroll
	if scroll < 0 {
		scroll = 0
	}
	maxScroll := len(rows) - contentH
	if maxScroll < 0 {
		maxScroll = 0
	}
	if scroll > maxScroll {
		scroll = maxScroll
	}

	end := scroll + contentH
	if end > len(rows) {
		end = len(rows)
	}
	visibleRows := rows[scroll:end]

	showAbove := scroll > 0
	showBelow := end < len(rows)

	var lines []string
	lines = append(lines, heading)
	lines = append(lines, separator)

	if showAbove {
		lines = append(lines, StyleDim.Render("^ more above"))
	} else {
		lines = append(lines, "")
	}

	for _, ri := range visibleRows {
		it := items[ri.itemIdx]
		if it.kind == kindHeader {
			if it.headerText == "" {
				lines = append(lines, "")
			} else {
				lines = append(lines, it.headerStyle.Render(it.headerText))
			}
		} else {
			m := v.Mutations[it.mutIdx]
			nameStyle := resourceStyle(m.BranchColor)
			flashing := v.MutationFlash == it.mutIdx
			hasCursor := v.MutationCursor == it.mutIdx

			switch ri.rowInItem {
			case 0: // name row
				namePart := m.Name
				if hasCursor {
					namePart = "> " + namePart
				}
				var nameStr string
				if flashing {
					nameStr = StyleFlash.Render(namePart)
				} else {
					nameStr = nameStyle.Render(namePart)
				}
				if m.OwnedCount > 0 {
					nameStr += "  " + StylePurchased.Render(fmt.Sprintf("x%d", m.OwnedCount))
				}
				lines = append(lines, nameStr)
			case 1: // description row
				lines = append(lines, StyleDim.Render(m.Description))
			case 2: // cost row
				costLine := fmt.Sprintf("Cost: %s", m.CostString)
				if m.CanAfford {
					lines = append(lines, StylePrimary.Render(costLine))
				} else {
					lines = append(lines, StyleDim.Render(costLine))
				}
			case 3: // blank row
				lines = append(lines, "")
			}
		}
	}

	if showBelow {
		lines = append(lines, StyleDim.Render("v more below"))
	} else {
		lines = append(lines, "")
	}

	content := strings.Join(lines, "\n")
	return lipgloss.NewStyle().
		Width(w).
		Height(h).
		PaddingLeft(2).
		PaddingRight(2).
		Render(content)
}

// --- Harvester Panel (right, when TabHarvesters active) ---

// renderHarvesterPanel renders 6 harvesters in 3 resource sections with names,
// rates, costs, owned counts, cursor indicator, flash, and scroll indicators.
func renderHarvesterPanel(v GameView, w, h int) string {
	heading := StyleHeading.Render("HARVESTERS")
	separator := strings.Repeat("-", w-4)

	headerRows := 2
	scrollIndicatorRows := 2
	contentH := h - headerRows - scrollIndicatorRows
	if contentH < 1 {
		contentH = 1
	}

	type itemKind int
	const (
		kindHeader itemKind = iota
		kindHarvester
	)
	type renderItem struct {
		kind        itemKind
		headerText  string
		headerStyle lipgloss.Style
		harvIdx     int
	}

	var items []renderItem
	sections := []struct {
		label string
		style lipgloss.Style
		start int
		end   int
	}{
		{"BLOOD HARVESTERS", StyleBlood, 0, 2},
		{"FLESH HARVESTERS", StyleFlesh, 2, 4},
		{"BONES HARVESTERS", StyleBones, 4, 6},
	}

	for si, sec := range sections {
		if si > 0 {
			// lg gap (2 blank rows) between sections
			items = append(items, renderItem{kind: kindHeader, headerText: "", headerStyle: StyleDim})
			items = append(items, renderItem{kind: kindHeader, headerText: "", headerStyle: StyleDim})
		}
		items = append(items, renderItem{kind: kindHeader, headerText: sec.label, headerStyle: sec.style})
		end := sec.end
		if end > len(v.Harvesters) {
			end = len(v.Harvesters)
		}
		for hi := sec.start; hi < end; hi++ {
			items = append(items, renderItem{kind: kindHarvester, harvIdx: hi})
		}
	}

	// Build row-indexed list for scrolling
	type rowItem struct {
		itemIdx   int
		rowInItem int
	}
	var rows []rowItem
	for i, it := range items {
		if it.kind == kindHarvester {
			for r := 0; r < 4; r++ {
				rows = append(rows, rowItem{i, r})
			}
		} else {
			rows = append(rows, rowItem{i, 0})
		}
	}

	// Clamp scroll
	scroll := v.HarvesterScroll
	if scroll < 0 {
		scroll = 0
	}
	maxScroll := len(rows) - contentH
	if maxScroll < 0 {
		maxScroll = 0
	}
	if scroll > maxScroll {
		scroll = maxScroll
	}

	end := scroll + contentH
	if end > len(rows) {
		end = len(rows)
	}
	visibleRows := rows[scroll:end]

	showAbove := scroll > 0
	showBelow := end < len(rows)

	var lines []string
	lines = append(lines, heading)
	lines = append(lines, separator)

	if showAbove {
		lines = append(lines, StyleDim.Render("^ more above"))
	} else {
		lines = append(lines, "")
	}

	for _, ri := range visibleRows {
		it := items[ri.itemIdx]
		if it.kind == kindHeader {
			if it.headerText == "" {
				lines = append(lines, "")
			} else {
				lines = append(lines, it.headerStyle.Render(it.headerText))
			}
		} else {
			harv := v.Harvesters[it.harvIdx]
			nameStyle := resourceStyle(harv.BranchColor)
			flashing := v.HarvesterFlash == it.harvIdx
			hasCursor := v.HarvesterCursor == it.harvIdx

			switch ri.rowInItem {
			case 0: // name row
				namePart := harv.Name
				if hasCursor {
					namePart = "> " + namePart
				}
				var nameStr string
				if flashing {
					nameStr = StyleFlash.Render(namePart)
				} else {
					nameStr = nameStyle.Render(namePart)
				}
				nameStr += "  " + StylePrimary.Render(fmt.Sprintf("owned: %d", harv.OwnedCount))
				lines = append(lines, nameStr)
			case 1: // rate row
				lines = append(lines, StyleDim.Render(harv.RateString))
			case 2: // cost row
				costLine := fmt.Sprintf("Cost: %s", harv.CostString)
				if harv.CanAfford {
					lines = append(lines, StylePrimary.Render(costLine))
				} else {
					lines = append(lines, StyleDim.Render(costLine))
				}
			case 3: // blank row
				lines = append(lines, "")
			}
		}
	}

	if showBelow {
		lines = append(lines, StyleDim.Render("v more below"))
	} else {
		lines = append(lines, "")
	}

	content := strings.Join(lines, "\n")
	return lipgloss.NewStyle().
		Width(w).
		Height(h).
		PaddingLeft(2).
		PaddingRight(2).
		Render(content)
}

// --- Status Bar ---

func renderStatusBar(width int) string {
	hint := "[1] zones  [2] mutations  [3] harvesters  [q] quit"
	return lipgloss.NewStyle().
		Width(width).
		Foreground(lipgloss.Color("8")).
		Render(hint)
}
