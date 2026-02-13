package commands

import (
	"fmt"

	"github.com/udisondev/la2go/internal/model"
)

// Invisible handles //invisible — toggles GM invisible mode.
//
// Java reference: AdminEffects.java
type Invisible struct{}

func (c *Invisible) Names() []string           { return []string{"invisible", "invis", "vis"} }
func (c *Invisible) RequiredAccessLevel() int32 { return 1 }

func (c *Invisible) Handle(player *model.Player, _ []string) error {
	current := player.IsInvisible()
	player.SetInvisible(!current)

	if player.IsInvisible() {
		player.SetLastAdminMessage("You are now invisible")
	} else {
		player.SetLastAdminMessage("You are now visible")
	}
	return nil
}

// Invul handles //invul — toggles GM invulnerable mode.
//
// Java reference: AdminEffects.java
type Invul struct{}

func (c *Invul) Names() []string           { return []string{"invul", "invulnerable", "setinvul"} }
func (c *Invul) RequiredAccessLevel() int32 { return 1 }

func (c *Invul) Handle(player *model.Player, _ []string) error {
	current := player.IsInvulnerable()
	player.SetInvulnerable(!current)

	if player.IsInvulnerable() {
		player.SetLastAdminMessage("You are now invulnerable")
	} else {
		player.SetLastAdminMessage(fmt.Sprintf("Invulnerable mode %s", "disabled"))
	}
	return nil
}
