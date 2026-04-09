package model

import (
	tea "github.com/charmbracelet/bubbletea"
)

// GameModel is the single source of truth for all game state.
// All state mutations happen exclusively in Update() — never mutate fields
// outside of Update (Bubbletea MVU pattern).
type GameModel struct {
	// Terminal dimensions — updated on tea.WindowSizeMsg
	width  int
	height int

	// tooSmall is true when the terminal is below the hard minimum (60x20).
	// View() returns the too-small error string immediately when true.
	tooSmall bool

	// Zone coordinate cache — recomputed on every WindowSizeMsg.
	// These are the authoritative hit-test boundaries; View() must match them.
	zonePanelLeft  int // first column of the right (zones) panel
	zonePanelRight int // one past last column of the right panel

	// Zone top rows (absolute row index, 0 = top of terminal).
	// Computed as: contentStartRow + ZoneXxxOffset.
	zoneBloodTop int
	zoneFleshTop int
	zoneBonesTop int

	// Game state
	Ledger    ResourceLedger
	ActiveTab TabID

	// flashZone is set on harvest click and cleared by the clearFlashMsg handler.
	// ZoneNone means no flash is active.
	flashZone ZoneType

	// ZoneUnlocked tracks which zones accept clicks. Index 0 (blood) is always
	// true. Flesh (1) and Bones (2) are unlocked by Phase 2 mutations.
	ZoneUnlocked [3]bool
}

// New returns a GameModel with safe defaults.
// width/height default to 80x24 so View() doesn't panic before the first
// tea.WindowSizeMsg arrives (Pitfall 3).
func New() GameModel {
	m := GameModel{
		width:  80,
		height: 24,
	}
	m.ZoneUnlocked[ResourceBlood] = true // blood is always unlocked (D-02)
	m = recalculateLayout(m)
	return m
}

// recalculateLayout recomputes all column widths and zone top rows from the
// current m.width and m.height. Call this every time width or height changes.
// Both the hit-test code and View() must use the values set here (Pitfall 2).
func recalculateLayout(m GameModel) GameModel {
	rightW := (m.width * 30) / 100
	m.zonePanelLeft = m.width - rightW
	m.zonePanelRight = m.width

	// Content area starts at row 3: row 0 is padding, row 1 is tab bar,
	// row 2 is separator. Zone offsets are defined in zones.go as constants.
	const contentStartRow = 3
	m.zoneBloodTop = contentStartRow + BloodZoneOffset
	m.zoneFleshTop = contentStartRow + FleshZoneOffset
	m.zoneBonesTop = contentStartRow + BonesZoneOffset

	return m
}

// Init returns the initial command set: start the 1 Hz tick loop.
func (m GameModel) Init() tea.Cmd {
	return doTick()
}

// Update handles all incoming messages and returns the updated model + next Cmd.
func (m GameModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		if msg.Width < 60 || msg.Height < 20 {
			m.tooSmall = true
			m.width, m.height = msg.Width, msg.Height
			return m, nil
		}
		m.tooSmall = false
		m.width, m.height = msg.Width, msg.Height
		m = recalculateLayout(m)
		return m, nil

	case tea.KeyMsg:
		return handleKey(m, msg)

	case tea.MouseMsg:
		// Filter motion events immediately — they fire dozens/sec and would
		// burn CPU. Only process press events (Pitfall 1, VISL-04).
		if msg.Action == tea.MouseActionMotion {
			return m, nil
		}
		if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft {
			return handleClick(m, msg.X, msg.Y)
		}
		return m, nil

	case TickMsg:
		// Phase 1: rates are all 0.0/s (no auto-harvesters).
		// Phase 2: apply m.Ledger.Rates[*] increments here.
		return m, doTick()

	case clearFlashMsg:
		m.flashZone = ZoneNone
		return m, nil
	}

	return m, nil
}

// View renders the current game state as a string. It must be a pure function
// of m — do not mutate m inside View (value receiver enforces this).
func (m GameModel) View() string {
	if m.tooSmall {
		return "terminal too small (min 60x20)"
	}
	return "(building...)" // replaced in Plan 03
}

// handleKey processes keyboard input and returns the updated model + Cmd.
// Keybind defaults per UI-SPEC: b=blood, f=flesh, n=bones, 1/2/3=tabs, q=quit.
func handleKey(m GameModel, msg tea.KeyMsg) (GameModel, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "1":
		m.ActiveTab = TabZones
	case "2":
		m.ActiveTab = TabMutations // no-op in Phase 1 (no content)
	case "3":
		m.ActiveTab = TabHarvesters // no-op in Phase 1 (no content)
	case "b":
		return harvestKey(m, ResourceBlood, ZoneBlood)
	case "f":
		return harvestKey(m, ResourceFlesh, ZoneFlesh)
	case "n":
		return harvestKey(m, ResourceBones, ZoneBones)
	}
	return m, nil
}

// harvestKey attempts to harvest resource r via keybind. No-op if zone locked.
func harvestKey(m GameModel, r ResourceType, z ZoneType) (GameModel, tea.Cmd) {
	if !m.ZoneUnlocked[r] {
		return m, nil
	}
	m.Ledger.Add(r, 1.0)
	m.flashZone = z
	return m, func() tea.Msg { return clearFlashMsg{} }
}

// handleClick routes a left-button click to the appropriate zone or tab.
func handleClick(m GameModel, x, y int) (GameModel, tea.Cmd) {
	// Tab bar is at row 1.
	if y == 1 {
		m = handleTabClick(m, x)
		return m, nil
	}
	// Zones panel hit-test.
	if x >= m.zonePanelLeft && x < m.zonePanelRight {
		switch {
		case y >= m.zoneBloodTop && y < m.zoneBloodTop+ZoneBoxHeight:
			if m.ZoneUnlocked[ResourceBlood] {
				m.Ledger.Add(ResourceBlood, 1.0)
				m.flashZone = ZoneBlood
				return m, func() tea.Msg { return clearFlashMsg{} }
			}
		case y >= m.zoneFleshTop && y < m.zoneFleshTop+ZoneBoxHeight:
			if m.ZoneUnlocked[ResourceFlesh] {
				m.Ledger.Add(ResourceFlesh, 1.0)
				m.flashZone = ZoneFlesh
				return m, func() tea.Msg { return clearFlashMsg{} }
			}
		case y >= m.zoneBonesTop && y < m.zoneBonesTop+ZoneBoxHeight:
			if m.ZoneUnlocked[ResourceBones] {
				m.Ledger.Add(ResourceBones, 1.0)
				m.flashZone = ZoneBones
				return m, func() tea.Msg { return clearFlashMsg{} }
			}
		}
	}
	return m, nil
}

// handleTabClick switches the active tab based on the click column.
// Tab label widths: ZONES=7, MUTATIONS=11, HARVESTERS=12, gaps=2 chars between.
func handleTabClick(m GameModel, x int) GameModel {
	// Tab layout (0-indexed columns): [ZONES]=cols 0-6, gap, [MUTATIONS]=cols 9-19, gap, [HARVESTERS]=cols 22-33
	switch {
	case x >= 0 && x <= 6:
		m.ActiveTab = TabZones
	case x >= 9 && x <= 19:
		m.ActiveTab = TabMutations
	case x >= 22 && x <= 33:
		m.ActiveTab = TabHarvesters
	}
	return m
}
