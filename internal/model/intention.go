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
	default:
		return "UNKNOWN"
	}
}
