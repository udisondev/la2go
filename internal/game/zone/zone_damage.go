package zone

import "strconv"

const defaultDamageInterval int32 = 3 // секунды

// DamageZone represents an area that deals periodic damage to characters inside it.
// Does NOT set any ZoneId flags (Java: DamageZone has no ZoneId in onEnter/onExit).
// Java reference: DamageZone.java
type DamageZone struct {
	*BaseZone
}

// NewDamageZone creates a DamageZone (no onEnter/onExit callbacks, damage only).
func NewDamageZone(base *BaseZone) *DamageZone {
	return &DamageZone{BaseZone: base}
}

// IsPeace returns false — damage zones are hostile.
func (z *DamageZone) IsPeace() bool { return false }

// AllowsPvP returns true — PvP is allowed in damage zones.
func (z *DamageZone) AllowsPvP() bool { return true }

// DamagePerSecond returns the amount of HP damage dealt per tick.
// Parsed from zone params["damagePerSecond"], defaults to 0.
func (z *DamageZone) DamagePerSecond() int32 {
	v, ok := z.params["damagePerSecond"]
	if !ok {
		return 0
	}

	n, err := strconv.Atoi(v)
	if err != nil {
		return 0
	}

	return int32(n)
}

// DamageInterval returns the interval in seconds between damage ticks.
// Parsed from zone params["damageInterval"], defaults to 3 seconds.
func (z *DamageZone) DamageInterval() int32 {
	v, ok := z.params["damageInterval"]
	if !ok {
		return defaultDamageInterval
	}

	n, err := strconv.Atoi(v)
	if err != nil {
		return defaultDamageInterval
	}

	return int32(n)
}
