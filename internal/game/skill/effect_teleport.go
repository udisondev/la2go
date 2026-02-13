package skill

import (
	"log/slog"
	"strconv"
)

// TeleportEffect teleports the target to specified coordinates.
// Params: "x", "y", "z" (int32) — destination coordinates.
// If no coordinates specified, teleports to recall point (town).
//
// Java reference: Teleport.java, EscapeSkill.java
type TeleportEffect struct {
	x, y, z int32
}

func NewTeleportEffect(params map[string]string) Effect {
	x, _ := strconv.Atoi(params["x"])
	y, _ := strconv.Atoi(params["y"])
	z, _ := strconv.Atoi(params["z"])
	return &TeleportEffect{
		x: int32(x),
		y: int32(y),
		z: int32(z),
	}
}

func (e *TeleportEffect) Name() string    { return "Teleport" }
func (e *TeleportEffect) IsInstant() bool { return true }

func (e *TeleportEffect) OnStart(casterObjID, targetObjID uint32) {
	slog.Debug("teleport",
		"caster", casterObjID,
		"target", targetObjID,
		"dest", [3]int32{e.x, e.y, e.z})
	// Actual teleportation handled by CastManager (needs World access)
}

func (e *TeleportEffect) OnActionTime(_, _ uint32) bool { return false }
func (e *TeleportEffect) OnExit(_, _ uint32)            {}

// Destination returns the teleport target coordinates.
func (e *TeleportEffect) Destination() (x, y, z int32) {
	return e.x, e.y, e.z
}
