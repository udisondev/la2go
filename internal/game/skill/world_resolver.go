package skill

import "github.com/udisondev/la2go/internal/model"

// worldResolver is the package-level callback for resolving objectID → WorldObject.
// Set via SetWorldResolver when CastManager is initialized.
// Effects use resolvePlayer/resolveCharacter helpers to access game objects.
//
// This follows the same pattern as Java's L2World.getInstance() but with
// explicit dependency injection instead of a singleton.
var worldResolver func(objectID uint32) (*model.WorldObject, bool)

// SetWorldResolver sets the package-level world object resolver.
// Called from CastManager.SetWorldObjectResolver to share the resolver with effects.
func SetWorldResolver(fn func(objectID uint32) (*model.WorldObject, bool)) {
	worldResolver = fn
}

// resolvePlayer resolves an objectID to a *model.Player.
// Returns nil if not found or not a Player.
func resolvePlayer(objectID uint32) *model.Player {
	if worldResolver == nil {
		return nil
	}
	obj, ok := worldResolver(objectID)
	if !ok || obj == nil || obj.Data == nil {
		return nil
	}
	p, ok := obj.Data.(*model.Player)
	if !ok {
		return nil
	}
	return p
}

// resolveCharacter resolves an objectID to a *model.Character.
// Works for Player, NPC, Monster, RaidBoss, GrandBoss.
// Returns nil if not found or type not recognized.
func resolveCharacter(objectID uint32) *model.Character {
	if worldResolver == nil {
		return nil
	}
	obj, ok := worldResolver(objectID)
	if !ok || obj == nil || obj.Data == nil {
		return nil
	}
	// Type assert order: most specific first (per MEMORY.md gotchas).
	switch d := obj.Data.(type) {
	case *model.Player:
		return d.Character
	case *model.RaidBoss:
		return d.Character
	case *model.GrandBoss:
		return d.Character
	case *model.Monster:
		return d.Character
	case *model.Npc:
		return d.Character
	default:
		return nil
	}
}

// targetMDef returns the target's magic defense.
// Players use GetMDef(); NPCs/Monsters use level-based formula.
func targetMDef(targetObjID uint32) int32 {
	if worldResolver == nil {
		return 100
	}
	obj, ok := worldResolver(targetObjID)
	if !ok || obj == nil || obj.Data == nil {
		return 100
	}
	if p, ok := obj.Data.(*model.Player); ok {
		return p.GetMDef()
	}
	// NPC fallback: level-based MDef (similar to GetBasePDef formula)
	char := resolveCharacter(targetObjID)
	if char == nil {
		return 100
	}
	level := char.Level()
	baseMDef := float64(40 + level*2)
	levelMod := float64(level+89) / 100.0
	return int32(baseMDef * levelMod)
}

// targetPDef returns the target's physical defense.
// Players use GetPDef(); NPCs/Monsters use Character.GetBasePDef().
func targetPDef(targetObjID uint32) int32 {
	if worldResolver == nil {
		return 100
	}
	obj, ok := worldResolver(targetObjID)
	if !ok || obj == nil || obj.Data == nil {
		return 100
	}
	if p, ok := obj.Data.(*model.Player); ok {
		return p.GetPDef()
	}
	char := resolveCharacter(targetObjID)
	if char == nil {
		return 100
	}
	return char.GetBasePDef()
}
