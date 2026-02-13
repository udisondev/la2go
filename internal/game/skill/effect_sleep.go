package skill

import "log/slog"

// SleepEffect puts target to sleep (no actions/movement, breaks on damage).
// Sets the sleeping flag on the target Character.
// No params needed.
//
// Phase 5.9.3+: Actual sleep flag management.
// Java reference: BlockActions.java (with SLEEP flag).
// Java: onStart() → effected.startSleeping(), onExit() → effected.stopSleeping(false)
type SleepEffect struct{}

func NewSleepEffect(_ map[string]string) Effect {
	return &SleepEffect{}
}

func (e *SleepEffect) Name() string    { return "Sleep" }
func (e *SleepEffect) IsInstant() bool { return false }

func (e *SleepEffect) OnStart(casterObjID, targetObjID uint32) {
	target := resolveCharacter(targetObjID)
	if target == nil {
		return
	}
	target.SetSleeping(true)
	slog.Debug("sleep applied", "target", targetObjID)
}

func (e *SleepEffect) OnActionTime(casterObjID, targetObjID uint32) bool {
	return true
}

func (e *SleepEffect) OnExit(casterObjID, targetObjID uint32) {
	target := resolveCharacter(targetObjID)
	if target != nil {
		target.SetSleeping(false)
	}
	slog.Debug("sleep removed", "target", targetObjID)
}
