package commands

import (
	"fmt"

	"github.com/udisondev/la2go/internal/model"
)

// Res handles //res [target] — resurrects a dead character.
//
// Java reference: AdminRes.java
type Res struct {
	clientMgr ClientManager
}

// NewRes creates the res command handler.
func NewRes(clientMgr ClientManager) *Res {
	return &Res{clientMgr: clientMgr}
}

func (c *Res) Names() []string           { return []string{"res", "resurrect"} }
func (c *Res) RequiredAccessLevel() int32 { return 1 }

func (c *Res) Handle(player *model.Player, args []string) error {
	if len(args) >= 2 {
		target := c.clientMgr.FindPlayerByName(args[1])
		if target == nil {
			return fmt.Errorf("player %q not found", args[1])
		}
		resCharacter(target.Character)
		player.SetLastAdminMessage(fmt.Sprintf("Resurrected player %s", target.Name()))
		return nil
	}

	target := player.Target()
	if target == nil {
		return fmt.Errorf("no target selected")
	}

	switch d := target.Data.(type) {
	case *model.Player:
		if !d.IsDead() {
			return fmt.Errorf("player %s is not dead", d.Name())
		}
		resCharacter(d.Character)
		player.SetLastAdminMessage(fmt.Sprintf("Resurrected player %s", d.Name()))
	case *model.Monster:
		resCharacter(d.Character)
		player.SetLastAdminMessage(fmt.Sprintf("Resurrected monster %s", d.Name()))
	case *model.Npc:
		resCharacter(d.Character)
		player.SetLastAdminMessage(fmt.Sprintf("Resurrected NPC %s", d.Name()))
	default:
		return fmt.Errorf("target cannot be resurrected (objectID: %d)", target.ObjectID())
	}

	return nil
}

func resCharacter(ch *model.Character) {
	ch.ResetDeathOnce()
	ch.SetCurrentHP(ch.MaxHP())
	ch.SetCurrentMP(ch.MaxMP())
	ch.SetCurrentCP(ch.MaxCP())
}
