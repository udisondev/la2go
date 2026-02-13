package quests

import (
	"fmt"

	"github.com/udisondev/la2go/internal/game/quest"
)

// NewQ00004 creates "Long Live the Pa'agrio Lord" quest.
// Level 2+, Orc only, one-time. Talk to 6 NPCs to collect gifts for Nakusin.
func NewQ00004() *quest.Quest {
	const (
		questID   int32 = 4
		questName       = "Q00004_LongLiveThePaagrioLord"

		nakusin int32 = 30578

		gift1 int32 = 1542
		gift2 int32 = 1541
		gift3 int32 = 1543
		gift4 int32 = 1544
		gift5 int32 = 1545
		gift6 int32 = 1546
	)

	// NPC → gift item mapping
	type npcGift struct {
		npcID  int32
		itemID int32
		name   string
	}

	giftNPCs := []npcGift{
		{30585, gift1, "Kunai"},
		{30566, gift2, "Uska"},
		{30562, gift3, "Grookin"},
		{30560, gift4, "Varkees"},
		{30559, gift5, "Tantus"},
		{30587, gift6, "Casian"},
	}

	allGifts := []int32{gift1, gift2, gift3, gift4, gift5, gift6}

	q := quest.NewQuest(questID, questName)
	for _, g := range allGifts {
		q.AddQuestItem(g)
	}

	q.AddTalkID(nakusin, func(e *quest.Event, qs *quest.QuestState) string {
		event := getEvent(e)

		if event == "accept" && qs.State() == quest.StateCreated {
			qs.SetState(quest.StateStarted)
			qs.SetCond(1)
			return "<html><body>Nakusin:<br>Visit our tribe's elders and collect their gifts. " +
				"Talk to all 6 of them and return to me.</body></html>"
		}

		switch qs.State() {
		case quest.StateCreated:
			if e.Player.RaceID() != RaceOrc {
				return "<html><body>Nakusin:<br>Only Orcs may honor Pa'agrio.</body></html>"
			}
			if e.Player.Level() < 2 {
				return "<html><body>Nakusin:<br>Grow stronger first.</body></html>"
			}
			return fmt.Sprintf("<html><body>Nakusin:<br>To honor Pa'agrio, "+
				"collect gifts from our tribe's 6 elders.<br>"+
				"<a action=\"bypass -h npc_%d_Quest %s accept\">Accept</a></body></html>",
				e.TargetID, questName)

		case quest.StateStarted:
			// Проверяем все подарки
			allCollected := true
			for _, g := range allGifts {
				if !hasItem(e.Player, g) {
					allCollected = false
					break
				}
			}
			if !allCollected {
				return "<html><body>Nakusin:<br>You haven't collected all the gifts yet. " +
					"Visit all 6 elders.</body></html>"
			}

			for _, g := range allGifts {
				takeItem(e.Player, g, -1)
			}
			qs.SetState(quest.StateCompleted)
			return "<html><body>Nakusin:<br>All gifts collected! " +
				"Pa'agrio is pleased! Here is your reward.</body></html>"

		case quest.StateCompleted:
			return "<html><body>Nakusin:<br>Pa'agrio blesses you.</body></html>"
		}

		return ""
	})

	// Регистрируем каждого NPC-дарителя
	for _, ng := range giftNPCs {
		npcItemID := ng.itemID
		npcName := ng.name
		q.AddTalkID(ng.npcID, func(e *quest.Event, qs *quest.QuestState) string {
			if !qs.IsStarted() {
				return ""
			}
			if hasItem(e.Player, npcItemID) {
				return fmt.Sprintf("<html><body>%s:<br>I already gave you my gift.</body></html>", npcName)
			}
			giveItem(e.Player, npcItemID, 1)
			return fmt.Sprintf("<html><body>%s:<br>Here is my gift for Pa'agrio. "+
				"Take it to Nakusin.</body></html>", npcName)
		})
	}

	return q
}
