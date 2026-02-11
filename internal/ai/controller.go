package ai

import "github.com/udisondev/la2go/internal/model"

// Controller represents AI controller interface for NPCs.
// Phase 5.7: Added NotifyDamage and Npc methods.
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

	// NotifyDamage notifies AI that NPC took damage from attacker.
	// Phase 5.7: Used to cancel spawn immunity and add attacker to hate list.
	NotifyDamage(attackerID uint32, damage int32)

	// Npc returns the underlying NPC for this controller.
	// Phase 5.7: Used by CombatManager to access NPC data.
	Npc() *model.Npc
}
