package skill

import "log/slog"

// ParalyzeEffect disables all actions (movement + skills + attack).
// No params needed.
//
// Phase 5.9.3: Effect Framework.
// Java reference: BlockActions.java (with PARALYZE)
type ParalyzeEffect struct{}

func NewParalyzeEffect(_ map[string]string) Effect {
	return &ParalyzeEffect{}
}

func (e *ParalyzeEffect) Name() string    { return "Paralyze" }
func (e *ParalyzeEffect) IsInstant() bool { return false }

func (e *ParalyzeEffect) OnStart(casterObjID, targetObjID uint32) {
	slog.Debug("paralyze applied", "target", targetObjID)
}

func (e *ParalyzeEffect) OnActionTime(casterObjID, targetObjID uint32) bool {
	return true
}

func (e *ParalyzeEffect) OnExit(casterObjID, targetObjID uint32) {
	slog.Debug("paralyze removed", "target", targetObjID)
}
