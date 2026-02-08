package ai

import "github.com/udisondev/la2go/internal/model"

// Controller represents AI controller interface for NPCs
type Controller interface {
	// Start starts AI controller
	Start()

	// Stop stops AI controller
	Stop()

	// SetIntention sets AI intention
	SetIntention(intention model.Intention)

	// CurrentIntention returns current AI intention
	CurrentIntention() model.Intention

	// Tick performs AI tick (called every second)
	Tick()
}
