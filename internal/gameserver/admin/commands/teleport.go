package commands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/udisondev/la2go/internal/model"
	"github.com/udisondev/la2go/internal/world"
)

// Teleport handles //teleport <x> <y> <z> and //goto <playerName>.
//
// Java reference: AdminTeleport.java
type Teleport struct {
	clientMgr ClientManager
}

// NewTeleport creates the teleport command handler.
func NewTeleport(clientMgr ClientManager) *Teleport {
	return &Teleport{clientMgr: clientMgr}
}

func (c *Teleport) Names() []string {
	return []string{"teleport", "goto", "recall", "move_to"}
}

func (c *Teleport) RequiredAccessLevel() int32 { return 1 }

func (c *Teleport) Handle(player *model.Player, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: //teleport <x> <y> <z> | //goto <player> | //recall <player>")
	}

	cmd := strings.ToLower(args[0])
	switch cmd {
	case "teleport", "move_to":
		return c.handleTeleportXYZ(player, args[1:])
	case "goto":
		return c.handleGoto(player, args[1])
	case "recall":
		return c.handleRecall(player, args[1])
	default:
		return fmt.Errorf("unknown teleport subcommand: %s", cmd)
	}
}

func (c *Teleport) handleTeleportXYZ(player *model.Player, args []string) error {
	if len(args) < 3 {
		return fmt.Errorf("usage: //teleport <x> <y> <z>")
	}

	x, err := strconv.ParseInt(args[0], 10, 32)
	if err != nil {
		return fmt.Errorf("invalid x coordinate %q: %w", args[0], err)
	}
	y, err := strconv.ParseInt(args[1], 10, 32)
	if err != nil {
		return fmt.Errorf("invalid y coordinate %q: %w", args[1], err)
	}
	z, err := strconv.ParseInt(args[2], 10, 32)
	if err != nil {
		return fmt.Errorf("invalid z coordinate %q: %w", args[2], err)
	}

	teleportPlayer(player, int32(x), int32(y), int32(z))
	player.SetLastAdminMessage(fmt.Sprintf("Teleported to (%d, %d, %d)", x, y, z))
	return nil
}

func (c *Teleport) handleGoto(player *model.Player, targetName string) error {
	target := c.clientMgr.FindPlayerByName(targetName)
	if target == nil {
		return fmt.Errorf("player %q not found", targetName)
	}

	loc := target.Location()
	teleportPlayer(player, loc.X, loc.Y, loc.Z)
	player.SetLastAdminMessage(fmt.Sprintf("Teleported to player %s at (%d, %d, %d)",
		target.Name(), loc.X, loc.Y, loc.Z))
	return nil
}

func (c *Teleport) handleRecall(player *model.Player, targetName string) error {
	target := c.clientMgr.FindPlayerByName(targetName)
	if target == nil {
		return fmt.Errorf("player %q not found", targetName)
	}

	loc := player.Location()
	teleportPlayer(target, loc.X, loc.Y, loc.Z)
	player.SetLastAdminMessage(fmt.Sprintf("Recalled player %s to your location", target.Name()))
	return nil
}

// teleportPlayer moves a player to new coordinates and updates world region.
func teleportPlayer(player *model.Player, x, y, z int32) {
	w := world.Instance()

	// Remove from old region
	w.RemoveObject(player.ObjectID())

	// Update location
	player.SetLocation(model.NewLocation(x, y, z, player.Heading()))

	// Add to new region
	_ = w.AddObject(player.WorldObject)
}
