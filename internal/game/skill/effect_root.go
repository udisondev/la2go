package skill

import "log/slog"

// RootEffect disables movement only (can still use skills/attack).
// Sets the rooted flag on the target Character.
// No params needed.
//
// Phase 5.9.3+: Actual root flag management.
// Java reference: BlockActions.java (with ROOT flag).
// Java: onStart() → effected.startRooting(), onExit() → effected.stopRooting(false)
type RootEffect struct{}

func NewRootEffect(_ map[string]string) Effect {
	return &RootEffect{}
}

func (e *RootEffect) Name() string    { return "Root" }
func (e *RootEffect) IsInstant() bool { return false }

func (e *RootEffect) OnStart(casterObjID, targetObjID uint32) {
	target := resolveCharacter(targetObjID)
	if target == nil {
		return
	}
	target.SetRooted(true)
	slog.Debug("root applied", "target", targetObjID)
}

func (e *RootEffect) OnActionTime(casterObjID, targetObjID uint32) bool {
	return true
}

func (e *RootEffect) OnExit(casterObjID, targetObjID uint32) {
	target := resolveCharacter(targetObjID)
	if target != nil {
		target.SetRooted(false)
	}
	slog.Debug("root removed", "target", targetObjID)
}
