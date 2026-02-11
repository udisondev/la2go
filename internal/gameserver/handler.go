package gameserver

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/udisondev/la2go/internal/constants"
	"github.com/udisondev/la2go/internal/game/combat"
	"github.com/udisondev/la2go/internal/gameserver/clientpackets"
	"github.com/udisondev/la2go/internal/gameserver/serverpackets"
	"github.com/udisondev/la2go/internal/login"
	"github.com/udisondev/la2go/internal/model"
	"github.com/udisondev/la2go/internal/protocol"
	"github.com/udisondev/la2go/internal/world"
)

// Handler processes game client packets.
type Handler struct {
	sessionManager *login.SessionManager
	clientManager  *ClientManager // Phase 4.5 PR4: register clients after auth
	charRepo       CharacterRepository // Phase 4.6: load characters for CharSelectionInfo
}

// CharacterRepository defines interface for loading characters from database.
// Used for dependency injection to keep handler testable.
type CharacterRepository interface {
	LoadByAccountName(ctx context.Context, accountName string) ([]*model.Player, error)
}

// NewHandler creates a new packet handler for game clients.
func NewHandler(sessionManager *login.SessionManager, clientManager *ClientManager, charRepo CharacterRepository) *Handler {
	return &Handler{
		sessionManager: sessionManager,
		clientManager:  clientManager,
		charRepo:       charRepo,
	}
}

// HandlePacket dispatches a decrypted packet to the appropriate handler.
// Writes response into buf. Returns: n — bytes written to buf (0 = nothing to send),
// ok — true if connection stays open (false = close after sending).
func (h *Handler) HandlePacket(
	ctx context.Context,
	client *GameClient,
	data, buf []byte,
) (int, bool, error) {
	if len(data) == 0 {
		return 0, false, fmt.Errorf("empty packet data")
	}

	opcode := data[0]
	body := data[1:]
	state := client.State()

	switch state {
	case ClientStateConnected:
		switch opcode {
		case clientpackets.OpcodeProtocolVersion:
			return handleProtocolVersion(client, body)
		default:
			slog.Warn("invalid opcode for state CONNECTED",
				"opcode", fmt.Sprintf("0x%02X", opcode),
				"client", client.IP())
			return 0, false, nil
		}

	case ClientStateAuthenticated, ClientStateEntering, ClientStateInGame:
		switch opcode {
		case clientpackets.OpcodeAuthLogin:
			return h.handleAuthLogin(ctx, client, body, buf)
		case clientpackets.OpcodeCharacterSelect:
			return h.handleCharacterSelect(ctx, client, body, buf)
		case clientpackets.OpcodeEnterWorld:
			return h.handleEnterWorld(ctx, client, body, buf)
		case clientpackets.OpcodeMoveToLocation:
			return h.handleMoveToLocation(ctx, client, body, buf)
		case clientpackets.OpcodeValidatePosition:
			return h.handleValidatePosition(ctx, client, body, buf)
		case clientpackets.OpcodeRequestAction:
			return h.handleRequestAction(ctx, client, body, buf)
		case clientpackets.OpcodeAttackRequest:
			return h.handleAttackRequest(ctx, client, body, buf)
		case clientpackets.OpcodeRequestPickup:
			return h.handleRequestPickup(ctx, client, body, buf)
		case clientpackets.OpcodeLogout:
			return h.handleLogout(ctx, client, body, buf)
		case clientpackets.OpcodeRequestRestart:
			return h.handleRequestRestart(ctx, client, body, buf)
		default:
			slog.Warn("unknown packet opcode",
				"opcode", fmt.Sprintf("0x%02X", opcode),
				"state", state,
				"client", client.IP())
			return 0, true, nil
		}

	default:
		return 0, false, fmt.Errorf("invalid state: %v", state)
	}
}

// handleProtocolVersion processes the ProtocolVersion packet (opcode 0x0E).
func handleProtocolVersion(client *GameClient, data []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseProtocolVersion(data)
	if err != nil {
		return 0, false, fmt.Errorf("parsing ProtocolVersion: %w", err)
	}

	if !pkt.IsValid() {
		slog.Warn("invalid protocol version",
			"expected", 0x0106,
			"got", pkt.ProtocolRevision,
			"client", client.IP())
		return 0, false, fmt.Errorf("invalid protocol revision: 0x%04X", pkt.ProtocolRevision)
	}

	slog.Debug("protocol version validated", "client", client.IP())

	// Protocol version is valid, wait for AuthLogin
	// No response packet
	return 0, true, nil
}

// handleAuthLogin processes the AuthLogin packet (opcode 0x08).
func (h *Handler) handleAuthLogin(ctx context.Context, client *GameClient, data, buf []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseAuthLogin(data)
	if err != nil {
		return 0, false, fmt.Errorf("parsing AuthLogin: %w", err)
	}

	// Validate SessionKey with SessionManager (shared with LoginServer)
	// showLicence=false because GameServer doesn't care about license state
	if !h.sessionManager.Validate(pkt.AccountName, pkt.SessionKey, false) {
		slog.Warn("session key validation failed",
			"account", pkt.AccountName,
			"client", client.IP())
		// TODO: Send AuthLoginFail packet
		return 0, false, fmt.Errorf("invalid session key for account %s", pkt.AccountName)
	}

	// SessionKey is valid, set client state
	client.SetAccountName(pkt.AccountName)
	client.SetSessionKey(&pkt.SessionKey)
	client.SetState(ClientStateAuthenticated)

	// Register client in ClientManager (Phase 4.5 PR4)
	h.clientManager.Register(pkt.AccountName, client)

	slog.Info("client authenticated",
		"account", pkt.AccountName,
		"client", client.IP())

	// Load characters for this account (Phase 4.6)
	// Phase 4.18: Use cached loader to eliminate redundant DB queries
	players, err := client.GetCharacters(pkt.AccountName, func(name string) ([]*model.Player, error) {
		return h.charRepo.LoadByAccountName(ctx, name)
	})
	if err != nil {
		return 0, false, fmt.Errorf("loading characters for account %s: %w", pkt.AccountName, err)
	}

	// Create and send CharSelectionInfo packet
	// SessionID is derived from SessionKey (use PlayOkID1)
	sessionID := pkt.SessionKey.PlayOkID1
	charSelInfo := serverpackets.NewCharSelectionInfoFromPlayers(pkt.AccountName, sessionID, players)

	packetData, err := charSelInfo.Write()
	if err != nil {
		return 0, false, fmt.Errorf("serializing CharSelectionInfo: %w", err)
	}

	// Copy packet data to response buffer
	n := copy(buf, packetData)
	if n != len(packetData) {
		return 0, false, fmt.Errorf("buffer too small: need %d bytes, have %d", len(packetData), len(buf))
	}

	slog.Debug("sent CharSelectionInfo",
		"account", pkt.AccountName,
		"character_count", len(players),
		"packet_size", n)

	return n, true, nil
}

// handleCharacterSelect processes the CharacterSelect packet (opcode 0x0D).
// Client sends this when user selects a character from the character list.
// Response: CharSelected packet with character data.
func (h *Handler) handleCharacterSelect(ctx context.Context, client *GameClient, data, buf []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseCharacterSelect(data)
	if err != nil {
		return 0, false, fmt.Errorf("parsing CharacterSelect: %w", err)
	}

	// Validate character slot (0-7)
	if pkt.CharSlot < 0 || pkt.CharSlot > 7 {
		slog.Warn("invalid character slot",
			"slot", pkt.CharSlot,
			"account", client.AccountName(),
			"client", client.IP())
		return 0, false, fmt.Errorf("invalid character slot: %d", pkt.CharSlot)
	}

	// Load characters for this account
	// Phase 4.18: Use cached loader (2nd call — cache hit expected)
	accountName := client.AccountName()
	players, err := client.GetCharacters(accountName, func(name string) ([]*model.Player, error) {
		return h.charRepo.LoadByAccountName(ctx, name)
	})
	if err != nil {
		return 0, false, fmt.Errorf("loading characters for account %s: %w", accountName, err)
	}

	// Validate slot index
	if int(pkt.CharSlot) >= len(players) {
		slog.Warn("character slot out of range",
			"slot", pkt.CharSlot,
			"character_count", len(players),
			"account", client.AccountName(),
			"client", client.IP())
		return 0, false, fmt.Errorf("character slot %d out of range (have %d characters)", pkt.CharSlot, len(players))
	}

	// Get selected character
	player := players[pkt.CharSlot]

	// Get PlayOkID1 from SessionKey for CharSelected packet
	sessionKey := client.SessionKey()
	if sessionKey == nil {
		slog.Error("no session key for authenticated client",
			"account", client.AccountName(),
			"client", client.IP())
		return 0, false, fmt.Errorf("missing session key")
	}

	// Store selected character slot
	client.SetSelectedCharacter(pkt.CharSlot)

	// Send CharSelected packet (Phase 4.17.1)
	charSelected := serverpackets.NewCharSelected(player, sessionKey.PlayOkID1)
	charSelectedData, err := charSelected.Write()
	if err != nil {
		return 0, false, fmt.Errorf("serializing CharSelected: %w", err)
	}

	n := copy(buf, charSelectedData)
	if n != len(charSelectedData) {
		return 0, false, fmt.Errorf("buffer too small for CharSelected")
	}

	// Transition to ENTERING state (Phase 4.17.2)
	client.SetState(ClientStateEntering)

	slog.Info("character selected",
		"account", client.AccountName(),
		"character", player.Name(),
		"slot", pkt.CharSlot,
		"level", player.Level(),
		"client", client.IP())

	return n, true, nil
}

// handleEnterWorld processes the EnterWorld packet (opcode 0x03).
// Client sends this after CharacterSelect to spawn in the world.
func (h *Handler) handleEnterWorld(ctx context.Context, client *GameClient, data, buf []byte) (int, bool, error) {
	_, err := clientpackets.ParseEnterWorld(data)
	if err != nil {
		return 0, false, fmt.Errorf("parsing EnterWorld: %w", err)
	}

	// Verify character was selected
	charSlot := client.SelectedCharacter()
	if charSlot < 0 {
		slog.Warn("EnterWorld without character selection",
			"account", client.AccountName(),
			"client", client.IP())
		return 0, false, fmt.Errorf("no character selected")
	}

	// Load characters for this account
	// Phase 4.18: Use cached loader (3rd call — cache hit expected)
	accountName := client.AccountName()
	players, err := client.GetCharacters(accountName, func(name string) ([]*model.Player, error) {
		return h.charRepo.LoadByAccountName(ctx, name)
	})
	if err != nil {
		return 0, false, fmt.Errorf("loading characters for account %s: %w", accountName, err)
	}

	// Validate slot index
	if int(charSlot) >= len(players) {
		slog.Warn("character slot out of range",
			"slot", charSlot,
			"character_count", len(players),
			"account", client.AccountName(),
			"client", client.IP())
		return 0, false, fmt.Errorf("character slot %d out of range (have %d characters)", charSlot, len(players))
	}

	// Get selected character
	player := players[charSlot]

	// Cache player in GameClient (Phase 4.8 part 2)
	client.SetActivePlayer(player)

	// Register player in World Grid (Phase 4.9)
	if err := world.Instance().AddObject(player.WorldObject); err != nil {
		return 0, false, fmt.Errorf("adding player to world: %w", err)
	}

	// Update client state
	client.SetState(ClientStateInGame)

	slog.Info("player entering world",
		"account", client.AccountName(),
		"character", player.Name(),
		"level", player.Level(),
		"client", client.IP())

	// Send multiple packets after EnterWorld (Phase 4.7)
	// Order is important: UserInfo must be first, then StatusUpdate, then others
	var totalBytes int

	// 1. UserInfo (spawns character in world)
	userInfo := serverpackets.NewUserInfo(player)
	userInfoData, err := userInfo.Write()
	if err != nil {
		return 0, false, fmt.Errorf("serializing UserInfo: %w", err)
	}
	n := copy(buf[totalBytes:], userInfoData)
	if n != len(userInfoData) {
		return 0, false, fmt.Errorf("buffer too small for UserInfo")
	}
	totalBytes += n

	// 2. StatusUpdate (HP/MP/CP bars)
	statusUpdate := serverpackets.NewStatusUpdate(player)
	statusData, err := statusUpdate.Write()
	if err != nil {
		return 0, false, fmt.Errorf("serializing StatusUpdate: %w", err)
	}
	n = copy(buf[totalBytes:], statusData)
	if n != len(statusData) {
		return 0, false, fmt.Errorf("buffer too small for StatusUpdate")
	}
	totalBytes += n

	// 3. InventoryItemList (empty for now)
	invList := serverpackets.NewInventoryItemList()
	invData, err := invList.Write()
	if err != nil {
		return 0, false, fmt.Errorf("serializing InventoryItemList: %w", err)
	}
	n = copy(buf[totalBytes:], invData)
	if n != len(invData) {
		return 0, false, fmt.Errorf("buffer too small for InventoryItemList")
	}
	totalBytes += n

	// 4. ShortCutInit (empty for now)
	shortcuts := serverpackets.NewShortCutInit()
	shortcutData, err := shortcuts.Write()
	if err != nil {
		return 0, false, fmt.Errorf("serializing ShortCutInit: %w", err)
	}
	n = copy(buf[totalBytes:], shortcutData)
	if n != len(shortcutData) {
		return 0, false, fmt.Errorf("buffer too small for ShortCutInit")
	}
	totalBytes += n

	// 5. SkillList (empty for now)
	skills := serverpackets.NewSkillList()
	skillData, err := skills.Write()
	if err != nil {
		return 0, false, fmt.Errorf("serializing SkillList: %w", err)
	}
	n = copy(buf[totalBytes:], skillData)
	if n != len(skillData) {
		return 0, false, fmt.Errorf("buffer too small for SkillList")
	}
	totalBytes += n

	// 6. QuestList (empty for now)
	quests := serverpackets.NewQuestList()
	questData, err := quests.Write()
	if err != nil {
		return 0, false, fmt.Errorf("serializing QuestList: %w", err)
	}
	n = copy(buf[totalBytes:], questData)
	if n != len(questData) {
		return 0, false, fmt.Errorf("buffer too small for QuestList")
	}
	totalBytes += n

	slog.Debug("sent spawn packets",
		"character", player.Name(),
		"total_bytes", totalBytes,
		"packets", "UserInfo+StatusUpdate+Inventory+Shortcuts+Skills+Quests")

	// Broadcast CharInfo to visible players (Phase 4.8 part 2)
	// This makes the spawned player visible to others
	charInfo := serverpackets.NewCharInfo(player)
	charInfoData, err := charInfo.Write()
	if err != nil {
		slog.Error("failed to serialize CharInfo",
			"character", player.Name(),
			"error", err)
		// Continue даже если broadcast failed (player still spawns)
	} else {
		// Broadcast to all visible players
		visibleCount := h.clientManager.BroadcastToVisible(player, charInfoData, len(charInfoData))
		if visibleCount > 0 {
			slog.Debug("broadcasted CharInfo",
				"character", player.Name(),
				"visible_players", visibleCount)
		}
	}

	// Send CharInfo + NpcInfo TO client for all visible objects (Phase 4.9 Part 2 + Phase 4.10)
	// This makes other players and NPCs visible to the spawned player
	if err := h.sendVisibleObjectsInfo(client, player); err != nil {
		slog.Error("failed to send info for visible objects",
			"character", player.Name(),
			"error", err)
		// Continue даже если некоторые packets failed
	}

	// TODO Phase 4.10: Add more packets:
	// - NpcInfo (for visible NPCs)
	// - ItemOnGround (for visible drops)

	return totalBytes, true, nil
}

// handleMoveToLocation processes the MoveToLocation packet (opcode 0x01).
// Client sends this when player clicks on ground to move.
func (h *Handler) handleMoveToLocation(ctx context.Context, client *GameClient, data, buf []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseMoveToLocation(data)
	if err != nil {
		return 0, false, fmt.Errorf("parsing MoveToLocation: %w", err)
	}

	// Verify character is in game
	if client.State() != ClientStateInGame {
		slog.Warn("MoveToLocation before entering world",
			"account", client.AccountName(),
			"client", client.IP())
		return 0, true, nil // Ignore silently
	}

	// Get cached player (Phase 4.18 Opt 3)
	player := client.ActivePlayer()
	if player == nil {
		slog.Warn("MoveToLocation without active player",
			"account", client.AccountName(),
			"client", client.IP())
		return 0, false, fmt.Errorf("no active player for account %s", client.AccountName())
	}

	// Phase 5.1: Validate movement (distance, Z-bounds)
	if err := ValidateMoveToLocation(player, pkt.TargetX, pkt.TargetY, pkt.TargetZ); err != nil {
		slog.Warn("movement validation failed",
			"character", player.Name(),
			"from", fmt.Sprintf("(%d,%d,%d)", pkt.OriginX, pkt.OriginY, pkt.OriginZ),
			"to", fmt.Sprintf("(%d,%d,%d)", pkt.TargetX, pkt.TargetY, pkt.TargetZ),
			"error", err)

		// Send ValidateLocation (force client to use server position)
		validateLoc := serverpackets.NewValidateLocation(player)
		validateData, err := validateLoc.Write()
		if err != nil {
			slog.Error("failed to serialize ValidateLocation",
				"character", player.Name(),
				"error", err)
			return 0, true, nil // Continue даже если failed
		}
		n := copy(buf, validateData)

		// Broadcast StopMove to visible players (Phase 5.1)
		stopMove := serverpackets.NewStopMove(player)
		stopData, err := stopMove.Write()
		if err != nil {
			slog.Error("failed to serialize StopMove",
				"character", player.Name(),
				"error", err)
		} else {
			// Phase 5.1: Use BroadcastToVisibleNear (LOD optimization, -90% packets)
			h.clientManager.BroadcastToVisibleNear(player, stopData, len(stopData))
		}

		return n, true, nil // Connection stays open
	}

	// Update player location (validated)
	newLoc := model.NewLocation(pkt.TargetX, pkt.TargetY, pkt.TargetZ, player.Location().Heading)
	player.SetLocation(newLoc)

	// Phase 5.1: Track last server-validated position
	player.Movement().SetLastServerPosition(pkt.TargetX, pkt.TargetY, pkt.TargetZ)

	slog.Debug("player moving",
		"character", player.Name(),
		"from", fmt.Sprintf("(%d,%d,%d)", pkt.OriginX, pkt.OriginY, pkt.OriginZ),
		"to", fmt.Sprintf("(%d,%d,%d)", pkt.TargetX, pkt.TargetY, pkt.TargetZ))

	// Broadcast movement to visible players
	movePkt := serverpackets.NewCharMoveToLocation(player, pkt.TargetX, pkt.TargetY, pkt.TargetZ)
	moveData, err := movePkt.Write()
	if err != nil {
		slog.Error("failed to serialize CharMoveToLocation",
			"character", player.Name(),
			"error", err)
		// Continue даже если broadcast failed
	} else {
		// Phase 5.1: Use BroadcastToVisibleNear (LOD optimization, -90% packets)
		visibleCount := h.clientManager.BroadcastToVisibleNear(player, moveData, len(moveData))
		if visibleCount > 0 {
			slog.Debug("broadcasted movement",
				"character", player.Name(),
				"visible_players", visibleCount)
		}
	}

	// No response packet to client (movement is client-predicted)
	return 0, true, nil
}

// sendVisibleObjectsInfo sends CharInfo + NpcInfo + ItemOnGround packets TO client for all visible objects.
// Phase 4.19: Parallel encryption implementation — encrypts packets in parallel and sends as batch.
//
// Uses ForEachVisibleObjectCached for efficient visibility queries.
// Handles Players (CharInfo), NPCs (NpcInfo), and Items (ItemOnGround).
//
// Performance improvement:
// - Before: sequential 450 packets × 50µs = 22.5ms per EnterWorld
// - After: parallel encryption (20 goroutines) + batched TCP send = ~1.6ms
// - Result: -92.9% latency reduction (22.5ms → 1.6ms)
//
// Thread-safety: EncryptInPlace() is safe after authentication (firstPacket=false).
func (h *Handler) sendVisibleObjectsInfo(client *GameClient, player *model.Player) error {
	// Thread-safe packet collection
	mu := sync.Mutex{}
	encryptedPackets := make([][]byte, 0, 450)
	var lastErr error

	// Semaphore to limit concurrent goroutines (avoid goroutine explosion)
	const maxConcurrent = 20
	semaphore := make(chan struct{}, maxConcurrent)

	var wg sync.WaitGroup
	var playerCount, npcCount, itemCount int

	world.ForEachVisibleObjectCached(player, func(obj *model.WorldObject) bool {
		objectID := obj.ObjectID()

		// Skip self
		if constants.IsPlayerObjectID(objectID) {
			otherClient := h.clientManager.GetClientByObjectID(objectID)
			if otherClient != nil {
				if otherPlayer := otherClient.ActivePlayer(); otherPlayer != nil {
					if otherPlayer.CharacterID() == player.CharacterID() {
						return true // Don't send CharInfo for self
					}
				}
			}
		}

		wg.Add(1)
		semaphore <- struct{}{} // Acquire

		go func(o *model.WorldObject) {
			defer wg.Done()
			defer func() { <-semaphore }() // Release

			// Serialize packet based on object type
			var payloadData []byte
			var packetType string
			var err error

			if constants.IsPlayerObjectID(o.ObjectID()) {
				// This is a Player — send CharInfo
				otherClient := h.clientManager.GetClientByObjectID(o.ObjectID())
				if otherClient == nil {
					return // Player offline, skip
				}

				otherPlayer := otherClient.ActivePlayer()
				if otherPlayer == nil {
					return // Player not in game yet, skip
				}

				payloadData, err = serverpackets.NewCharInfo(otherPlayer).Write()
				packetType = "CharInfo"

				mu.Lock()
				playerCount++
				mu.Unlock()

			} else if constants.IsNpcObjectID(o.ObjectID()) {
				// This is an NPC — send NpcInfo
				npc, ok := world.Instance().GetNpc(o.ObjectID())
				if !ok {
					return // NPC not found or despawned, skip
				}

				payloadData, err = serverpackets.NewNpcInfo(npc).Write()
				packetType = "NpcInfo"

				mu.Lock()
				npcCount++
				mu.Unlock()

			} else if constants.IsItemObjectID(o.ObjectID()) {
				// This is a dropped item — send ItemOnGround
				droppedItem, ok := world.Instance().GetItem(o.ObjectID())
				if !ok {
					return // Item not found or picked up, skip
				}

				payloadData, err = serverpackets.NewItemOnGround(droppedItem).Write()
				packetType = "ItemOnGround"

				mu.Lock()
				itemCount++
				mu.Unlock()

			} else {
				return // Unknown object type, skip
			}

			if err != nil {
				slog.Error("failed to serialize packet",
					"packet_type", packetType,
					"object_id", o.ObjectID(),
					"error", err)
				mu.Lock()
				if lastErr == nil {
					lastErr = err
				}
				mu.Unlock()
				return
			}

			// Allocate buffer for this packet (header + payload + padding)
			buf := make([]byte, constants.PacketHeaderSize+len(payloadData)+constants.PacketBufferPadding)
			copy(buf[constants.PacketHeaderSize:], payloadData)

			// Encrypt in-place (thread-safe after authentication)
			encSize, err := protocol.EncryptInPlace(client.Encryption(), buf, len(payloadData))
			if err != nil {
				slog.Error("failed to encrypt packet",
					"packet_type", packetType,
					"object_id", o.ObjectID(),
					"error", err)
				mu.Lock()
				if lastErr == nil {
					lastErr = err
				}
				mu.Unlock()
				return
			}

			// Add encrypted packet to collection (mutex-protected)
			mu.Lock()
			encryptedPackets = append(encryptedPackets, buf[:encSize])
			mu.Unlock()
		}(obj)

		return true // Continue iteration
	})

	// Wait for all goroutines to complete
	wg.Wait()

	// Check for errors during packet creation/encryption
	if lastErr != nil {
		return fmt.Errorf("failed to create visible objects info packets: %w", lastErr)
	}

	// Send all packets in single batch (single TCP syscall)
	if len(encryptedPackets) > 0 {
		if err := protocol.WriteBatch(client.Conn(), encryptedPackets); err != nil {
			return fmt.Errorf("failed to send visible objects info batch: %w", err)
		}

		slog.Debug("sent info for visible objects",
			"character", player.Name(),
			"visible_players", playerCount,
			"visible_npcs", npcCount,
			"visible_items", itemCount,
			"total_packets", len(encryptedPackets))
	}

	return nil
}

// handleLogout processes the Logout packet (opcode 0x09).
// Client sends this when user clicks Exit button.
//
// Phase 4.17.5: MVP implementation with basic logout flow.
// TODO Phase 5.x: Add offline-trade mode support.
//
// Reference: L2J_Mobius Logout.java (53-107)
func (h *Handler) handleLogout(ctx context.Context, client *GameClient, data, buf []byte) (int, bool, error) {
	_, err := clientpackets.ParseLogout(data)
	if err != nil {
		return 0, false, fmt.Errorf("parsing Logout: %w", err)
	}

	// Get active player
	player := client.ActivePlayer()
	if player == nil {
		slog.Warn("Logout without active player",
			"account", client.AccountName(),
			"client", client.IP())
		// Close connection даже если player nil
		client.MarkForDisconnection()
		return 0, true, nil
	}

	// Check if can logout
	if !player.CanLogout() {
		slog.Info("logout denied (cannot logout)",
			"account", client.AccountName(),
			"character", player.Name(),
			"client", client.IP())

		// TODO: Send SystemMessage "YOU_CANNOT_EXIT_WHILE_IN_COMBAT"
		// For now, just ignore the request
		return 0, true, nil
	}

	slog.Info("player logging out",
		"account", client.AccountName(),
		"character", player.Name(),
		"level", player.Level(),
		"client", client.IP())

	// TODO Phase 4.17.6: Remove from boss zone, Olympiad unregister
	// player.RemoveFromBossZone()
	// olympiad.Unregister(player)

	// TODO Phase 4.17.6: Instance cleanup (if RESTORE_PLAYER_INSTANCE = false)
	// if !config.RestorePlayerInstance {
	//     // Save exit location, remove from instance
	// }

	// TODO Phase 5.x: Check offline-trade mode
	// if offlineTrader.EnteredOfflineMode(player) {
	//     return 0, true, nil
	// }

	// TODO Phase 4.17.6: Save player to DB (location, inventory, skills, quests, etc.)
	// For MVP, we skip DB save to keep logout simple
	// if err := h.charRepo.Save(ctx, player); err != nil {
	//     slog.Error("failed to save player on logout", "error", err)
	// }

	// Remove from world (Phase 4.17.5)
	world.Instance().RemoveObject(player.ObjectID())

	// Clear active player from client
	client.SetActivePlayer(nil)

	// Send LeaveWorld packet (Phase 4.17.3)
	leaveWorld := serverpackets.NewLeaveWorld()
	leaveWorldData, err := leaveWorld.Write()
	if err != nil {
		slog.Error("failed to serialize LeaveWorld",
			"character", player.Name(),
			"error", err)
		// Continue with disconnect даже если packet failed
		client.MarkForDisconnection()
		return 0, true, nil
	}

	n := copy(buf, leaveWorldData)
	if n != len(leaveWorldData) {
		slog.Error("buffer too small for LeaveWorld",
			"character", player.Name(),
			"size", len(leaveWorldData),
			"buffer_size", len(buf))
		// Continue with disconnect
		client.MarkForDisconnection()
		return 0, true, nil
	}

	// Mark client for disconnection (server.go will close TCP after sending LeaveWorld)
	client.MarkForDisconnection()

	slog.Info("player logged out successfully",
		"account", client.AccountName(),
		"character", player.Name())

	return n, true, nil
}

// handleRequestRestart processes the RequestRestart packet (opcode 0x46).
// Client sends this when user clicks "Restart" to return to character selection screen.
// Unlike Logout, RequestRestart does NOT close TCP connection — client returns to char selection.
//
// Phase 4.17.6: MVP implementation with basic restart flow.
// TODO Phase 5.x: Add full checks (enchant, class change, store mode, festival).
//
// Reference: L2J_Mobius RequestRestart.java (60-173)
func (h *Handler) handleRequestRestart(ctx context.Context, client *GameClient, data, buf []byte) (int, bool, error) {
	_, err := clientpackets.ParseRequestRestart(data)
	if err != nil {
		return 0, false, fmt.Errorf("parsing RequestRestart: %w", err)
	}

	// Verify client is in game
	if client.State() != ClientStateInGame {
		slog.Warn("RequestRestart from non-ingame state",
			"account", client.AccountName(),
			"state", client.State(),
			"client", client.IP())

		// Send denial
		restartResp := serverpackets.NewRestartResponse(false)
		respData, _ := restartResp.Write()
		copy(buf, respData)
		return len(respData), true, nil
	}

	// Get active player
	player := client.ActivePlayer()
	if player == nil {
		slog.Warn("RequestRestart without active player",
			"account", client.AccountName(),
			"client", client.IP())

		// Send denial
		restartResp := serverpackets.NewRestartResponse(false)
		respData, _ := restartResp.Write()
		copy(buf, respData)
		return len(respData), true, nil
	}

	// TODO Phase 5.x: Check active enchant
	// if player.GetActiveEnchantItemID() != 0 {
	//     return sendRestartDenied(buf)
	// }

	// TODO Phase 5.x: Check class change
	// if player.IsChangingClass() {
	//     return sendRestartDenied(buf)
	// }

	// Check if in trade/store mode
	if player.IsTrading() {
		slog.Info("restart denied (trading)",
			"account", client.AccountName(),
			"character", player.Name(),
			"client", client.IP())

		restartResp := serverpackets.NewRestartResponse(false)
		respData, _ := restartResp.Write()
		copy(buf, respData)
		return len(respData), true, nil
	}

	// Check if can logout (includes attack stance check)
	if !player.CanLogout() {
		slog.Info("restart denied (cannot logout)",
			"account", client.AccountName(),
			"character", player.Name(),
			"client", client.IP())

		restartResp := serverpackets.NewRestartResponse(false)
		respData, _ := restartResp.Write()
		copy(buf, respData)
		return len(respData), true, nil
	}

	// TODO Phase 5.x: Check festival participant
	// if player.IsFestivalParticipant() {
	//     if sevenSignsFestival.IsInitialized() {
	//         return sendRestartDenied(buf)
	//     }
	// }

	slog.Info("player restarting to character selection",
		"account", client.AccountName(),
		"character", player.Name(),
		"level", player.Level(),
		"client", client.IP())

	// TODO Phase 4.17.7: Remove from boss zone, Olympiad unregister
	// player.RemoveFromBossZone()
	// olympiad.Unregister(player)

	// TODO Phase 4.17.7: Instance cleanup (if RESTORE_PLAYER_INSTANCE = false)
	// if !config.RestorePlayerInstance {
	//     // Save exit location, remove from instance
	// }

	// TODO Phase 5.x: Check offline-trade mode
	// if offlineTrader.EnteredOfflineMode(player) {
	//     return sendRestartSuccess(client, buf)
	// }

	// TODO Phase 4.17.7: Full player save (inventory, skills, quests, etc.)
	// For MVP, we skip DB save to keep restart simple
	// if err := h.charRepo.Save(ctx, player); err != nil {
	//     slog.Error("failed to save player on restart", "error", err)
	// }

	// Remove from world (Phase 4.17.6)
	world.Instance().RemoveObject(player.ObjectID())

	// Clear active player from client
	client.SetActivePlayer(nil)
	client.SetSelectedCharacter(-1)

	// Transition to AUTHENTICATED state (Phase 4.17.6)
	// This allows client to access CharacterSelect, CharacterCreate, CharacterDelete packets
	client.SetState(ClientStateAuthenticated)

	slog.Info("player returned to character selection",
		"account", client.AccountName(),
		"character", player.Name())

	// Send response packets
	var totalBytes int

	// 1. RestartResponse(true) — confirms restart success
	restartResp := serverpackets.NewRestartResponse(true)
	respData, err := restartResp.Write()
	if err != nil {
		slog.Error("failed to serialize RestartResponse",
			"character", player.Name(),
			"error", err)
		return 0, false, fmt.Errorf("serializing RestartResponse: %w", err)
	}
	n := copy(buf[totalBytes:], respData)
	if n != len(respData) {
		return 0, false, fmt.Errorf("buffer too small for RestartResponse")
	}
	totalBytes += n

	// 2. CharSelectionInfo — sends list of characters for account
	// Get SessionKey PlayOkID1 for CharSelectionInfo
	sessionKey := client.SessionKey()
	if sessionKey == nil {
		slog.Error("no session key for authenticated client",
			"account", client.AccountName(),
			"client", client.IP())
		return 0, false, fmt.Errorf("missing session key")
	}

	// Load characters for this account (Phase 4.17.6)
	players, err := h.charRepo.LoadByAccountName(ctx, client.AccountName())
	if err != nil {
		slog.Error("failed to load characters for restart",
			"account", client.AccountName(),
			"error", err)
		return 0, false, fmt.Errorf("loading characters: %w", err)
	}

	charList := serverpackets.NewCharSelectionInfoFromPlayers(client.AccountName(), sessionKey.PlayOkID1, players)
	charListData, err := charList.Write()
	if err != nil {
		slog.Error("failed to serialize CharSelectionInfo",
			"account", client.AccountName(),
			"error", err)
		return 0, false, fmt.Errorf("serializing CharSelectionInfo: %w", err)
	}
	n = copy(buf[totalBytes:], charListData)
	if n != len(charListData) {
		return 0, false, fmt.Errorf("buffer too small for CharSelectionInfo")
	}
	totalBytes += n

	slog.Info("restart completed successfully",
		"account", client.AccountName(),
		"total_bytes", totalBytes)

	return totalBytes, true, nil
}

// handleValidatePosition processes ValidatePosition packet (opcode 0x48).
// Client sends this periodically (~200ms) to report current position.
// Server validates and corrects if desynced.
//
// Phase 5.1: Movement validation — desync detection and correction.
//
// Reference: L2J_Mobius ValidatePosition.java
func (h *Handler) handleValidatePosition(ctx context.Context, client *GameClient, data, buf []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseValidatePosition(data)
	if err != nil {
		return 0, false, fmt.Errorf("parsing ValidatePosition: %w", err)
	}

	// Verify character is in game
	if client.State() != ClientStateInGame {
		return 0, true, nil // Ignore silently
	}

	// Get active player
	player := client.ActivePlayer()
	if player == nil {
		return 0, true, nil // Ignore silently
	}

	// Z-bounds check (prevent flying/underground exploits)
	// Reference: L2J_Mobius ValidatePosition.java:76-82
	if pkt.Z < MinZCoordinate || pkt.Z > MaxZCoordinate {
		slog.Warn("abnormal Z coordinate from client",
			"character", player.Name(),
			"z", pkt.Z,
			"allowed_range", fmt.Sprintf("[%d..%d]", MinZCoordinate, MaxZCoordinate))

		// Teleport player to last server-validated position
		lastX, lastY, lastZ := player.Movement().LastServerPosition()
		player.SetLocation(model.NewLocation(lastX, lastY, lastZ, player.Location().Heading))

		// Send ValidateLocation to force correction
		validateLoc := serverpackets.NewValidateLocation(player)
		validateData, err := validateLoc.Write()
		if err != nil {
			slog.Error("failed to serialize ValidateLocation",
				"character", player.Name(),
				"error", err)
			return 0, true, nil
		}

		n := copy(buf, validateData)
		return n, true, nil
	}

	// Update client-reported position
	player.Movement().SetClientPosition(pkt.X, pkt.Y, pkt.Z, pkt.Heading)

	// Check desync between client and server positions
	needsCorrection, diffSq := ValidatePositionDesync(player, pkt.X, pkt.Y, pkt.Z)
	if needsCorrection {
		slog.Info("position desync detected",
			"character", player.Name(),
			"diff_squared", diffSq,
			"client", fmt.Sprintf("(%d,%d,%d)", pkt.X, pkt.Y, pkt.Z),
			"server", fmt.Sprintf("(%d,%d,%d)", player.Location().X, player.Location().Y, player.Location().Z))

		// Send ValidateLocation to correct client
		validateLoc := serverpackets.NewValidateLocation(player)
		validateData, err := validateLoc.Write()
		if err != nil {
			slog.Error("failed to serialize ValidateLocation",
				"character", player.Name(),
				"error", err)
			return 0, true, nil
		}

		n := copy(buf, validateData)
		return n, true, nil
	}

	// Position synchronized, no response needed
	return 0, true, nil
}

// handleRequestAction processes RequestAction packet (opcode 0x04).
// Client sends this when player clicks on an object (target selection or attack intent).
//
// Phase 5.2: Target System.
//
// Reference: L2J_Mobius RequestActionUse.java
func (h *Handler) handleRequestAction(ctx context.Context, client *GameClient, data, buf []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseRequestAction(data)
	if err != nil {
		return 0, false, fmt.Errorf("parsing RequestAction: %w", err)
	}

	// Verify character is in game
	if client.State() != ClientStateInGame {
		return 0, true, nil // Ignore silently
	}

	// Get active player
	player := client.ActivePlayer()
	if player == nil {
		return 0, true, nil // Ignore silently
	}

	// Validate target selection
	worldInst := world.Instance()
	target, err := ValidateTargetSelection(player, uint32(pkt.ObjectID), worldInst)
	if err != nil {
		slog.Debug("target selection failed",
			"character", player.Name(),
			"targetID", pkt.ObjectID,
			"error", err)
		// Silent failure — client will not change target
		return 0, true, nil
	}

	// Set target
	player.SetTarget(target)

	slog.Debug("target selected",
		"character", player.Name(),
		"targetID", target.ObjectID(),
		"targetName", target.Name(),
		"attackIntent", pkt.IsAttackIntent())

	// Prepare response buffer
	totalBytes := 0

	// 1. Send MyTargetSelected (highlight target + show HP bar)
	myTargetSel := serverpackets.NewMyTargetSelected(target.ObjectID())
	targetSelData, err := myTargetSel.Write()
	if err != nil {
		slog.Error("failed to serialize MyTargetSelected",
			"character", player.Name(),
			"error", err)
		return 0, true, nil
	}
	n := copy(buf[totalBytes:], targetSelData)
	totalBytes += n

	// 2. Send StatusUpdate (HP/MP/CP values for target)
	// Check if target is a Character (has HP/MP/CP)
	if character := getCharacterFromObject(target, worldInst); character != nil {
		statusUpdate := serverpackets.NewStatusUpdateForTarget(character)
		statusData, err := statusUpdate.Write()
		if err != nil {
			slog.Error("failed to serialize StatusUpdate",
				"character", player.Name(),
				"error", err)
		} else {
			n = copy(buf[totalBytes:], statusData)
			totalBytes += n
		}
	}

	// TODO Phase 5.3: If attack intent (shift+click), start auto-attack

	return totalBytes, true, nil
}

// getCharacterFromObject attempts to extract Character from WorldObject.
// Returns nil if object is not a Character (e.g., dropped item).
//
// Phase 5.2: Helper для получения Character из WorldObject.
func getCharacterFromObject(obj *model.WorldObject, worldInst *world.World) *model.Character {
	objectID := obj.ObjectID()

	// Check if it's an NPC
	if npc, ok := worldInst.GetNpc(objectID); ok {
		return npc.Character
	}

	// Check if it's a Player
	// Note: World doesn't have GetPlayer(), so we need to cast via ObjectIDRange
	// Phase 4.15: Player IDs start with 0x10000000
	if objectID >= 0x10000000 && objectID < 0x20000000 {
		// This is a player objectID
		// For now, we don't have a way to get Player from World by objectID
		// TODO Phase 5.3: Add World.GetPlayer() or use a player registry
		// Fallback: return nil (won't send StatusUpdate for players)
		return nil
	}

	return nil
}

// handleAttackRequest processes AttackRequest packet (opcode 0x0A).
// Client sends this when player clicks on enemy to initiate auto-attack.
//
// Workflow:
//  1. Validate target exists in world
//  2. Validate attack (range, dead, etc)
//  3. Start auto-attack via player.DoAttack(target)
//
// Phase 5.3: Basic Combat System.
// Java reference: AttackRequest.java (runImpl, line 53-129).
func (h *Handler) handleAttackRequest(ctx context.Context, client *GameClient, data, buf []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseAttackRequest(data)
	if err != nil {
		return 0, false, fmt.Errorf("parsing AttackRequest: %w", err)
	}

	if client.State() != ClientStateInGame {
		return 0, true, nil // Ignore silently
	}

	player := client.ActivePlayer()
	if player == nil {
		return 0, true, nil
	}

	// Get target from world
	worldInst := world.Instance()
	target, exists := worldInst.GetObject(pkt.ObjectID)
	if !exists {
		// Target not found — send ActionFailed
		actionFailed := serverpackets.NewActionFailed()
		failedData, _ := actionFailed.Write()
		n := copy(buf, failedData)
		return n, true, nil
	}

	// Validate attack (range, dead, etc)
	if err := combat.ValidateAttack(player, target); err != nil {
		slog.Warn("attack validation failed",
			"character", player.Name(),
			"target", target.ObjectID(),
			"error", err)

		// Send ActionFailed
		actionFailed := serverpackets.NewActionFailed()
		failedData, _ := actionFailed.Write()
		n := copy(buf, failedData)
		return n, true, nil
	}

	// Phase 5.6: PvP + PvE combat (Player vs Player/NPC)
	// ExecuteAttack handles type assertion internally
	if combat.CombatMgr != nil {
		combat.CombatMgr.ExecuteAttack(player, target)
	}

	// No response to client (Attack packet sent via broadcast)
	return 0, true, nil
}

// handleRequestPickup processes RequestPickup packet (opcode 0x14).
// Client sends this when player clicks on dropped item to pick it up.
//
// Workflow:
//  1. Validate item exists in world
//  2. Validate pickup range (200 units max)
//  3. Add item to player's inventory
//  4. Remove DroppedItem from world
//  5. Broadcast DeleteObject to visible players
//
// Phase 5.7: Loot System MVP.
// Java reference: RequestGetItem.java (runImpl, line 59-188).
func (h *Handler) handleRequestPickup(ctx context.Context, client *GameClient, data, buf []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseRequestPickup(data)
	if err != nil {
		return 0, false, fmt.Errorf("parsing RequestPickup: %w", err)
	}

	if client.State() != ClientStateInGame {
		return 0, true, nil // Ignore silently
	}

	player := client.ActivePlayer()
	if player == nil {
		return 0, true, nil
	}

	// Get DroppedItem from world
	worldInst := world.Instance()
	obj, exists := worldInst.GetObject(uint32(pkt.ObjectID))
	if !exists {
		slog.Warn("pickup failed: item not found",
			"character", player.Name(),
			"objectID", pkt.ObjectID)

		// Send ActionFailed
		actionFailed := serverpackets.NewActionFailed()
		failedData, _ := actionFailed.Write()
		n := copy(buf, failedData)
		return n, true, nil
	}

	// Type assert to DroppedItem
	droppedItem, ok := obj.Data.(*model.DroppedItem)
	if !ok {
		slog.Warn("pickup failed: object is not DroppedItem",
			"character", player.Name(),
			"objectID", pkt.ObjectID)

		actionFailed := serverpackets.NewActionFailed()
		failedData, _ := actionFailed.Write()
		n := copy(buf, failedData)
		return n, true, nil
	}

	// Validate pickup range (200 units max)
	const MaxItemPickupRange = 200
	const MaxItemPickupRangeSquared = MaxItemPickupRange * MaxItemPickupRange

	playerLoc := player.Location()
	itemLoc := droppedItem.Location()

	dx := int64(playerLoc.X - itemLoc.X)
	dy := int64(playerLoc.Y - itemLoc.Y)
	distSq := dx*dx + dy*dy

	if distSq > MaxItemPickupRangeSquared {
		slog.Warn("pickup failed: out of range",
			"character", player.Name(),
			"objectID", pkt.ObjectID,
			"distance_sq", distSq,
			"max", MaxItemPickupRangeSquared)

		actionFailed := serverpackets.NewActionFailed()
		failedData, _ := actionFailed.Write()
		n := copy(buf, failedData)
		return n, true, nil
	}

	// Get Item from DroppedItem
	item := droppedItem.Item()
	if item == nil {
		slog.Error("pickup failed: DroppedItem has nil item",
			"character", player.Name(),
			"objectID", pkt.ObjectID)

		actionFailed := serverpackets.NewActionFailed()
		failedData, _ := actionFailed.Write()
		n := copy(buf, failedData)
		return n, true, nil
	}

	// Add item to player's inventory
	if err := player.Inventory().AddItem(item); err != nil {
		slog.Error("pickup failed: cannot add to inventory",
			"character", player.Name(),
			"objectID", pkt.ObjectID,
			"itemID", item.ItemID(),
			"error", err)

		actionFailed := serverpackets.NewActionFailed()
		failedData, _ := actionFailed.Write()
		n := copy(buf, failedData)
		return n, true, nil
	}

	// Remove DroppedItem from world
	worldInst.RemoveObject(uint32(pkt.ObjectID))

	// Broadcast DeleteObject to visible players
	deleteObj := serverpackets.NewDeleteObject(pkt.ObjectID)
	deleteData, err := deleteObj.Write()
	if err != nil {
		slog.Error("failed to serialize DeleteObject",
			"objectID", pkt.ObjectID,
			"error", err)
		// Continue — item already picked up, just failed to broadcast
	} else {
		h.clientManager.BroadcastToVisible(player, deleteData, len(deleteData))
	}

	slog.Info("item picked up",
		"character", player.Name(),
		"itemID", item.ItemID(),
		"itemName", item.Template().Name,
		"count", item.Count(),
		"objectID", pkt.ObjectID)

	// TODO Phase 5.8: Send InventoryUpdate packet to client
	// For MVP, item added to inventory but client doesn't see visual update
	// Client will see item after restart/re-login

	// No response to client (InventoryUpdate to be implemented in Phase 5.8)
	return 0, true, nil
}
