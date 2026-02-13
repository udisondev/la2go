package quests

import (
	"fmt"

	"github.com/udisondev/la2go/internal/game/quest"
)

// NewQ00293 creates "The Hidden Veins" quest.
// Level 6+, Dwarf only, repeatable. Kill Utuku Orcs for Chrysolite Ore and Map Fragments.
// Chinchirin exchanges 4 Map Fragments → 1 Hidden Vein Map.
func NewQ00293() *quest.Quest {
	const (
		questID   int32 = 293
		questName       = "Q00293_TheHiddenVeins"

		filaur    int32 = 30535
		chinchrin int32 = 30539

		chrysoliteOre   int32 = 1488
		tornMapFragment int32 = 1489
		hiddenVeinMap   int32 = 1490
	)

	q := quest.NewQuest(questID, questName)
	q.AddQuestItem(chrysoliteOre)
	q.AddQuestItem(tornMapFragment)
	q.AddQuestItem(hiddenVeinMap)

	// Filaur: start NPC + turn-in
	q.AddTalkID(filaur, func(e *quest.Event, qs *quest.QuestState) string {
		event := getEvent(e)

		if event == "accept" {
			qs.SetState(quest.StateStarted)
			qs.SetCond(1)
			return "<html><body>Miner Filaur:<br>Hunt the Utuku Orcs in the mines. " +
				"Bring me Chrysolite Ore and Hidden Vein Maps.</body></html>"
		}

		if event == "exit" {
			qs.SetState(quest.StateCreated)
			return "<html><body>Miner Filaur:<br>Very well. Come back anytime.</body></html>"
		}

		switch qs.State() {
		case quest.StateCreated:
			if e.Player.RaceID() != RaceDwarf {
				return "<html><body>Miner Filaur:<br>Only Dwarves can help me with this.</body></html>"
			}
			if e.Player.Level() < 6 {
				return "<html><body>Miner Filaur:<br>You need to be level 6.</body></html>"
			}
			return fmt.Sprintf("<html><body>Miner Filaur:<br>The Utuku Orcs invaded our mines!<br>"+
				"Kill them and bring me Chrysolite Ore. Also look for Map Fragments — "+
				"Chinchirin can combine them into Hidden Vein Maps, which are worth more.<br>"+
				"<a action=\"bypass -h npc_%d_Quest %s accept\">Accept</a></body></html>",
				e.TargetID, questName)

		case quest.StateStarted:
			ores := getItemCount(e.Player, chrysoliteOre)
			maps := getItemCount(e.Player, hiddenVeinMap)

			if ores+maps == 0 {
				return "<html><body>Miner Filaur:<br>Nothing yet? Keep looking!</body></html>"
			}

			reward := int32(ores*5 + maps*500)
			if ores >= 10 {
				reward += 2000
			}
			if maps >= 10 {
				reward += 2000
			}

			takeItem(e.Player, chrysoliteOre, -1)
			takeItem(e.Player, hiddenVeinMap, -1)
			giveAdena(e, reward)

			return fmt.Sprintf("<html><body>Miner Filaur:<br>Excellent! "+
				"Here are %d adena.</body></html>", reward)
		}
		return ""
	})

	// Chinchirin: exchanges 4 Torn Map Fragments → 1 Hidden Vein Map
	q.AddTalkID(chinchrin, func(e *quest.Event, qs *quest.QuestState) string {
		if !qs.IsStarted() {
			return ""
		}

		event := getEvent(e)

		if event == "exchange" {
			fragments := getItemCount(e.Player, tornMapFragment)
			if fragments >= 4 {
				takeItem(e.Player, tornMapFragment, 4)
				giveItem(e.Player, hiddenVeinMap, 1)
				return "<html><body>Chinchirin:<br>Here is a Hidden Vein Map!</body></html>"
			}
			return "<html><body>Chinchirin:<br>You need 4 Map Fragments.</body></html>"
		}

		fragments := getItemCount(e.Player, tornMapFragment)
		if fragments >= 4 {
			return fmt.Sprintf("<html><body>Chinchirin:<br>You have %d Map Fragments. "+
				"I can combine 4 into a Hidden Vein Map!<br>"+
				"<a action=\"bypass -h npc_%d_Quest %s exchange\">Exchange</a></body></html>",
				fragments, e.TargetID, questName)
		}

		return fmt.Sprintf("<html><body>Chinchirin:<br>Bring me 4 Torn Map Fragments "+
			"and I'll make a Hidden Vein Map. You have %d.</body></html>", fragments)
	})

	// onKill: Utuku Orcs
	killHook := func(e *quest.Event, qs *quest.QuestState) string {
		if !qs.IsStarted() {
			return ""
		}

		chance := getRandom(100)
		if chance > 50 {
			giveItem(e.Player, chrysoliteOre, 1)
		} else if chance < 5 {
			giveItem(e.Player, tornMapFragment, 1)
		}
		return ""
	}

	for _, npcID := range []int32{20446, 20447, 20448} {
		q.AddKillID(npcID, killHook)
	}

	return q
}
