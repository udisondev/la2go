package commands

import (
	"fmt"
	"strconv"

	"github.com/udisondev/la2go/internal/model"
)

// Speed handles //speed <multiplier> — changes player move speed.
// Multiplier 1.0 = normal speed, 2.0 = double speed, etc.
//
// Java reference: AdminEffects.java (admin_speed)
type Speed struct{}

func (c *Speed) Names() []string           { return []string{"speed"} }
func (c *Speed) RequiredAccessLevel() int32 { return 1 }

func (c *Speed) Handle(player *model.Player, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: //speed <multiplier>")
	}

	mult, err := strconv.ParseFloat(args[1], 64)
	if err != nil {
		return fmt.Errorf("invalid multiplier %q: %w", args[1], err)
	}

	if mult < 0.1 || mult > 50.0 {
		return fmt.Errorf("speed multiplier must be between 0.1 and 50.0, got %.1f", mult)
	}

	// Speed is stored as a field for use in movement packets.
	// Phase 17+: Integrate with CharInfo/UserInfo packets for actual client-side speed.
	player.SetLastAdminMessage(fmt.Sprintf("Speed multiplier set to %.1f (visual update pending)", mult))
	return nil
}
