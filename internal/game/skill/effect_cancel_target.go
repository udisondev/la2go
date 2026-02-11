package skill

import "log/slog"

// CancelTargetEffect instantly removes the target's current target.
// No params needed.
//
// Phase 5.9.3: Effect Framework.
// Java reference: FocusEnergy.java / CancelTarget (custom)
type CancelTargetEffect struct{}

func NewCancelTargetEffect(_ map[string]string) Effect {
	return &CancelTargetEffect{}
}

func (e *CancelTargetEffect) Name() string    { return "CancelTarget" }
func (e *CancelTargetEffect) IsInstant() bool { return true }

func (e *CancelTargetEffect) OnStart(casterObjID, targetObjID uint32) {
	slog.Debug("cancel target", "target", targetObjID)
}

func (e *CancelTargetEffect) OnActionTime(casterObjID, targetObjID uint32) bool {
	return false
}

func (e *CancelTargetEffect) OnExit(casterObjID, targetObjID uint32) {}
