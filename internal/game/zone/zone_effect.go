package zone

// EffectZone represents an area that applies special effects to characters.
// Java reference: EffectZone.java
type EffectZone struct {
	*BaseZone
}

// NewEffectZone creates an EffectZone (stub — no ZoneId flags).
func NewEffectZone(base *BaseZone) *EffectZone {
	return &EffectZone{BaseZone: base}
}

// IsPeace returns false — effect zones are not peace zones.
func (z *EffectZone) IsPeace() bool { return false }

// AllowsPvP returns true — PvP is allowed in effect zones.
func (z *EffectZone) AllowsPvP() bool { return true }
