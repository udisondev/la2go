package skill

// Effect interface for all skill effects.
// Instant effects do work in OnStart and return true from IsInstant().
// Continuous effects tick via OnActionTime and clean up in OnExit.
//
// Phase 5.9.3: Effect Framework.
// Java reference: AbstractEffect.java
type Effect interface {
	Name() string
	IsInstant() bool
	OnStart(casterObjID, targetObjID uint32)
	OnActionTime(casterObjID, targetObjID uint32) bool // returns true to continue
	OnExit(casterObjID, targetObjID uint32)
}

// ActiveEffect tracks a running effect on a character.
// Created when a skill is cast and stored in EffectManager.
//
// Phase 5.9.3: Effect Framework.
// Java reference: BuffInfo.java
type ActiveEffect struct {
	CasterObjID   uint32
	TargetObjID   uint32
	SkillID       int32
	SkillLevel    int32
	Effect        Effect
	RemainingMs   int32
	PeriodMs      int32
	AbnormalType  string
	AbnormalLevel int32
}

// IsExpired returns true if the effect duration has elapsed.
func (ae *ActiveEffect) IsExpired() bool {
	return ae.RemainingMs <= 0
}

// Tick decrements remaining time by deltaMs.
// Returns true if effect is still active, false if expired.
func (ae *ActiveEffect) Tick(deltaMs int32) bool {
	ae.RemainingMs -= deltaMs
	return ae.RemainingMs > 0
}
