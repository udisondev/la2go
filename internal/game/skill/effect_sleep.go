package skill

import "log/slog"

// SleepEffect disables all actions until hit.
// No params needed.
//
// Phase 5.9.3: Effect Framework.
// Java reference: BlockActions.java (with SLEEP)
type SleepEffect struct{}

func NewSleepEffect(_ map[string]string) Effect {
	return &SleepEffect{}
}

func (e *SleepEffect) Name() string    { return "Sleep" }
func (e *SleepEffect) IsInstant() bool { return false }

func (e *SleepEffect) OnStart(casterObjID, targetObjID uint32) {
	slog.Debug("sleep applied", "target", targetObjID)
}

func (e *SleepEffect) OnActionTime(casterObjID, targetObjID uint32) bool {
	return true
}

func (e *SleepEffect) OnExit(casterObjID, targetObjID uint32) {
	slog.Debug("sleep removed", "target", targetObjID)
}
