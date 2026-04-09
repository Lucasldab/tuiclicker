package model

import "testing"

func TestFormatOfflineMsgEmpty(t *testing.T) {
	got := FormatOfflineMsg([3]float64{0, 0, 0})
	if got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}

func TestFormatOfflineMsgSingleResource(t *testing.T) {
	got := FormatOfflineMsg([3]float64{42.5, 0, 0})
	want := "While idle: +42.5 blood"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestFormatOfflineMsgAllResources(t *testing.T) {
	got := FormatOfflineMsg([3]float64{10.0, 5.0, 3.3})
	want := "While idle: +10.0 blood, +5.0 flesh, +3.3 bones"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestFormatOfflineMsgBelowThreshold(t *testing.T) {
	got := FormatOfflineMsg([3]float64{0, 0, 0.0009})
	if got != "" {
		t.Errorf("sub-threshold delta should produce empty string, got %q", got)
	}
}

func TestWithOfflineMsgEmpty(t *testing.T) {
	m := New()
	m.OfflineCreditMsg = "prior"
	result := WithOfflineMsg(m, "")
	if result.OfflineCreditMsg != "prior" {
		t.Errorf("WithOfflineMsg('') should be no-op, got %q", result.OfflineCreditMsg)
	}
}

func TestWithOfflineMsgSetsField(t *testing.T) {
	m := New()
	result := WithOfflineMsg(m, "While idle: +10.0 blood")
	if result.OfflineCreditMsg != "While idle: +10.0 blood" {
		t.Errorf("unexpected OfflineCreditMsg: %q", result.OfflineCreditMsg)
	}
}
