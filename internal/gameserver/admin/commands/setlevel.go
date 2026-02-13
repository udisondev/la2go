package commands

import (
	"fmt"
	"strconv"

	"github.com/udisondev/la2go/internal/model"
)

// SetLevel handles //setlevel <level> — changes target player's level.
//
// Java reference: AdminLevel.java
type SetLevel struct{}

func (c *SetLevel) Names() []string           { return []string{"setlevel", "set_level"} }
func (c *SetLevel) RequiredAccessLevel() int32 { return 2 }

func (c *SetLevel) Handle(player *model.Player, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: //setlevel <level>")
	}

	level, err := strconv.ParseInt(args[1], 10, 32)
	if err != nil {
		return fmt.Errorf("invalid level %q: %w", args[1], err)
	}

	if level < 1 || level > 80 {
		return fmt.Errorf("level must be between 1 and 80, got %d", level)
	}

	// Target player or self
	var target *model.Player

	tgt := player.Target()
	if tgt != nil {
		if p, ok := tgt.Data.(*model.Player); ok {
			target = p
		}
	}
	if target == nil {
		target = player
	}

	if err := target.SetLevel(int32(level)); err != nil {
		return fmt.Errorf("set level: %w", err)
	}

	player.SetLastAdminMessage(fmt.Sprintf("Set %s level to %d", target.Name(), level))
	return nil
}
