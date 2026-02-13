package gameserver

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/udisondev/la2go/internal/game/bbs"
	"github.com/udisondev/la2go/internal/gameserver/clientpackets"
	"github.com/udisondev/la2go/internal/gameserver/serverpackets"
	"github.com/udisondev/la2go/internal/model"
)

// handleRequestShowBoard processes RequestShowBoard (C2S 0x57).
// Client pressed ALT+B — open Community Board with default home page.
//
// Flow:
//  1. Parse packet (unused int32)
//  2. Get default BBS command (_bbshome)
//  3. Generate HTML via bbsHandler
//  4. Split into chunks and send ShowBoard packets
//
// Phase 30: Community Board.
func (h *Handler) handleRequestShowBoard(_ context.Context, client *GameClient, data, _ []byte) (int, bool, error) {
	if _, err := clientpackets.ParseRequestShowBoard(data); err != nil {
		return 0, true, fmt.Errorf("parsing RequestShowBoard: %w", err)
	}

	player := client.ActivePlayer()
	if player == nil {
		return 0, true, nil
	}

	if h.bbsHandler == nil {
		slog.Debug("community board not initialized")
		return 0, true, nil
	}

	html := h.bbsHandler.HandleCommand(bbs.DefaultCommand, player.CharacterID(), player.Name())
	h.sendShowBoard(player, html)
	return 0, true, nil
}

// handleRequestBBSwrite processes RequestBBSwrite (C2S 0x22).
// Client submitted a form on the Community Board.
//
// Flow:
//  1. Parse packet (URL + 5 args)
//  2. Route to bbsHandler.HandleWrite
//  3. Send response HTML via ShowBoard
//
// Phase 30: Community Board.
func (h *Handler) handleRequestBBSwrite(_ context.Context, client *GameClient, data, _ []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseRequestBBSwrite(data)
	if err != nil {
		return 0, true, fmt.Errorf("parsing RequestBBSwrite: %w", err)
	}

	player := client.ActivePlayer()
	if player == nil {
		return 0, true, nil
	}

	if h.bbsHandler == nil {
		return 0, true, nil
	}

	html := h.bbsHandler.HandleWrite(player.CharacterID(), player.Name(), pkt.URL, pkt.Args)
	if html != "" {
		h.sendShowBoard(player, html)
	}
	return 0, true, nil
}

// handleBBSBypass processes Community Board bypass commands (_bbs*, bbs_*).
// Called from handleRequestBypassToServer.
//
// Phase 30: Community Board.
func (h *Handler) handleBBSBypass(client *GameClient, player *model.Player, bypass string, _ []byte) (int, bool, error) {
	if h.bbsHandler == nil {
		slog.Debug("community board not initialized")
		return 0, true, nil
	}

	html := h.bbsHandler.HandleCommand(bypass, player.CharacterID(), player.Name())
	if html == "" {
		slog.Debug("community board: unhandled bypass", "bypass", bypass)
		return 0, true, nil
	}

	h.sendShowBoard(player, html)
	return 0, true, nil
}

// sendShowBoard splits HTML into chunks and sends ShowBoard packets.
// Java: HtmlUtil.sendCBHtml — 3 chunks of 4090 bytes each.
func (h *Handler) sendShowBoard(player *model.Player, html string) {
	chunks := bbs.SplitHTML(html)

	for _, chunk := range chunks {
		var pkt serverpackets.ShowBoard
		if chunk.Content != "" {
			pkt = serverpackets.NewShowBoard(chunk.ID, chunk.Content)
		} else {
			pkt = serverpackets.NewShowBoardHide()
		}

		pktData, err := pkt.Write()
		if err != nil {
			slog.Error("writing ShowBoard",
				"chunkID", chunk.ID, "error", err)
			continue
		}

		if err := h.clientManager.SendToPlayer(player.ObjectID(), pktData, len(pktData)); err != nil {
			slog.Error("sending ShowBoard",
				"player", player.Name(),
				"chunkID", chunk.ID,
				"error", err)
			return
		}
	}

	slog.Debug("community board sent",
		"player", player.Name(),
		"chunks", len(chunks))
}
