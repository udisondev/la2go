package gameserver

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/udisondev/la2go/internal/constants"
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
		case clientpackets.OpcodeLogout:
			return h.handleLogout(ctx, client, body, buf)
		case clientpackets.OpcodeRequestRestart:
			return h.handleRequestRestart(ctx, client, body, buf)
		// TODO: Add more packet handlers (ValidatePosition 0x48, etc.)
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
	players, err := h.charRepo.LoadByAccountName(ctx, pkt.AccountName)
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
	players, err := h.charRepo.LoadByAccountName(ctx, client.AccountName())
	if err != nil {
		return 0, false, fmt.Errorf("loading characters for account %s: %w", client.AccountName(), err)
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
	players, err := h.charRepo.LoadByAccountName(ctx, client.AccountName())
	if err != nil {
		return 0, false, fmt.Errorf("loading characters for account %s: %w", client.AccountName(), err)
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

	// Get cached player (Phase 4.8 part 2)
	player := client.ActivePlayer()
	if player == nil {
		slog.Warn("MoveToLocation without active player",
			"account", client.AccountName(),
			"client", client.IP())
		return 0, false, fmt.Errorf("no active player for account %s", client.AccountName())
	}

	// Update player location (simplified — no pathfinding yet)
	// TODO Phase 4.9: Add pathfinding, collision detection, speed validation
	newLoc := model.NewLocation(pkt.TargetX, pkt.TargetY, pkt.TargetZ, player.Location().Heading)
	player.SetLocation(newLoc)

	slog.Debug("player moving",
		"character", player.Name(),
		"from", fmt.Sprintf("(%d,%d,%d)", pkt.OriginX, pkt.OriginY, pkt.OriginZ),
		"to", fmt.Sprintf("(%d,%d,%d)", pkt.TargetX, pkt.TargetY, pkt.TargetZ))

	// Broadcast movement to visible players (Phase 4.8 part 2)
	movePkt := serverpackets.NewCharMoveToLocation(player, pkt.TargetX, pkt.TargetY, pkt.TargetZ)
	moveData, err := movePkt.Write()
	if err != nil {
		slog.Error("failed to serialize CharMoveToLocation",
			"character", player.Name(),
			"error", err)
		// Continue даже если broadcast failed
	} else {
		// Broadcast to all visible players (except self)
		visibleCount := h.clientManager.BroadcastToVisibleExcept(player, player, moveData, len(moveData))
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
// Called after player enters world (Phase 4.9 Part 2 + Phase 4.10).
// Uses ForEachVisibleObjectCached for efficient visibility queries.
// Handles Players (CharInfo), NPCs (NpcInfo), and Items (ItemOnGround).
func (h *Handler) sendVisibleObjectsInfo(client *GameClient, player *model.Player) error {
	var playerCount, npcCount, itemCount int
	var lastErr error

	// Allocate buffer for packets (reused for each packet)
	// CharInfo ~512 bytes, NpcInfo ~256 bytes, ItemOnGround ~128 bytes
	// Phase 4.11 Tier 1 Opt 3: Increased to 2048 to reduce grows (-50 allocs expected)
	buf := make([]byte, 2048) // + header + padding

	world.ForEachVisibleObjectCached(player, func(obj *model.WorldObject) bool {
		objectID := obj.ObjectID()

		// Determine object type by ObjectID range
		if constants.IsPlayerObjectID(objectID) {
			// This is a Player — send CharInfo
			otherClient := h.clientManager.GetClientByObjectID(objectID)
			if otherClient == nil {
				return true // Player offline, skip
			}

			otherPlayer := otherClient.ActivePlayer()
			if otherPlayer == nil {
				return true // Player not in game yet, skip
			}

			// Don't send CharInfo for self
			if otherPlayer.CharacterID() == player.CharacterID() {
				return true
			}

			// Create CharInfo packet
			charInfo := serverpackets.NewCharInfo(otherPlayer)
			charInfoData, err := charInfo.Write()
			if err != nil {
				slog.Error("failed to serialize CharInfo",
					"character", player.Name(),
					"visible_character", otherPlayer.Name(),
					"error", err)
				lastErr = err
				return true
			}

			// Send CharInfo packet
			if err := h.sendPacketToClient(client, buf, charInfoData, "CharInfo", otherPlayer.Name()); err != nil {
				lastErr = err
				return true
			}

			playerCount++

		} else if constants.IsNpcObjectID(objectID) {
			// This is an NPC — send NpcInfo (Phase 4.10 Part 2)
			npc, ok := world.Instance().GetNpc(objectID)
			if !ok {
				return true // NPC not found or despawned, skip
			}

			// Create NpcInfo packet
			npcInfo := serverpackets.NewNpcInfo(npc)
			npcInfoData, err := npcInfo.Write()
			if err != nil {
				slog.Error("failed to serialize NpcInfo",
					"character", player.Name(),
					"visible_npc", npc.Name(),
					"error", err)
				lastErr = err
				return true
			}

			// Send NpcInfo packet
			if err := h.sendPacketToClient(client, buf, npcInfoData, "NpcInfo", npc.Name()); err != nil {
				lastErr = err
				return true
			}

			npcCount++

		} else if constants.IsItemObjectID(objectID) {
			// This is a dropped item — send ItemOnGround (Phase 4.10 Part 3)
			droppedItem, ok := world.Instance().GetItem(objectID)
			if !ok {
				return true // Item not found or picked up, skip
			}

			// Create ItemOnGround packet
			itemPkt := serverpackets.NewItemOnGround(droppedItem)
			itemData, err := itemPkt.Write()
			if err != nil {
				slog.Error("failed to serialize ItemOnGround",
					"character", player.Name(),
					"item_object_id", objectID,
					"error", err)
				lastErr = err
				return true
			}

			// Send ItemOnGround packet
			if err := h.sendPacketToClient(client, buf, itemData, "ItemOnGround", fmt.Sprintf("item_%d", objectID)); err != nil {
				lastErr = err
				return true
			}

			itemCount++
		}

		return true // Continue iteration
	})

	if playerCount > 0 || npcCount > 0 || itemCount > 0 {
		slog.Debug("sent info for visible objects",
			"character", player.Name(),
			"visible_players", playerCount,
			"visible_npcs", npcCount,
			"visible_items", itemCount)
	}

	return lastErr
}

// sendPacketToClient is a helper to send a packet to client with error handling.
// Reuses buf for packet header, copies packetData to buf, and calls WritePacket.
func (h *Handler) sendPacketToClient(client *GameClient, buf, packetData []byte, packetType, targetName string) error {
	// Check buffer size
	if len(packetData) > len(buf[constants.PacketHeaderSize:]) {
		slog.Error("packet too large for buffer",
			"packet_type", packetType,
			"target", targetName,
			"size", len(packetData),
			"buffer_size", len(buf)-constants.PacketHeaderSize)
		return fmt.Errorf("packet too large: %d > %d", len(packetData), len(buf)-constants.PacketHeaderSize)
	}

	// Copy packet data to buffer
	copy(buf[constants.PacketHeaderSize:], packetData)

	// Send packet
	if err := protocol.WritePacket(client.Conn(), client.Encryption(), buf, len(packetData)); err != nil {
		slog.Error("failed to send packet",
			"packet_type", packetType,
			"target", targetName,
			"error", err)
		return fmt.Errorf("writing %s packet: %w", packetType, err)
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

// TODO: Add more packet handlers:
// - handleValidatePosition (opcode 0x48)
