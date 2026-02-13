package commands

import (
	"fmt"
	"strconv"

	"github.com/udisondev/la2go/internal/model"
	"github.com/udisondev/la2go/internal/world"
)

// Jail handles //jail <playerName> <minutes> — teleports player to jail.
//
// Jail location: -114356, -249645, -2984 (L2J default jail location).
// Java reference: AdminJail.java
type Jail struct {
	clientMgr ClientManager
}

// Jail coordinates (L2J default).
const (
	jailX = int32(-114356)
	jailY = int32(-249645)
	jailZ = int32(-2984)
)

// NewJail creates the jail command handler.
func NewJail(clientMgr ClientManager) *Jail {
	return &Jail{clientMgr: clientMgr}
}

func (c *Jail) Names() []string           { return []string{"jail"} }
func (c *Jail) RequiredAccessLevel() int32 { return 1 }

func (c *Jail) Handle(player *model.Player, args []string) error {
	if len(args) < 3 {
		return fmt.Errorf("usage: //jail <playerName> <minutes>")
	}

	targetName := args[1]
	minutes, err := strconv.ParseInt(args[2], 10, 32)
	if err != nil {
		return fmt.Errorf("invalid minutes %q: %w", args[2], err)
	}

	if minutes < 1 {
		return fmt.Errorf("jail time must be positive, got %d", minutes)
	}

	target := c.clientMgr.FindPlayerByName(targetName)
	if target == nil {
		return fmt.Errorf("player %q not found online", targetName)
	}

	// Teleport to jail
	w := world.Instance()
	w.RemoveObject(target.ObjectID())
	target.SetLocation(model.NewLocation(jailX, jailY, jailZ, 0))
	_ = w.AddObject(target.WorldObject)

	player.SetLastAdminMessage(fmt.Sprintf("Jailed player %s for %d minutes", targetName, minutes))
	return nil
}
