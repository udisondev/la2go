package gameserver

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/udisondev/la2go/internal/gameserver/clientpackets"
	"github.com/udisondev/la2go/internal/gameserver/serverpackets"
	"github.com/udisondev/la2go/internal/login"
	"github.com/udisondev/la2go/internal/model"
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
		// TODO: Add more packet handlers (Logout, ValidatePosition, etc.)
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

	// Store selected character slot
	client.SetSelectedCharacter(pkt.CharSlot)
	client.SetState(ClientStateEntering)

	slog.Info("character selected",
		"account", client.AccountName(),
		"slot", pkt.CharSlot,
		"client", client.IP())

	// No response packet for CharacterSelect (client waits for EnterWorld)
	return 0, true, nil
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

	// TODO Phase 4.8: Add more packets:
	// - CharInfo (for other visible players)
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

	// Load player from DB (TODO Phase 4.9: cache Player in GameClient)
	charSlot := client.SelectedCharacter()
	if charSlot < 0 {
		slog.Warn("MoveToLocation without character selection",
			"account", client.AccountName(),
			"client", client.IP())
		return 0, true, nil
	}

	players, err := h.charRepo.LoadByAccountName(ctx, client.AccountName())
	if err != nil {
		return 0, false, fmt.Errorf("loading characters for account %s: %w", client.AccountName(), err)
	}

	if int(charSlot) >= len(players) {
		slog.Warn("character slot out of range",
			"slot", charSlot,
			"character_count", len(players),
			"account", client.AccountName(),
			"client", client.IP())
		return 0, true, nil
	}

	player := players[charSlot]

	// Update player location (simplified — no pathfinding yet)
	// TODO Phase 4.9: Add pathfinding, collision detection, speed validation
	newLoc := model.NewLocation(pkt.TargetX, pkt.TargetY, pkt.TargetZ, player.Location().Heading)
	player.SetLocation(newLoc)

	slog.Debug("player moving",
		"character", player.Name(),
		"from", fmt.Sprintf("(%d,%d,%d)", pkt.OriginX, pkt.OriginY, pkt.OriginZ),
		"to", fmt.Sprintf("(%d,%d,%d)", pkt.TargetX, pkt.TargetY, pkt.TargetZ))

	// Broadcast movement to visible players
	// TODO Phase 4.9: Use BroadcastToVisible with CharMoveToLocation packet
	// For now, just return empty response (no broadcast)

	// No response packet to client (movement is client-predicted)
	return 0, true, nil
}

// TODO: Add more packet handlers:
// - handleLogout (opcode 0x09)
// - handleRequestRestart (opcode 0x46)
// - handleValidatePosition (opcode 0x48)
