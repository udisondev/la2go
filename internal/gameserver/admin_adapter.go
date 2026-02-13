package gameserver

import (
	"strings"

	"github.com/udisondev/la2go/internal/model"
)

// AdminClientAdapter adapts ClientManager to the commands.ClientManager interface.
// This avoids import cycle between gameserver ↔ admin/commands packages.
// Phase 17: Admin Commands.
type AdminClientAdapter struct {
	cm *ClientManager
}

// NewAdminClientAdapter creates a new adapter.
func NewAdminClientAdapter(cm *ClientManager) *AdminClientAdapter {
	return &AdminClientAdapter{cm: cm}
}

// FindPlayerByName finds an online player by name (case-insensitive).
func (a *AdminClientAdapter) FindPlayerByName(name string) *model.Player {
	nameLower := strings.ToLower(name)
	var found *model.Player

	a.cm.ForEachPlayer(func(player *model.Player, _ *GameClient) bool {
		if strings.ToLower(player.Name()) == nameLower {
			found = player
			return false // stop iteration
		}
		return true
	})

	return found
}

// ForEachPlayer iterates over all online players.
func (a *AdminClientAdapter) ForEachPlayer(fn func(*model.Player) bool) {
	a.cm.ForEachPlayer(func(player *model.Player, _ *GameClient) bool {
		return fn(player)
	})
}

// PlayerCount returns number of online players.
func (a *AdminClientAdapter) PlayerCount() int {
	return a.cm.PlayerCount()
}

// KickPlayer disconnects a player by name. Returns true if found.
func (a *AdminClientAdapter) KickPlayer(name string) bool {
	client := a.cm.FindClientByPlayerName(name)
	if client == nil {
		return false
	}
	client.Close()
	return true
}
