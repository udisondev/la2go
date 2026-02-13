package quests

import (
	"fmt"

	"github.com/udisondev/la2go/internal/game/quest"
)

// NewQ00257 creates "The Guard is Busy" quest.
// Level 6+, any race, repeatable. Kill orcs/werewolves near Gludio, turn in drops for adena.
func NewQ00257() *quest.Quest {
	const (
		questID   int32 = 257
		questName       = "Q00257_TheGuardIsBusy"

		gilbert int32 = 30039

		orcAmulet    int32 = 752
		orcNecklace  int32 = 1085
		werewolfFang int32 = 1086
	)

	q := quest.NewQuest(questID, questName)
	q.AddQuestItem(orcAmulet)
	q.AddQuestItem(orcNecklace)
	q.AddQuestItem(werewolfFang)

	// onTalk + onEvent: Gilbert
	q.AddTalkID(gilbert, func(e *quest.Event, qs *quest.QuestState) string {
		event := getEvent(e)

		// onEvent: accept quest
		if event == "30039-03.htm" {
			qs.SetState(quest.StateStarted)
			qs.SetCond(1)
			return "<html><body>Guard Gilbert:<br>Good. Go hunt Orcs and Werewolves around Gludio.<br>" +
				"Bring me Orc Amulets, Orc Necklaces and Werewolf Fangs.</body></html>"
		}

		// onEvent: exit quest
		if event == "30039-06.htm" {
			qs.SetState(quest.StateCreated)
			return "<html><body>Guard Gilbert:<br>Come back if you change your mind.</body></html>"
		}

		switch qs.State() {
		case quest.StateCreated:
			if e.Player.Level() < 6 {
				return "<html><body>Guard Gilbert:<br>You are too weak. Come back when you reach level 6.</body></html>"
			}
			return fmt.Sprintf("<html><body>Guard Gilbert:<br>The Gludio area is overrun with Orcs and Werewolves!<br>"+
				"Help me clear them out and I'll pay you for their trophies.<br>"+
				"<a action=\"bypass -h npc_%d_Quest %s 30039-03.htm\">Accept quest</a></body></html>",
				e.TargetID, questName)

		case quest.StateStarted:
			amulets := getItemCount(e.Player, orcAmulet)
			necklaces := getItemCount(e.Player, orcNecklace)
			fangs := getItemCount(e.Player, werewolfFang)
			total := amulets + necklaces + fangs

			if total == 0 {
				return "<html><body>Guard Gilbert:<br>You haven't brought anything yet. " +
					"Go hunt Orcs and Werewolves!</body></html>"
			}

			reward := int32(amulets*10 + necklaces*20 + fangs*20)
			if total >= 10 {
				reward += 1000
			}

			takeItem(e.Player, orcAmulet, -1)
			takeItem(e.Player, orcNecklace, -1)
			takeItem(e.Player, werewolfFang, -1)
			giveAdena(e, reward)

			return fmt.Sprintf("<html><body>Guard Gilbert:<br>Good work! "+
				"Here is your payment of %d adena.<br>"+
				"Come back if you want to continue hunting.</body></html>", reward)
		}

		return ""
	})

	// onKill: various orcs and werewolves
	killHook := func(e *quest.Event, qs *quest.QuestState) string {
		if !qs.IsStarted() {
			return ""
		}

		var itemID int32
		var dropRate int

		switch e.NpcID {
		case 20006, 20130, 20131:
			itemID, dropRate = orcAmulet, 50
		case 20093, 20096, 20098:
			itemID, dropRate = orcNecklace, 50
		case 20342:
			itemID, dropRate = werewolfFang, 20
		case 20343:
			itemID, dropRate = werewolfFang, 40
		case 20132:
			itemID, dropRate = werewolfFang, 50
		default:
			return ""
		}

		if getRandom(100) < dropRate {
			giveItem(e.Player, itemID, 1)
		}
		return ""
	}

	for _, npcID := range []int32{20006, 20130, 20131, 20093, 20096, 20098, 20342, 20343, 20132} {
		q.AddKillID(npcID, killHook)
	}

	return q
}
