package model

// Intention represents AI state for NPCs and Creatures
type Intention int32

const (
	// IntentionIdle - NPC is standing idle, no active behavior
	IntentionIdle Intention = iota
	// IntentionActive - NPC is actively wandering or performing random actions
	IntentionActive
	// IntentionAttack - NPC is attacking a target
	IntentionAttack
	// IntentionCast - NPC is casting a skill
	IntentionCast
	// IntentionMoveTo - NPC is moving to a specific location
	IntentionMoveTo
	// IntentionFollow - Summon is following its owner (Phase 19)
	IntentionFollow
)

// String returns human-readable intention name
func (i Intention) String() string {
	switch i {
	case IntentionIdle:
		return "IDLE"
	case IntentionActive:
		return "ACTIVE"
	case IntentionAttack:
		return "ATTACK"
	case IntentionCast:
		return "CAST"
	case IntentionMoveTo:
		return "MOVE_TO"
	case IntentionFollow:
		return "FOLLOW"
	default:
		return "UNKNOWN"
	}
}
