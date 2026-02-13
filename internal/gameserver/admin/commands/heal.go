package commands

import (
	"fmt"

	"github.com/udisondev/la2go/internal/model"
)

// Heal handles //heal [target] — fully heals the target.
//
// Java reference: AdminHeal.java
type Heal struct {
	clientMgr ClientManager
}

// NewHeal creates the heal command handler.
func NewHeal(clientMgr ClientManager) *Heal {
	return &Heal{clientMgr: clientMgr}
}

func (c *Heal) Names() []string           { return []string{"heal"} }
func (c *Heal) RequiredAccessLevel() int32 { return 1 }

func (c *Heal) Handle(player *model.Player, args []string) error {
	// If player name specified, heal that player
	if len(args) >= 2 {
		target := c.clientMgr.FindPlayerByName(args[1])
		if target == nil {
			return fmt.Errorf("player %q not found", args[1])
		}
		healCharacter(target.Character)
		player.SetLastAdminMessage(fmt.Sprintf("Healed player %s", target.Name()))
		return nil
	}

	// Otherwise heal current target
	target := player.Target()
	if target == nil {
		// Heal self
		healCharacter(player.Character)
		player.SetLastAdminMessage("Healed self")
		return nil
	}

	switch d := target.Data.(type) {
	case *model.Player:
		healCharacter(d.Character)
		player.SetLastAdminMessage(fmt.Sprintf("Healed player %s", d.Name()))
	case *model.Monster:
		healCharacter(d.Character)
		player.SetLastAdminMessage(fmt.Sprintf("Healed monster %s", d.Name()))
	case *model.Npc:
		healCharacter(d.Character)
		player.SetLastAdminMessage(fmt.Sprintf("Healed NPC %s", d.Name()))
	default:
		return fmt.Errorf("target cannot be healed (objectID: %d)", target.ObjectID())
	}

	return nil
}

func healCharacter(ch *model.Character) {
	ch.SetCurrentHP(ch.MaxHP())
	ch.SetCurrentMP(ch.MaxMP())
	ch.SetCurrentCP(ch.MaxCP())
}
