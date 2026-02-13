package combat

import (
	"math"
	"sync"
	"testing"

	"github.com/udisondev/la2go/internal/data"
	"github.com/udisondev/la2go/internal/model"
)

func newTestNpcWithExp(level int32, baseExp int64, baseSP int32) *model.Npc {
	template := model.NewNpcTemplate(
		1000, "TestMob", "Monster",
		level, 1000, 500,
		100, 50, 80, 40,
		300, 120, 253,
		30, 60,
		baseExp, baseSP,
	)
	npc := model.NewNpc(0x20000001, 1000, template)
	return npc
}

func newTestPlayerForExp(t *testing.T, level int32, exp int64) *model.Player {
	t.Helper()
	player, err := model.NewPlayer(1, 1, 1, "TestPlayer", level, 0, 0)
	if err != nil {
		t.Fatalf("NewPlayer: %v", err)
	}
	player.SetExperience(exp)
	return player
}

func newTestPlayerWithID(t *testing.T, objectID uint32, name string, level int32) *model.Player {
	t.Helper()
	player, err := model.NewPlayer(objectID, int64(objectID), 1, name, level, 0, 0)
	if err != nil {
		t.Fatalf("NewPlayer: %v", err)
	}
	return player
}

func TestRewardExpAndSp_BasicReward(t *testing.T) {
	player := newTestPlayerForExp(t, 1, 0)
	npc := newTestNpcWithExp(5, 250, 25)

	var sentPackets []uint32
	var mu sync.Mutex
	sendFn := func(objectID uint32, pktData []byte, size int) {
		mu.Lock()
		sentPackets = append(sentPackets, objectID)
		mu.Unlock()
	}
	broadcastFn := func(source *model.Player, pktData []byte, size int) {}

	RewardExpAndSp(player, npc, sendFn, broadcastFn)

	if player.Experience() != 250 {
		t.Errorf("Experience() = %d, want 250", player.Experience())
	}
	if player.SP() != 25 {
		t.Errorf("SP() = %d, want 25", player.SP())
	}

	// Should have sent at least one packet (exp message)
	if len(sentPackets) == 0 {
		t.Error("expected at least one packet sent to player")
	}
}

func TestRewardExpAndSp_LevelUp(t *testing.T) {
	// Level 1 → Level 2 requires 68 XP
	player := newTestPlayerForExp(t, 1, 0)
	npc := newTestNpcWithExp(5, 100, 10) // 100 XP > 68 threshold

	var sentCount int
	var broadcastCount int
	sendFn := func(objectID uint32, pktData []byte, size int) {
		sentCount++
	}
	broadcastFn := func(source *model.Player, pktData []byte, size int) {
		broadcastCount++
	}

	RewardExpAndSp(player, npc, sendFn, broadcastFn)

	if player.Level() != 2 {
		t.Errorf("Level() = %d, want 2 after level-up", player.Level())
	}
	if player.Experience() != 100 {
		t.Errorf("Experience() = %d, want 100", player.Experience())
	}

	// Should have sent: exp message + level-up message + UserInfo = 3 packets
	// Plus broadcast: SocialAction = 1
	if sentCount < 3 {
		t.Errorf("sentCount = %d, want >= 3 (exp msg + level msg + UserInfo)", sentCount)
	}
	if broadcastCount < 1 {
		t.Errorf("broadcastCount = %d, want >= 1 (SocialAction)", broadcastCount)
	}
}

func TestRewardExpAndSp_MultipleLevelUp(t *testing.T) {
	// Level 1 with enough XP to jump to level 5 (requires 2884)
	player := newTestPlayerForExp(t, 1, 0)
	npc := newTestNpcWithExp(10, 3000, 50) // 3000 > 2884 (level 5)

	sendFn := func(objectID uint32, pktData []byte, size int) {}
	broadcastFn := func(source *model.Player, pktData []byte, size int) {}

	RewardExpAndSp(player, npc, sendFn, broadcastFn)

	if player.Level() != 5 {
		t.Errorf("Level() = %d, want 5 after multi-level-up", player.Level())
	}
	if player.Experience() != 3000 {
		t.Errorf("Experience() = %d, want 3000", player.Experience())
	}
}

func TestRewardExpAndSp_NoRewardForZeroExp(t *testing.T) {
	player := newTestPlayerForExp(t, 1, 0)
	npc := newTestNpcWithExp(5, 0, 0) // zero rewards

	var sentCount int
	sendFn := func(objectID uint32, pktData []byte, size int) {
		sentCount++
	}
	broadcastFn := func(source *model.Player, pktData []byte, size int) {}

	RewardExpAndSp(player, npc, sendFn, broadcastFn)

	if player.Experience() != 0 {
		t.Errorf("Experience() = %d, want 0 for zero-reward NPC", player.Experience())
	}
	if sentCount != 0 {
		t.Errorf("sentCount = %d, want 0 (no packets for zero reward)", sentCount)
	}
}

func TestRewardExpAndSp_NoLevelUpAtMaxLevel(t *testing.T) {
	// Level 80 player should not level up
	maxExp := data.GetExpForLevel(80)
	player := newTestPlayerForExp(t, 80, maxExp)
	npc := newTestNpcWithExp(50, 999999, 100)

	sendFn := func(objectID uint32, pktData []byte, size int) {}
	broadcastFn := func(source *model.Player, pktData []byte, size int) {}

	RewardExpAndSp(player, npc, sendFn, broadcastFn)

	if player.Level() != 80 {
		t.Errorf("Level() = %d, want 80 (max level, no level-up)", player.Level())
	}
}

func TestRewardExpAndSp_SPOnly(t *testing.T) {
	player := newTestPlayerForExp(t, 1, 0)
	npc := newTestNpcWithExp(5, 0, 50) // only SP, no exp

	var sentCount int
	sendFn := func(objectID uint32, pktData []byte, size int) {
		sentCount++
	}
	broadcastFn := func(source *model.Player, pktData []byte, size int) {}

	RewardExpAndSp(player, npc, sendFn, broadcastFn)

	if player.SP() != 50 {
		t.Errorf("SP() = %d, want 50", player.SP())
	}
	if player.Experience() != 0 {
		t.Errorf("Experience() = %d, want 0", player.Experience())
	}
	// Should send SP message
	if sentCount != 1 {
		t.Errorf("sentCount = %d, want 1 (SP message)", sentCount)
	}
}

func TestRewardExpAndSp_HPRestoredOnLevelUp(t *testing.T) {
	player := newTestPlayerForExp(t, 1, 0)
	// Reduce HP to simulate combat
	player.SetCurrentHP(1)

	npc := newTestNpcWithExp(5, 100, 10) // enough for level 2

	sendFn := func(objectID uint32, pktData []byte, size int) {}
	broadcastFn := func(source *model.Player, pktData []byte, size int) {}

	RewardExpAndSp(player, npc, sendFn, broadcastFn)

	if player.Level() != 2 {
		t.Errorf("Level() = %d, want 2", player.Level())
	}
	// HP should be restored to max
	if player.CurrentHP() != player.MaxHP() {
		t.Errorf("CurrentHP() = %d, want MaxHP = %d", player.CurrentHP(), player.MaxHP())
	}
}

// TestRewardExpAndSp_PartyLevelSquaredDistribution verifies that XP is distributed
// proportionally to level² among party members (Java: Party.distributeXpAndSp).
func TestRewardExpAndSp_PartyLevelSquaredDistribution(t *testing.T) {
	// Create 2 party members: level 10 and level 5 (gap=5, within maxLevelGap=9)
	// level 10: 10² = 100
	// level 5:  5²  = 25
	// sqLevelSum = 125
	// level 10 share = 100/125 = 0.80
	// level 5 share  = 25/125  = 0.20
	p1 := newTestPlayerWithID(t, 101, "Warrior", 10)
	p2 := newTestPlayerWithID(t, 102, "Mage", 5)

	party := model.NewParty(1, p1, 0)
	if err := party.AddMember(p2); err != nil {
		t.Fatalf("AddMember: %v", err)
	}
	p1.SetParty(party)
	p2.SetParty(party)

	// NPC with 1000 baseExp and 100 SP
	npc := newTestNpcWithExp(10, 1000, 100)

	sendFn := func(objectID uint32, pktData []byte, size int) {}
	broadcastFn := func(source *model.Player, pktData []byte, size int) {}

	RewardExpAndSp(p1, npc, sendFn, broadcastFn)

	// Party bonus for 2 members: 1.10 (Java BONUS_EXP_SP[1])
	// totalExp = 1000 * 1.10 = 1100
	// totalSP = 100 * 1.10 = 110
	// p1 (level 10): round(1100 * 100/125) = round(880) = 880
	// p2 (level 5):  round(1100 * 25/125)  = round(220) = 220
	expectedP1Exp := int64(math.Round(1100.0 * 100.0 / 125.0))
	expectedP2Exp := int64(math.Round(1100.0 * 25.0 / 125.0))
	expectedP1SP := int64(math.Round(110.0 * 100.0 / 125.0))
	expectedP2SP := int64(math.Round(110.0 * 25.0 / 125.0))

	if p1.Experience() != expectedP1Exp {
		t.Errorf("p1 Experience() = %d, want %d", p1.Experience(), expectedP1Exp)
	}
	if p2.Experience() != expectedP2Exp {
		t.Errorf("p2 Experience() = %d, want %d", p2.Experience(), expectedP2Exp)
	}
	if p1.SP() != expectedP1SP {
		t.Errorf("p1 SP() = %d, want %d", p1.SP(), expectedP1SP)
	}
	if p2.SP() != expectedP2SP {
		t.Errorf("p2 SP() = %d, want %d", p2.SP(), expectedP2SP)
	}

	// Verify total XP is preserved (no loss/gain)
	totalAwarded := p1.Experience() + p2.Experience()
	totalExpected := int64(1100) // 1000 * 1.10
	if math.Abs(float64(totalAwarded-totalExpected)) > 1 {
		t.Errorf("total XP awarded = %d, want ~%d (rounding tolerance 1)", totalAwarded, totalExpected)
	}
}

// TestRewardExpAndSp_PartyLevelGapPenalty verifies that members with level gap > 20
// from the highest-level member receive 0 XP/SP.
func TestRewardExpAndSp_PartyLevelGapPenalty(t *testing.T) {
	// Level 40 and level 19: gap = 21 > maxLevelGap(20) → level 19 gets nothing
	p1 := newTestPlayerWithID(t, 201, "HighLevel", 40)
	p2 := newTestPlayerWithID(t, 202, "LowLevel", 19)

	party := model.NewParty(1, p1, 0)
	if err := party.AddMember(p2); err != nil {
		t.Fatalf("AddMember: %v", err)
	}
	p1.SetParty(party)
	p2.SetParty(party)

	npc := newTestNpcWithExp(30, 1000, 100)

	sendFn := func(objectID uint32, pktData []byte, size int) {}
	broadcastFn := func(source *model.Player, pktData []byte, size int) {}

	RewardExpAndSp(p1, npc, sendFn, broadcastFn)

	// p2 should get 0 XP and 0 SP (level gap = 21 > 20)
	if p2.Experience() != 0 {
		t.Errorf("p2 Experience() = %d, want 0 (level gap penalty)", p2.Experience())
	}
	if p2.SP() != 0 {
		t.Errorf("p2 SP() = %d, want 0 (level gap penalty)", p2.SP())
	}

	// p1 gets ALL the XP (only valid member)
	// Party bonus for 2 members: 1.10, but only 1 valid member in sqLevelSum
	// totalExp = 1000 * 1.10 = 1100
	// p1 share = 40² / 40² = 1.0 → 1100
	if p1.Experience() != 1100 {
		t.Errorf("p1 Experience() = %d, want 1100", p1.Experience())
	}
}

// TestRewardExpAndSp_PartyLevelGapBorderline verifies that level gap exactly 20 is OK.
func TestRewardExpAndSp_PartyLevelGapBorderline(t *testing.T) {
	// Level 40 and level 20: gap = 20 = maxLevelGap → level 20 DOES get XP
	p1 := newTestPlayerWithID(t, 301, "HighLvl", 40)
	p2 := newTestPlayerWithID(t, 302, "MidLvl", 20)

	party := model.NewParty(1, p1, 0)
	if err := party.AddMember(p2); err != nil {
		t.Fatalf("AddMember: %v", err)
	}
	p1.SetParty(party)
	p2.SetParty(party)

	npc := newTestNpcWithExp(30, 1000, 100)

	sendFn := func(objectID uint32, pktData []byte, size int) {}
	broadcastFn := func(source *model.Player, pktData []byte, size int) {}

	RewardExpAndSp(p1, npc, sendFn, broadcastFn)

	// Both should get XP (gap = 20, not > 20)
	if p2.Experience() <= 0 {
		t.Errorf("p2 Experience() = %d, want > 0 (gap=20 is borderline OK)", p2.Experience())
	}
	if p1.Experience() <= 0 {
		t.Errorf("p1 Experience() = %d, want > 0", p1.Experience())
	}

	// p1 (40²=1600) should get more than p2 (20²=400)
	if p1.Experience() <= p2.Experience() {
		t.Errorf("p1 Experience() = %d should be > p2 Experience() = %d (level²-proportional)",
			p1.Experience(), p2.Experience())
	}
}

// TestRewardExpAndSp_PartyEqualLevels verifies equal-level members get equal shares.
func TestRewardExpAndSp_PartyEqualLevels(t *testing.T) {
	p1 := newTestPlayerWithID(t, 401, "P1", 40)
	p2 := newTestPlayerWithID(t, 402, "P2", 40)

	party := model.NewParty(1, p1, 0)
	if err := party.AddMember(p2); err != nil {
		t.Fatalf("AddMember: %v", err)
	}
	p1.SetParty(party)
	p2.SetParty(party)

	npc := newTestNpcWithExp(30, 1000, 100)

	sendFn := func(objectID uint32, pktData []byte, size int) {}
	broadcastFn := func(source *model.Player, pktData []byte, size int) {}

	RewardExpAndSp(p1, npc, sendFn, broadcastFn)

	// Equal levels → equal shares
	// totalExp = 1000 * 1.10 = 1100
	// Each: round(1100 * 0.5) = 550
	if p1.Experience() != p2.Experience() {
		t.Errorf("unequal XP for equal levels: p1=%d, p2=%d", p1.Experience(), p2.Experience())
	}
	if p1.SP() != p2.SP() {
		t.Errorf("unequal SP for equal levels: p1=%d, p2=%d", p1.SP(), p2.SP())
	}
}
