package gameserver

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/udisondev/la2go/internal/data"
	"github.com/udisondev/la2go/internal/gameserver/clientpackets"
	"github.com/udisondev/la2go/internal/gameserver/serverpackets"
)

// ─── Miscellaneous handlers ─────────────────────────────────────────────────

// handleRequestRecordInfo processes 0xCF — client requests UserInfo refresh.
func (h *Handler) handleRequestRecordInfo(_ context.Context, client *GameClient, _, buf []byte) (int, bool, error) {
	player := client.ActivePlayer()
	if player == nil {
		return 0, true, nil
	}

	// Send UserInfo to refresh client state
	userInfo := serverpackets.NewUserInfo(player)
	uiData, err := userInfo.Write()
	if err != nil {
		slog.Error("serializing UserInfo for record info", "error", err)
		return 0, true, nil
	}
	n := copy(buf, uiData)
	return n, true, nil
}

// handleRequestShowMiniMap processes 0xCD — client opens minimap (no-op server-side).
func (h *Handler) handleRequestShowMiniMap(_ context.Context, _ *GameClient, _ []byte, _ []byte) (int, bool, error) {
	return 0, true, nil
}

// handleObserverReturn processes 0xB8 — exit observer mode.
// Stub until observer/spectator system is implemented.
func (h *Handler) handleObserverReturn(_ context.Context, _ *GameClient, _ []byte, _ []byte) (int, bool, error) {
	slog.Debug("observer return (stub)")
	return 0, true, nil
}

// handleRequestRecipeBookDestroy processes 0xAD — delete recipe from book.
func (h *Handler) handleRequestRecipeBookDestroy(_ context.Context, client *GameClient, dataBytes, buf []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseRequestRecipeBookDestroy(dataBytes)
	if err != nil {
		return 0, false, fmt.Errorf("parsing RequestRecipeBookDestroy: %w", err)
	}

	player := client.ActivePlayer()
	if player == nil {
		return 0, false, fmt.Errorf("no active player for recipe book destroy")
	}

	recipeID := pkt.RecipeListID
	if recipeID <= 0 {
		return 0, true, nil
	}

	// Determine recipe type
	recipeTmpl := data.GetRecipeTemplate(recipeID)
	isDwarven := false
	if recipeTmpl != nil {
		isDwarven = recipeTmpl.IsDwarven
	}

	if err := player.ForgetRecipe(recipeID, isDwarven); err != nil {
		slog.Debug("recipe book destroy: not learned",
			"player", player.Name(),
			"recipeID", recipeID)
		return 0, true, nil
	}

	slog.Info("player deleted recipe",
		"player", player.Name(),
		"recipeID", recipeID,
		"isDwarven", isDwarven)

	// Send updated recipe book
	recipeIDs := player.GetRecipeBook(isDwarven)
	resp := &serverpackets.RecipeBookItemList{
		IsDwarvenCraft: isDwarven,
		MaxMP:          player.MaxMP(),
		RecipeIDs:      recipeIDs,
	}
	respData, writeErr := resp.Write()
	if writeErr != nil {
		return 0, false, fmt.Errorf("serializing RecipeBookItemList: %w", writeErr)
	}
	n := copy(buf, respData)
	return n, true, nil
}

// handleRequestEvaluate processes 0xB9 — give recommendation to player.
// Stub: acknowledged, needs recommendation system.
func (h *Handler) handleRequestEvaluate(_ context.Context, _ *GameClient, _ []byte, _ []byte) (int, bool, error) {
	slog.Debug("evaluate (stub)")
	return 0, true, nil
}

// handleRequestPartyMatchConfig processes 0x6F — party matching window config.
func (h *Handler) handleRequestPartyMatchConfig(_ context.Context, _ *GameClient, _ []byte, _ []byte) (int, bool, error) {
	return 0, true, nil
}

// handleRequestPartyMatchList processes 0x70 — party matching list.
func (h *Handler) handleRequestPartyMatchList(_ context.Context, _ *GameClient, _ []byte, _ []byte) (int, bool, error) {
	return 0, true, nil
}

// handleRequestPartyMatchDetail processes 0x71 — party matching detail.
func (h *Handler) handleRequestPartyMatchDetail(_ context.Context, _ *GameClient, _ []byte, _ []byte) (int, bool, error) {
	return 0, true, nil
}
