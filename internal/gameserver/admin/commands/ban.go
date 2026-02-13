package commands

import (
	"fmt"

	"github.com/udisondev/la2go/internal/model"
)

// Ban handles //ban <playerName> — sets player's access level to -100 (banned).
// Optionally kicks the player if online.
//
// Java reference: AdminBan.java
type Ban struct {
	clientMgr ClientManager
}

// NewBan creates the ban command handler.
func NewBan(clientMgr ClientManager) *Ban {
	return &Ban{clientMgr: clientMgr}
}

func (c *Ban) Names() []string           { return []string{"ban"} }
func (c *Ban) RequiredAccessLevel() int32 { return 1 }

func (c *Ban) Handle(player *model.Player, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: //ban <playerName>")
	}

	targetName := args[1]
	target := c.clientMgr.FindPlayerByName(targetName)
	if target == nil {
		// Player offline — ban needs DB update (Phase 17 MVP: only online players)
		return fmt.Errorf("player %q not found online (offline ban requires DB)", targetName)
	}

	// Set access level to -100 (banned)
	target.SetAccessLevel(-100)

	// Kick the player
	c.clientMgr.KickPlayer(targetName)

	player.SetLastAdminMessage(fmt.Sprintf("Banned player %s (set access_level = -100)", targetName))
	return nil
}
