package gameserver

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/udisondev/la2go/internal/game/hall"
	"github.com/udisondev/la2go/internal/gameserver/serverpackets"
)

// handleRequestClanHallInfo sends AgitDecoInfo for the player's clan hall.
//
// Phase 22: Clan Halls.
func (h *Handler) handleRequestClanHallInfo(_ context.Context, client *GameClient, buf []byte) (int, error) {
	player := client.ActivePlayer()
	if player == nil || h.hallTable == nil {
		return 0, nil
	}

	clanID := player.ClanID()
	if clanID == 0 {
		return 0, nil
	}

	ch := h.hallTable.HallByOwner(clanID)
	if ch == nil {
		return 0, nil
	}

	info := h.buildAgitDecoInfo(ch)
	pktData, err := info.Write()
	if err != nil {
		return 0, fmt.Errorf("writing AgitDecoInfo: %w", err)
	}
	n := copy(buf, pktData)
	return n, nil
}

// handleHallBypass processes clan hall NPC bypass commands.
// Returns true if the command was handled.
//
// Bypass commands:
//   - ClanHallDecoInfo <hallID>  — sends AgitDecoInfo
//   - ClanHallFunctions <hallID> — lists active functions
//
// Phase 22: Clan Halls.
func (h *Handler) handleHallBypass(client *GameClient, cmdName string, args []string, buf []byte) (int, bool, error) {
	if h.hallTable == nil {
		return 0, false, nil
	}

	switch cmdName {
	case "ClanHallDecoInfo":
		return h.bypassClanHallDecoInfo(client, args, buf)
	default:
		return 0, false, nil
	}
}

// bypassClanHallDecoInfo sends AgitDecoInfo for a specific hall.
func (h *Handler) bypassClanHallDecoInfo(client *GameClient, args []string, buf []byte) (int, bool, error) {
	player := client.ActivePlayer()
	if player == nil {
		return 0, true, nil
	}

	if len(args) == 0 {
		return 0, true, nil
	}

	var hallID int32
	if _, err := fmt.Sscan(args[0], &hallID); err != nil {
		slog.Debug("hall bypass: invalid hall ID", "arg", args[0])
		return 0, true, nil
	}

	ch := h.hallTable.Hall(hallID)
	if ch == nil {
		slog.Debug("hall bypass: hall not found", "hall_id", hallID)
		return 0, true, nil
	}

	// Только владелец клан-холла может просматривать декорации.
	if player.ClanID() == 0 || ch.OwnerClanID() != player.ClanID() {
		return 0, true, nil
	}

	info := h.buildAgitDecoInfo(ch)
	pktData, err := info.Write()
	if err != nil {
		return 0, true, fmt.Errorf("writing AgitDecoInfo: %w", err)
	}
	n := copy(buf, pktData)
	return n, true, nil
}

// buildAgitDecoInfo constructs an AgitDecoInfo packet from a clan hall.
func (h *Handler) buildAgitDecoInfo(ch *hall.ClanHall) *serverpackets.AgitDecoInfo {
	return &serverpackets.AgitDecoInfo{
		HallID:       ch.ID(),
		HPLevel:      ch.FunctionLevel(hall.FuncRestoreHP),
		MPLevel:      ch.FunctionLevel(hall.FuncRestoreMP),
		ExpLevel:     ch.FunctionLevel(hall.FuncRestoreExp),
		SPLevel:      0, // Always 0 in Interlude
		TeleLevel:    ch.FunctionLevel(hall.FuncTeleport),
		CurtainLevel: 0, // Not implemented in Interlude
		FrontLevel:   0, // Not implemented in Interlude
		ItemLevel:    ch.FunctionLevel(hall.FuncItemCreate),
		SupportLevel: ch.FunctionLevel(hall.FuncSupport),
	}
}
