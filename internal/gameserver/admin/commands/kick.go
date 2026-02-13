package commands

import (
	"fmt"

	"github.com/udisondev/la2go/internal/model"
)

// Kick handles //kick <playerName> — disconnects a player.
//
// Java reference: AdminKick.java
type Kick struct {
	clientMgr ClientManager
}

// NewKick creates the kick command handler.
func NewKick(clientMgr ClientManager) *Kick {
	return &Kick{clientMgr: clientMgr}
}

func (c *Kick) Names() []string           { return []string{"kick"} }
func (c *Kick) RequiredAccessLevel() int32 { return 1 }

func (c *Kick) Handle(player *model.Player, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: //kick <playerName>")
	}

	targetName := args[1]
	if !c.clientMgr.KickPlayer(targetName) {
		return fmt.Errorf("player %q not found or already disconnected", targetName)
	}

	player.SetLastAdminMessage(fmt.Sprintf("Kicked player %s", targetName))
	return nil
}
