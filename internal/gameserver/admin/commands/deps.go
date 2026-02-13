package commands

import "github.com/udisondev/la2go/internal/model"

// ClientManager provides player lookup for admin commands.
// Interface to avoid import cycle with gameserver package.
type ClientManager interface {
	// FindPlayerByName finds an online player by name (case-insensitive).
	FindPlayerByName(name string) *model.Player
	// ForEachPlayer iterates over all online players.
	ForEachPlayer(fn func(*model.Player) bool)
	// PlayerCount returns number of online players.
	PlayerCount() int
	// KickPlayer disconnects a player by name. Returns true if found.
	KickPlayer(name string) bool
}
