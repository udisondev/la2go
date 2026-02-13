package quests

import (
	"fmt"

	"github.com/udisondev/la2go/internal/game/quest"
)

// NewQ00152 creates "Shards of Golem" quest.
// Level 10+, any race, one-time. Collect 5 Golem Shards for Harris via Altran.
func NewQ00152() *quest.Quest {
	const (
		questID   int32 = 152
		questName       = "Q00152_ShardsOfGolem"

		harris int32 = 30035
		altran int32 = 30283

		harrisReceipt1 int32 = 1008
		harrisReceipt2 int32 = 1009
		golemShard     int32 = 1010
		toolBox        int32 = 1011

		requiredShards int64 = 5
	)

	q := quest.NewQuest(questID, questName)
	q.AddQuestItem(harrisReceipt1)
	q.AddQuestItem(harrisReceipt2)
	q.AddQuestItem(golemShard)
	q.AddQuestItem(toolBox)

	// Harris: start + finish
	q.AddTalkID(harris, func(e *quest.Event, qs *quest.QuestState) string {
		event := getEvent(e)

		if event == "accept" && qs.State() == quest.StateCreated {
			qs.SetState(quest.StateStarted)
			qs.SetCond(1)
			giveItem(e.Player, harrisReceipt1, 1)
			return "<html><body>Blacksmith Harris:<br>Take this receipt to Altran.</body></html>"
		}

		switch qs.State() {
		case quest.StateCreated:
			if e.Player.Level() < 10 {
				return "<html><body>Blacksmith Harris:<br>Come back at level 10.</body></html>"
			}
			return fmt.Sprintf("<html><body>Blacksmith Harris:<br>I need Stone Golem shards.<br>"+
				"Take this receipt to Blacksmith Altran — he'll explain what to do.<br>"+
				"<a action=\"bypass -h npc_%d_Quest %s accept\">Accept</a></body></html>",
				e.TargetID, questName)

		case quest.StateStarted:
			switch qs.GetCond() {
			case 1:
				return "<html><body>Blacksmith Harris:<br>Go to Altran with the receipt.</body></html>"
			case 2:
				return "<html><body>Blacksmith Harris:<br>Altran needs the Golem Shards.</body></html>"
			case 3:
				return "<html><body>Blacksmith Harris:<br>Bring the shards to Altran.</body></html>"
			case 4:
				if hasItem(e.Player, toolBox) {
					takeItem(e.Player, toolBox, 1)
					takeItem(e.Player, harrisReceipt2, 1)
					qs.SetState(quest.StateCompleted)
					giveExp(e, 5000)
					return "<html><body>Blacksmith Harris:<br>The tool box! " +
						"Here is a Wooden Breastplate as thanks.</body></html>"
				}
				return "<html><body>Blacksmith Harris:<br>Altran has the tool box.</body></html>"
			}

		case quest.StateCompleted:
			return "<html><body>Blacksmith Harris:<br>Thank you again.</body></html>"
		}
		return ""
	})

	// Altran: intermediate NPC
	q.AddTalkID(altran, func(e *quest.Event, qs *quest.QuestState) string {
		if !qs.IsStarted() {
			return ""
		}

		switch qs.GetCond() {
		case 1:
			takeItem(e.Player, harrisReceipt1, 1)
			giveItem(e.Player, harrisReceipt2, 1)
			qs.SetCond(2)
			return "<html><body>Blacksmith Altran:<br>Harris needs 5 Golem Shards. " +
				"Kill Stone Golems and bring them to me.</body></html>"
		case 2:
			return "<html><body>Blacksmith Altran:<br>Kill Stone Golems for shards.</body></html>"
		case 3:
			shards := getItemCount(e.Player, golemShard)
			if shards < requiredShards {
				return fmt.Sprintf("<html><body>Blacksmith Altran:<br>"+
					"You need %d more shards.</body></html>", requiredShards-shards)
			}
			takeItem(e.Player, golemShard, -1)
			giveItem(e.Player, toolBox, 1)
			qs.SetCond(4)
			return "<html><body>Blacksmith Altran:<br>Excellent! Here is the Tool Box. " +
				"Take it to Harris.</body></html>"
		case 4:
			return "<html><body>Blacksmith Altran:<br>Bring the Tool Box to Harris.</body></html>"
		}
		return ""
	})

	// onKill: Stone Golem (30% drop)
	q.AddKillID(20016, func(e *quest.Event, qs *quest.QuestState) string {
		if !qs.IsStarted() || qs.GetCond() != 2 {
			return ""
		}

		count := getItemCount(e.Player, golemShard)
		if count >= requiredShards {
			return ""
		}

		if getRandom(100) < 30 {
			giveItem(e.Player, golemShard, 1)
			if getItemCount(e.Player, golemShard) >= requiredShards {
				qs.SetCond(3)
			}
		}
		return ""
	})

	return q
}
