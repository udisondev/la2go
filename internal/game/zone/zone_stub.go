package zone

// This file contains stub zone types that don't set any flags
// but must be recognized by the factory to avoid "unknown zone type" warnings.

// --- ConditionZone ---
// Generic condition zone (1080 entries in data). No flags.

type ConditionZone struct{ *BaseZone }

func NewConditionZone(base *BaseZone) *ConditionZone {
	return &ConditionZone{BaseZone: base}
}
func (z *ConditionZone) IsPeace() bool   { return false }
func (z *ConditionZone) AllowsPvP() bool { return true }

// --- SiegableHallZone ---
// Siegeable clan hall zone. Similar to SiegeZone but for clan halls.

type SiegableHallZone struct{ *BaseZone }

func NewSiegableHallZone(base *BaseZone) *SiegableHallZone {
	return &SiegableHallZone{BaseZone: base}
}
func (z *SiegableHallZone) IsPeace() bool   { return false }
func (z *SiegableHallZone) AllowsPvP() bool { return true }

// --- ResidenceTeleportZone ---
// Teleport zone for residences (castles). No flags, data-only.

type ResidenceTeleportZone struct{ *BaseZone }

func NewResidenceTeleportZone(base *BaseZone) *ResidenceTeleportZone {
	return &ResidenceTeleportZone{BaseZone: base}
}
func (z *ResidenceTeleportZone) IsPeace() bool   { return false }
func (z *ResidenceTeleportZone) AllowsPvP() bool { return true }

// --- ResidenceHallTeleportZone ---
// Teleport zone for residence halls. No flags, data-only.

type ResidenceHallTeleportZone struct{ *BaseZone }

func NewResidenceHallTeleportZone(base *BaseZone) *ResidenceHallTeleportZone {
	return &ResidenceHallTeleportZone{BaseZone: base}
}
func (z *ResidenceHallTeleportZone) IsPeace() bool   { return false }
func (z *ResidenceHallTeleportZone) AllowsPvP() bool { return true }
