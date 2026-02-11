package skill

import (
	"testing"
	"time"

	"github.com/udisondev/la2go/internal/data"
	"github.com/udisondev/la2go/internal/model"
)

func init() {
	// Load skill data for tests
	if err := data.LoadSkills(); err != nil {
		panic("failed to load skills: " + err.Error())
	}
}

func makeTestPlayer(t *testing.T, objectID uint32, skillID, skillLevel int32, passive bool) *model.Player {
	t.Helper()
	p, err := model.NewPlayer(objectID, 1, 1, "TestCaster", 20, 0, 0)
	if err != nil {
		t.Fatalf("creating test player: %v", err)
	}
	p.AddSkill(skillID, skillLevel, passive)
	return p
}

func makeCastManager() (*CastManager, *[][]byte) {
	var sentPackets [][]byte
	sendFn := func(objectID uint32, d []byte, size int) {
		cp := make([]byte, len(d))
		copy(cp, d)
		sentPackets = append(sentPackets, cp)
	}
	broadcastFn := func(source *model.Player, d []byte, size int) {
		cp := make([]byte, len(d))
		copy(cp, d)
		sentPackets = append(sentPackets, cp)
	}
	em := NewEffectManager()
	getEM := func(objectID uint32) *EffectManager {
		return em
	}

	cm := NewCastManager(sendFn, broadcastFn, getEM)
	return cm, &sentPackets
}

func TestUseMagic_BasicCast(t *testing.T) {
	cm, packets := makeCastManager()

	// Power Strike (id=3) is instant A1, no HitTime
	player := makeTestPlayer(t, 100, 3, 1, false)

	if err := cm.UseMagic(player, 3, false, false); err != nil {
		t.Fatalf("UseMagic failed: %v", err)
	}

	// Should have sent at least MagicSkillUse packets (broadcast + self)
	if len(*packets) == 0 {
		t.Fatal("expected packets to be sent")
	}

	// First packet should be MagicSkillUse (opcode 0x48)
	if (*packets)[0][0] != 0x48 {
		t.Errorf("expected opcode 0x48, got 0x%02x", (*packets)[0][0])
	}
}

func TestUseMagic_SkillNotLearned(t *testing.T) {
	cm, _ := makeCastManager()

	player := makeTestPlayer(t, 100, 3, 1, false) // only has skill 3

	err := cm.UseMagic(player, 999, false, false)
	if err == nil {
		t.Fatal("expected error for unlearned skill")
	}
}

func TestUseMagic_PassiveCannotCast(t *testing.T) {
	cm, _ := makeCastManager()

	// Light Armor Mastery (id=228) is passive
	player := makeTestPlayer(t, 100, 228, 1, true)

	err := cm.UseMagic(player, 228, false, false)
	if err == nil {
		t.Fatal("expected error for passive skill cast")
	}
}

func TestUseMagic_OnCooldown(t *testing.T) {
	cm, _ := makeCastManager()

	player := makeTestPlayer(t, 100, 3, 1, false)

	// First cast succeeds
	if err := cm.UseMagic(player, 3, false, false); err != nil {
		t.Fatalf("first cast failed: %v", err)
	}

	// Force cooldown
	cm.cooldowns.Store(cooldownKey(100, 3), time.Now().Add(10*time.Second))

	// Second cast should fail on cooldown
	err := cm.UseMagic(player, 3, false, false)
	if err == nil {
		t.Fatal("expected cooldown error")
	}
}

func TestUseMagic_NotEnoughMP(t *testing.T) {
	cm, _ := makeCastManager()

	// Wind Strike (id=56) consumes MP
	player := makeTestPlayer(t, 100, 56, 1, false)

	// Set MP to 0
	player.SetCurrentMP(0)

	err := cm.UseMagic(player, 56, false, false)
	if err == nil {
		t.Fatal("expected not enough MP error")
	}
}

func TestIsOnCooldown(t *testing.T) {
	cm, _ := makeCastManager()

	if cm.IsOnCooldown(100, 3) {
		t.Fatal("should not be on cooldown initially")
	}

	cm.cooldowns.Store(cooldownKey(100, 3), time.Now().Add(10*time.Second))

	if !cm.IsOnCooldown(100, 3) {
		t.Fatal("should be on cooldown")
	}
}

func TestUseMagic_ConsumesMP(t *testing.T) {
	cm, _ := makeCastManager()

	// Wind Strike (id=56) has MpConsume > 0
	player := makeTestPlayer(t, 100, 56, 1, false)
	player.SetCurrentMP(500) // set enough MP

	mpBefore := player.CurrentMP()

	if err := cm.UseMagic(player, 56, false, false); err != nil {
		t.Fatalf("UseMagic failed: %v", err)
	}

	tmpl := data.GetSkillTemplate(56, 1)
	if tmpl == nil {
		t.Fatal("skill template 56 L1 not found")
	}

	expectedMP := mpBefore - tmpl.MpConsume
	if player.CurrentMP() != expectedMP {
		t.Errorf("MP after cast: got %d, want %d", player.CurrentMP(), expectedMP)
	}
}
