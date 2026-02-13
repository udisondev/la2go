package quests

import (
	"fmt"

	"github.com/udisondev/la2go/internal/game/quest"
)

// NewQ00102 creates "Sea of Spores Fever" quest.
// Level 12+, Elf only, one-time. Kill spore creatures, deliver medicines to 4 NPCs.
func NewQ00102() *quest.Quest {
	const (
		questID   int32 = 102
		questName       = "Q00102_SeaOfSporesFever"

		alberius    int32 = 30284
		cobendell   int32 = 30156
		berros      int32 = 30217
		veltress    int32 = 30219
		rayen       int32 = 30221
		gartrandell int32 = 30285

		alberiusLetter    int32 = 964
		evergreenAmulet   int32 = 965
		dryadTears        int32 = 966
		alberiusList      int32 = 746
		cobendellMed1     int32 = 1130
		cobendellMed2     int32 = 1131
		cobendellMed3     int32 = 1132
		cobendellMed4     int32 = 1133
		cobendellMed5     int32 = 1134
		requiredTears     int64 = 10
	)

	q := quest.NewQuest(questID, questName)
	for _, id := range []int32{alberiusLetter, evergreenAmulet, dryadTears, alberiusList,
		cobendellMed1, cobendellMed2, cobendellMed3, cobendellMed4, cobendellMed5} {
		q.AddQuestItem(id)
	}

	// Alberius: start NPC + final reward
	q.AddTalkID(alberius, func(e *quest.Event, qs *quest.QuestState) string {
		event := getEvent(e)

		if event == "accept" && qs.State() == quest.StateCreated {
			qs.SetState(quest.StateStarted)
			qs.SetCond(1)
			giveItem(e.Player, alberiusLetter, 1)
			return "<html><body>Alberius:<br>Take this letter to Cobendell.</body></html>"
		}

		switch qs.State() {
		case quest.StateCreated:
			if e.Player.RaceID() != RaceElf {
				return "<html><body>Alberius:<br>Only Elves may help with this.</body></html>"
			}
			if e.Player.Level() < 12 {
				return "<html><body>Alberius:<br>Come back at level 12.</body></html>"
			}
			return fmt.Sprintf("<html><body>Alberius:<br>A fever plagues the Sea of Spores.<br>"+
				"I need your help to gather the cure.<br>"+
				"<a action=\"bypass -h npc_%d_Quest %s accept\">Accept</a></body></html>",
				e.TargetID, questName)

		case quest.StateStarted:
			switch qs.GetCond() {
			case 1:
				return "<html><body>Alberius:<br>Bring the letter to Cobendell.</body></html>"
			case 2, 3:
				return "<html><body>Alberius:<br>Help Cobendell with the cure.</body></html>"
			case 4:
				takeItem(e.Player, cobendellMed1, 1)
				giveItem(e.Player, alberiusList, 1)
				qs.Set("medicines", "4")
				qs.SetCond(5)
				return "<html><body>Alberius:<br>One medicine for me. Now deliver the rest:<br>" +
					"Berros, Veltress, Rayen, and Gartrandell each need one.</body></html>"
			case 5:
				return "<html><body>Alberius:<br>Deliver medicines to the other elders.</body></html>"
			case 6:
				takeItem(e.Player, alberiusList, 1)
				qs.SetState(quest.StateCompleted)
				giveExp(e, 30000)
				giveAdena(e, 12400)
				return "<html><body>Alberius:<br>All medicines delivered! " +
					"The village is saved. Here is your reward!</body></html>"
			}

		case quest.StateCompleted:
			return "<html><body>Alberius:<br>Thank you for saving us.</body></html>"
		}
		return ""
	})

	// Cobendell: creates medicines
	q.AddTalkID(cobendell, func(e *quest.Event, qs *quest.QuestState) string {
		if !qs.IsStarted() {
			return ""
		}
		switch qs.GetCond() {
		case 1:
			takeItem(e.Player, alberiusLetter, 1)
			giveItem(e.Player, evergreenAmulet, 1)
			qs.SetCond(2)
			return "<html><body>Cobendell:<br>A letter from Alberius! " +
				"Take this amulet to the Sea of Spores and collect 10 Dryad Tears.</body></html>"
		case 2:
			return "<html><body>Cobendell:<br>Collect 10 Dryad Tears.</body></html>"
		case 3:
			takeItem(e.Player, dryadTears, -1)
			takeItem(e.Player, evergreenAmulet, 1)
			giveItem(e.Player, cobendellMed1, 1)
			giveItem(e.Player, cobendellMed2, 1)
			giveItem(e.Player, cobendellMed3, 1)
			giveItem(e.Player, cobendellMed4, 1)
			giveItem(e.Player, cobendellMed5, 1)
			qs.SetCond(4)
			return "<html><body>Cobendell:<br>Excellent! I've made 5 medicines. " +
				"Take them to Alberius.</body></html>"
		case 4:
			return "<html><body>Cobendell:<br>Give the medicines to Alberius first.</body></html>"
		}
		return ""
	})

	// Medicine delivery NPCs (each takes one medicine, decrements counter)
	deliverMedicine := func(medID int32) quest.HookFunc {
		return func(e *quest.Event, qs *quest.QuestState) string {
			if !qs.IsStarted() || qs.GetCond() != 5 {
				return ""
			}
			if !hasItem(e.Player, medID) {
				return "<html><body>I've already received my medicine. Thank you.</body></html>"
			}
			takeItem(e.Player, medID, 1)

			medsLeft := 0
			if v := qs.Get("medicines"); v != "" {
				for _, c := range v {
					medsLeft = medsLeft*10 + int(c-'0')
				}
			}
			medsLeft--

			if medsLeft <= 0 {
				qs.SetCond(6)
			} else {
				qs.Set("medicines", intToStr(medsLeft))
			}

			return "<html><body>Thank you for the medicine!</body></html>"
		}
	}

	q.AddTalkID(berros, deliverMedicine(cobendellMed2))
	q.AddTalkID(veltress, deliverMedicine(cobendellMed3))
	q.AddTalkID(rayen, deliverMedicine(cobendellMed4))
	q.AddTalkID(gartrandell, deliverMedicine(cobendellMed5))

	// onKill: spore creatures (30% drop rate for Dryad Tears)
	killHook := func(e *quest.Event, qs *quest.QuestState) string {
		if !qs.IsStarted() || qs.GetCond() != 2 {
			return ""
		}
		if getRandom(10) >= 3 {
			return ""
		}
		giveItem(e.Player, dryadTears, 1)
		if getItemCount(e.Player, dryadTears) >= requiredTears {
			qs.SetCond(3)
		}
		return ""
	}

	q.AddKillID(20013, killHook)
	q.AddKillID(20019, killHook)

	return q
}

// intToStr converts int to string without importing strconv.
func intToStr(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [10]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}
