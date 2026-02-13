package gameserver

import (
	"context"
	"log/slog"
	"time"

	"github.com/udisondev/la2go/internal/model"
	"github.com/udisondev/la2go/internal/world"
)

const (
	// CombatTime is the delay before removing disconnected player from world if in combat.
	// Prevents "combat logging" — instant disappearance when disconnecting during fight.
	//
	// Reference: L2J_Mobius AttackStanceTaskManager.COMBAT_TIME (15000ms)
	CombatTime = 15 * time.Second
)

// OfflineTradeService defines the interface for offline trade operations.
// Used by OnDisconnection to check if a player should enter offline trade mode.
type OfflineTradeService interface {
	Enabled() bool
	EnteredOfflineMode(ctx context.Context, player *model.Player, objectID uint32, accountName string) error
}

// OnDisconnection handles player disconnection (TCP connection lost).
// Implements delayed removal if player cannot logout immediately (combat stance).
//
// Flow:
// 1. If player is nil → return (already cleaned up)
// 2. If offline trade enabled and player in store mode → enter offline trade mode
// 3. If player.CanLogout() → immediate storeAndDelete()
// 4. If player.CanLogout() == false → delayed storeAndDelete() after CombatTime (15 seconds)
//
// The 15-second delay prevents "combat logging" — disconnected player remains in world
// vulnerable to attacks for 15 seconds after disconnect.
//
// Phase 6.0: Added persister parameter for DB save on disconnect.
// Phase 31: Added offlineSvc parameter for offline trade mode.
//
// Reference: L2J_Mobius Disconnection.onDisconnection() (lines 155-176)
func OnDisconnection(ctx context.Context, client *GameClient, persister PlayerPersister, offlineSvc OfflineTradeService) {
	player := client.ActivePlayer()
	if player == nil {
		// No player associated — already cleaned up or never entered world
		return
	}

	// Phase 31: Check if player qualifies for offline trade mode
	if offlineSvc != nil && offlineSvc.Enabled() && player.IsInStoreMode() {
		// Игрок в активном store mode → переходим в offline trade вместо удаления
		if err := offlineSvc.EnteredOfflineMode(ctx, player, player.ObjectID(), client.AccountName()); err != nil {
			slog.Warn("offline trade mode failed, proceeding with normal disconnect",
				"character", player.Name(),
				"error", err)
		} else {
			// Успешно вошли в offline trade:
			// - Player остаётся в мире (WorldObject не удаляется)
			// - Клиент помечается как detached
			client.Detach()
			client.ClearCharacterCache()
			client.SetActivePlayer(nil)

			slog.Info("player entered offline trade mode on disconnect",
				"account", client.AccountName(),
				"character", player.Name(),
				"objectID", player.ObjectID())
			return
		}
	}

	// Break client-player link (Phase 4.17.7)
	// Prevents double-processing if onDisconnection called multiple times
	client.SetActivePlayer(nil)

	// Clear character cache to free memory (Phase 4.18 Optimization 3)
	client.ClearCharacterCache()

	// Check if player can logout immediately
	if player.CanLogout() {
		// Immediate removal: no combat stance, no active tasks
		slog.Info("Player disconnected (immediate cleanup)",
			"account", client.AccountName(),
			"character", player.Name(),
			"objectID", player.ObjectID(),
		)
		storeAndDelete(ctx, player, persister)
		return
	}

	// Delayed removal: player in combat stance
	// Remains in world for CombatTime (15 seconds) to prevent combat logging
	slog.Info("Player disconnected (delayed cleanup due to combat stance)",
		"account", client.AccountName(),
		"character", player.Name(),
		"objectID", player.ObjectID(),
		"delay", CombatTime,
	)

	// Schedule delayed removal.
	// Use fresh context (not the connection ctx which may already be cancelled).
	time.AfterFunc(CombatTime, func() {
		// Check if player still in world (may have reconnected or already removed)
		w := world.Instance()
		obj, exists := w.GetObject(player.ObjectID())
		if !exists || obj == nil {
			slog.Debug("Delayed disconnect: player already removed from world",
				"objectID", player.ObjectID(),
			)
			return
		}

		// Player still in world after 15 seconds — perform cleanup
		slog.Info("Delayed disconnect: removing player from world after combat timeout",
			"character", player.Name(),
			"objectID", player.ObjectID(),
		)
		saveCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		storeAndDelete(saveCtx, player, persister)
	})
}

// storeAndDelete saves player to DB and removes from world.
//
// Phase 6.0: Saves player data (character, items, skills) before removing from world.
// Phase 7.3: Removes player from party on disconnect.
//
// Reference: L2J_Mobius Disconnection.storeAndDelete() (lines 110-134)
func storeAndDelete(ctx context.Context, player *model.Player, persister PlayerPersister) {
	// Phase 8.1: Close private store on disconnect
	if player.IsTrading() {
		player.ClosePrivateStore()
	}

	// Phase 7.3: Remove from party on disconnect
	if party := player.GetParty(); party != nil {
		party.RemoveMember(player.ObjectID())
		player.SetParty(nil)
	}

	// Phase 6.0: Save player to DB
	if persister != nil {
		if err := persister.SavePlayer(ctx, player); err != nil {
			slog.Error("failed to save player on disconnect",
				"character", player.Name(),
				"error", err)
		}
	}

	// Remove from world
	w := world.Instance()
	w.RemoveObject(player.ObjectID())

	slog.Info("Player removed from world",
		"character", player.Name(),
		"objectID", player.ObjectID(),
	)
}
