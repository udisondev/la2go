package gameserver

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"sync"

	"github.com/udisondev/la2go/internal/constants"
	skilldata "github.com/udisondev/la2go/internal/data"
	"github.com/udisondev/la2go/internal/db"
	"github.com/udisondev/la2go/internal/game/combat"
	"github.com/udisondev/la2go/internal/game/skill"
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
	clientManager  *ClientManager      // Phase 4.5 PR4: register clients after auth
	charRepo       CharacterRepository // Phase 4.6: load characters for CharSelectionInfo
	persister      PlayerPersister     // Phase 6.0: DB persistence
}

// CharacterRepository defines interface for loading characters from database.
// Used for dependency injection to keep handler testable.
type CharacterRepository interface {
	LoadByAccountName(ctx context.Context, accountName string) ([]*model.Player, error)
}

// PlayerPersister defines interface for saving/loading player data.
// Phase 6.0: DB Persistence.
type PlayerPersister interface {
	SavePlayer(ctx context.Context, player *model.Player) error
	LoadPlayerData(ctx context.Context, charID int64) ([]db.ItemRow, []*model.SkillInfo, error)
}

// NewHandler creates a new packet handler for game clients.
func NewHandler(sessionManager *login.SessionManager, clientManager *ClientManager, charRepo CharacterRepository, persister PlayerPersister) *Handler {
	return &Handler{
		sessionManager: sessionManager,
		clientManager:  clientManager,
		charRepo:       charRepo,
		persister:      persister,
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
		case clientpackets.OpcodeRequestMagicSkillUse:
			return h.handleRequestMagicSkillUse(ctx, client, body, buf)
		case clientpackets.OpcodeSay2:
			return h.handleSay2(ctx, client, body, buf)
		case clientpackets.OpcodeRequestBypassToServer:
			return h.handleRequestBypassToServer(ctx, client, body, buf)
		case clientpackets.OpcodeRequestBuyItem:
			return h.handleRequestBuyItem(ctx, client, body, buf)
		case clientpackets.OpcodeRequestSellItem:
			return h.handleRequestSellItem(ctx, client, body, buf)
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

	// Phase 5.9.5: Apply auto-get skills for current level
	autoSkills := skilldata.GetAutoGetSkills(player.ClassID(), player.Level())
	for _, sl := range autoSkills {
		isPassive := false
		if tmpl := skilldata.GetSkillTemplate(sl.SkillID, sl.SkillLevel); tmpl != nil {
			isPassive = tmpl.IsPassive()
		}
		player.AddSkill(sl.SkillID, sl.SkillLevel, isPassive)
	}

	// Phase 6.0: Load items and skills from DB
	itemRows, skillInfos, err := h.persister.LoadPlayerData(ctx, player.CharacterID())
	if err != nil {
		slog.Error("load player data",
			"characterID", player.CharacterID(),
			"err", err)
		// Continue without — not fatal
	}

	// Restore skills from DB (override auto-get with saved levels)
	for _, si := range skillInfos {
		isPassive := false
		if tmpl := skilldata.GetSkillTemplate(si.SkillID, si.Level); tmpl != nil {
			isPassive = tmpl.IsPassive()
		}
		player.AddSkill(si.SkillID, si.Level, isPassive)
	}

	// Restore items to inventory
	for _, row := range itemRows {
		template := db.ItemDefToTemplate(row.ItemTypeID)
		if template == nil {
			slog.Warn("item template not found, skipping",
				"itemTypeID", row.ItemTypeID,
				"characterID", player.CharacterID())
			continue
		}
		objectID := world.IDGenerator().NextItemID()
		item, itemErr := model.NewItem(objectID, row.ItemTypeID, player.CharacterID(), row.Count, template)
		if itemErr != nil {
			slog.Error("restore item failed",
				"itemTypeID", row.ItemTypeID,
				"error", itemErr)
			continue
		}
		if row.Enchant > 0 {
			if enchErr := item.SetEnchant(row.Enchant); enchErr != nil {
				slog.Error("set enchant failed",
					"itemTypeID", row.ItemTypeID,
					"error", enchErr)
			}
		}
		if addErr := player.Inventory().AddItem(item); addErr != nil {
			slog.Error("add item to inventory failed",
				"itemTypeID", row.ItemTypeID,
				"error", addErr)
			continue
		}
		if model.ItemLocation(row.Location) == model.ItemLocationPaperdoll && row.SlotID >= 0 {
			if equipErr := player.Inventory().EquipItem(item, row.SlotID); equipErr != nil {
				slog.Error("equip item failed",
					"itemTypeID", row.ItemTypeID,
					"slot", row.SlotID,
					"error", equipErr)
			}
		}
	}

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

	// 3. InventoryItemList (Phase 6.0: send real items)
	invList := serverpackets.NewInventoryItemList(player.Inventory().GetItems())
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

	// 5. SkillList (player's learned skills)
	skills := serverpackets.NewSkillList(player.Skills())
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
// Phase 4.19: Parallel encryption implementation — encrypts packets in parallel.
// Phase 7.0: Sends via client.Send() (writePump batches and writes).
//
// Uses ForEachVisibleObjectCached for efficient visibility queries.
// Handles Players (CharInfo), NPCs (NpcInfo), and Items (ItemOnGround).
//
// Thread-safety: Encryption is safe after authentication (firstPacket=false).
const maxConcurrent = 20

func (h *Handler) sendVisibleObjectsInfo(client *GameClient, player *model.Player) error {
	// Thread-safe packet collection
	var (
		mu                               sync.Mutex
		lastErr                          error
		wg                               sync.WaitGroup
		playerCount, npcCount, itemCount int
		encryptedPackets                 = make([][]byte, 0, 450)
	)

	// Semaphore to limit concurrent goroutines (avoid goroutine explosion)
	semaphore := make(chan struct{}, maxConcurrent)

	writePool := h.clientManager.writePool

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

		semaphore <- struct{}{} // Acquire
		wg.Go(func() {
			defer func() { <-semaphore }() // Release

			// Serialize packet based on object type
			var payloadData []byte
			var packetType string
			var err error

			if constants.IsPlayerObjectID(obj.ObjectID()) {
				// This is a Player — send CharInfo
				otherClient := h.clientManager.GetClientByObjectID(obj.ObjectID())
				if otherClient == nil {
					return // Player offline, skip
				}

				otherPlayer := otherClient.ActivePlayer()
				if otherPlayer == nil {
					return // Player not in game yet, skip
				}

				charInfoPkt := serverpackets.NewCharInfo(otherPlayer)
				payloadData, err = charInfoPkt.Write()
				packetType = "CharInfo"

				mu.Lock()
				playerCount++
				mu.Unlock()

			} else if constants.IsNpcObjectID(obj.ObjectID()) {
				// This is an NPC — send NpcInfo
				npc, ok := world.Instance().GetNpc(obj.ObjectID())
				if !ok {
					return // NPC not found or despawned, skip
				}

				npcInfoPkt := serverpackets.NewNpcInfo(npc)
				payloadData, err = npcInfoPkt.Write()
				packetType = "NpcInfo"

				mu.Lock()
				npcCount++
				mu.Unlock()

			} else if constants.IsItemObjectID(obj.ObjectID()) {
				// This is a dropped item — send ItemOnGround
				droppedItem, ok := world.Instance().GetItem(obj.ObjectID())
				if !ok {
					return // Item not found or picked up, skip
				}

				itemOnGroundPkt := serverpackets.NewItemOnGround(droppedItem)
				payloadData, err = itemOnGroundPkt.Write()
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
					"object_id", obj.ObjectID(),
					"error", err)
				mu.Lock()
				if lastErr == nil {
					lastErr = err
				}
				mu.Unlock()
				return
			}

			// Encrypt into pool buffer (zero-alloc in steady state)
			var encPkt []byte
			if writePool != nil {
				encPkt, err = writePool.EncryptToPooled(client.Encryption(), payloadData, len(payloadData))
			} else {
				// Fallback: allocate buffer (for tests without writePool)
				buf := make([]byte, constants.PacketHeaderSize+len(payloadData)+constants.PacketBufferPadding)
				copy(buf[constants.PacketHeaderSize:], payloadData)
				var encSize int
				encSize, err = protocol.EncryptInPlace(client.Encryption(), buf, len(payloadData))
				if err == nil {
					encPkt = buf[:encSize]
				}
			}

			if err != nil {
				slog.Error("failed to encrypt packet",
					"packet_type", packetType,
					"object_id", obj.ObjectID(),
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
			encryptedPackets = append(encryptedPackets, encPkt)
			mu.Unlock()
		})

		return true // Continue iteration
	})

	// Wait for all goroutines to complete
	wg.Wait()

	// Check for errors during packet creation/encryption
	if lastErr != nil {
		return fmt.Errorf("creating visible objects info packets: %w", lastErr)
	}

	// Send all packets via write queue (writePump will batch via drain loop)
	if len(encryptedPackets) > 0 {
		for _, pkt := range encryptedPackets {
			if err := client.Send(pkt); err != nil {
				return fmt.Errorf("queueing visible object packet: %w", err)
			}
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

	// Phase 6.0: Save player to DB (location, inventory, skills)
	if err := h.persister.SavePlayer(ctx, player); err != nil {
		slog.Error("failed to save player on logout",
			"character", player.Name(),
			"error", err)
	}

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
		respData, err := restartResp.Write()
		if err != nil {
			return 0, false, fmt.Errorf("serializing RestartResponse: %w", err)
		}
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
		respData, err := restartResp.Write()
		if err != nil {
			return 0, false, fmt.Errorf("serializing RestartResponse: %w", err)
		}
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
		respData, err := restartResp.Write()
		if err != nil {
			return 0, false, fmt.Errorf("serializing RestartResponse: %w", err)
		}
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
		respData, err := restartResp.Write()
		if err != nil {
			return 0, false, fmt.Errorf("serializing RestartResponse: %w", err)
		}
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

	// Phase 6.0: Save player to DB (location, inventory, skills)
	if err := h.persister.SavePlayer(ctx, player); err != nil {
		slog.Error("failed to save player on restart",
			"character", player.Name(),
			"error", err)
	}

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

	// Phase 8.2: NPC Dialogues — show chat window on simple click for talkable NPC
	if pkt.ActionType == clientpackets.ActionSimpleClick {
		if npc, ok := worldInst.GetNpc(uint32(pkt.ObjectID)); ok {
			npcDef := skilldata.GetNpcDef(npc.TemplateID())
			if npcDef != nil && isNpcTalkable(npcDef.NpcType()) {
				htmlN := h.buildNpcDefaultHtml(npc)
				htmlMsg := serverpackets.NewNpcHtmlMessage(int32(npc.ObjectID()), htmlN)
				htmlData, err := htmlMsg.Write()
				if err != nil {
					slog.Error("failed to serialize NpcHtmlMessage",
						"character", player.Name(),
						"npcID", npc.TemplateID(),
						"error", err)
				} else {
					n = copy(buf[totalBytes:], htmlData)
					totalBytes += n
				}
			}
		}
	}

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
		failedData, err := actionFailed.Write()
		if err != nil {
			return 0, false, fmt.Errorf("serializing ActionFailed: %w", err)
		}
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
		failedData, err := actionFailed.Write()
		if err != nil {
			return 0, false, fmt.Errorf("serializing ActionFailed: %w", err)
		}
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
		failedData, err := actionFailed.Write()
		if err != nil {
			return 0, false, fmt.Errorf("serializing ActionFailed: %w", err)
		}
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
		failedData, err := actionFailed.Write()
		if err != nil {
			return 0, false, fmt.Errorf("serializing ActionFailed: %w", err)
		}
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
		failedData, err := actionFailed.Write()
		if err != nil {
			return 0, false, fmt.Errorf("serializing ActionFailed: %w", err)
		}
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
		failedData, err := actionFailed.Write()
		if err != nil {
			return 0, false, fmt.Errorf("serializing ActionFailed: %w", err)
		}
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
		failedData, err := actionFailed.Write()
		if err != nil {
			return 0, false, fmt.Errorf("serializing ActionFailed: %w", err)
		}
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

// handleRequestMagicSkillUse processes RequestMagicSkillUse packet (opcode 0x2F).
// Client sends this when player uses a skill from the skill bar.
//
// Phase 5.9.4: Cast Flow & Packets.
// Java reference: RequestMagicSkillUse.java
func (h *Handler) handleRequestMagicSkillUse(_ context.Context, client *GameClient, data, buf []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseRequestMagicSkillUse(data)
	if err != nil {
		return 0, false, fmt.Errorf("parsing RequestMagicSkillUse: %w", err)
	}

	if client.State() != ClientStateInGame {
		return 0, true, nil
	}

	player := client.ActivePlayer()
	if player == nil {
		return 0, true, nil
	}

	if skill.CastMgr == nil {
		slog.Warn("CastManager not initialized, ignoring skill use")
		return 0, true, nil
	}

	if err := skill.CastMgr.UseMagic(player, pkt.SkillID, pkt.CtrlPressed, pkt.ShiftPressed); err != nil {
		slog.Debug("skill use failed",
			"player", player.Name(),
			"skillID", pkt.SkillID,
			"error", err)

		// Send ActionFailed
		actionFailed := serverpackets.NewActionFailed()
		failedData, err := actionFailed.Write()
		if err != nil {
			return 0, false, fmt.Errorf("serializing ActionFailed: %w", err)
		}
		n := copy(buf, failedData)
		return n, true, nil
	}

	return 0, true, nil
}

// handleSay2 processes the Say2 packet (opcode 0x38).
// Client sends this when player types a chat message.
//
// Phase 5.11: Chat System.
// Channels supported: GENERAL (radius), SHOUT (all), WHISPER (1 player), TRADE (all).
// Java reference: Say2.java, CreatureSay.java.
func (h *Handler) handleSay2(_ context.Context, client *GameClient, data, buf []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseSay2(data)
	if err != nil {
		return 0, false, fmt.Errorf("parsing Say2: %w", err)
	}

	if client.State() != ClientStateInGame {
		return 0, true, nil
	}

	player := client.ActivePlayer()
	if player == nil {
		return 0, true, nil
	}

	chatType := ChatType(pkt.ChatType)

	// Validate chat type
	if !chatType.IsValid() {
		slog.Warn("invalid chat type",
			"character", player.Name(),
			"chatType", pkt.ChatType,
			"client", client.IP())
		return 0, false, nil // disconnect
	}

	// Validate empty message
	if len(pkt.Text) == 0 {
		slog.Warn("empty chat message",
			"character", player.Name(),
			"chatType", pkt.ChatType,
			"client", client.IP())
		return 0, false, nil // disconnect
	}

	// Validate message length (max 105 chars for non-GM)
	if len([]rune(pkt.Text)) > MaxMessageLength {
		slog.Info("chat message too long",
			"character", player.Name(),
			"length", len([]rune(pkt.Text)),
			"max", MaxMessageLength)

		// Send system message: exceeded chat text limit
		sysMsg := serverpackets.NewSystemMessage(serverpackets.SysMsgYouHaveExceededTheChatTextLimit)
		sysMsgData, err := sysMsg.Write()
		if err != nil {
			return 0, false, fmt.Errorf("serializing SystemMessage: %w", err)
		}
		n := copy(buf, sysMsgData)
		return n, true, nil
	}

	// Route by chat type
	switch chatType {
	case ChatGeneral:
		return h.handleChatGeneral(client, player, pkt.Text, buf)
	case ChatShout:
		return h.handleChatShout(player, pkt.Text, buf)
	case ChatWhisper:
		return h.handleChatWhisper(client, player, pkt.Text, pkt.Target, buf)
	case ChatTrade:
		return h.handleChatTrade(player, pkt.Text, buf)
	default:
		slog.Warn("unsupported chat type",
			"character", player.Name(),
			"chatType", pkt.ChatType)
		return 0, true, nil
	}
}

// handleChatGeneral broadcasts a GENERAL message to nearby visible players.
// Radius is LODNear (~1250 units, same region).
func (h *Handler) handleChatGeneral(client *GameClient, player *model.Player, text string, buf []byte) (int, bool, error) {
	say := serverpackets.NewCreatureSay(int32(player.ObjectID()), int32(ChatGeneral), player.Name(), text)
	sayData, err := say.Write()
	if err != nil {
		return 0, false, fmt.Errorf("serializing CreatureSay GENERAL: %w", err)
	}

	// Send to sender
	n := copy(buf, sayData)

	// Broadcast to nearby visible players
	h.clientManager.BroadcastToVisibleNear(player, sayData, len(sayData))

	return n, true, nil
}

// handleChatShout broadcasts a SHOUT message to all connected players.
func (h *Handler) handleChatShout(player *model.Player, text string, buf []byte) (int, bool, error) {
	say := serverpackets.NewCreatureSay(int32(player.ObjectID()), int32(ChatShout), player.Name(), text)
	sayData, err := say.Write()
	if err != nil {
		return 0, false, fmt.Errorf("serializing CreatureSay SHOUT: %w", err)
	}

	// Send to sender (included in BroadcastToAll but also return in response buffer)
	n := copy(buf, sayData)

	// Broadcast to all players
	h.clientManager.BroadcastToAll(sayData, len(sayData))

	return n, true, nil
}

// handleChatWhisper sends a WHISPER message to a specific player by name.
func (h *Handler) handleChatWhisper(senderClient *GameClient, sender *model.Player, text, targetName string, buf []byte) (int, bool, error) {
	if targetName == "" {
		return 0, true, nil
	}

	targetClient := h.clientManager.FindClientByPlayerName(targetName)
	if targetClient == nil {
		// Target not found — send system message
		sysMsg := serverpackets.NewSystemMessage(serverpackets.SysMsgTargetIsNotFound).AddString(targetName)
		sysMsgData, err := sysMsg.Write()
		if err != nil {
			return 0, false, fmt.Errorf("serializing SystemMessage: %w", err)
		}
		n := copy(buf, sysMsgData)
		return n, true, nil
	}

	targetPlayer := targetClient.ActivePlayer()
	if targetPlayer == nil {
		sysMsg := serverpackets.NewSystemMessage(serverpackets.SysMsgTargetIsNotFound).AddString(targetName)
		sysMsgData, err := sysMsg.Write()
		if err != nil {
			return 0, false, fmt.Errorf("serializing SystemMessage: %w", err)
		}
		n := copy(buf, sysMsgData)
		return n, true, nil
	}

	// Send message to target
	sayToTarget := serverpackets.NewCreatureSay(int32(sender.ObjectID()), int32(ChatWhisper), sender.Name(), text)
	sayToTargetData, err := sayToTarget.Write()
	if err != nil {
		return 0, false, fmt.Errorf("serializing CreatureSay WHISPER: %w", err)
	}

	if err := h.clientManager.SendToPlayer(targetPlayer.ObjectID(), sayToTargetData, len(sayToTargetData)); err != nil {
		slog.Warn("failed to send whisper to target",
			"sender", sender.Name(),
			"target", targetName,
			"error", err)
	}

	// Echo to sender: "-> targetName: text"
	sayToSender := serverpackets.NewCreatureSay(int32(sender.ObjectID()), int32(ChatWhisper), sender.Name(), "->"+targetPlayer.Name()+": "+text)
	sayToSenderData, err := sayToSender.Write()
	if err != nil {
		return 0, false, fmt.Errorf("serializing CreatureSay WHISPER echo: %w", err)
	}

	n := copy(buf, sayToSenderData)
	return n, true, nil
}

// handleChatTrade broadcasts a TRADE message to all connected players.
func (h *Handler) handleChatTrade(player *model.Player, text string, buf []byte) (int, bool, error) {
	say := serverpackets.NewCreatureSay(int32(player.ObjectID()), int32(ChatTrade), player.Name(), text)
	sayData, err := say.Write()
	if err != nil {
		return 0, false, fmt.Errorf("serializing CreatureSay TRADE: %w", err)
	}

	// Send to sender
	n := copy(buf, sayData)

	// Broadcast to all players
	h.clientManager.BroadcastToAll(sayData, len(sayData))

	return n, true, nil
}

// --- Phase 8: NPC Interaction ---

// NPC interaction distance limit (game units).
const maxNpcInteractionDistance = 150

// maxNpcInteractionDistanceSquared is squared for performance (avoid sqrt).
const maxNpcInteractionDistanceSquared = maxNpcInteractionDistance * maxNpcInteractionDistance

// isNpcTalkable returns true for NPC types that can show dialog.
// Phase 8.2: NPC Dialogues.
func isNpcTalkable(npcType string) bool {
	switch npcType {
	case "Folk", "Merchant", "Guard", "Teleporter", "Warehouse":
		return true
	default:
		return false
	}
}

// buildNpcDefaultHtml builds default HTML dialog for NPC.
// Shows available actions based on NPC type (Shop, Quest, Teleport, etc.).
//
// Phase 8.2: NPC Dialogues.
func (h *Handler) buildNpcDefaultHtml(npc *model.Npc) string {
	templateID := npc.TemplateID()
	npcDef := skilldata.GetNpcDef(templateID)
	if npcDef == nil {
		return "<html><body>I have nothing to say.</body></html>"
	}

	var sb strings.Builder
	sb.WriteString("<html><body>")
	sb.WriteString(npcDef.Name())
	sb.WriteString(":<br>")

	npcType := npcDef.NpcType()

	// Check if this NPC has buylists → show Shop link
	if buylists := skilldata.GetBuylistsByNpc(templateID); len(buylists) > 0 {
		sb.WriteString("<a action=\"bypass -h npc_")
		sb.WriteString(strconv.FormatUint(uint64(npc.ObjectID()), 10))
		sb.WriteString("_Shop\">Shop</a><br>")
	}

	// Merchant type NPCs always show sell option
	if npcType == "Merchant" {
		sb.WriteString("<a action=\"bypass -h npc_")
		sb.WriteString(strconv.FormatUint(uint64(npc.ObjectID()), 10))
		sb.WriteString("_Sell\">Sell</a><br>")
	}

	// Teleporter NPCs — placeholder for Phase 11
	if npcType == "Teleporter" {
		sb.WriteString("I can teleport you. (Coming soon)<br>")
	}

	// Warehouse NPCs — placeholder for Phase 9
	if npcType == "Warehouse" {
		sb.WriteString("I can store your items. (Coming soon)<br>")
	}

	sb.WriteString("</body></html>")
	return sb.String()
}

// handleRequestBypassToServer processes RequestBypassToServer packet (opcode 0x21).
// Client sends this when player clicks a link in NPC HTML dialog.
//
// Bypass routing:
//   - "npc_%objectId%_Shop" → send BuyList
//   - "npc_%objectId%_Sell" → send SellList
//   - "_bbshome", "_bbsgetfav" → Community Board (Phase 11+)
//
// Phase 8.2: NPC Dialogues.
func (h *Handler) handleRequestBypassToServer(_ context.Context, client *GameClient, data, buf []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseRequestBypassToServer(data)
	if err != nil {
		return 0, false, fmt.Errorf("parsing RequestBypassToServer: %w", err)
	}

	player := client.ActivePlayer()
	if player == nil {
		return 0, true, nil
	}

	bypass := pkt.Bypass
	slog.Debug("bypass received", "character", player.Name(), "bypass", bypass)

	// Route NPC bypass commands: "npc_<objectID>_<command>"
	if strings.HasPrefix(bypass, "npc_") {
		return h.handleNpcBypass(player, bypass, buf)
	}

	// Community Board (placeholder)
	if strings.HasPrefix(bypass, "_bbs") {
		slog.Debug("community board bypass (not implemented)", "bypass", bypass)
		return 0, true, nil
	}

	slog.Warn("unknown bypass command", "bypass", bypass, "character", player.Name())
	return 0, true, nil
}

// handleNpcBypass routes NPC-specific bypass commands.
// Format: "npc_<objectID>_<command>"
//
// Phase 8.2/8.3: NPC Dialogues + Shops.
func (h *Handler) handleNpcBypass(player *model.Player, bypass string, buf []byte) (int, bool, error) {
	// Parse: "npc_<objectID>_<command>"
	parts := strings.SplitN(bypass, "_", 3)
	if len(parts) < 3 {
		slog.Warn("malformed npc bypass", "bypass", bypass)
		return 0, true, nil
	}

	npcObjectID, err := strconv.ParseUint(parts[1], 10, 32)
	if err != nil {
		slog.Warn("invalid npc objectID in bypass", "bypass", bypass, "error", err)
		return 0, true, nil
	}

	command := parts[2]
	worldInst := world.Instance()

	// Validate NPC exists and is within interaction distance
	npc, ok := worldInst.GetNpc(uint32(npcObjectID))
	if !ok {
		slog.Warn("bypass target NPC not found", "objectID", npcObjectID)
		return 0, true, nil
	}

	playerLoc := player.Location()
	npcLoc := npc.Location()
	distSq := playerLoc.DistanceSquared(npcLoc)
	if distSq > maxNpcInteractionDistanceSquared {
		slog.Debug("NPC too far for bypass interaction",
			"character", player.Name(),
			"npcID", npc.TemplateID(),
			"distSq", distSq)
		af := serverpackets.NewActionFailed()
		afData, _ := af.Write()
		n := copy(buf, afData)
		return n, true, nil
	}

	switch command {
	case "Shop":
		return h.handleNpcShop(player, npc, buf)
	case "Sell":
		return h.handleNpcSell(player, npc, buf)
	default:
		slog.Debug("unhandled NPC bypass command",
			"command", command,
			"npcID", npc.TemplateID(),
			"character", player.Name())
		return 0, true, nil
	}
}

// handleNpcShop sends BuyList packet for NPC's shop.
//
// Phase 8.3: NPC Shops.
func (h *Handler) handleNpcShop(player *model.Player, npc *model.Npc, buf []byte) (int, bool, error) {
	templateID := npc.TemplateID()

	buylistIDs := skilldata.GetBuylistsByNpc(templateID)
	if len(buylistIDs) == 0 {
		slog.Debug("NPC has no buylists", "npcID", templateID)
		af := serverpackets.NewActionFailed()
		afData, _ := af.Write()
		n := copy(buf, afData)
		return n, true, nil
	}

	// Use the first buylist for this NPC
	listID := buylistIDs[0]

	// Build products list
	products := buildBuyListProducts(listID)

	playerAdena := player.Inventory().GetAdena()
	buyListPkt := serverpackets.NewBuyList(playerAdena, listID, products)

	pktData, err := buyListPkt.Write()
	if err != nil {
		return 0, false, fmt.Errorf("serializing BuyList: %w", err)
	}

	slog.Debug("sent BuyList",
		"character", player.Name(),
		"npcID", templateID,
		"listID", listID,
		"products", len(products))

	n := copy(buf, pktData)
	return n, true, nil
}

// handleNpcSell sends SellList packet with player's sellable items.
//
// Phase 8.3: NPC Shops.
func (h *Handler) handleNpcSell(player *model.Player, npc *model.Npc, buf []byte) (int, bool, error) {
	sellableItems := player.Inventory().GetSellableItems()

	var items []serverpackets.SellListItem
	for _, item := range sellableItems {
		itemDef := skilldata.GetItemDef(item.ItemID())
		sellPrice := int64(0)
		if itemDef != nil {
			sellPrice = itemDef.Price() / 2 // Sell at 50% of base price
		}
		if sellPrice <= 0 {
			continue // Skip items with no sell value
		}

		items = append(items, serverpackets.SellListItem{
			Item:      item,
			SellPrice: sellPrice,
		})
	}

	playerAdena := player.Inventory().GetAdena()
	sellListPkt := serverpackets.NewSellList(playerAdena, items)

	pktData, err := sellListPkt.Write()
	if err != nil {
		return 0, false, fmt.Errorf("serializing SellList: %w", err)
	}

	slog.Debug("sent SellList",
		"character", player.Name(),
		"npcID", npc.TemplateID(),
		"items", len(items))

	n := copy(buf, pktData)
	return n, true, nil
}

// handleRequestBuyItem processes RequestBuyItem packet (opcode 0x1F).
// Player confirms purchase of items from NPC shop.
//
// Phase 8.3: NPC Shops.
func (h *Handler) handleRequestBuyItem(_ context.Context, client *GameClient, data, buf []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseRequestBuyItem(data)
	if err != nil {
		return 0, false, fmt.Errorf("parsing RequestBuyItem: %w", err)
	}

	player := client.ActivePlayer()
	if player == nil {
		return 0, true, nil
	}

	// Validate buylist exists
	if skilldata.GetBuylistProducts(pkt.ListID) == nil {
		slog.Warn("invalid buylist ID", "listID", pkt.ListID, "character", player.Name())
		af := serverpackets.NewActionFailed()
		afData, _ := af.Write()
		n := copy(buf, afData)
		return n, true, nil
	}

	// Calculate total cost and validate items
	var totalCost int64
	for _, entry := range pkt.Items {
		product := skilldata.FindProductInBuylist(pkt.ListID, entry.ItemID)
		if product == nil {
			slog.Warn("item not in buylist",
				"itemID", entry.ItemID,
				"listID", pkt.ListID,
				"character", player.Name())
			af := serverpackets.NewActionFailed()
			afData, _ := af.Write()
			n := copy(buf, afData)
			return n, true, nil
		}

		price := product.Price
		if price <= 0 {
			// Use item's base price from item data
			if itemDef := skilldata.GetItemDef(entry.ItemID); itemDef != nil {
				price = itemDef.Price()
			}
		}

		totalCost += price * int64(entry.Count)
	}

	// Check Adena
	playerAdena := player.Inventory().GetAdena()
	if playerAdena < totalCost {
		slog.Debug("not enough adena for purchase",
			"character", player.Name(),
			"have", playerAdena,
			"need", totalCost)
		af := serverpackets.NewActionFailed()
		afData, _ := af.Write()
		n := copy(buf, afData)
		return n, true, nil
	}

	// Deduct Adena
	if err := player.Inventory().RemoveAdena(int32(totalCost)); err != nil {
		slog.Error("failed to remove adena", "error", err, "character", player.Name())
		af := serverpackets.NewActionFailed()
		afData, _ := af.Write()
		n := copy(buf, afData)
		return n, true, nil
	}

	// Create items
	for _, entry := range pkt.Items {
		tmpl := db.ItemDefToTemplate(entry.ItemID)
		if tmpl == nil {
			slog.Error("item template not found", "itemID", entry.ItemID)
			continue
		}

		objectID := world.IDGenerator().NextItemID()
		item, err := model.NewItem(objectID, entry.ItemID, int64(player.CharacterID()), entry.Count, tmpl)
		if err != nil {
			slog.Error("failed to create item",
				"itemID", entry.ItemID,
				"error", err)
			continue
		}

		if err := player.Inventory().AddItem(item); err != nil {
			slog.Error("failed to add item to inventory",
				"itemID", entry.ItemID,
				"error", err)
			continue
		}

		slog.Debug("item purchased",
			"character", player.Name(),
			"itemID", entry.ItemID,
			"count", entry.Count)
	}

	// Send updated inventory
	totalBytes := 0
	invPkt := serverpackets.NewInventoryItemList(player.Inventory().GetItems())
	invPkt.ShowWindow = false
	invData, err := invPkt.Write()
	if err != nil {
		slog.Error("failed to serialize InventoryItemList", "error", err)
	} else {
		n := copy(buf[totalBytes:], invData)
		totalBytes += n
	}

	slog.Info("purchase completed",
		"character", player.Name(),
		"items", len(pkt.Items),
		"totalCost", totalCost)

	return totalBytes, true, nil
}

// handleRequestSellItem processes RequestSellItem packet (opcode 0x1E).
// Player confirms selling items to NPC.
//
// Phase 8.3: NPC Shops.
func (h *Handler) handleRequestSellItem(_ context.Context, client *GameClient, data, buf []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseRequestSellItem(data)
	if err != nil {
		return 0, false, fmt.Errorf("parsing RequestSellItem: %w", err)
	}

	player := client.ActivePlayer()
	if player == nil {
		return 0, true, nil
	}

	// Calculate total sell value and validate items
	var totalValue int64
	type sellEntry struct {
		item      *model.Item
		sellPrice int64
		count     int32
	}
	var entries []sellEntry

	for _, entry := range pkt.Items {
		item := player.Inventory().GetItem(uint32(entry.ObjectID))
		if item == nil {
			slog.Warn("sell: item not in inventory",
				"objectID", entry.ObjectID,
				"character", player.Name())
			af := serverpackets.NewActionFailed()
			afData, _ := af.Write()
			n := copy(buf, afData)
			return n, true, nil
		}

		// Cannot sell equipped items
		if item.IsEquipped() {
			slog.Debug("sell: cannot sell equipped item",
				"objectID", entry.ObjectID,
				"character", player.Name())
			af := serverpackets.NewActionFailed()
			afData, _ := af.Write()
			n := copy(buf, afData)
			return n, true, nil
		}

		// Cannot sell Adena
		if item.ItemID() == model.AdenaItemID {
			continue
		}

		// Validate count
		if entry.Count > item.Count() {
			slog.Warn("sell: not enough items",
				"objectID", entry.ObjectID,
				"have", item.Count(),
				"want", entry.Count,
				"character", player.Name())
			af := serverpackets.NewActionFailed()
			afData, _ := af.Write()
			n := copy(buf, afData)
			return n, true, nil
		}

		// Calculate sell price
		itemDef := skilldata.GetItemDef(item.ItemID())
		sellPrice := int64(0)
		if itemDef != nil {
			sellPrice = itemDef.Price() / 2
		}

		totalValue += sellPrice * int64(entry.Count)
		entries = append(entries, sellEntry{
			item:      item,
			sellPrice: sellPrice,
			count:     entry.Count,
		})
	}

	// Process sales
	for _, se := range entries {
		if se.count >= se.item.Count() {
			// Remove entire item
			player.Inventory().RemoveItem(se.item.ObjectID())
		} else {
			// Decrease count (stackable items)
			if err := se.item.SetCount(se.item.Count() - se.count); err != nil {
				slog.Error("failed to decrease item count", "error", err)
			}
		}
	}

	// Add Adena
	if totalValue > 0 {
		if err := player.Inventory().AddAdena(int32(totalValue)); err != nil {
			// If no Adena item exists yet, create one
			tmpl := db.ItemDefToTemplate(model.AdenaItemID)
			if tmpl != nil {
				objectID := world.IDGenerator().NextItemID()
				adenaItem, err := model.NewItem(objectID, model.AdenaItemID, int64(player.CharacterID()), int32(totalValue), tmpl)
				if err != nil {
					slog.Error("failed to create adena item", "error", err)
				} else {
					if err := player.Inventory().AddItem(adenaItem); err != nil {
						slog.Error("failed to add adena to inventory", "error", err)
					}
				}
			}
		}
	}

	// Send updated inventory
	totalBytes := 0
	invPkt := serverpackets.NewInventoryItemList(player.Inventory().GetItems())
	invPkt.ShowWindow = false
	invData, err := invPkt.Write()
	if err != nil {
		slog.Error("failed to serialize InventoryItemList", "error", err)
	} else {
		n := copy(buf[totalBytes:], invData)
		totalBytes += n
	}

	slog.Info("items sold",
		"character", player.Name(),
		"items", len(entries),
		"totalValue", totalValue)

	return totalBytes, true, nil
}

// buildBuyListProducts converts buylist products from data package into
// BuyListProduct slice for the BuyList server packet.
//
// Phase 8.3: NPC Shops.
func buildBuyListProducts(listID int32) []serverpackets.BuyListProduct {
	dataProducts := skilldata.GetBuylistProducts(listID)
	if dataProducts == nil {
		return nil
	}

	products := make([]serverpackets.BuyListProduct, 0, len(dataProducts))

	for _, dp := range dataProducts {
		itemDef := skilldata.GetItemDef(dp.ItemID)
		if itemDef == nil {
			continue
		}

		price := dp.Price
		if price <= 0 {
			price = itemDef.Price()
		}

		tmpl := db.ItemDefToTemplate(dp.ItemID)
		var type1, type2 int16
		var bodyPart int32
		if tmpl != nil {
			type1, type2, bodyPart = getItemPacketTypes(tmpl)
		}

		products = append(products, serverpackets.BuyListProduct{
			ItemID:       dp.ItemID,
			Price:        price,
			Count:        dp.Count,
			RestockDelay: dp.RestockDelay,
			Type1:        type1,
			Type2:        type2,
			BodyPart:     bodyPart,
			Weight:       itemDef.Weight(),
		})
	}

	return products
}

// getItemPacketTypes returns type1, type2, and bodyPart mask for item template.
// Reuses logic from InventoryItemList for consistency.
//
// Phase 8.3: NPC Shops.
func getItemPacketTypes(tmpl *model.ItemTemplate) (int16, int16, int32) {
	var type1, type2 int16
	var bodyPart int32

	switch tmpl.Type {
	case model.ItemTypeWeapon:
		type1 = 0
		type2 = 0
		bodyPart = 0x4000 // rhand
	case model.ItemTypeArmor:
		if tmpl.BodyPart == model.ArmorSlotNeck ||
			tmpl.BodyPart == model.ArmorSlotEar ||
			tmpl.BodyPart == model.ArmorSlotFinger {
			type1 = 2
			type2 = 2
		} else {
			type1 = 1
			type2 = 1
		}
		switch tmpl.BodyPart {
		case model.ArmorSlotChest:
			bodyPart = 0x0400
		case model.ArmorSlotLegs:
			bodyPart = 0x0800
		case model.ArmorSlotHead:
			bodyPart = 0x0040
		case model.ArmorSlotFeet:
			bodyPart = 0x1000
		case model.ArmorSlotGloves:
			bodyPart = 0x0200
		case model.ArmorSlotNeck:
			bodyPart = 0x0008
		case model.ArmorSlotEar:
			bodyPart = 0x0006
		case model.ArmorSlotFinger:
			bodyPart = 0x0030
		}
	default:
		type1 = 5
		type2 = 5
	}

	return type1, type2, bodyPart
}
