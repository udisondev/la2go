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

// OnDisconnection handles player disconnection (TCP connection lost).
// Implements delayed removal if player cannot logout immediately (combat stance).
//
// Flow:
// 1. If player is nil → return (already cleaned up)
// 2. If player.CanLogout() → immediate storeAndDelete()
// 3. If player.CanLogout() == false → delayed storeAndDelete() after CombatTime (15 seconds)
//
// The 15-second delay prevents "combat logging" — disconnected player remains in world
// vulnerable to attacks for 15 seconds after disconnect.
//
// Phase 6.0: Added persister parameter for DB save on disconnect.
//
// Reference: L2J_Mobius Disconnection.onDisconnection() (lines 155-176)
func OnDisconnection(ctx context.Context, client *GameClient, persister PlayerPersister) {
	player := client.ActivePlayer()
	if player == nil {
		// No player associated — already cleaned up or never entered world
		return
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

	// Schedule delayed removal
	time.AfterFunc(CombatTime, func() {
		// Check if player still in world (may have reconnected or already removed)
		w := world.Instance()
		obj, exists := w.GetObject(player.ObjectID())
		if !exists || obj == nil {
			// Player already removed from world (reconnected or manual cleanup)
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
		storeAndDelete(ctx, player, persister)
	})
}

// storeAndDelete saves player to DB and removes from world.
//
// Phase 6.0: Saves player data (character, items, skills) before removing from world.
//
// Reference: L2J_Mobius Disconnection.storeAndDelete() (lines 110-134)
func storeAndDelete(ctx context.Context, player *model.Player, persister PlayerPersister) {
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
