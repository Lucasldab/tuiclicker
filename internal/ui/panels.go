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

// GameView is the data-transfer type panels receive. No model dependency.
type GameView struct {
	Width        int
	Height       int
	ActiveTab    TabID
	FlashZone    ZoneID
	ZoneUnlocked [3]bool // index 0=blood, 1=flesh, 2=bones
	Resources    [3]ResourceView
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

func renderCreaturePanel(_ GameView, w, h int) string {
	heading := StyleHeading.Render("THE ABERRATION")

	// Phase 1 static placeholder — Phase 2 replaces with mutation-driven art.
	// Rendered in dim (ANSI 8) to signal placeholder state per UI-SPEC.
	placeholder := []string{
		"    .  .  .",
		"   ( o  o )",
		"    >  --  <",
		"   /|      |\\",
		"  * |      | *",
	}
	art := StyleDim.Render(strings.Join(placeholder, "\n"))

	content := lipgloss.JoinVertical(lipgloss.Center,
		heading,
		"",
		art,
	)

	return lipgloss.NewStyle().
		Width(w).
		Height(h).
		AlignHorizontal(lipgloss.Center).
		Render(content)
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

// --- Status Bar ---

func renderStatusBar(width int) string {
	hint := "[1] zones  [2] mutations  [3] harvesters  [q] quit"
	return lipgloss.NewStyle().
		Width(width).
		Foreground(lipgloss.Color("8")).
		Render(hint)
}
