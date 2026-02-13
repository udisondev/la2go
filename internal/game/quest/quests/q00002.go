package quests

import (
	"fmt"

	"github.com/udisondev/la2go/internal/game/quest"
)

// NewQ00002 creates "What Women Want" quest.
// Level 2+, Human/Elf only, one-time. Deliver letters for Arujien. Choose: earring or 450 adena.
func NewQ00002() *quest.Quest {
	const (
		questID   int32 = 2
		questName       = "Q00002_WhatWomenWant"

		arujien int32 = 30223
		mirabel int32 = 30146
		herbiel int32 = 30150
		greenis int32 = 30157

		arujienLetter1 int32 = 1092
		arujienLetter2 int32 = 1093
		arujienLetter3 int32 = 1094
		poetryBook     int32 = 689
		greenisLetter  int32 = 693
	)

	q := quest.NewQuest(questID, questName)
	q.AddQuestItem(arujienLetter1)
	q.AddQuestItem(arujienLetter2)
	q.AddQuestItem(arujienLetter3)
	q.AddQuestItem(poetryBook)
	q.AddQuestItem(greenisLetter)

	q.AddTalkID(arujien, func(e *quest.Event, qs *quest.QuestState) string {
		event := getEvent(e)

		if event == "accept" && qs.State() == quest.StateCreated {
			qs.SetState(quest.StateStarted)
			qs.SetCond(1)
			giveItem(e.Player, arujienLetter1, 1)
			return "<html><body>Arujien:<br>Take this letter to Mirabel.</body></html>"
		}

		if event == "poetry" && qs.GetCond() == 3 {
			takeItem(e.Player, arujienLetter3, 1)
			giveItem(e.Player, poetryBook, 1)
			qs.SetCond(4)
			return "<html><body>Arujien:<br>Give this poetry book to Greenis.</body></html>"
		}

		if event == "adena" && qs.GetCond() == 3 {
			takeItem(e.Player, arujienLetter3, 1)
			giveAdena(e, 450)
			qs.SetState(quest.StateCompleted)
			return "<html><body>Arujien:<br>Here are 450 adena. Thank you!</body></html>"
		}

		switch qs.State() {
		case quest.StateCreated:
			if e.Player.RaceID() != RaceHuman && e.Player.RaceID() != RaceElf {
				return "<html><body>Arujien:<br>Only Humans and Elves can help me.</body></html>"
			}
			if e.Player.Level() < 2 {
				return "<html><body>Arujien:<br>Come back at level 2.</body></html>"
			}
			return fmt.Sprintf("<html><body>Arujien:<br>I need help delivering letters.<br>"+
				"<a action=\"bypass -h npc_%d_Quest %s accept\">Help Arujien</a></body></html>",
				e.TargetID, questName)

		case quest.StateStarted:
			switch qs.GetCond() {
			case 1:
				return "<html><body>Arujien:<br>Take the letter to Mirabel.</body></html>"
			case 2:
				takeItem(e.Player, arujienLetter2, 1)
				giveItem(e.Player, arujienLetter3, 1)
				qs.SetCond(3)
				return "<html><body>Arujien:<br>Mirabel responded! Now take this to Herbiel.</body></html>"
			case 3:
				if hasItem(e.Player, arujienLetter3) {
					return "<html><body>Arujien:<br>Take Herbiel's reply back to me.</body></html>"
				}
				return fmt.Sprintf("<html><body>Arujien:<br>Herbiel wrote back! "+
					"Would you like a poetry book for Greenis, or 450 adena?<br>"+
					"<a action=\"bypass -h npc_%d_Quest %s poetry\">Poetry book</a><br>"+
					"<a action=\"bypass -h npc_%d_Quest %s adena\">450 adena</a></body></html>",
					e.TargetID, questName, e.TargetID, questName)
			case 4:
				return "<html><body>Arujien:<br>Give the poetry book to Greenis.</body></html>"
			case 5:
				takeItem(e.Player, greenisLetter, 1)
				qs.SetState(quest.StateCompleted)
				return "<html><body>Arujien:<br>Greenis liked my poetry! " +
					"Here, take this Mystic's Earring!</body></html>"
			}

		case quest.StateCompleted:
			return "<html><body>Arujien:<br>Thank you for everything!</body></html>"
		}

		return ""
	})

	q.AddTalkID(mirabel, func(e *quest.Event, qs *quest.QuestState) string {
		if !qs.IsStarted() || qs.GetCond() != 1 {
			return ""
		}
		takeItem(e.Player, arujienLetter1, 1)
		giveItem(e.Player, arujienLetter2, 1)
		qs.SetCond(2)
		return "<html><body>Mirabel:<br>A letter from Arujien? Here's my reply.</body></html>"
	})

	q.AddTalkID(herbiel, func(e *quest.Event, qs *quest.QuestState) string {
		if !qs.IsStarted() || qs.GetCond() != 3 {
			return ""
		}
		if !hasItem(e.Player, arujienLetter3) {
			return ""
		}
		takeItem(e.Player, arujienLetter3, 1)
		return "<html><body>Herbiel:<br>Interesting. Take this reply back to Arujien.</body></html>"
	})

	q.AddTalkID(greenis, func(e *quest.Event, qs *quest.QuestState) string {
		if !qs.IsStarted() || qs.GetCond() != 4 {
			return ""
		}
		takeItem(e.Player, poetryBook, 1)
		giveItem(e.Player, greenisLetter, 1)
		qs.SetCond(5)
		return "<html><body>Greenis:<br>Poetry from Arujien? How lovely! " +
			"Give him this response.</body></html>"
	})

	return q
}
