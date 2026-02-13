package commands

import (
	"fmt"
	"strings"

	"github.com/udisondev/la2go/internal/model"
)

// Announce handles //announce <text> — broadcasts message to all players.
//
// Java reference: AdminAnnouncements.java
type Announce struct {
	clientMgr ClientManager
}

// NewAnnounce creates the announce command handler.
func NewAnnounce(clientMgr ClientManager) *Announce {
	return &Announce{clientMgr: clientMgr}
}

func (c *Announce) Names() []string           { return []string{"announce", "ann"} }
func (c *Announce) RequiredAccessLevel() int32 { return 1 }

func (c *Announce) Handle(player *model.Player, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: //announce <text>")
	}

	text := strings.Join(args[1:], " ")
	if len(text) == 0 {
		return fmt.Errorf("announcement text cannot be empty")
	}

	// Store announcement text on player so handler can broadcast it
	// using CreatureSay with ChatAnnounce type
	player.SetLastAdminMessage("ANNOUNCE:" + text)
	return nil
}
