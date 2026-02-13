package commands

import (
	"fmt"

	"github.com/udisondev/la2go/internal/model"
	"github.com/udisondev/la2go/internal/world"
)

// Delete handles //delete — removes the targeted NPC from the world.
//
// Java reference: AdminDelete.java
type Delete struct{}

func (c *Delete) Names() []string           { return []string{"delete"} }
func (c *Delete) RequiredAccessLevel() int32 { return 2 }

func (c *Delete) Handle(player *model.Player, _ []string) error {
	target := player.Target()
	if target == nil {
		return fmt.Errorf("no target selected")
	}

	// Only allow deleting NPCs, not players
	switch target.Data.(type) {
	case *model.Npc, *model.Monster:
		// ok
	default:
		return fmt.Errorf("target is not an NPC (objectID: %d)", target.ObjectID())
	}

	w := world.Instance()
	w.RemoveObject(target.ObjectID())
	player.ClearTarget()

	player.SetLastAdminMessage(fmt.Sprintf("Deleted NPC %s (objectID: %d)",
		target.Name(), target.ObjectID()))
	return nil
}
