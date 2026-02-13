package quests

import (
	"fmt"

	"github.com/udisondev/la2go/internal/game/quest"
)

// NewQ00261 creates "Collector's Dream" quest.
// Level 15+, any race, repeatable. Collect 8 Giant Spider Legs for 1000 adena + 2000 EXP.
func NewQ00261() *quest.Quest {
	const (
		questID   int32 = 261
		questName       = "Q00261_CollectorsDream"

		alshupes       int32 = 30222
		giantSpiderLeg int32 = 1087
		requiredCount  int64 = 8
	)

	q := quest.NewQuest(questID, questName)
	q.AddQuestItem(giantSpiderLeg)

	q.AddTalkID(alshupes, func(e *quest.Event, qs *quest.QuestState) string {
		event := getEvent(e)

		if event == "accept" {
			qs.SetState(quest.StateStarted)
			qs.SetCond(1)
			return "<html><body>Trader Alshupes:<br>Bring me 8 Giant Spider Legs. " +
				"I'll pay you well.</body></html>"
		}

		switch qs.State() {
		case quest.StateCreated:
			if e.Player.Level() < 15 {
				return "<html><body>Trader Alshupes:<br>You need to be level 15 to help me.</body></html>"
			}
			return fmt.Sprintf("<html><body>Trader Alshupes:<br>I collect rare spider parts.<br>"+
				"Bring me 8 Giant Spider Legs from the nearby forest.<br>"+
				"<a action=\"bypass -h npc_%d_Quest %s accept\">Accept quest</a></body></html>",
				e.TargetID, questName)

		case quest.StateStarted:
			count := getItemCount(e.Player, giantSpiderLeg)
			if count < requiredCount {
				return fmt.Sprintf("<html><body>Trader Alshupes:<br>You have %d of %d Giant Spider Legs. "+
					"Keep hunting!</body></html>", count, requiredCount)
			}

			takeItem(e.Player, giantSpiderLeg, -1)
			giveAdena(e, 1000)
			giveExp(e, 2000)

			// Repeatable — сбрасываем квест
			qs.SetState(quest.StateCreated)

			return "<html><body>Trader Alshupes:<br>Wonderful! Here are 1000 adena and some experience. " +
				"Come back if you find more!</body></html>"
		}

		return ""
	})

	killHook := func(e *quest.Event, qs *quest.QuestState) string {
		if !qs.IsStarted() || qs.GetCond() != 1 {
			return ""
		}

		count := getItemCount(e.Player, giantSpiderLeg)
		if count >= requiredCount {
			return ""
		}

		giveItem(e.Player, giantSpiderLeg, 1)
		if getItemCount(e.Player, giantSpiderLeg) >= requiredCount {
			qs.SetCond(2)
		}
		return ""
	}

	for _, npcID := range []int32{20308, 20460, 20466} {
		q.AddKillID(npcID, killHook)
	}

	return q
}
