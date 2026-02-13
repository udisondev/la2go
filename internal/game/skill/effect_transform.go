package skill

import (
	"log/slog"
	"strconv"
)

// TransformEffect transforms the character into a different model.
// Params: "transformID" (int32).
//
// Java reference: Transformation.java — changes player appearance and abilities.
type TransformEffect struct {
	transformID int32
}

func NewTransformEffect(params map[string]string) Effect {
	tid, _ := strconv.Atoi(params["transformID"])
	return &TransformEffect{transformID: int32(tid)}
}

func (e *TransformEffect) Name() string    { return "Transform" }
func (e *TransformEffect) IsInstant() bool { return false }

func (e *TransformEffect) OnStart(casterObjID, targetObjID uint32) {
	slog.Debug("transform started",
		"transformID", e.transformID,
		"target", targetObjID)
	// Actual transformation handled by CastManager (needs UserInfo broadcast)
}

func (e *TransformEffect) OnActionTime(_, _ uint32) bool {
	return true // Continues until duration expires or manually cancelled
}

func (e *TransformEffect) OnExit(_, targetObjID uint32) {
	slog.Debug("transform ended", "target", targetObjID)
	// Revert transformation
}

// TransformID returns the transformation template ID.
func (e *TransformEffect) TransformID() int32 { return e.transformID }
