package commands

import (
	"fmt"
	"strconv"

	"github.com/udisondev/la2go/internal/data"
	"github.com/udisondev/la2go/internal/model"
	"github.com/udisondev/la2go/internal/world"
)

// Spawn handles //spawn <npcID> [count].
//
// Java reference: AdminSpawn.java
type Spawn struct{}

func (c *Spawn) Names() []string           { return []string{"spawn"} }
func (c *Spawn) RequiredAccessLevel() int32 { return 2 }

func (c *Spawn) Handle(player *model.Player, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: //spawn <npcID> [count]")
	}

	npcID, err := strconv.ParseInt(args[1], 10, 32)
	if err != nil {
		return fmt.Errorf("invalid npcID %q: %w", args[1], err)
	}

	count := int32(1)
	if len(args) >= 3 {
		c, err := strconv.ParseInt(args[2], 10, 32)
		if err != nil {
			return fmt.Errorf("invalid count %q: %w", args[2], err)
		}
		count = int32(c)
	}

	if count < 1 || count > 100 {
		return fmt.Errorf("count must be between 1 and 100, got %d", count)
	}

	npcDef := data.GetNpcDef(int32(npcID))
	if npcDef == nil {
		return fmt.Errorf("NPC template %d not found", npcID)
	}

	loc := player.Location()
	w := world.Instance()

	spawned := int32(0)
	for range count {
		tmpl := model.NewNpcTemplate(
			int32(npcID),
			npcDef.Name(), npcDef.Title(),
			npcDef.Level(),
			int32(npcDef.HP()), int32(npcDef.MP()),
			int32(npcDef.PAtk()), int32(npcDef.PDef()),
			int32(npcDef.MAtk()), int32(npcDef.MDef()),
			npcDef.AggroRange(), npcDef.RunSpeed(), npcDef.AtkSpeed(),
			0, 0, // respawnMin, respawnMax (admin spawn = no respawn)
			npcDef.BaseExp(), npcDef.BaseSP(),
		)

		objectID := world.IDGenerator().NextNpcID()
		npc := model.NewNpc(objectID, int32(npcID), tmpl)
		npc.SetLocation(model.NewLocation(loc.X, loc.Y, loc.Z, loc.Heading))
		npc.WorldObject.Data = npc

		if err := w.AddNpc(npc); err != nil {
			return fmt.Errorf("adding NPC to world: %w", err)
		}
		spawned++
	}

	player.SetLastAdminMessage(fmt.Sprintf("Spawned %d %s (ID: %d) at (%d, %d, %d)",
		spawned, npcDef.Name(), npcID, loc.X, loc.Y, loc.Z))
	return nil
}
