package zone

// RespawnZone marks areas with race-specific respawn points.
// Java reference: RespawnZone.java — does NOT set any ZoneId flags.
// This is a data-only zone used by the respawn system.
type RespawnZone struct {
	*BaseZone
}

// NewRespawnZone creates a RespawnZone (no onEnter/onExit callbacks).
func NewRespawnZone(base *BaseZone) *RespawnZone {
	return &RespawnZone{BaseZone: base}
}

// IsPeace returns false.
func (z *RespawnZone) IsPeace() bool { return false }

// AllowsPvP returns true.
func (z *RespawnZone) AllowsPvP() bool { return true }
