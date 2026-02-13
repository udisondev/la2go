package quests

import (
	"fmt"

	"github.com/udisondev/la2go/internal/game/quest"
	"github.com/udisondev/la2go/internal/model"
)

// NewQ00003 creates "Will The Seal Be Broken" quest.
// Level 16+, Dark Elf only, one-time. Collect 3 monster parts for an Enchant Scroll.
func NewQ00003() *quest.Quest {
	const (
		questID   int32 = 3
		questName       = "Q00003_WillTheSealBeBroken"

		talloth int32 = 30141

		omenBeastEye int32 = 1081
		taintStone   int32 = 1082
		succubusBlood int32 = 1083
	)

	q := quest.NewQuest(questID, questName)
	q.AddQuestItem(omenBeastEye)
	q.AddQuestItem(taintStone)
	q.AddQuestItem(succubusBlood)

	q.AddTalkID(talloth, func(e *quest.Event, qs *quest.QuestState) string {
		event := getEvent(e)

		if event == "accept" && qs.State() == quest.StateCreated {
			qs.SetState(quest.StateStarted)
			qs.SetCond(1)
			return "<html><body>Talloth:<br>Bring me an Omen Beast Eye, a Taint Stone, " +
				"and Succubus Blood.</body></html>"
		}

		switch qs.State() {
		case quest.StateCreated:
			if e.Player.RaceID() != RaceDarkElf {
				return "<html><body>Talloth:<br>Only a Dark Elf can help with this ritual.</body></html>"
			}
			if e.Player.Level() < 16 {
				return "<html><body>Talloth:<br>You need to be level 16.</body></html>"
			}
			return fmt.Sprintf("<html><body>Talloth:<br>I need 3 rare components for a ritual.<br>"+
				"An Omen Beast Eye, a Taint Stone, and Succubus Blood.<br>"+
				"<a action=\"bypass -h npc_%d_Quest %s accept\">Accept</a></body></html>",
				e.TargetID, questName)

		case quest.StateStarted:
			if hasItem(e.Player, omenBeastEye) && hasItem(e.Player, taintStone) && hasItem(e.Player, succubusBlood) {
				takeItem(e.Player, omenBeastEye, -1)
				takeItem(e.Player, taintStone, -1)
				takeItem(e.Player, succubusBlood, -1)
				qs.SetState(quest.StateCompleted)
				return "<html><body>Talloth:<br>You found them all! " +
					"Here is a Scroll of Enchant as reward.</body></html>"
			}
			return "<html><body>Talloth:<br>You still need: " +
				itemStatus("Omen Beast Eye", e.Player, omenBeastEye) +
				itemStatus("Taint Stone", e.Player, taintStone) +
				itemStatus("Succubus Blood", e.Player, succubusBlood) +
				"</body></html>"

		case quest.StateCompleted:
			return "<html><body>Talloth:<br>The ritual is complete. Thank you.</body></html>"
		}

		return ""
	})

	// Omen Beast
	q.AddKillID(20031, func(e *quest.Event, qs *quest.QuestState) string {
		if !qs.IsStarted() || hasItem(e.Player, omenBeastEye) {
			return ""
		}
		if getRandom(10) < 3 {
			giveItem(e.Player, omenBeastEye, 1)
			checkAllCollected(e.Player, qs, omenBeastEye, taintStone, succubusBlood)
		}
		return ""
	})

	// Tainted/Stink Zombies
	zombieKill := func(e *quest.Event, qs *quest.QuestState) string {
		if !qs.IsStarted() || hasItem(e.Player, taintStone) {
			return ""
		}
		if getRandom(10) < 3 {
			giveItem(e.Player, taintStone, 1)
			checkAllCollected(e.Player, qs, omenBeastEye, taintStone, succubusBlood)
		}
		return ""
	}
	q.AddKillID(20041, zombieKill)
	q.AddKillID(20046, zombieKill)

	// Succubi
	succubusKill := func(e *quest.Event, qs *quest.QuestState) string {
		if !qs.IsStarted() || hasItem(e.Player, succubusBlood) {
			return ""
		}
		if getRandom(10) < 3 {
			giveItem(e.Player, succubusBlood, 1)
			checkAllCollected(e.Player, qs, omenBeastEye, taintStone, succubusBlood)
		}
		return ""
	}
	q.AddKillID(20048, succubusKill)
	q.AddKillID(20052, succubusKill)
	q.AddKillID(20057, succubusKill)

	return q
}

// itemStatus returns HTML status for a quest item.
func itemStatus(name string, player *model.Player, itemID int32) string {
	if hasItem(player, itemID) {
		return "<br>" + name + ": <font color=\"00FF00\">Found</font>"
	}
	return "<br>" + name + ": <font color=\"FF0000\">Not found</font>"
}

// checkAllCollected sets cond=2 if all 3 items are collected.
func checkAllCollected(player *model.Player, qs *quest.QuestState, items ...int32) {
	for _, id := range items {
		if !hasItem(player, id) {
			return
		}
	}
	qs.SetCond(2)
}
