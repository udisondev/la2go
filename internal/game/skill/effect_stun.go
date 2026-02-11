package skill

import "log/slog"

// StunEffect disables movement and actions for duration.
// No params needed.
//
// Phase 5.9.3: Effect Framework.
// Java reference: BlockActions.java (with STUN)
type StunEffect struct{}

func NewStunEffect(_ map[string]string) Effect {
	return &StunEffect{}
}

func (e *StunEffect) Name() string    { return "Stun" }
func (e *StunEffect) IsInstant() bool { return false }

func (e *StunEffect) OnStart(casterObjID, targetObjID uint32) {
	slog.Debug("stun applied", "target", targetObjID)
}

func (e *StunEffect) OnActionTime(casterObjID, targetObjID uint32) bool {
	return true
}

func (e *StunEffect) OnExit(casterObjID, targetObjID uint32) {
	slog.Debug("stun removed", "target", targetObjID)
}
