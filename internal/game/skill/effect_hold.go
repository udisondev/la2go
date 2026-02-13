package skill

import "log/slog"

// HoldEffect applies fear (target forced to flee from caster).
// Sets the feared flag on the target Character.
// No params needed.
//
// Phase 5.9.3+: Actual fear flag management.
// Java reference: Fear.java.
// Java: onStart() → effected.startFear(), onExit() → effected.stopFear(false)
type HoldEffect struct{}

func NewHoldEffect(_ map[string]string) Effect {
	return &HoldEffect{}
}

func (e *HoldEffect) Name() string    { return "Hold" }
func (e *HoldEffect) IsInstant() bool { return false }

func (e *HoldEffect) OnStart(casterObjID, targetObjID uint32) {
	target := resolveCharacter(targetObjID)
	if target == nil {
		return
	}
	target.SetFeared(true)
	slog.Debug("fear/hold applied", "target", targetObjID)
}

func (e *HoldEffect) OnActionTime(casterObjID, targetObjID uint32) bool {
	return true
}

func (e *HoldEffect) OnExit(casterObjID, targetObjID uint32) {
	target := resolveCharacter(targetObjID)
	if target != nil {
		target.SetFeared(false)
	}
	slog.Debug("fear/hold removed", "target", targetObjID)
}
