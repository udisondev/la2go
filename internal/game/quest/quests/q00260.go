package quests

import (
	"fmt"

	"github.com/udisondev/la2go/internal/game/quest"
)

// NewQ00260 creates "Hunt the Orcs" quest.
// Level 6+, Elf only, repeatable. Kill Kaboo Orcs near Elven Village for adena.
func NewQ00260() *quest.Quest {
	const (
		questID   int32 = 260
		questName       = "Q00260_OrcHunting"

		rayen int32 = 30221

		orcAmulet   int32 = 1114
		orcNecklace int32 = 1115
	)

	q := quest.NewQuest(questID, questName)
	q.AddQuestItem(orcAmulet)
	q.AddQuestItem(orcNecklace)

	q.AddTalkID(rayen, func(e *quest.Event, qs *quest.QuestState) string {
		event := getEvent(e)

		if event == "accept" {
			qs.SetState(quest.StateStarted)
			qs.SetCond(1)
			return "<html><body>Sentinel Rayen:<br>Good. Hunt the Kaboo Orcs and bring me their trophies.</body></html>"
		}

		if event == "exit" {
			qs.SetState(quest.StateCreated)
			return "<html><body>Sentinel Rayen:<br>Come back anytime.</body></html>"
		}

		switch qs.State() {
		case quest.StateCreated:
			if e.Player.RaceID() != RaceElf {
				return "<html><body>Sentinel Rayen:<br>This quest is only for Elves.</body></html>"
			}
			if e.Player.Level() < 6 {
				return "<html><body>Sentinel Rayen:<br>Come back when you are level 6.</body></html>"
			}
			return fmt.Sprintf("<html><body>Sentinel Rayen:<br>The Kaboo Orcs threaten our village.<br>"+
				"Hunt them and bring me Orc Amulets and Orc Necklaces.<br>"+
				"<a action=\"bypass -h npc_%d_Quest %s accept\">Accept quest</a></body></html>",
				e.TargetID, questName)

		case quest.StateStarted:
			amulets := getItemCount(e.Player, orcAmulet)
			necklaces := getItemCount(e.Player, orcNecklace)
			total := amulets + necklaces

			if total == 0 {
				return "<html><body>Sentinel Rayen:<br>Hunt the Kaboo Orcs! " +
					"You haven't brought anything yet.</body></html>"
			}

			reward := int32(amulets*12 + necklaces*30)
			if total >= 10 {
				reward += 1000
			}

			takeItem(e.Player, orcAmulet, -1)
			takeItem(e.Player, orcNecklace, -1)
			giveAdena(e, reward)

			return fmt.Sprintf("<html><body>Sentinel Rayen:<br>Excellent work! "+
				"Here are %d adena for your trouble.</body></html>", reward)
		}

		return ""
	})

	killHook := func(e *quest.Event, qs *quest.QuestState) string {
		if !qs.IsStarted() {
			return ""
		}
		var itemID int32
		switch e.NpcID {
		case 20468, 20469, 20470:
			itemID = orcAmulet
		case 20471, 20472, 20473:
			itemID = orcNecklace
		default:
			return ""
		}
		if getRandom(100) < 50 {
			giveItem(e.Player, itemID, 1)
		}
		return ""
	}

	for _, npcID := range []int32{20468, 20469, 20470, 20471, 20472, 20473} {
		q.AddKillID(npcID, killHook)
	}

	return q
}
