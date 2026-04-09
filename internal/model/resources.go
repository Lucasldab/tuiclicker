package model

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
