package instance

import "errors"

// Sentinel errors for the instance system.
var (
	ErrInvalidTemplateID = errors.New("invalid template ID")
	ErrEmptyTemplateName = errors.New("empty template name")
	ErrInvalidMaxPlayers = errors.New("invalid max players")
	ErrInvalidLevel      = errors.New("invalid level range")
	ErrTemplateNotFound  = errors.New("instance template not found")
	ErrInstanceNotFound  = errors.New("instance not found")
	ErrInstanceFull      = errors.New("instance is full")
	ErrAlreadyInInstance = errors.New("player already in an instance")
	ErrNotInInstance     = errors.New("player not in any instance")
	ErrOnCooldown        = errors.New("instance reentry on cooldown")
	ErrLevelTooLow       = errors.New("player level too low")
	ErrLevelTooHigh      = errors.New("player level too high")
	ErrInstanceExpired   = errors.New("instance has expired")
	ErrInstanceDestroyed = errors.New("instance is destroyed")
)
