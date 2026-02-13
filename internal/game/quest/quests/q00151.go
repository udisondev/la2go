package quests

import (
	"fmt"

	"github.com/udisondev/la2go/internal/game/quest"
)

// NewQ00151 creates "Cure for Fever Disease" quest.
// Level 15+, any race, one-time. Kill monsters for Poison Sac, deliver medicine between NPCs.
func NewQ00151() *quest.Quest {
	const (
		questID   int32 = 151
		questName       = "Q00151_CureForFever"

		elias   int32 = 30050
		yohanes int32 = 30032

		poisonSac     int32 = 703
		feverMedicine int32 = 704
	)

	q := quest.NewQuest(questID, questName)
	q.AddQuestItem(poisonSac)
	q.AddQuestItem(feverMedicine)

	// Elias: start NPC + finish
	q.AddTalkID(elias, func(e *quest.Event, qs *quest.QuestState) string {
		event := getEvent(e)

		if event == "accept" && qs.State() == quest.StateCreated {
			qs.SetState(quest.StateStarted)
			qs.SetCond(1)
			return "<html><body>Grocer Elias:<br>My wife is sick with fever! " +
				"Kill Marsh Spiders and bring me a Poison Sac. " +
				"Then I'll send you to Yohanes for the cure.</body></html>"
		}

		switch qs.State() {
		case quest.StateCreated:
			if e.Player.Level() < 15 {
				return "<html><body>Grocer Elias:<br>You need to be level 15.</body></html>"
			}
			return fmt.Sprintf("<html><body>Grocer Elias:<br>My wife has a terrible fever!<br>"+
				"I need a Poison Sac from Marsh Spiders to make a cure.<br>"+
				"<a action=\"bypass -h npc_%d_Quest %s accept\">Help Elias</a></body></html>",
				e.TargetID, questName)

		case quest.StateStarted:
			switch qs.GetCond() {
			case 1:
				return "<html><body>Grocer Elias:<br>Kill Marsh Spiders until you find a Poison Sac.</body></html>"
			case 2:
				return "<html><body>Grocer Elias:<br>You have a Poison Sac! " +
					"Take it to Yohanes — he can make the medicine.</body></html>"
			case 3:
				if hasItem(e.Player, feverMedicine) {
					takeItem(e.Player, feverMedicine, 1)
					qs.SetState(quest.StateCompleted)
					return "<html><body>Grocer Elias:<br>The medicine! " +
						"Thank you so much! Take this as a reward.</body></html>"
				}
				return "<html><body>Grocer Elias:<br>Get the medicine from Yohanes.</body></html>"
			}

		case quest.StateCompleted:
			return "<html><body>Grocer Elias:<br>My wife recovered thanks to you!</body></html>"
		}
		return ""
	})

	// Yohanes: crafts medicine from Poison Sac
	q.AddTalkID(yohanes, func(e *quest.Event, qs *quest.QuestState) string {
		if !qs.IsStarted() || qs.GetCond() != 2 {
			return ""
		}
		if !hasItem(e.Player, poisonSac) {
			return "<html><body>Yohanes:<br>Bring me a Poison Sac.</body></html>"
		}
		takeItem(e.Player, poisonSac, 1)
		giveItem(e.Player, feverMedicine, 1)
		qs.SetCond(3)
		return "<html><body>Yohanes:<br>Here is the Fever Medicine. " +
			"Take it to Elias quickly!</body></html>"
	})

	// onKill: marsh spiders (20% drop Poison Sac)
	killHook := func(e *quest.Event, qs *quest.QuestState) string {
		if !qs.IsStarted() || qs.GetCond() != 1 {
			return ""
		}
		if hasItem(e.Player, poisonSac) {
			return ""
		}
		if getRandom(100) < 20 {
			giveItem(e.Player, poisonSac, 1)
			qs.SetCond(2)
		}
		return ""
	}

	q.AddKillID(20103, killHook)
	q.AddKillID(20106, killHook)
	q.AddKillID(20108, killHook)

	return q
}
