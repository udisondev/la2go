package zone

// FishingZone marks areas where fishing is allowed.
// Java reference: FishingZone.java — does NOT set any ZoneId flags.
// Used only for coordinate checks when player attempts to fish.
type FishingZone struct {
	*BaseZone
}

// NewFishingZone creates a FishingZone (no onEnter/onExit callbacks).
func NewFishingZone(base *BaseZone) *FishingZone {
	return &FishingZone{BaseZone: base}
}

// IsPeace returns false.
func (z *FishingZone) IsPeace() bool { return false }

// AllowsPvP returns true.
func (z *FishingZone) AllowsPvP() bool { return true }

// WaterZ returns the water surface Z coordinate (top of zone).
func (z *FishingZone) WaterZ() int32 { return z.maxZ }
