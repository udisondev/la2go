package commands

import (
	"fmt"
	"time"

	"github.com/udisondev/la2go/internal/model"
)

// Loc handles /loc — shows current coordinates.
//
// Java reference: Loc.java (voiced command handler)
type Loc struct{}

func (c *Loc) Names() []string { return []string{"loc", "location"} }

func (c *Loc) Handle(player *model.Player, _ string) error {
	loc := player.Location()
	player.SetLastAdminMessage(fmt.Sprintf("Location: X=%d Y=%d Z=%d Heading=%d",
		loc.X, loc.Y, loc.Z, loc.Heading))
	return nil
}

// GameTime handles /time — shows server game time.
//
// Java reference: Time.java (voiced command handler)
type GameTime struct{}

func (c *GameTime) Names() []string { return []string{"time", "gametime"} }

func (c *GameTime) Handle(player *model.Player, _ string) error {
	// L2 game time: 1 real second = 6 game seconds (10 min real = 1 hour in-game)
	// Day starts at 06:00, night starts at 00:00
	now := time.Now()
	gameSeconds := now.Unix() * 6
	gameMinutes := (gameSeconds / 60) % 60
	gameHours := (gameSeconds / 3600) % 24

	dayNight := "Day"
	if gameHours < 6 || gameHours >= 22 {
		dayNight = "Night"
	}

	player.SetLastAdminMessage(fmt.Sprintf("Game time: %02d:%02d (%s) | Server time: %s",
		gameHours, gameMinutes, dayNight, now.Format("15:04:05")))
	return nil
}

// Unstuck handles /unstuck — teleport to nearest town (5 min cast).
// MVP: instant teleport to Talking Island (will be improved later).
//
// Java reference: Unstuck.java (voiced command handler)
type Unstuck struct{}

// Default unstuck location: Talking Island Village
const (
	unstuckX = int32(-83968)
	unstuckY = int32(244634)
	unstuckZ = int32(-3730)
)

func (c *Unstuck) Names() []string { return []string{"unstuck"} }

func (c *Unstuck) Handle(player *model.Player, _ string) error {
	if player.IsDead() {
		return fmt.Errorf("cannot use /unstuck while dead")
	}

	if player.HasAttackStance() {
		return fmt.Errorf("cannot use /unstuck during combat")
	}

	// MVP: instant teleport. Phase 17+: add 5 min cast time.
	player.SetLocation(model.NewLocation(unstuckX, unstuckY, unstuckZ, 0))
	player.SetLastAdminMessage("Teleported to Talking Island Village")
	return nil
}

// Online handles /online — shows number of online players.
type Online struct {
	clientMgr ClientManager
}

// NewOnline creates the online command handler.
func NewOnline(clientMgr ClientManager) *Online {
	return &Online{clientMgr: clientMgr}
}

func (c *Online) Names() []string { return []string{"online", "players"} }

func (c *Online) Handle(player *model.Player, _ string) error {
	count := c.clientMgr.PlayerCount()
	player.SetLastAdminMessage(fmt.Sprintf("Online: %d players", count))
	return nil
}
