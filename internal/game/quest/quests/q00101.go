package quests

import (
	"fmt"

	"github.com/udisondev/la2go/internal/game/quest"
)

// NewQ00101 creates "Sword of Solidarity" quest.
// Level 9+, Human only, one-time. Collect broken blade parts, forge the Sword of Solidarity.
func NewQ00101() *quest.Quest {
	const (
		questID   int32 = 101
		questName       = "Q00101_SwordOfSolidarity"

		roien  int32 = 30008
		altran int32 = 30283

		roiensLetter     int32 = 796
		dirToRuins       int32 = 937
		brokenBladeTop   int32 = 741
		brokenBladeBtm   int32 = 740
		altransNote      int32 = 742
		brokenSwordHandle int32 = 739
	)

	q := quest.NewQuest(questID, questName)
	q.AddQuestItem(roiensLetter)
	q.AddQuestItem(dirToRuins)
	q.AddQuestItem(brokenBladeTop)
	q.AddQuestItem(brokenBladeBtm)
	q.AddQuestItem(altransNote)
	q.AddQuestItem(brokenSwordHandle)

	// Roien: start NPC
	q.AddTalkID(roien, func(e *quest.Event, qs *quest.QuestState) string {
		event := getEvent(e)

		if event == "accept" && qs.State() == quest.StateCreated {
			qs.SetState(quest.StateStarted)
			qs.SetCond(1)
			giveItem(e.Player, roiensLetter, 1)
			return "<html><body>Captain Roien:<br>Take this letter to Blacksmith Altran. " +
				"He'll tell you what to do next.</body></html>"
		}

		switch qs.State() {
		case quest.StateCreated:
			if e.Player.RaceID() != RaceHuman {
				return "<html><body>Captain Roien:<br>Only Humans may take this quest.</body></html>"
			}
			if e.Player.Level() < 9 {
				return "<html><body>Captain Roien:<br>Come back at level 9.</body></html>"
			}
			return fmt.Sprintf("<html><body>Captain Roien:<br>I need a brave adventurer "+
				"to help recover a legendary sword.<br>"+
				"<a action=\"bypass -h npc_%d_Quest %s accept\">Accept</a></body></html>",
				e.TargetID, questName)

		case quest.StateStarted:
			switch qs.GetCond() {
			case 1:
				return "<html><body>Captain Roien:<br>Go see Blacksmith Altran.</body></html>"
			case 2:
				return "<html><body>Captain Roien:<br>Find the broken blade pieces in the ruins.</body></html>"
			case 3:
				return "<html><body>Captain Roien:<br>Return the blade pieces to Altran.</body></html>"
			case 4:
				takeItem(e.Player, altransNote, 1)
				giveItem(e.Player, brokenSwordHandle, 1)
				qs.SetCond(5)
				return "<html><body>Captain Roien:<br>Good work! " +
					"Here is the sword handle. Take it back to Altran.</body></html>"
			case 5:
				return "<html><body>Captain Roien:<br>Bring the handle to Altran.</body></html>"
			}

		case quest.StateCompleted:
			return "<html><body>Captain Roien:<br>The sword has been restored!</body></html>"
		}
		return ""
	})

	// Altran: blacksmith
	q.AddTalkID(altran, func(e *quest.Event, qs *quest.QuestState) string {
		if !qs.IsStarted() {
			if qs.State() == quest.StateCompleted {
				return "<html><body>Blacksmith Altran:<br>The sword serves you well, I hope.</body></html>"
			}
			return ""
		}

		switch qs.GetCond() {
		case 1:
			takeItem(e.Player, roiensLetter, 1)
			giveItem(e.Player, dirToRuins, 1)
			qs.SetCond(2)
			return "<html><body>Blacksmith Altran:<br>A letter from Roien! " +
				"Go to the ruins and find two pieces of a broken blade.</body></html>"
		case 2:
			return "<html><body>Blacksmith Altran:<br>Hunt the monsters in the ruins " +
				"for the Broken Blade Top and Bottom.</body></html>"
		case 3:
			takeItem(e.Player, dirToRuins, 1)
			takeItem(e.Player, brokenBladeTop, 1)
			takeItem(e.Player, brokenBladeBtm, 1)
			giveItem(e.Player, altransNote, 1)
			qs.SetCond(4)
			return "<html><body>Blacksmith Altran:<br>Both pieces! " +
				"Take this note back to Roien.</body></html>"
		case 4:
			return "<html><body>Blacksmith Altran:<br>Take the note to Roien.</body></html>"
		case 5:
			takeItem(e.Player, brokenSwordHandle, 1)
			qs.SetState(quest.StateCompleted)
			giveExp(e, 25747)
			giveAdena(e, 6981)
			return "<html><body>Blacksmith Altran:<br>The Sword of Solidarity is complete! " +
				"Take it with honor.</body></html>"
		}
		return ""
	})

	// onKill: ruins monsters (20% drop for each blade piece)
	killHook := func(e *quest.Event, qs *quest.QuestState) string {
		if !qs.IsStarted() || qs.GetCond() != 2 {
			return ""
		}

		if getRandom(5) != 0 {
			return ""
		}

		if !hasItem(e.Player, brokenBladeTop) {
			giveItem(e.Player, brokenBladeTop, 1)
			if hasItem(e.Player, brokenBladeBtm) {
				qs.SetCond(3)
			}
		} else if !hasItem(e.Player, brokenBladeBtm) {
			giveItem(e.Player, brokenBladeBtm, 1)
			if hasItem(e.Player, brokenBladeTop) {
				qs.SetCond(3)
			}
		}
		return ""
	}

	q.AddKillID(20361, killHook)
	q.AddKillID(20362, killHook)

	return q
}
