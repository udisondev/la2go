package gameserver

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/udisondev/la2go/internal/game/sevensigns"
	"github.com/udisondev/la2go/internal/gameserver/clientpackets"
	"github.com/udisondev/la2go/internal/gameserver/serverpackets"
	"github.com/udisondev/la2go/internal/model"
)

// handleRequestSSQStatus processes RequestSSQStatus (C2S 0xC7).
// Sends the Seven Signs status page to the client.
//
// Phase 25: Seven Signs.
func (h *Handler) handleRequestSSQStatus(_ context.Context, client *GameClient, data, buf []byte) (int, bool, error) {
	player := client.ActivePlayer()
	if player == nil || h.sevenSignsMgr == nil {
		return 0, true, nil
	}

	pkt, err := clientpackets.ParseRequestSSQStatus(data)
	if err != nil {
		return 0, true, fmt.Errorf("parsing RequestSSQStatus: %w", err)
	}

	page := pkt.Page
	if page < 1 || page > 4 {
		slog.Debug("ssq status: invalid page", "page", page, "character", player.Name())
		return 0, true, nil
	}

	// Page 4 blocked during Seal Validation and Results periods.
	period := h.sevenSignsMgr.CurrentPeriod()
	if page == 4 && (period == sevensigns.PeriodSealValidation || period == sevensigns.PeriodResults) {
		slog.Debug("ssq status: page 4 blocked during this period",
			"period", period,
			"character", player.Name())
		return 0, true, nil
	}

	status := h.sevenSignsMgr.Status()
	pd := h.sevenSignsMgr.PlayerData(player.CharacterID())

	ssq := h.buildSSQStatus(page, status, pd)
	pktData, err := ssq.Write()
	if err != nil {
		return 0, true, fmt.Errorf("writing SSQStatus: %w", err)
	}

	n := copy(buf, pktData)
	return n, true, nil
}

// buildSSQStatus constructs the SSQStatus server packet from manager state.
func (h *Handler) buildSSQStatus(page byte, status sevensigns.Status, pd *sevensigns.PlayerData) *serverpackets.SSQStatus {
	ssq := &serverpackets.SSQStatus{
		Page:          page,
		CurrentPeriod: status.ActivePeriod,
		CurrentCycle:  status.CurrentCycle,
	}

	switch page {
	case 1:
		if pd != nil {
			ssq.PlayerCabal = pd.Cabal
			ssq.PlayerSeal = pd.Seal
			ssq.PlayerStones = int32(pd.ContributionScore)
			ssq.PlayerAdena = int32(pd.AncientAdena)
		}
		ssq.DawnStoneScore = float64(status.DawnStoneScore)
		ssq.DuskStoneScore = float64(status.DuskStoneScore)
		ssq.DawnFestival = status.DawnFestivalScore
		ssq.DuskFestival = status.DuskFestivalScore

	case 3:
		for i := int32(1); i <= 3; i++ {
			seal := sevensigns.Seal(i)
			ssq.SealOwners[i] = status.SealOwner(seal)
		}
		ssq.TotalDawnMembers = status.AvariceDawnScore + status.GnosisDawnScore + status.StrifeDawnScore
		ssq.TotalDuskMembers = status.AvariceDuskScore + status.GnosisDuskScore + status.StrifeDuskScore
		ssq.DawnSealMembers[1] = status.AvariceDawnScore
		ssq.DawnSealMembers[2] = status.GnosisDawnScore
		ssq.DawnSealMembers[3] = status.StrifeDawnScore
		ssq.DuskSealMembers[1] = status.AvariceDuskScore
		ssq.DuskSealMembers[2] = status.GnosisDuskScore
		ssq.DuskSealMembers[3] = status.StrifeDuskScore

	case 4:
		ssq.WinnerCabal = status.PreviousWinner
		for i := int32(1); i <= 3; i++ {
			seal := sevensigns.Seal(i)
			ssq.SealOwners[i] = status.SealOwner(seal)
		}
	}

	return ssq
}

// handleSevenSignsBypass processes Seven Signs NPC bypass commands.
// Returns (n, handled, error). If handled==false, the caller should try other bypass handlers.
//
// Bypass commands:
//   - SSQJoin <cabal>       — join Dawn or Dusk cabal
//   - SSQSeal <sealID>      — choose a seal
//   - SSQContribute <count> — contribute stones
//   - SSQStatus             — show status page 1
//
// Phase 25: Seven Signs.
func (h *Handler) handleSevenSignsBypass(player *model.Player, npc *model.Npc, cmdName, cmdArg string, buf []byte) (int, bool, error) {
	if h.sevenSignsMgr == nil {
		return 0, false, nil
	}

	switch cmdName {
	case "SSQJoin":
		return h.bypassSSQJoin(player, cmdArg, buf)
	case "SSQSeal":
		return h.bypassSSQSeal(player, cmdArg, buf)
	case "SSQContribute":
		return h.bypassSSQContribute(player, cmdArg, buf)
	case "SSQStatus":
		return h.bypassSSQStatus(player, buf)
	default:
		return 0, false, nil
	}
}

// bypassSSQJoin handles joining a cabal during Recruitment period.
func (h *Handler) bypassSSQJoin(player *model.Player, arg string, buf []byte) (int, bool, error) {
	cabal := sevensigns.ParseCabal(strings.ToLower(arg))
	if cabal == sevensigns.CabalNull {
		slog.Debug("ssq join: invalid cabal", "arg", arg, "character", player.Name())
		return 0, true, nil
	}

	if !h.sevenSignsMgr.JoinCabal(player.CharacterID(), cabal) {
		slog.Debug("ssq join failed",
			"character", player.Name(),
			"cabal", sevensigns.CabalShortName(cabal))
		return 0, true, nil
	}

	slog.Info("player joined Seven Signs cabal",
		"character", player.Name(),
		"cabal", sevensigns.CabalShortName(cabal))

	return h.sendSSQPage(player, 1, buf)
}

// bypassSSQSeal handles choosing a seal.
func (h *Handler) bypassSSQSeal(player *model.Player, arg string, buf []byte) (int, bool, error) {
	sealID, err := strconv.ParseInt(arg, 10, 32)
	if err != nil || sealID < 1 || sealID > 3 {
		slog.Debug("ssq seal: invalid seal ID", "arg", arg, "character", player.Name())
		return 0, true, nil
	}

	seal := sevensigns.Seal(sealID)
	if !h.sevenSignsMgr.ChooseSeal(player.CharacterID(), seal) {
		slog.Debug("ssq seal failed",
			"character", player.Name(),
			"seal", sealID)
		return 0, true, nil
	}

	return h.sendSSQPage(player, 1, buf)
}

// bypassSSQContribute handles contributing stones.
// For simplicity, treats the count as blue stones contribution.
func (h *Handler) bypassSSQContribute(player *model.Player, arg string, buf []byte) (int, bool, error) {
	count, err := strconv.ParseInt(arg, 10, 32)
	if err != nil || count <= 0 {
		slog.Debug("ssq contribute: invalid count", "arg", arg, "character", player.Name())
		return 0, true, nil
	}

	contrib := h.sevenSignsMgr.ContributeStones(player.CharacterID(), int32(count), 0, 0)
	if contrib == 0 {
		slog.Debug("ssq contribute: no contribution",
			"character", player.Name(),
			"count", count)
		return 0, true, nil
	}

	return h.sendSSQPage(player, 1, buf)
}

// bypassSSQStatus sends SSQ status page 1.
func (h *Handler) bypassSSQStatus(player *model.Player, buf []byte) (int, bool, error) {
	return h.sendSSQPage(player, 1, buf)
}

// sendSSQPage builds and sends an SSQ status page.
func (h *Handler) sendSSQPage(player *model.Player, page byte, buf []byte) (int, bool, error) {
	status := h.sevenSignsMgr.Status()
	pd := h.sevenSignsMgr.PlayerData(player.CharacterID())

	ssq := h.buildSSQStatus(page, status, pd)
	pktData, err := ssq.Write()
	if err != nil {
		return 0, true, fmt.Errorf("writing SSQStatus: %w", err)
	}

	n := copy(buf, pktData)
	return n, true, nil
}
