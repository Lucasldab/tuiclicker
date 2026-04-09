package game_test

import (
	"testing"

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
