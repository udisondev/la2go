package quests

import (
	"fmt"

	"github.com/udisondev/la2go/internal/game/quest"
)

// NewQ00001 creates "Letters of Love" quest.
// Level 2+, any race, one-time. Deliver letters between Darin, Roxxy, and Baulro.
func NewQ00001() *quest.Quest {
	const (
		questID   int32 = 1
		questName       = "Q00001_LettersOfLove"

		darin  int32 = 30048
		roxxy  int32 = 30006
		baulro int32 = 30033

		darinLetter  int32 = 687
		roxxyKerch   int32 = 688
		darinReceipt int32 = 1079
		baulroPotion int32 = 1080

		rewardNecklace int32 = 906
	)

	q := quest.NewQuest(questID, questName)
	q.AddQuestItem(darinLetter)
	q.AddQuestItem(roxxyKerch)
	q.AddQuestItem(darinReceipt)
	q.AddQuestItem(baulroPotion)

	// Darin: start + cond transitions
	q.AddTalkID(darin, func(e *quest.Event, qs *quest.QuestState) string {
		event := getEvent(e)

		if event == "accept" && qs.State() == quest.StateCreated {
			qs.SetState(quest.StateStarted)
			qs.SetCond(1)
			giveItem(e.Player, darinLetter, 1)
			return "<html><body>Darin:<br>Please deliver this letter to Roxxy for me.</body></html>"
		}

		switch qs.State() {
		case quest.StateCreated:
			if e.Player.Level() < 2 {
				return "<html><body>Darin:<br>You must be level 2.</body></html>"
			}
			return fmt.Sprintf("<html><body>Darin:<br>I need someone to deliver a letter to Roxxy.<br>"+
				"<a action=\"bypass -h npc_%d_Quest %s accept\">I'll do it</a></body></html>",
				e.TargetID, questName)

		case quest.StateStarted:
			switch qs.GetCond() {
			case 1:
				return "<html><body>Darin:<br>Please take the letter to Roxxy.</body></html>"
			case 2:
				// Got kerchief from Roxxy
				takeItem(e.Player, roxxyKerch, 1)
				giveItem(e.Player, darinReceipt, 1)
				qs.SetCond(3)
				return "<html><body>Darin:<br>A kerchief from Roxxy! Here, take this receipt to Baulro.</body></html>"
			case 3:
				return "<html><body>Darin:<br>Bring me Baulro's potion.</body></html>"
			case 4:
				// Got potion from Baulro
				takeItem(e.Player, baulroPotion, 1)
				giveAdena(e, 0) // Reward is item, not adena
				qs.SetState(quest.StateCompleted)
				return "<html><body>Darin:<br>The potion! Thank you! " +
					"Here, take this necklace as a reward.</body></html>"
			}

		case quest.StateCompleted:
			return "<html><body>Darin:<br>Thank you again for your help!</body></html>"
		}

		return ""
	})

	// Roxxy: give kerchief
	q.AddTalkID(roxxy, func(e *quest.Event, qs *quest.QuestState) string {
		if !qs.IsStarted() || qs.GetCond() != 1 {
			return ""
		}

		if !hasItem(e.Player, darinLetter) {
			return ""
		}

		takeItem(e.Player, darinLetter, 1)
		giveItem(e.Player, roxxyKerch, 1)
		qs.SetCond(2)
		return "<html><body>Roxxy:<br>A letter from Darin? How sweet! " +
			"Give him this kerchief from me.</body></html>"
	})

	// Baulro: give potion
	q.AddTalkID(baulro, func(e *quest.Event, qs *quest.QuestState) string {
		if !qs.IsStarted() || qs.GetCond() != 3 {
			return ""
		}

		if !hasItem(e.Player, darinReceipt) {
			return ""
		}

		takeItem(e.Player, darinReceipt, 1)
		giveItem(e.Player, baulroPotion, 1)
		qs.SetCond(4)
		return "<html><body>Baulro:<br>A receipt from Darin? Here's his potion. " +
			"Take it back to him.</body></html>"
	})

	return q
}
