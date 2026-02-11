package combat

import (
	"log/slog"

	"github.com/udisondev/la2go/internal/data"
	"github.com/udisondev/la2go/internal/gameserver/serverpackets"
	"github.com/udisondev/la2go/internal/model"
)

// RewardExpAndSp calculates and awards XP/SP to killer for NPC death.
// Sends SystemMessage packets to killer and handles level-up if threshold reached.
//
// sendPacketFunc sends a packet to a specific player (by objectID).
// broadcastFunc broadcasts to visible players around killer.
func RewardExpAndSp(
	player *model.Player,
	npc *model.Npc,
	sendPacketFunc func(objectID uint32, data []byte, size int),
	broadcastFunc func(source *model.Player, data []byte, size int),
) {
	baseExp := npc.Template().BaseExp()
	baseSP := int64(npc.Template().BaseSP())

	if baseExp <= 0 && baseSP <= 0 {
		return
	}

	// Add experience and SP
	player.AddExperience(baseExp)
	player.AddSP(baseSP)

	// Send "You earned X exp and Y SP" message
	sendExpMessage(player, baseExp, baseSP, sendPacketFunc)

	// Check level-up
	oldLevel := player.Level()
	newLevel := data.GetLevelForExp(player.Experience(), oldLevel)

	if newLevel > oldLevel {
		if err := player.SetLevel(newLevel); err != nil {
			slog.Error("failed to set level after level-up",
				"player", player.Name(),
				"newLevel", newLevel,
				"error", err)
			return
		}

		// Restore HP/MP/CP to max on level-up
		player.SetCurrentHP(player.MaxHP())
		player.SetCurrentMP(player.MaxMP())
		player.SetCurrentCP(player.MaxCP())

		// Send "Your level has increased!" message
		levelUpMsg := serverpackets.NewSystemMessage(serverpackets.SysMsgYourLevelHasIncreased)
		levelUpData, err := levelUpMsg.Write()
		if err != nil {
			slog.Error("failed to write level-up SystemMessage",
				"player", player.Name(),
				"error", err)
		} else {
			sendPacketFunc(player.ObjectID(), levelUpData, len(levelUpData))
		}

		// Broadcast SocialAction (level-up animation)
		socialAction := serverpackets.NewSocialAction(int32(player.ObjectID()), serverpackets.SocialActionLevelUp)
		socialData, err := socialAction.Write()
		if err != nil {
			slog.Error("failed to write SocialAction packet",
				"player", player.Name(),
				"error", err)
		} else {
			broadcastFunc(player, socialData, len(socialData))
		}

		// Send UserInfo (full stat refresh)
		userInfo := serverpackets.NewUserInfo(player)
		userInfoData, err := userInfo.Write()
		if err != nil {
			slog.Error("failed to write UserInfo after level-up",
				"player", player.Name(),
				"error", err)
		} else {
			sendPacketFunc(player.ObjectID(), userInfoData, len(userInfoData))
		}

		slog.Info("player leveled up",
			"player", player.Name(),
			"oldLevel", oldLevel,
			"newLevel", newLevel,
			"exp", player.Experience())
	}
}

// sendExpMessage sends the appropriate exp/sp system message to player.
func sendExpMessage(
	player *model.Player,
	exp int64,
	sp int64,
	sendPacketFunc func(objectID uint32, data []byte, size int),
) {
	var msg *serverpackets.SystemMessage

	if exp > 0 && sp > 0 {
		msg = serverpackets.NewSystemMessage(serverpackets.SysMsgYouEarnedS1ExpAndS2SP).
			AddNumber(exp).
			AddNumber(sp)
	} else if exp > 0 {
		msg = serverpackets.NewSystemMessage(serverpackets.SysMsgYouEarnedS1Exp).
			AddNumber(exp)
	} else {
		msg = serverpackets.NewSystemMessage(serverpackets.SysMsgYouAcquiredS1SP).
			AddNumber(sp)
	}

	msgData, err := msg.Write()
	if err != nil {
		slog.Error("failed to write exp SystemMessage",
			"player", player.Name(),
			"error", err)
		return
	}

	sendPacketFunc(player.ObjectID(), msgData, len(msgData))
}
