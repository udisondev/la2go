package integration

import (
	"testing"
	"testing/synctest"

	"github.com/udisondev/la2go/internal/game/combat"
	"github.com/udisondev/la2go/internal/gameserver"
	"github.com/udisondev/la2go/internal/model"
	"github.com/udisondev/la2go/internal/world"
)

// TestWeaponPAtk_Combat verifies weapon pAtk affects combat damage.
// Phase 5.5: Weapon & Equipment System.
func TestWeaponPAtk_Combat(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		// Setup combat managers
		clientMgr := gameserver.NewClientManager()

		attackStanceMgr := combat.NewAttackStanceManager(nil)
		combat.AttackStanceMgr = attackStanceMgr
		attackStanceMgr.Start()
		defer attackStanceMgr.Stop()

		broadcastFunc := func(source *model.Player, data []byte, size int) {
			clientMgr.BroadcastToVisibleNear(source, data, size)
		}
		combatMgr := combat.NewCombatManager(broadcastFunc, nil, nil)
		combat.CombatMgr = combatMgr

		// Get world instance
		worldInst := world.Instance()

		// Create attacker level 10 Human Fighter
		attackerOID := nextOID()
		attacker, err := model.NewPlayer(attackerOID, 100, 200, "Attacker", 10, 0, 0)
		if err != nil {
			t.Fatalf("NewPlayer failed: %v", err)
		}
		attacker.SetLocation(model.NewLocation(0, 0, 0, 0))

		// Create target level 10
		targetOID := nextOID()
		target, err := model.NewPlayer(targetOID, 101, 201, "Target", 10, 0, 0)
		if err != nil {
			t.Fatalf("NewPlayer failed: %v", err)
		}
		target.SetLocation(model.NewLocation(50, 0, 0, 0))

		// Add to world
		if err := worldInst.AddObject(attacker.WorldObject); err != nil {
			t.Fatalf("AddObject attacker failed: %v", err)
		}
		defer worldInst.RemoveObject(attacker.ObjectID())

		if err := worldInst.AddObject(target.WorldObject); err != nil {
			t.Fatalf("AddObject target failed: %v", err)
		}
		defer worldInst.RemoveObject(target.ObjectID())

		// Attack without weapon — use AttackUntilHit for deterministic normal hit
		resultNoWeapon, attempts := combat.AttackUntilHit(combatMgr, attacker, target.WorldObject, 20)
		if resultNoWeapon.Miss {
			t.Fatalf("could not land a hit without weapon in %d attempts", attempts)
		}
		damageNoWeapon := resultNoWeapon.Damage

		t.Logf("Attack without weapon: damage=%d (attempts=%d)", damageNoWeapon, attempts)

		// Reset target HP
		target.SetCurrentHP(target.MaxHP())

		// Equip sword (pAtk=8)
		swordTemplate := &model.ItemTemplate{
			ItemID:      1,
			Name:        "Short Sword",
			Type:        model.ItemTypeWeapon,
			PAtk:        8,
			AttackRange: 40,
		}

		sword, _ := model.NewItem(1000, 1, 100, 1, swordTemplate)
		attacker.Inventory().AddItem(sword)
		attacker.Inventory().EquipItem(sword, model.PaperdollRHand)

		// Attack with weapon — use AttackUntilHit for deterministic normal hit
		resultWithWeapon, attempts2 := combat.AttackUntilHit(combatMgr, attacker, target.WorldObject, 20)
		if resultWithWeapon.Miss {
			t.Fatalf("could not land a hit with weapon in %d attempts", attempts2)
		}
		damageWithWeapon := resultWithWeapon.Damage

		t.Logf("Attack with sword (pAtk=8): damage=%d (attempts=%d)", damageWithWeapon, attempts2)

		// Weapon should increase damage
		if damageWithWeapon <= damageNoWeapon {
			t.Errorf("Weapon should increase damage: with=%d vs without=%d", damageWithWeapon, damageNoWeapon)
		}

		// Damage increase should be significant (proportional to pAtk increase)
		// Attacker pAtk: no weapon ~4 → with sword ~14 (3.5× increase)
		// Damage should increase proportionally
		minExpectedIncrease := float64(damageNoWeapon) * 2.0 // at least 2× damage
		if float64(damageWithWeapon) < minExpectedIncrease {
			t.Errorf("Weapon damage=%d should be >= %.0f (2× no-weapon damage)", damageWithWeapon, minExpectedIncrease)
		}

		t.Log("Weapon pAtk integration test passed!")
	})
}

// TestArmorPDef_Combat verifies armor pDef reduces incoming damage.
// Phase 5.5: Weapon & Equipment System.
func TestArmorPDef_Combat(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		// Setup combat managers
		clientMgr := gameserver.NewClientManager()

		attackStanceMgr := combat.NewAttackStanceManager(nil)
		combat.AttackStanceMgr = attackStanceMgr
		attackStanceMgr.Start()
		defer attackStanceMgr.Stop()

		broadcastFunc := func(source *model.Player, data []byte, size int) {
			clientMgr.BroadcastToVisibleNear(source, data, size)
		}
		combatMgr := combat.NewCombatManager(broadcastFunc, nil, nil)
		combat.CombatMgr = combatMgr

		// Get world instance
		worldInst := world.Instance()

		// Create attacker with weapon (to deal meaningful damage)
		attackerOID := nextOID()
		attacker, _ := model.NewPlayer(attackerOID, 100, 200, "Attacker", 10, 0, 0)
		attacker.SetLocation(model.NewLocation(0, 0, 0, 0))

		swordTemplate := &model.ItemTemplate{
			ItemID:      1,
			Name:        "Sword",
			Type:        model.ItemTypeWeapon,
			PAtk:        10,
			AttackRange: 40,
		}
		sword, _ := model.NewItem(1000, 1, 100, 1, swordTemplate)
		attacker.Inventory().AddItem(sword)
		attacker.Inventory().EquipItem(sword, model.PaperdollRHand)

		// Create defender
		defenderOID := nextOID()
		defender, _ := model.NewPlayer(defenderOID, 101, 201, "Defender", 10, 0, 0)
		defender.SetLocation(model.NewLocation(50, 0, 0, 0))

		// Add to world
		worldInst.AddObject(attacker.WorldObject)
		defer worldInst.RemoveObject(attacker.ObjectID())
		worldInst.AddObject(defender.WorldObject)
		defer worldInst.RemoveObject(defender.ObjectID())

		// Attack defender without armor — use AttackUntilHit for deterministic normal hit
		resultNude, attemptsNude := combat.AttackUntilHit(combatMgr, attacker, defender.WorldObject, 20)
		if resultNude.Miss {
			t.Fatalf("could not land a hit without armor in %d attempts", attemptsNude)
		}
		damageNude := resultNude.Damage

		t.Logf("Defender nude: damage=%d (attempts=%d)", damageNude, attemptsNude)

		// Reset defender HP
		defender.SetCurrentHP(defender.MaxHP())

		// Equip armor (pDef=100)
		armorTemplate := &model.ItemTemplate{
			ItemID:   2,
			Name:     "Full Plate Armor",
			Type:     model.ItemTypeArmor,
			PDef:     100,
			BodyPart: model.ArmorSlotChest,
		}
		armor, _ := model.NewItem(1001, 2, 101, 1, armorTemplate)
		defender.Inventory().AddItem(armor)
		defender.Inventory().EquipItem(armor, model.PaperdollChest)

		// Attack defender with armor — use AttackUntilHit for deterministic normal hit
		resultArmor, attemptsArmor := combat.AttackUntilHit(combatMgr, attacker, defender.WorldObject, 20)
		if resultArmor.Miss {
			t.Fatalf("could not land a hit with armor in %d attempts", attemptsArmor)
		}
		damageWithArmor := resultArmor.Damage

		t.Logf("Defender with armor (pDef=100): damage=%d (attempts=%d)", damageWithArmor, attemptsArmor)

		// Armor should reduce damage
		if damageWithArmor >= damageNude {
			t.Errorf("Armor should reduce damage: with=%d vs nude=%d", damageWithArmor, damageNude)
		}

		// Damage reduction should be significant
		reductionPercent := (1.0 - float64(damageWithArmor)/float64(damageNude)) * 100.0
		t.Logf("Armor reduces damage by %.1f%%", reductionPercent)

		if reductionPercent < 10.0 {
			t.Errorf("Armor damage reduction=%.1f%%, expected >= 10%%", reductionPercent)
		}

		t.Log("Armor pDef integration test passed!")
	})
}
