package commands

import (
	"fmt"

	"github.com/udisondev/la2go/internal/model"
)

// Kill handles //kill [target] — kills the selected target or specified player.
//
// Java reference: AdminKill.java
type Kill struct {
	clientMgr ClientManager
}

// NewKill creates the kill command handler.
func NewKill(clientMgr ClientManager) *Kill {
	return &Kill{clientMgr: clientMgr}
}

func (c *Kill) Names() []string           { return []string{"kill"} }
func (c *Kill) RequiredAccessLevel() int32 { return 2 }

func (c *Kill) Handle(player *model.Player, args []string) error {
	// If player name specified, find that player
	if len(args) >= 2 {
		target := c.clientMgr.FindPlayerByName(args[1])
		if target == nil {
			return fmt.Errorf("player %q not found", args[1])
		}
		killCharacter(target.Character)
		player.SetLastAdminMessage(fmt.Sprintf("Killed player %s", target.Name()))
		return nil
	}

	// Otherwise kill current target
	target := player.Target()
	if target == nil {
		return fmt.Errorf("no target selected")
	}

	switch d := target.Data.(type) {
	case *model.Player:
		killCharacter(d.Character)
		player.SetLastAdminMessage(fmt.Sprintf("Killed player %s", d.Name()))
	case *model.Monster:
		killCharacter(d.Character)
		player.SetLastAdminMessage(fmt.Sprintf("Killed monster %s", d.Name()))
	case *model.Npc:
		killCharacter(d.Character)
		player.SetLastAdminMessage(fmt.Sprintf("Killed NPC %s", d.Name()))
	default:
		return fmt.Errorf("target cannot be killed (objectID: %d)", target.ObjectID())
	}

	return nil
}

func killCharacter(ch *model.Character) {
	ch.SetCurrentHP(0)
	ch.DoDie(nil)
}
