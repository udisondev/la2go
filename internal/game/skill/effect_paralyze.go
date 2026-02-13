package skill

import "log/slog"

// ParalyzeEffect disables all movement and actions.
// Sets the paralyzed flag on the target Character.
// No params needed.
//
// Phase 5.9.3+: Actual paralyze flag management.
// Java reference: Paralyze.java.
// Java: onStart() → effected.startParalyze(), onExit() → effected.stopParalyze()
type ParalyzeEffect struct{}

func NewParalyzeEffect(_ map[string]string) Effect {
	return &ParalyzeEffect{}
}

func (e *ParalyzeEffect) Name() string    { return "Paralyze" }
func (e *ParalyzeEffect) IsInstant() bool { return false }

func (e *ParalyzeEffect) OnStart(casterObjID, targetObjID uint32) {
	target := resolveCharacter(targetObjID)
	if target == nil {
		return
	}
	target.SetParalyzed(true)
	slog.Debug("paralyze applied", "target", targetObjID)
}

func (e *ParalyzeEffect) OnActionTime(casterObjID, targetObjID uint32) bool {
	return true
}

func (e *ParalyzeEffect) OnExit(casterObjID, targetObjID uint32) {
	target := resolveCharacter(targetObjID)
	if target != nil {
		target.SetParalyzed(false)
	}
	slog.Debug("paralyze removed", "target", targetObjID)
}
