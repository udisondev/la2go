package gameserver

import (
	"sync"

	"github.com/udisondev/la2go/internal/model"
)

// ClientManager manages all connected game clients.
// Provides registration, lookup, and broadcast functionality.
// Thread-safe for concurrent access.
type ClientManager struct {
	mu      sync.RWMutex
	clients map[string]*GameClient // key: accountName

	// playerClients maps Player to GameClient for efficient broadcast
	// Updated when player enters/leaves world
	playerClients map[*model.Player]*GameClient
}

// NewClientManager creates a new client manager.
func NewClientManager() *ClientManager {
	return &ClientManager{
		clients:       make(map[string]*GameClient, 1000),       // pre-allocate for 1K players
		playerClients: make(map[*model.Player]*GameClient, 1000), // pre-allocate for 1K players
	}
}

// Register adds a client to the manager.
// Called when client completes authentication (after AuthLogin).
func (cm *ClientManager) Register(accountName string, client *GameClient) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.clients[accountName] = client
}

// Unregister removes a client from the manager.
// Called when client disconnects or logs out.
func (cm *ClientManager) Unregister(accountName string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Remove from clients map
	delete(cm.clients, accountName)

	// Remove from playerClients map (find and delete)
	for player, client := range cm.playerClients {
		if client.AccountName() == accountName {
			delete(cm.playerClients, player)
			break
		}
	}
}

// RegisterPlayer associates a Player with a GameClient.
// Called when player enters world (after character selection).
func (cm *ClientManager) RegisterPlayer(player *model.Player, client *GameClient) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.playerClients[player] = client
}

// UnregisterPlayer removes Player→Client association.
// Called when player leaves world or logs out.
func (cm *ClientManager) UnregisterPlayer(player *model.Player) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	delete(cm.playerClients, player)
}

// GetClient returns the client for given account name.
// Returns nil if not found.
func (cm *ClientManager) GetClient(accountName string) *GameClient {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.clients[accountName]
}

// GetClientByPlayer returns the client for given player.
// Returns nil if not found.
func (cm *ClientManager) GetClientByPlayer(player *model.Player) *GameClient {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.playerClients[player]
}

// GetClientByObjectID returns the client for given object ID.
// Iterates playerClients to find matching player.ObjectID().
// Returns nil if not found or object is not a player.
// Phase 4.9: Used to resolve WorldObject → Player → GameClient.
func (cm *ClientManager) GetClientByObjectID(objectID uint32) *GameClient {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	for player, client := range cm.playerClients {
		if uint32(player.CharacterID()) == objectID {
			return client
		}
	}
	return nil
}

// Count returns total number of connected clients.
func (cm *ClientManager) Count() int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return len(cm.clients)
}

// PlayerCount returns number of players in world.
func (cm *ClientManager) PlayerCount() int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return len(cm.playerClients)
}

// ForEachClient iterates over all connected clients.
// fn receives GameClient pointer. If fn returns false, iteration stops.
func (cm *ClientManager) ForEachClient(fn func(*GameClient) bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	for _, client := range cm.clients {
		if !fn(client) {
			return
		}
	}
}

// ForEachPlayer iterates over all players in world.
// fn receives Player and GameClient pointers. If fn returns false, iteration stops.
func (cm *ClientManager) ForEachPlayer(fn func(*model.Player, *GameClient) bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	for player, client := range cm.playerClients {
		if !fn(player, client) {
			return
		}
	}
}
