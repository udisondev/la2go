package skill

import "log/slog"

// StunEffect disables movement and actions for duration.
// Sets the stunned flag on the target Character.
// No params needed.
//
// Phase 5.9.3+: Actual stun flag management.
// Java reference: BlockActions.java (with STUN flag).
// Java: onStart() → effected.startStunning(), onExit() → effected.stopStunning(false)
type StunEffect struct{}

func NewStunEffect(_ map[string]string) Effect {
	return &StunEffect{}
}

func (e *StunEffect) Name() string    { return "Stun" }
func (e *StunEffect) IsInstant() bool { return false }

func (e *StunEffect) OnStart(casterObjID, targetObjID uint32) {
	target := resolveCharacter(targetObjID)
	if target == nil {
		return
	}
	target.SetStunned(true)
	slog.Debug("stun applied", "target", targetObjID)
}

func (e *StunEffect) OnActionTime(casterObjID, targetObjID uint32) bool {
	return true
}

func (e *StunEffect) OnExit(casterObjID, targetObjID uint32) {
	target := resolveCharacter(targetObjID)
	if target != nil {
		target.SetStunned(false)
	}
	slog.Debug("stun removed", "target", targetObjID)
}
