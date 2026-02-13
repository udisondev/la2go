package quests

import (
	"fmt"

	"github.com/udisondev/la2go/internal/game/quest"
)

// NewQ00005 creates "Miner's Favor" quest.
// Level 2+, any race, one-time. Collect 4 items from NPCs for Bolter the miner.
func NewQ00005() *quest.Quest {
	const (
		questID   int32 = 5
		questName       = "Q00005_MinersFavor"

		bolter int32 = 30554
		shari  int32 = 30517
		garita int32 = 30518
		reed   int32 = 30520
		brunon int32 = 30526

		boltersList    int32 = 1547
		miningBoots    int32 = 1548
		minersPick     int32 = 1549
		boomboomPowder int32 = 1550
		redstoneBeer   int32 = 1551
		smellySocks    int32 = 1552
	)

	q := quest.NewQuest(questID, questName)
	q.AddQuestItem(boltersList)
	q.AddQuestItem(miningBoots)
	q.AddQuestItem(minersPick)
	q.AddQuestItem(boomboomPowder)
	q.AddQuestItem(redstoneBeer)
	q.AddQuestItem(smellySocks)

	q.AddTalkID(bolter, func(e *quest.Event, qs *quest.QuestState) string {
		event := getEvent(e)

		if event == "accept" && qs.State() == quest.StateCreated {
			qs.SetState(quest.StateStarted)
			qs.SetCond(1)
			giveItem(e.Player, boltersList, 1)
			giveItem(e.Player, smellySocks, 1)
			return "<html><body>Bolter:<br>Here's my shopping list and... my socks for Brunon. " +
				"Visit Garita, Shari, Reed and Brunon.</body></html>"
		}

		switch qs.State() {
		case quest.StateCreated:
			if e.Player.Level() < 2 {
				return "<html><body>Bolter:<br>Come back at level 2.</body></html>"
			}
			return fmt.Sprintf("<html><body>Bolter:<br>I need supplies from the village!<br>"+
				"Can you pick them up for me?<br>"+
				"<a action=\"bypass -h npc_%d_Quest %s accept\">Help Bolter</a></body></html>",
				e.TargetID, questName)

		case quest.StateStarted:
			if hasItem(e.Player, miningBoots) && hasItem(e.Player, minersPick) &&
				hasItem(e.Player, boomboomPowder) && hasItem(e.Player, redstoneBeer) {
				takeItem(e.Player, boltersList, -1)
				takeItem(e.Player, miningBoots, -1)
				takeItem(e.Player, minersPick, -1)
				takeItem(e.Player, boomboomPowder, -1)
				takeItem(e.Player, redstoneBeer, -1)
				qs.SetState(quest.StateCompleted)
				return "<html><body>Bolter:<br>All the supplies! " +
					"Here, take this necklace as thanks!</body></html>"
			}
			return "<html><body>Bolter:<br>Still waiting for the supplies. " +
				"Check with Garita, Shari, Reed and Brunon.</body></html>"

		case quest.StateCompleted:
			return "<html><body>Bolter:<br>Thanks again!</body></html>"
		}
		return ""
	})

	q.AddTalkID(garita, func(e *quest.Event, qs *quest.QuestState) string {
		if !qs.IsStarted() || hasItem(e.Player, miningBoots) {
			return ""
		}
		giveItem(e.Player, miningBoots, 1)
		return "<html><body>Garita:<br>Mining boots for Bolter? Here you go.</body></html>"
	})

	q.AddTalkID(shari, func(e *quest.Event, qs *quest.QuestState) string {
		if !qs.IsStarted() || hasItem(e.Player, boomboomPowder) {
			return ""
		}
		giveItem(e.Player, boomboomPowder, 1)
		return "<html><body>Shari:<br>Boom-boom powder? Sure, here.</body></html>"
	})

	q.AddTalkID(reed, func(e *quest.Event, qs *quest.QuestState) string {
		if !qs.IsStarted() || hasItem(e.Player, redstoneBeer) {
			return ""
		}
		giveItem(e.Player, redstoneBeer, 1)
		return "<html><body>Reed:<br>Redstone beer for Bolter? Take it.</body></html>"
	})

	q.AddTalkID(brunon, func(e *quest.Event, qs *quest.QuestState) string {
		if !qs.IsStarted() || hasItem(e.Player, minersPick) {
			return ""
		}
		if !hasItem(e.Player, smellySocks) {
			return "<html><body>Brunon:<br>Bring me Bolter's payment first.</body></html>"
		}
		takeItem(e.Player, smellySocks, 1)
		giveItem(e.Player, minersPick, 1)
		return "<html><body>Brunon:<br>Ugh, his socks... Here's the pick.</body></html>"
	})

	return q
}
