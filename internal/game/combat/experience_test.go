package combat

import (
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
