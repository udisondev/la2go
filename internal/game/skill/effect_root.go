package skill

import "log/slog"

// RootEffect disables movement only (can still use skills/attack).
// No params needed.
//
// Phase 5.9.3: Effect Framework.
// Java reference: BlockActions.java (with ROOT)
type RootEffect struct{}

func NewRootEffect(_ map[string]string) Effect {
	return &RootEffect{}
}

func (e *RootEffect) Name() string    { return "Root" }
func (e *RootEffect) IsInstant() bool { return false }

func (e *RootEffect) OnStart(casterObjID, targetObjID uint32) {
	slog.Debug("root applied", "target", targetObjID)
}

func (e *RootEffect) OnActionTime(casterObjID, targetObjID uint32) bool {
	return true
}

func (e *RootEffect) OnExit(casterObjID, targetObjID uint32) {
	slog.Debug("root removed", "target", targetObjID)
}
