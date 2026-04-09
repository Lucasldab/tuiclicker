package model

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/lucasldab/tuiclicker/internal/balance"
	"github.com/lucasldab/tuiclicker/internal/ui"
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

	// Phase 2 — mutation data
	MutationDefs   []MutationDef
	MutationStates []MutationState

	// Phase 2 — harvester data
	HarvesterDefs   []HarvesterDef
	HarvesterStates []HarvesterState

	// Phase 2 — UI cursor and scroll state
	MutationCursor  int
	MutationScroll  int
	HarvesterCursor int
	HarvesterScroll int

	// Phase 2 — purchase flash (one-render effect, -1 = no flash)
	flashMutationIdx  int
	flashHarvesterIdx int
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

	// Phase 2 initialization
	m.MutationDefs = AllMutations
	m.MutationStates = make([]MutationState, len(AllMutations))
	m.HarvesterDefs = AllHarvesters
	m.HarvesterStates = make([]HarvesterState, len(AllHarvesters))
	m.MutationCursor = -1  // no cursor until user presses up/down
	m.HarvesterCursor = -1
	m.flashMutationIdx = -1
	m.flashHarvesterIdx = -1

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

	// Content area starts at row 2: row 0 is tab bar, row 1 is separator.
	// Zone offsets (defined in zones.go) include heading rows within the panel.
	const contentStartRow = 2
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
		if msg.Action == tea.MouseActionPress {
			switch msg.Button {
			case tea.MouseButtonWheelUp:
				m = handleScrollUp(m)
				return m, nil
			case tea.MouseButtonWheelDown:
				m = handleScrollDown(m)
				return m, nil
			case tea.MouseButtonLeft:
				return handleClick(m, msg.X, msg.Y)
			}
		}
		return m, nil

	case TickMsg:
		// Apply one second of harvester income (AUTO-03)
		for i, def := range m.HarvesterDefs {
			buff := BranchHarvesterBuff(m.MutationDefs, m.MutationStates, def.Branch)
			rate := def.EffectiveRate(m.HarvesterStates[i], buff)
			m.Ledger.Amounts[def.Resource] += rate
		}
		// Rebuild rates from scratch (Pitfall 1)
		m.Ledger = RecalcAllRates(m)
		return m, tea.Batch(doTick(), saveCmd(m))

	case clearFlashMsg:
		m.flashZone = ZoneNone
		return m, nil

	case clearMutationFlashMsg:
		m.flashMutationIdx = -1
		return m, nil

	case clearHarvesterFlashMsg:
		m.flashHarvesterIdx = -1
		return m, nil

	case SaveResultMsg:
		// Silent: save errors do not affect game state.
		// Errors are intentionally swallowed here; stderr logging omitted
		// to keep the TUI clean.
		return m, nil
	}

	return m, nil
}

// FlashZone returns the currently active flash zone (ZoneNone if no flash).
// Exported for testing only — do not use in production rendering logic.
func (m GameModel) FlashZone() ZoneType { return m.flashZone }

// TooSmall returns true when the terminal is below the minimum supported size.
// Exported for testing only.
func (m GameModel) TooSmall() bool { return m.tooSmall }

// View renders the current game state as a string. It must be a pure function
// of m — do not mutate m inside View (value receiver enforces this).
func (m GameModel) View() string {
	if m.tooSmall {
		return "terminal too small (min 60x20)"
	}
	return ui.BuildLayout(m.toGameView())
}

// toGameView converts GameModel to the ui.GameView data-transfer type.
// This adapter prevents the ui package from importing internal/model
// (which would create a circular import).
func (m GameModel) toGameView() ui.GameView {
	// Build mutations slice
	mutations := make([]ui.MutationView, len(m.MutationDefs))
	for i, def := range m.MutationDefs {
		costs := def.CurrentCost(m.MutationStates[i])
		mutations[i] = ui.MutationView{
			Name:        def.Name,
			Description: def.Description,
			CostString:  FormatCosts(costs),
			OwnedCount:  m.MutationStates[i].PurchaseCount,
			CanAfford:   CanAfford(costs, m.Ledger),
			BranchColor: ui.ZoneID(def.Branch + 1), // BranchBlood(0)+1=ZoneBlood(1), etc.
		}
	}

	// Build harvesters slice
	harvesters := make([]ui.HarvesterView, len(m.HarvesterDefs))
	for i, def := range m.HarvesterDefs {
		costs := def.CurrentCost(m.HarvesterStates[i])
		harvesters[i] = ui.HarvesterView{
			Name:        def.Name,
			RateString:  fmt.Sprintf("+%.1f/s each", def.BaseRate),
			CostString:  FormatCosts(costs),
			OwnedCount:  m.HarvesterStates[i].Owned,
			CanAfford:   CanAfford(costs, m.Ledger),
			BranchColor: ui.ZoneID(def.Branch + 1), // BranchBlood(0)+1=ZoneBlood(1), etc.
		}
	}

	return ui.GameView{
		Width:           m.width,
		Height:          m.height,
		ActiveTab:       ui.TabID(m.ActiveTab),
		FlashZone:       ui.ZoneID(m.flashZone),
		ZoneUnlocked:    m.ZoneUnlocked,
		Resources: [3]ui.ResourceView{
			{
				Label:  "BLOOD",
				Amount: FormatAmount(m.Ledger.Amounts[ResourceBlood]),
				Rate:   FormatRate(m.Ledger.Rates[ResourceBlood]),
				ZoneID: ui.ZoneBlood,
			},
			{
				Label:  "FLESH",
				Amount: FormatAmount(m.Ledger.Amounts[ResourceFlesh]),
				Rate:   FormatRate(m.Ledger.Rates[ResourceFlesh]),
				ZoneID: ui.ZoneFlesh,
			},
			{
				Label:  "BONES",
				Amount: FormatAmount(m.Ledger.Amounts[ResourceBones]),
				Rate:   FormatRate(m.Ledger.Rates[ResourceBones]),
				ZoneID: ui.ZoneBones,
			},
		},
		Mutations:       mutations,
		Harvesters:      harvesters,
		CreatureTier:    m.creatureTier(),
		DominantBranch:  m.dominantBranch(),
		MutationScroll:  m.MutationScroll,
		HarvesterScroll: m.HarvesterScroll,
		MutationCursor:  m.MutationCursor,
		HarvesterCursor: m.HarvesterCursor,
		MutationFlash:   m.flashMutationIdx,
		HarvesterFlash:  m.flashHarvesterIdx,
	}
}

// creatureTier returns the visual tier of the creature (0-3) based on
// total mutations purchased.
func (m GameModel) creatureTier() int {
	total := m.totalMutationsPurchased()
	switch {
	case total == 0:
		return 0
	case total < balance.CreatureTier2Threshold:
		return 1
	case total < balance.CreatureTier3Threshold:
		return 2
	default:
		return 3
	}
}

// totalMutationsPurchased returns the sum of all mutation purchase counts.
func (m GameModel) totalMutationsPurchased() int {
	total := 0
	for _, s := range m.MutationStates {
		total += s.PurchaseCount
	}
	return total
}

// dominantBranch returns the index (0=blood, 1=flesh, 2=bones) of the branch
// with the most total mutation purchases. Blood wins ties.
func (m GameModel) dominantBranch() int {
	counts := [3]int{}
	for i, def := range m.MutationDefs {
		counts[int(def.Branch)] += m.MutationStates[i].PurchaseCount
	}
	best := 0
	for b := 1; b < 3; b++ {
		if counts[b] > counts[best] {
			best = b
		}
	}
	return best
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
		m.ActiveTab = TabMutations
	case "3":
		m.ActiveTab = TabHarvesters
	case "b":
		return harvestKey(m, ResourceBlood, ZoneBlood)
	case "f":
		return harvestKey(m, ResourceFlesh, ZoneFlesh)
	case "n":
		return harvestKey(m, ResourceBones, ZoneBones)
	case "up", "k":
		m = handleListUp(m)
	case "down", "j":
		m = handleListDown(m)
	case "enter":
		return handleListEnter(m)
	}
	return m, nil
}

// harvestKey attempts to harvest resource r via keybind. No-op if zone locked.
// Applies BranchYieldMultiplier to scale the harvest amount.
func harvestKey(m GameModel, r ResourceType, z ZoneType) (GameModel, tea.Cmd) {
	if !m.ZoneUnlocked[r] {
		return m, nil
	}
	branch := BranchID(r) // ResourceBlood=0=BranchBlood, ResourceFlesh=1=BranchFlesh, etc.
	mult := BranchYieldMultiplier(m.MutationDefs, m.MutationStates, branch)
	m.Ledger.Add(r, balance.HarvestYield*mult)
	m.flashZone = z
	return m, func() tea.Msg { return clearFlashMsg{} }
}

// handleListUp moves the cursor up in the active list (mutations or harvesters).
// Clamps cursor to 0 minimum (T-02-08).
func handleListUp(m GameModel) GameModel {
	switch m.ActiveTab {
	case TabMutations:
		if m.MutationCursor < 0 {
			// Initialize cursor to last visible item
			m.MutationCursor = len(m.MutationDefs) - 1
		} else if m.MutationCursor > 0 {
			m.MutationCursor--
		}
		// Adjust scroll if cursor moved above visible window
		if m.MutationCursor < m.MutationScroll {
			m.MutationScroll = m.MutationCursor
		}
	case TabHarvesters:
		if m.HarvesterCursor < 0 {
			m.HarvesterCursor = len(m.HarvesterDefs) - 1
		} else if m.HarvesterCursor > 0 {
			m.HarvesterCursor--
		}
		if m.HarvesterCursor < m.HarvesterScroll {
			m.HarvesterScroll = m.HarvesterCursor
		}
	}
	return m
}

// handleListDown moves the cursor down in the active list (mutations or harvesters).
// Clamps cursor to len-1 maximum (T-02-08).
func handleListDown(m GameModel) GameModel {
	switch m.ActiveTab {
	case TabMutations:
		if m.MutationCursor < 0 {
			m.MutationCursor = 0
		} else if m.MutationCursor < len(m.MutationDefs)-1 {
			m.MutationCursor++
		}
	case TabHarvesters:
		if m.HarvesterCursor < 0 {
			m.HarvesterCursor = 0
		} else if m.HarvesterCursor < len(m.HarvesterDefs)-1 {
			m.HarvesterCursor++
		}
	}
	return m
}

// handleListEnter attempts to purchase the item at the current cursor position.
func handleListEnter(m GameModel) (GameModel, tea.Cmd) {
	switch m.ActiveTab {
	case TabMutations:
		if m.MutationCursor >= 0 {
			return tryPurchaseAndFlashMutation(m, m.MutationCursor)
		}
	case TabHarvesters:
		if m.HarvesterCursor >= 0 {
			return tryPurchaseAndFlashHarvester(m, m.HarvesterCursor)
		}
	}
	return m, nil
}

// handleScrollUp scrolls the active list up by one item.
func handleScrollUp(m GameModel) GameModel {
	switch m.ActiveTab {
	case TabMutations:
		if m.MutationScroll > 0 {
			m.MutationScroll--
		}
	case TabHarvesters:
		if m.HarvesterScroll > 0 {
			m.HarvesterScroll--
		}
	}
	return m
}

// handleScrollDown scrolls the active list down by one item.
func handleScrollDown(m GameModel) GameModel {
	switch m.ActiveTab {
	case TabMutations:
		if m.MutationScroll < len(m.MutationDefs)-1 {
			m.MutationScroll++
		}
	case TabHarvesters:
		if m.HarvesterScroll < len(m.HarvesterDefs)-1 {
			m.HarvesterScroll++
		}
	}
	return m
}

// tryPurchaseAndFlashMutation attempts to purchase a mutation at idx.
// On success: flashes the purchased item and rebuilds harvester rates
// (mutation buffs affect harvester output).
func tryPurchaseAndFlashMutation(m GameModel, idx int) (GameModel, tea.Cmd) {
	updated, ok := TryPurchaseMutation(m, idx)
	if !ok {
		return m, nil
	}
	// Rebuild rates — mutation buffs may have changed (Pitfall 1)
	updated.Ledger = RecalcAllRates(updated)
	updated.flashMutationIdx = idx
	return updated, func() tea.Msg { return clearMutationFlashMsg{} }
}

// tryPurchaseAndFlashHarvester attempts to purchase a harvester at idx.
// On success: flashes the purchased item.
func tryPurchaseAndFlashHarvester(m GameModel, idx int) (GameModel, tea.Cmd) {
	updated, ok := TryPurchaseHarvester(m, idx)
	if !ok {
		return m, nil
	}
	updated.flashHarvesterIdx = idx
	return updated, func() tea.Msg { return clearHarvesterFlashMsg{} }
}

// handleClick routes a left-button click to the appropriate zone, tab, or list item.
func handleClick(m GameModel, x, y int) (GameModel, tea.Cmd) {
	// Tab bar is at row 1.
	if y == 1 {
		m = handleTabClick(m, x)
		return m, nil
	}

	// Mutations panel hit-test (when on mutations tab, right panel shows mutations)
	if m.ActiveTab == TabMutations && x >= m.zonePanelLeft {
		// Each item occupies itemRowHeight rows; content starts at contentStartRow.
		const contentStartRow = 2
		const itemRowHeight = 4
		if y >= contentStartRow {
			idx := (y-contentStartRow)/itemRowHeight + m.MutationScroll
			// Bounds-check guard (T-02-07)
			if idx >= 0 && idx < len(m.MutationDefs) {
				return tryPurchaseAndFlashMutation(m, idx)
			}
		}
		return m, nil
	}

	// Harvesters panel hit-test (when on harvesters tab, right panel shows harvesters)
	if m.ActiveTab == TabHarvesters && x >= m.zonePanelLeft {
		const contentStartRow = 2
		const itemRowHeight = 4
		if y >= contentStartRow {
			idx := (y-contentStartRow)/itemRowHeight + m.HarvesterScroll
			// Bounds-check guard (T-02-07)
			if idx >= 0 && idx < len(m.HarvesterDefs) {
				return tryPurchaseAndFlashHarvester(m, idx)
			}
		}
		return m, nil
	}

	// Zones panel hit-test (zones tab or non-list-panel area)
	if x >= m.zonePanelLeft && x < m.zonePanelRight {
		switch {
		case y >= m.zoneBloodTop && y < m.zoneBloodTop+ZoneBoxHeight:
			if m.ZoneUnlocked[ResourceBlood] {
				branch := BranchID(ResourceBlood)
				mult := BranchYieldMultiplier(m.MutationDefs, m.MutationStates, branch)
				m.Ledger.Add(ResourceBlood, balance.HarvestYield*mult)
				m.flashZone = ZoneBlood
				return m, func() tea.Msg { return clearFlashMsg{} }
			}
		case y >= m.zoneFleshTop && y < m.zoneFleshTop+ZoneBoxHeight:
			if m.ZoneUnlocked[ResourceFlesh] {
				branch := BranchID(ResourceFlesh)
				mult := BranchYieldMultiplier(m.MutationDefs, m.MutationStates, branch)
				m.Ledger.Add(ResourceFlesh, balance.HarvestYield*mult)
				m.flashZone = ZoneFlesh
				return m, func() tea.Msg { return clearFlashMsg{} }
			}
		case y >= m.zoneBonesTop && y < m.zoneBonesTop+ZoneBoxHeight:
			if m.ZoneUnlocked[ResourceBones] {
				branch := BranchID(ResourceBones)
				mult := BranchYieldMultiplier(m.MutationDefs, m.MutationStates, branch)
				m.Ledger.Add(ResourceBones, balance.HarvestYield*mult)
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
