package skill

import (
	"testing"

	"github.com/udisondev/la2go/internal/data"
	"github.com/udisondev/la2go/internal/model"
)

// TestSkillCast_FullFlow tests the full cast flow: validation → MP consume → effects → packets.
func TestSkillCast_FullFlow(t *testing.T) {
	cm, packets := makeCastManager()

	// Create player with Triple Slash (id=1, A1 instant)
	player, err := model.NewPlayer(100, 1, 1, "Caster", 20, 0, 0)
	if err != nil {
		t.Fatalf("creating player: %v", err)
	}
	player.AddSkill(1, 1, false)

	tmpl := data.GetSkillTemplate(1, 1)
	if tmpl == nil {
		t.Fatal("Triple Slash template not found")
	}

	// Set enough MP
	player.SetCurrentMP(500)

	mpBefore := player.CurrentMP()

	// Cast skill
	if err := cm.UseMagic(player, 1, false, false); err != nil {
		t.Fatalf("UseMagic failed: %v", err)
	}

	// Verify MP consumed
	if tmpl.MpConsume > 0 {
		expectedMP := mpBefore - tmpl.MpConsume
		if player.CurrentMP() != expectedMP {
			t.Errorf("MP: got %d, want %d", player.CurrentMP(), expectedMP)
		}
	}

	// Verify packets sent (at least MagicSkillUse)
	if len(*packets) == 0 {
		t.Fatal("no packets sent")
	}
	// First two packets should be MagicSkillUse (broadcast + self)
	if (*packets)[0][0] != 0x48 {
		t.Errorf("expected MagicSkillUse (0x48), got 0x%02x", (*packets)[0][0])
	}
}

// TestAutoGetSkills_OnLevelUp tests that skills are auto-granted when leveling up.
func TestAutoGetSkills_OnLevelUp(t *testing.T) {
	if err := data.LoadSkillTrees(); err != nil {
		t.Fatalf("LoadSkillTrees failed: %v", err)
	}

	player, err := model.NewPlayer(200, 2, 1, "Leveler", 1, 0, 0)
	if err != nil {
		t.Fatalf("creating player: %v", err)
	}

	// Apply auto-get skills for level 1 (like enter world)
	autoSkills := data.GetAutoGetSkills(0, 1) // Human Fighter classID=0
	for _, sl := range autoSkills {
		isPassive := false
		if tmpl := data.GetSkillTemplate(sl.SkillID, sl.SkillLevel); tmpl != nil {
			isPassive = tmpl.IsPassive()
		}
		player.AddSkill(sl.SkillID, sl.SkillLevel, isPassive)
	}

	// Should have Lucky (id=194) — auto-get for Human Fighter at level 1
	if !player.HasSkill(194) {
		t.Error("player should have Lucky (id=194) at level 1")
	}

	// Simulate level-up to 20 (Human Fighter gets Weight Limit id=239 autoGet at 20)
	newSkills := data.GetNewAutoGetSkills(0, 20) // Human Fighter classID=0
	for _, sl := range newSkills {
		isPassive := false
		if tmpl := data.GetSkillTemplate(sl.SkillID, sl.SkillLevel); tmpl != nil {
			isPassive = tmpl.IsPassive()
		}
		player.AddSkill(sl.SkillID, sl.SkillLevel, isPassive)
	}

	// Should have Weight Limit (id=239) after level 20
	if !player.HasSkill(239) {
		t.Error("player should have Weight Limit (id=239) at level 20 for Human Fighter")
	}
}

// TestBuffApply_StatChange tests that applying a buff changes stat bonuses.
func TestBuffApply_StatChange(t *testing.T) {
	em := NewEffectManager()

	// Create a buff effect that gives +100 pAtk
	buffEffect := NewBuffEffect(map[string]string{
		"stat":  "pAtk",
		"type":  "ADD",
		"value": "100",
	})

	ae := &ActiveEffect{
		CasterObjID:   1,
		TargetObjID:   2,
		SkillID:       275, // Greater Might
		SkillLevel:    1,
		Effect:        buffEffect,
		RemainingMs:   20000,
		AbnormalType:  "MIGHT",
		AbnormalLevel: 1,
	}

	em.AddBuff(ae)

	// Check stat bonus
	bonus := em.GetStatBonus("pAtk")
	if bonus != 100 {
		t.Errorf("expected 100 pAtk bonus, got %.1f", bonus)
	}

	// Remove buff
	em.RemoveEffect("MIGHT")

	// Bonus should be gone
	bonus = em.GetStatBonus("pAtk")
	if bonus != 0 {
		t.Errorf("expected 0 pAtk after remove, got %.1f", bonus)
	}
}

// TestEffectManager_LinkedToPlayer tests that EffectManager can be linked to Player.
func TestEffectManager_LinkedToPlayer(t *testing.T) {
	player, err := model.NewPlayer(300, 3, 1, "Buffed", 20, 0, 0)
	if err != nil {
		t.Fatalf("creating player: %v", err)
	}

	em := NewEffectManager()
	player.SetEffectManager(em)

	// Verify round-trip
	retrieved := player.EffectManager()
	if retrieved == nil {
		t.Fatal("EffectManager should not be nil after SetEffectManager")
	}

	// Verify it implements StatBonusProvider
	bonus := retrieved.GetStatBonus("pAtk")
	if bonus != 0 {
		t.Errorf("expected 0 bonus initially, got %.1f", bonus)
	}
}
