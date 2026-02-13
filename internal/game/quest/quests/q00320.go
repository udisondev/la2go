package quests

import (
	"fmt"

	"github.com/udisondev/la2go/internal/game/quest"
)

// NewQ00320 creates "Bones Tell the Future" quest.
// Level 10+, Dark Elf only, repeatable. Collect 10 Bone Fragments for 8470 adena.
func NewQ00320() *quest.Quest {
	const (
		questID   int32 = 320
		questName       = "Q00320_BonesTellTheFuture"

		kaitar        int32 = 30359
		boneFragment  int32 = 809
		requiredCount int64 = 10
	)

	q := quest.NewQuest(questID, questName)
	q.AddQuestItem(boneFragment)

	q.AddTalkID(kaitar, func(e *quest.Event, qs *quest.QuestState) string {
		event := getEvent(e)

		if event == "accept" {
			qs.SetState(quest.StateStarted)
			qs.SetCond(1)
			return "<html><body>Seer Kaitar:<br>Collect 10 Bone Fragments from the undead.</body></html>"
		}

		switch qs.State() {
		case quest.StateCreated:
			if e.Player.RaceID() != RaceDarkElf {
				return "<html><body>Seer Kaitar:<br>Only Dark Elves may undertake this task.</body></html>"
			}
			if e.Player.Level() < 10 {
				return "<html><body>Seer Kaitar:<br>Come back at level 10.</body></html>"
			}
			return fmt.Sprintf("<html><body>Seer Kaitar:<br>I need Bone Fragments for my divinations.<br>"+
				"Bring me 10 from the nearby undead.<br>"+
				"<a action=\"bypass -h npc_%d_Quest %s accept\">Accept</a></body></html>",
				e.TargetID, questName)

		case quest.StateStarted:
			count := getItemCount(e.Player, boneFragment)
			if count < requiredCount {
				return fmt.Sprintf("<html><body>Seer Kaitar:<br>You have %d of %d Bone Fragments.</body></html>",
					count, requiredCount)
			}

			takeItem(e.Player, boneFragment, -1)
			giveAdena(e, 8470)

			// Repeatable
			qs.SetState(quest.StateCreated)

			return "<html><body>Seer Kaitar:<br>These are perfect. " +
				"Here are 8470 adena.</body></html>"
		}

		return ""
	})

	// onKill: undead monsters
	killHook := func(e *quest.Event, qs *quest.QuestState) string {
		if !qs.IsStarted() || qs.GetCond() != 1 {
			return ""
		}

		count := getItemCount(e.Player, boneFragment)
		if count >= requiredCount {
			return ""
		}

		dropRate := 18
		if e.NpcID == 20518 {
			dropRate = 20
		}

		if getRandom(100) < dropRate {
			giveItem(e.Player, boneFragment, 1)
			if getItemCount(e.Player, boneFragment) >= requiredCount {
				qs.SetCond(2)
			}
		}
		return ""
	}

	q.AddKillID(20517, killHook)
	q.AddKillID(20518, killHook)

	return q
}
