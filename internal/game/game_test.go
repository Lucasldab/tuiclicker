package game_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/lucasldab/tuiclicker/internal/model"
)

// --- FormatAmount ---

func TestFormatAmountSmall(t *testing.T) {
	cases := []struct {
		in  float64
		out string
	}{
		{0, "0"}, {42, "42"}, {999, "999"},
	}
	for _, c := range cases {
		got := model.FormatAmount(c.in)
		if got != c.out {
			t.Errorf("FormatAmount(%v) = %q, want %q", c.in, got, c.out)
		}
	}
}

func TestFormatAmountThousands(t *testing.T) {
	cases := []struct {
		in  float64
		out string
	}{
		{1000, "1,000"}, {1234, "1,234"}, {999999, "999,999"},
	}
	for _, c := range cases {
		got := model.FormatAmount(c.in)
		if got != c.out {
			t.Errorf("FormatAmount(%v) = %q, want %q", c.in, got, c.out)
		}
	}
}

func TestFormatAmountMillions(t *testing.T) {
	cases := []struct {
		in  float64
		out string
	}{
		{1_000_000, "1.00M"}, {1_234_567, "1.23M"},
	}
	for _, c := range cases {
		got := model.FormatAmount(c.in)
		if got != c.out {
			t.Errorf("FormatAmount(%v) = %q, want %q", c.in, got, c.out)
		}
	}
}

func TestFormatAmountBillions(t *testing.T) {
	cases := []struct {
		in  float64
		out string
	}{
		{1_000_000_000, "1.00B"}, {4_567_890_123, "4.57B"},
	}
	for _, c := range cases {
		got := model.FormatAmount(c.in)
		if got != c.out {
			t.Errorf("FormatAmount(%v) = %q, want %q", c.in, got, c.out)
		}
	}
}

func TestFormatRate(t *testing.T) {
	cases := []struct {
		in  float64
		out string
	}{
		{0.0, "+0.0/s"}, {1.5, "+1.5/s"}, {12.3, "+12.3/s"},
	}
	for _, c := range cases {
		got := model.FormatRate(c.in)
		if got != c.out {
			t.Errorf("FormatRate(%v) = %q, want %q", c.in, got, c.out)
		}
	}
}

// --- Resource harvesting ---

func TestHarvestBloodIncrementsLedger(t *testing.T) {
	m := model.New()
	before := m.Ledger.Amounts[model.ResourceBlood]
	m.Ledger.Add(model.ResourceBlood, 1.0)
	if m.Ledger.Amounts[model.ResourceBlood] != before+1.0 {
		t.Errorf("blood ledger not incremented")
	}
}

func TestHarvestFleshDoesNotAffectBlood(t *testing.T) {
	m := model.New()
	m.Ledger.Add(model.ResourceFlesh, 5.0)
	if m.Ledger.Amounts[model.ResourceBlood] != 0 {
		t.Errorf("flesh harvest should not affect blood")
	}
	if m.Ledger.Amounts[model.ResourceFlesh] != 5.0 {
		t.Errorf("flesh ledger incorrect")
	}
}

func TestAllThreeResourcesIndependent(t *testing.T) {
	m := model.New()
	m.Ledger.Add(model.ResourceBlood, 10)
	m.Ledger.Add(model.ResourceFlesh, 20)
	m.Ledger.Add(model.ResourceBones, 30)
	if m.Ledger.Amounts[model.ResourceBlood] != 10 {
		t.Error("blood wrong")
	}
	if m.Ledger.Amounts[model.ResourceFlesh] != 20 {
		t.Error("flesh wrong")
	}
	if m.Ledger.Amounts[model.ResourceBones] != 30 {
		t.Error("bones wrong")
	}
}

// --- Click handling ---

func TestClickBloodZoneHarvests(t *testing.T) {
	m := model.New()
	// Blood zone: x within right panel, y = contentStartRow + BloodZoneOffset
	// At 80-wide: rightW = (80*30)/100 = 24, zonePanelLeft = 80-24 = 56.
	// Use x=60 (within [56,80)). zoneBloodTop = 3 + 0 = 3.
	updated, cmd := m.Update(makeMouse(60, 3))
	gm := updated.(model.GameModel)
	if gm.Ledger.Amounts[model.ResourceBlood] != 1.0 {
		t.Errorf("blood should be 1.0 after click, got %v", gm.Ledger.Amounts[model.ResourceBlood])
	}
	if gm.FlashZone() != model.ZoneBlood {
		t.Errorf("flashZone should be ZoneBlood after click")
	}
	if cmd == nil {
		t.Errorf("should return clearFlashMsg cmd after click")
	}
}

func TestClickFleshZoneLockedNoOp(t *testing.T) {
	m := model.New()
	// zoneFleshTop = 3 + 7 = 10. x=60 (within panel [56,80)).
	updated, _ := m.Update(makeMouse(60, 10))
	gm := updated.(model.GameModel)
	if gm.Ledger.Amounts[model.ResourceFlesh] != 0 {
		t.Errorf("flesh should not increment when locked")
	}
	if gm.FlashZone() != model.ZoneNone {
		t.Errorf("no flash when zone locked")
	}
}

func TestClickBonesZoneLockedNoOp(t *testing.T) {
	m := model.New()
	// zoneBonesTop = 3 + 14 = 17. x=60 (within panel [56,80)).
	updated, _ := m.Update(makeMouse(60, 17))
	gm := updated.(model.GameModel)
	if gm.Ledger.Amounts[model.ResourceBones] != 0 {
		t.Errorf("bones should not increment when locked")
	}
}

func TestClearFlashMsgResetsFlash(t *testing.T) {
	m := model.New()
	// Simulate a click that sets flash
	m.Ledger.Add(model.ResourceBlood, 0) // no-op, just to use the var
	updated1, _ := m.Update(makeMouse(60, 3))
	gm1 := updated1.(model.GameModel)
	if gm1.FlashZone() != model.ZoneBlood {
		t.Skip("flash not set — check TestClickBloodZoneHarvests first")
	}
	// Now send clearFlashMsg
	updated2, _ := gm1.Update(model.ClearFlashMsg())
	gm2 := updated2.(model.GameModel)
	if gm2.FlashZone() != model.ZoneNone {
		t.Errorf("flash should clear after clearFlashMsg")
	}
}

func TestKeybindBHarvestsBlood(t *testing.T) {
	m := model.New()
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("b")})
	gm := updated.(model.GameModel)
	if gm.Ledger.Amounts[model.ResourceBlood] != 1.0 {
		t.Errorf("keybind b should harvest blood")
	}
}

func TestKeybindFLockedNoOp(t *testing.T) {
	m := model.New()
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("f")})
	gm := updated.(model.GameModel)
	if gm.Ledger.Amounts[model.ResourceFlesh] != 0 {
		t.Errorf("keybind f should be no-op when flesh locked")
	}
}

func TestWindowSizeSmall(t *testing.T) {
	m := model.New()
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 59, Height: 20})
	gm := updated.(model.GameModel)
	if !gm.TooSmall() {
		t.Errorf("should be tooSmall when width < 60")
	}
}

func TestWindowSizeNormal(t *testing.T) {
	m := model.New()
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	gm := updated.(model.GameModel)
	if gm.TooSmall() {
		t.Errorf("should not be tooSmall at 80x24")
	}
}

// --- helpers ---

func makeMouse(x, y int) tea.MouseMsg {
	return tea.MouseMsg{
		X:      x,
		Y:      y,
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonLeft,
	}
}
