package gameserver

import (
	"fmt"
	"log/slog"

	"github.com/udisondev/la2go/internal/model"
	"github.com/udisondev/la2go/internal/world"
)

// BroadcastToAll sends packet to all connected clients.
// WARNING: This is SLOW — sends to ALL clients regardless of visibility.
// Use BroadcastToVisible for gameplay packets (player movement, skill casts, etc).
// Parameters:
//   - payload: raw packet data (opcode + fields)
//   - payloadLen: length of payload
func (cm *ClientManager) BroadcastToAll(payload []byte, payloadLen int) int {
	sent := 0

	cm.ForEachClient(func(client *GameClient) bool {
		// Skip clients not yet authenticated
		if client.State() < ClientStateAuthenticated {
			return true
		}

		if cm.writePool == nil {
			slog.Warn("writePool not initialized for broadcast")
			return true
		}

		// Pool buffer: encrypt payload → pool buf → sendCh → writePump → pool.Put
		encPkt, err := cm.writePool.EncryptToPooled(client.Encryption(), payload[:payloadLen], payloadLen)
		if err != nil {
			slog.Warn("failed to encrypt broadcast", "account", client.AccountName(), "error", err)
			return true
		}
		// Send takes ownership of encPkt (pool buffer)
		if err := client.Send(encPkt); err != nil {
			return true // slow client disconnected, buffer already returned to pool
		}

		sent++
		return true
	})

	return sent
}

// BroadcastToVisible sends packet to all players who can see sourcePlayer (all LOD levels).
// FAST PATH — uses visibility cache (Phase 4.5 PR3) to filter clients.
// Only sends to visible players, dramatically reducing broadcast cost.
// Parameters:
//   - payload: raw packet data (opcode + fields)
//   - payloadLen: length of payload
func (cm *ClientManager) BroadcastToVisible(sourcePlayer *model.Player, payload []byte, payloadLen int) int {
	return cm.BroadcastToVisibleByLOD(sourcePlayer, world.LODAll, payload, payloadLen)
}

// BroadcastToVisibleByLOD sends packet to all players who can see sourcePlayer at given LOD level.
// Phase 4.13: LOD-aware broadcast for optimized packet filtering.
// Parameters:
//   - payload: raw packet data (opcode + fields)
//   - payloadLen: length of payload
func (cm *ClientManager) BroadcastToVisibleByLOD(sourcePlayer *model.Player, lodLevel world.LODLevel, payload []byte, payloadLen int) int {
	sent := 0

	// Phase 5.1: Check VisibilityManager nil (for unit tests without World setup)
	if cm.visibilityManager == nil {
		return 0 // Skip broadcast if VisibilityManager not initialized
	}

	observerIDs := cm.visibilityManager.GetObservers(sourcePlayer.ObjectID())
	if observerIDs == nil {
		return 0
	}

	// Iterate only through observers (M=~100 players, not N=100K)
	for _, playerID := range observerIDs {
		targetClient := cm.GetClientByObjectID(playerID)
		if targetClient == nil {
			continue // Player offline
		}

		// Skip if client not in game
		if targetClient.State() != ClientStateInGame {
			continue
		}

		targetPlayer := targetClient.ActivePlayer()
		if targetPlayer == nil {
			continue // Player not loaded yet
		}

		// Skip if target is source
		if targetPlayer.ObjectID() == sourcePlayer.ObjectID() {
			continue
		}

		// Filter by LOD level
		canSeeAtLOD := false
		world.ForEachVisibleObjectByLOD(targetPlayer, lodLevel, func(obj *model.WorldObject) bool {
			if obj.ObjectID() == sourcePlayer.ObjectID() {
				canSeeAtLOD = true
				return false // stop iteration
			}
			return true
		})

		if !canSeeAtLOD {
			continue
		}

		if cm.writePool == nil {
			continue
		}

		// Pool buffer: encrypt payload → pool buf → sendCh → writePump → pool.Put
		encPkt, err := cm.writePool.EncryptToPooled(targetClient.Encryption(), payload[:payloadLen], payloadLen)
		if err != nil {
			slog.Warn("failed to encrypt broadcast to visible player",
				"source", sourcePlayer.Name(),
				"target", targetPlayer.Name(),
				"error", err)
			continue
		}

		if err := targetClient.Send(encPkt); err != nil {
			continue
		}

		sent++
	}

	return sent
}

// BroadcastToVisibleNear sends packet to players in same region as sourcePlayer.
// Phase 4.13: Convenience wrapper for most critical events (movement, combat, spell cast).
// Parameters:
//   - payload: raw packet data (opcode + fields)
//   - payloadLen: length of payload
func (cm *ClientManager) BroadcastToVisibleNear(sourcePlayer *model.Player, payload []byte, payloadLen int) int {
	return cm.BroadcastToVisibleByLOD(sourcePlayer, world.LODNear, payload, payloadLen)
}

// BroadcastToVisibleMedium sends packet to players in same or adjacent regions.
// Phase 4.13: For zone-level events (NPC spawn, skill AOE, etc).
// Parameters:
//   - payload: raw packet data (opcode + fields)
//   - payloadLen: length of payload
func (cm *ClientManager) BroadcastToVisibleMedium(sourcePlayer *model.Player, payload []byte, payloadLen int) int {
	return cm.BroadcastToVisibleByLOD(sourcePlayer, world.LODMedium, payload, payloadLen)
}

// BroadcastToVisibleExcept sends packet to all players who can see sourcePlayer, except excluded player.
// Parameters:
//   - payload: raw packet data (opcode + fields)
//   - payloadLen: length of payload
func (cm *ClientManager) BroadcastToVisibleExcept(sourcePlayer *model.Player, excludePlayer *model.Player, payload []byte, payloadLen int) int {
	return cm.BroadcastToVisibleByLODExcept(sourcePlayer, excludePlayer, world.LODAll, payload, payloadLen)
}

// BroadcastToVisibleByLODExcept sends packet to all players who can see sourcePlayer at given LOD level, except excluded player.
// Parameters:
//   - payload: raw packet data (opcode + fields)
//   - payloadLen: length of payload
func (cm *ClientManager) BroadcastToVisibleByLODExcept(sourcePlayer *model.Player, excludePlayer *model.Player, lodLevel world.LODLevel, payload []byte, payloadLen int) int {
	sent := 0

	cm.ForEachPlayer(func(targetPlayer *model.Player, targetClient *GameClient) bool {
		// Skip if target is excluded
		if targetPlayer == excludePlayer {
			return true
		}

		// Skip if client not in game
		if targetClient.State() != ClientStateInGame {
			return true
		}

		// Check if targetPlayer can see sourcePlayer at given LOD level
		canSee := false
		world.ForEachVisibleObjectByLOD(targetPlayer, lodLevel, func(obj *model.WorldObject) bool {
			if obj.ObjectID() == sourcePlayer.ObjectID() {
				canSee = true
				return false
			}
			return true
		})

		if !canSee {
			return true
		}

		if cm.writePool == nil {
			return true
		}

		encPkt, err := cm.writePool.EncryptToPooled(targetClient.Encryption(), payload[:payloadLen], payloadLen)
		if err != nil {
			slog.Warn("failed to encrypt broadcast to visible player",
				"source", sourcePlayer.Name(),
				"target", targetPlayer.Name(),
				"error", err)
			return true
		}

		if err := targetClient.Send(encPkt); err != nil {
			return true
		}

		sent++
		return true
	})

	return sent
}

// BroadcastToVisibleNearExcept sends packet to players in same region, except excluded player.
// Parameters:
//   - payload: raw packet data (opcode + fields)
//   - payloadLen: length of payload
func (cm *ClientManager) BroadcastToVisibleNearExcept(sourcePlayer *model.Player, excludePlayer *model.Player, payload []byte, payloadLen int) int {
	return cm.BroadcastToVisibleByLODExcept(sourcePlayer, excludePlayer, world.LODNear, payload, payloadLen)
}

// BroadcastFromPosition sends packet to all players who can see the given position.
// Used for NPC actions (attack, skill) where no Player source exists.
// Parameters:
//   - x, y: world coordinates of the NPC/source
//   - payload: raw packet data (opcode + fields)
//   - payloadLen: length of payload
func (cm *ClientManager) BroadcastFromPosition(x, y int32, payload []byte, payloadLen int) int {
	sent := 0

	// Iterate visible objects around position to find players
	world.ForEachVisibleObject(world.Instance(), x, y, func(obj *model.WorldObject) bool {
		// Check if this object is a Player (via objectID index)
		client := cm.GetClientByObjectID(obj.ObjectID())
		if client == nil {
			return true // not a player or offline
		}

		if client.State() != ClientStateInGame {
			return true
		}

		if cm.writePool == nil {
			return true
		}

		encPkt, err := cm.writePool.EncryptToPooled(client.Encryption(), payload[:payloadLen], payloadLen)
		if err != nil {
			slog.Warn("failed to encrypt broadcast from position",
				"x", x, "y", y,
				"target", obj.Name(),
				"error", err)
			return true
		}

		if err := client.Send(encPkt); err != nil {
			return true
		}

		sent++
		return true
	})

	return sent
}

// SendToPlayer sends a packet directly to a specific player by objectID.
// Used for personal messages (SystemMessage, exp gain, level-up).
// Returns error if player not found or send fails.
func (cm *ClientManager) SendToPlayer(objectID uint32, payload []byte, payloadLen int) error {
	client := cm.GetClientByObjectID(objectID)
	if client == nil {
		return fmt.Errorf("player objectID=%d not found", objectID)
	}

	if client.State() != ClientStateInGame {
		return fmt.Errorf("player objectID=%d not in game", objectID)
	}

	if cm.writePool == nil {
		return fmt.Errorf("writePool not initialized")
	}

	encPkt, err := cm.writePool.EncryptToPooled(client.Encryption(), payload[:payloadLen], payloadLen)
	if err != nil {
		return fmt.Errorf("encrypting packet for player objectID=%d: %w", objectID, err)
	}

	if err := client.Send(encPkt); err != nil {
		return fmt.Errorf("sending packet to player objectID=%d: %w", objectID, err)
	}

	return nil
}

// BroadcastToRegion sends packet to all players in given region.
// Parameters:
//   - payload: raw packet data (opcode + fields)
//   - payloadLen: length of payload
func (cm *ClientManager) BroadcastToRegion(regionX, regionY int32, payload []byte, payloadLen int) int {
	sent := 0

	cm.ForEachPlayer(func(player *model.Player, client *GameClient) bool {
		// Skip if client not in game
		if client.State() != ClientStateInGame {
			return true
		}

		// Check if player is in target region
		loc := player.Location()
		playerRegionX, playerRegionY := world.CoordToRegionIndex(loc.X, loc.Y)

		if playerRegionX != regionX || playerRegionY != regionY {
			return true
		}

		if cm.writePool == nil {
			return true
		}

		encPkt, err := cm.writePool.EncryptToPooled(client.Encryption(), payload[:payloadLen], payloadLen)
		if err != nil {
			slog.Warn("failed to encrypt broadcast to region",
				"region", [2]int32{regionX, regionY},
				"player", player.Name(),
				"error", err)
			return true
		}

		if err := client.Send(encPkt); err != nil {
			return true
		}

		sent++
		return true
	})

	return sent
}
