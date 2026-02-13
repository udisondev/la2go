package combat

import (
	"log/slog"
	"math"

	"github.com/udisondev/la2go/internal/data"
	"github.com/udisondev/la2go/internal/gameserver/serverpackets"
	"github.com/udisondev/la2go/internal/model"
)

// PartyXPRangeSq is the squared distance for party XP sharing (1600 world units).
// Java reference: Party.java PARTY_DISTRIBUTION_RANGE = 1600.
const PartyXPRangeSq int64 = 1600 * 1600

// maxLevelGap is the maximum allowed level difference from the highest-level member.
// Members exceeding this gap receive 0 XP/SP (Java: PARTY_XP_CUTOFF_LEVEL default=20).
const maxLevelGap int32 = 20

// RewardExpAndSp calculates and awards XP/SP to killer for NPC death.
// If the killer is in a party, XP is distributed among nearby party members with bonus.
// Sends SystemMessage packets to each recipient and handles level-up if threshold reached.
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

	// Phase 7.3: Party XP distribution (level²-proportional, Java: Party.distributeXpAndSp)
	if party := player.GetParty(); party != nil {
		loc := player.Location()
		nearbyMembers := party.MembersInRange(loc.X, loc.Y, loc.Z, PartyXPRangeSq)
		if len(nearbyMembers) > 1 {
			distributePartyXPAndSP(nearbyMembers, baseExp, baseSP, party, sendPacketFunc, broadcastFunc)
			return
		}
	}

	// Solo reward (no party or only killer in range)
	player.AddExperience(baseExp)
	player.AddSP(baseSP)

	// Send "You earned X exp and Y SP" message
	sendExpMessage(player, baseExp, baseSP, sendPacketFunc)

	// Check level-up
	checkLevelUp(player, sendPacketFunc, broadcastFunc)
}

// distributePartyXPAndSP distributes XP/SP among party members proportionally to level²
// with level gap filtering. Matches Java Party.distributeXpAndSp().
//
// Algorithm:
//  1. Find highest level among nearby members (topLevel).
//  2. Filter valid members: alive AND (topLevel - memberLevel) <= maxLevelGap.
//  3. Apply party bonus based on valid member count.
//  4. Calculate sqLevelSum = sum(level²) for valid members.
//  5. Each valid member receives: totalXP * (level² / sqLevelSum).
//  6. Invalid/dead members receive nothing.
func distributePartyXPAndSP(
	members []*model.Player,
	baseExp int64,
	baseSP int64,
	party *model.Party,
	sendPacketFunc func(objectID uint32, data []byte, size int),
	broadcastFunc func(source *model.Player, data []byte, size int),
) {
	// Find highest level among nearby members
	var topLevel int32
	for _, m := range members {
		if lvl := m.Level(); lvl > topLevel {
			topLevel = lvl
		}
	}

	// Filter valid members: alive and within level gap
	validMembers := make([]*model.Player, 0, len(members))
	for _, m := range members {
		if m.IsDead() {
			continue
		}
		if topLevel-m.Level() > maxLevelGap {
			continue
		}
		validMembers = append(validMembers, m)
	}

	if len(validMembers) == 0 {
		return
	}

	// Apply party bonus based on valid member count
	bonus := party.GetXPBonus()
	totalExp := float64(baseExp) * bonus
	totalSP := float64(baseSP) * bonus

	// Calculate sum of level² for proportional distribution
	var sqLevelSum int64
	for _, m := range validMembers {
		lvl := int64(m.Level())
		sqLevelSum += lvl * lvl
	}

	if sqLevelSum == 0 {
		return
	}

	// Distribute proportionally to level² (Java: sqLevel / sqLevelSum * preCalculation)
	for _, m := range members {
		if m.IsDead() {
			continue
		}

		// Check if member is valid (within level gap)
		isValid := false
		for _, v := range validMembers {
			if v.ObjectID() == m.ObjectID() {
				isValid = true
				break
			}
		}

		if !isValid {
			// Level gap too large — Java calls addExpAndSp(0, 0)
			continue
		}

		lvl := int64(m.Level())
		sqLevel := float64(lvl * lvl)
		share := sqLevel / float64(sqLevelSum)

		memberExp := int64(math.Round(totalExp * share))
		memberSP := int64(math.Round(totalSP * share))

		m.AddExperience(memberExp)
		m.AddSP(memberSP)

		if memberExp > 0 || memberSP > 0 {
			sendExpMessage(m, memberExp, memberSP, sendPacketFunc)
		}
		checkLevelUp(m, sendPacketFunc, broadcastFunc)
	}
}

// checkLevelUp checks if player reached a new level and handles level-up effects.
func checkLevelUp(
	player *model.Player,
	sendPacketFunc func(objectID uint32, data []byte, size int),
	broadcastFunc func(source *model.Player, data []byte, size int),
) {
	oldLevel := player.Level()
	newLevel := data.GetLevelForExp(player.Experience(), oldLevel)

	if newLevel <= oldLevel {
		return
	}

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

	// Grant auto-get skills for new level
	newAutoSkills := data.GetNewAutoGetSkills(player.ClassID(), newLevel)
	for _, sl := range newAutoSkills {
		isPassive := false
		if tmpl := data.GetSkillTemplate(sl.SkillID, sl.SkillLevel); tmpl != nil {
			isPassive = tmpl.IsPassive()
		}
		player.AddSkill(sl.SkillID, sl.SkillLevel, isPassive)
	}

	// Send SkillList if new skills were granted
	if len(newAutoSkills) > 0 {
		skillList := serverpackets.NewSkillList(player.Skills())
		skillListData, err := skillList.Write()
		if err != nil {
			slog.Error("failed to write SkillList after level-up",
				"player", player.Name(),
				"error", err)
		} else {
			sendPacketFunc(player.ObjectID(), skillListData, len(skillListData))
		}
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
		"exp", player.Experience(),
		"newSkills", len(newAutoSkills))
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
