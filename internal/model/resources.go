package model

import (
	"fmt"
	"math"
)

// ResourceType identifies one of the three harvestable resources.
type ResourceType int

const (
	ResourceBlood ResourceType = iota
	ResourceFlesh
	ResourceBones
	resourceCount // sentinel — array size only, not a valid resource
)

// ResourceLedger holds current amounts and per-second rates for all resources.
// Rates are 0.0/s in Phase 1 (no auto-harvesters). Phase 2 populates Rates.
type ResourceLedger struct {
	Amounts [resourceCount]float64
	Rates   [resourceCount]float64
}

// Add increases the amount of resource r by delta.
func (l *ResourceLedger) Add(r ResourceType, delta float64) {
	l.Amounts[r] += delta
}

// FormatAmount formats a resource count per UI-SPEC number formatting rules:
//
//	0–999: integer, 1000–999999: comma-separated, 1M+: abbreviated with 2dp.
func FormatAmount(v float64) string {
	switch {
	case v >= 1_000_000_000:
		return fmt.Sprintf("%.2fB", v/1_000_000_000)
	case v >= 1_000_000:
		return fmt.Sprintf("%.2fM", v/1_000_000)
	case v >= 1_000:
		// Format with comma separator
		i := int(math.Floor(v))
		return fmt.Sprintf("%d,%03d", i/1000, i%1000)
	default:
		return fmt.Sprintf("%d", int(math.Floor(v)))
	}
}

// FormatRate formats a per-second rate per UI-SPEC: always 1 decimal place,
// always prefixed with "+".
func FormatRate(v float64) string {
	return fmt.Sprintf("+%.1f/s", v)
}
