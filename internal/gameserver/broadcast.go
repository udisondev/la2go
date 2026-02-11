package gameserver

import (
	"fmt"
	"log/slog"

	"github.com/udisondev/la2go/internal/model"
	"github.com/udisondev/la2go/internal/protocol"
	"github.com/udisondev/la2go/internal/world"
)

// BroadcastToAll sends packet to all connected clients.
// WARNING: This is SLOW — sends to ALL clients regardless of visibility.
// Use BroadcastToVisible for gameplay packets (player movement, skill casts, etc).
// Parameters:
//   - packetBuf: buffer with PacketHeaderSize + payload + PacketBufferPadding
//   - payloadLen: length of payload (without header)
func (cm *ClientManager) BroadcastToAll(packetBuf []byte, payloadLen int) int {
	sent := 0

	cm.ForEachClient(func(client *GameClient) bool {
		// Skip clients not yet authenticated
		if client.State() < ClientStateAuthenticated {
			return true
		}

		// Send packet (each client has unique encryption key)
		if err := protocol.WritePacket(client.Conn(), client.Encryption(), packetBuf, payloadLen); err != nil {
			slog.Warn("failed to broadcast to client", "account", client.AccountName(), "error", err)
			return true
		}

		sent++
		return true
	})

	return sent
}

// BroadcastToVisible sends packet to all players who can see sourcePlayer (all LOD levels).
// FAST PATH — uses visibility cache (Phase 4.5 PR3) to filter clients.
// Only sends to visible players, dramatically reducing broadcast cost.
// Phase 4.5 PR4: -99.5% broadcast cost for typical scenarios.
// Phase 4.13: Uses LODAll for backward compatibility.
// For critical events (movement, combat), use BroadcastToVisibleNear for -89% packet reduction.
// Parameters:
//   - packetBuf: buffer with PacketHeaderSize + payload + PacketBufferPadding
//   - payloadLen: length of payload (without header)
func (cm *ClientManager) BroadcastToVisible(sourcePlayer *model.Player, packetBuf []byte, payloadLen int) int {
	return cm.BroadcastToVisibleByLOD(sourcePlayer, world.LODAll, packetBuf, payloadLen)
}

// BroadcastToVisibleByLOD sends packet to all players who can see sourcePlayer at given LOD level.
// Phase 4.13: LOD-aware broadcast for optimized packet filtering.
// LOD levels:
//   - LODNear: same region (~50 objects, -89% vs LODAll) — critical events (movement, combat)
//   - LODMedium: adjacent regions (~200 objects, -56% vs LODAll) — zone events
//   - LODFar: diagonal regions (~200 objects)
//   - LODAll: all visible objects (~450 objects) — global events, backward compat
// Parameters:
//   - packetBuf: buffer with PacketHeaderSize + payload + PacketBufferPadding
//   - payloadLen: length of payload (without header)
func (cm *ClientManager) BroadcastToVisibleByLOD(sourcePlayer *model.Player, lodLevel world.LODLevel, packetBuf []byte, payloadLen int) int {
	sent := 0

	// Phase 4.18 Optimization 1: Use reverse visibility map for O(M) instead of O(N×M)
	// Before: O(N×M) = 100K players × 100 objects = 10M operations
	// After: O(M) = 100 observer lookups
	// Expected: -99.999% improvement (100,000× faster)
	//
	// GetObservers returns ALL observers (near+medium+far), then we filter by LOD level
	// Alternative: Build 3 separate reverse caches (near/medium/far) — 3× memory overhead (540MB)
	// Decision: Filter approach is memory-efficient and still O(M) << O(N×M)

	// Phase 5.1: Check VisibilityManager nil (for unit tests without World setup)
	if cm.visibilityManager == nil {
		return 0 // Skip broadcast if VisibilityManager not initialized
	}

	observerIDs := cm.visibilityManager.GetObservers(sourcePlayer.ObjectID())
	if observerIDs == nil {
		// Reverse cache not initialized yet (first batch update pending)
		// Fallback to empty broadcast (acceptable during server startup)
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

		// Filter by LOD level: check if sourcePlayer is visible at requested LOD
		// This is still O(M) per observer, but M is small (~100 objects vs 100K players)
		canSeeAtLOD := false
		world.ForEachVisibleObjectByLOD(targetPlayer, lodLevel, func(obj *model.WorldObject) bool {
			if obj.ObjectID() == sourcePlayer.ObjectID() {
				canSeeAtLOD = true
				return false // stop iteration
			}
			return true
		})

		if !canSeeAtLOD {
			continue // Observer can see source, but not at requested LOD level
		}

		// Send packet to visible player
		if err := protocol.WritePacket(targetClient.Conn(), targetClient.Encryption(), packetBuf, payloadLen); err != nil {
			slog.Warn("failed to broadcast to visible player",
				"source", sourcePlayer.Name(),
				"target", targetPlayer.Name(),
				"error", err)
			continue
		}

		sent++
	}

	return sent
}

// BroadcastToVisibleNear sends packet to players in same region as sourcePlayer.
// Phase 4.13: Convenience wrapper for most critical events (movement, combat, spell cast).
// Expected packet reduction: -89% vs BroadcastToVisible (50 vs 450 objects).
// Trade-off: players in adjacent/diagonal regions won't see action immediately.
// This is ACCEPTABLE for L2 Interlude (not competitive FPS).
// Parameters:
//   - packetBuf: buffer with PacketHeaderSize + payload + PacketBufferPadding
//   - payloadLen: length of payload (without header)
func (cm *ClientManager) BroadcastToVisibleNear(sourcePlayer *model.Player, packetBuf []byte, payloadLen int) int {
	return cm.BroadcastToVisibleByLOD(sourcePlayer, world.LODNear, packetBuf, payloadLen)
}

// BroadcastToVisibleMedium sends packet to players in same or adjacent regions.
// Phase 4.13: For zone-level events (NPC spawn, skill AOE, etc).
// Expected packet reduction: -56% vs BroadcastToVisible (200 vs 450 objects).
// Parameters:
//   - packetBuf: buffer with PacketHeaderSize + payload + PacketBufferPadding
//   - payloadLen: length of payload (without header)
func (cm *ClientManager) BroadcastToVisibleMedium(sourcePlayer *model.Player, packetBuf []byte, payloadLen int) int {
	return cm.BroadcastToVisibleByLOD(sourcePlayer, world.LODMedium, packetBuf, payloadLen)
}

// BroadcastToVisibleExcept sends packet to all players who can see sourcePlayer, except excluded player.
// Useful for broadcasting sourcePlayer's actions to others (e.g., player movement, skill cast).
// Phase 4.13: Uses LODAll for backward compatibility.
// For critical events, use BroadcastToVisibleNearExcept for -89% packet reduction.
// Parameters:
//   - packetBuf: buffer with PacketHeaderSize + payload + PacketBufferPadding
//   - payloadLen: length of payload (without header)
func (cm *ClientManager) BroadcastToVisibleExcept(sourcePlayer *model.Player, excludePlayer *model.Player, packetBuf []byte, payloadLen int) int {
	return cm.BroadcastToVisibleByLODExcept(sourcePlayer, excludePlayer, world.LODAll, packetBuf, payloadLen)
}

// BroadcastToVisibleByLODExcept sends packet to all players who can see sourcePlayer at given LOD level, except excluded player.
// Phase 4.13: LOD-aware broadcast with exclusion for optimized packet filtering.
// Parameters:
//   - packetBuf: buffer with PacketHeaderSize + payload + PacketBufferPadding
//   - payloadLen: length of payload (without header)
func (cm *ClientManager) BroadcastToVisibleByLODExcept(sourcePlayer *model.Player, excludePlayer *model.Player, lodLevel world.LODLevel, packetBuf []byte, payloadLen int) int {
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

		// Send packet
		if err := protocol.WritePacket(targetClient.Conn(), targetClient.Encryption(), packetBuf, payloadLen); err != nil {
			slog.Warn("failed to broadcast to visible player",
				"source", sourcePlayer.Name(),
				"target", targetPlayer.Name(),
				"error", err)
			return true
		}

		sent++
		return true
	})

	return sent
}

// BroadcastToVisibleNearExcept sends packet to players in same region, except excluded player.
// Phase 4.13: Most common use case — broadcast sourcePlayer's movement to nearby players.
// Expected packet reduction: -89% vs BroadcastToVisibleExcept (50 vs 450 objects).
// Parameters:
//   - packetBuf: buffer with PacketHeaderSize + payload + PacketBufferPadding
//   - payloadLen: length of payload (without header)
func (cm *ClientManager) BroadcastToVisibleNearExcept(sourcePlayer *model.Player, excludePlayer *model.Player, packetBuf []byte, payloadLen int) int {
	return cm.BroadcastToVisibleByLODExcept(sourcePlayer, excludePlayer, world.LODNear, packetBuf, payloadLen)
}

// BroadcastFromPosition sends packet to all players who can see the given position.
// Used for NPC actions (attack, skill) where no Player source exists.
// Phase 5.7: NPC Aggro & Basic AI.
// Parameters:
//   - x, y: world coordinates of the NPC/source
//   - packetBuf: buffer with PacketHeaderSize + payload + PacketBufferPadding
//   - payloadLen: length of payload (without header)
func (cm *ClientManager) BroadcastFromPosition(x, y int32, packetBuf []byte, payloadLen int) int {
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

		// Send packet
		if err := protocol.WritePacket(client.Conn(), client.Encryption(), packetBuf, payloadLen); err != nil {
			slog.Warn("failed to broadcast from position",
				"x", x, "y", y,
				"target", obj.Name(),
				"error", err)
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
func (cm *ClientManager) SendToPlayer(objectID uint32, packetBuf []byte, payloadLen int) error {
	client := cm.GetClientByObjectID(objectID)
	if client == nil {
		return fmt.Errorf("player objectID=%d not found", objectID)
	}

	if client.State() != ClientStateInGame {
		return fmt.Errorf("player objectID=%d not in game", objectID)
	}

	if err := protocol.WritePacket(client.Conn(), client.Encryption(), packetBuf, payloadLen); err != nil {
		return fmt.Errorf("sending packet to player objectID=%d: %w", objectID, err)
	}

	return nil
}

// BroadcastToRegion sends packet to all players in given region.
// Useful for area-of-effect announcements (e.g., castle siege start, boss spawn).
// Parameters:
//   - packetBuf: buffer with PacketHeaderSize + payload + PacketBufferPadding
//   - payloadLen: length of payload (without header)
func (cm *ClientManager) BroadcastToRegion(regionX, regionY int32, packetBuf []byte, payloadLen int) int {
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

		// Send packet
		if err := protocol.WritePacket(client.Conn(), client.Encryption(), packetBuf, payloadLen); err != nil {
			slog.Warn("failed to broadcast to region",
				"region", [2]int32{regionX, regionY},
				"player", player.Name(),
				"error", err)
			return true
		}

		sent++
		return true
	})

	return sent
}
