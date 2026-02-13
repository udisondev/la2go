package clan

import (
	"testing"
)

func TestClanWar_DeclareAndAccept(t *testing.T) {
	attacker := New(1, "Attackers", 100)
	attacker.SetLevel(5)

	defender := New(2, "Defenders", 200)
	defender.SetLevel(4)

	// Объявляем войну
	if err := attacker.DeclareWar(defender.ID()); err != nil {
		t.Fatalf("DeclareWar: %v", err)
	}

	// Враг принимает (в Interlude автоматически)
	defender.AcceptWar(attacker.ID())

	// Проверяем состояние атакующего
	if !attacker.IsAtWarWith(defender.ID()) {
		t.Error("attacker.IsAtWarWith(defender) = false, want true")
	}
	if attacker.IsUnderAttack(defender.ID()) {
		t.Error("attacker.IsUnderAttack(defender) = true, want false")
	}

	// Проверяем состояние защитника
	if defender.IsUnderAttack(attacker.ID()) != true {
		t.Error("defender.IsUnderAttack(attacker) = false, want true")
	}
	if defender.IsAtWarWith(attacker.ID()) {
		t.Error("defender.IsAtWarWith(attacker) = true, want false")
	}

	// Проверяем списки
	attackerWars := attacker.WarList()
	if len(attackerWars) != 1 || attackerWars[0] != defender.ID() {
		t.Errorf("attacker.WarList() = %v, want [%d]", attackerWars, defender.ID())
	}

	defenderAttackers := defender.AttackerList()
	if len(defenderAttackers) != 1 || defenderAttackers[0] != attacker.ID() {
		t.Errorf("defender.AttackerList() = %v, want [%d]", defenderAttackers, attacker.ID())
	}
}

func TestClanWar_DuplicateDeclaration(t *testing.T) {
	c1 := New(1, "Clan1", 100)
	c2 := New(2, "Clan2", 200)

	if err := c1.DeclareWar(c2.ID()); err != nil {
		t.Fatalf("first DeclareWar: %v", err)
	}

	// Повторное объявление войны должно вернуть ошибку
	if err := c1.DeclareWar(c2.ID()); err != ErrAlreadyAtWar {
		t.Errorf("duplicate DeclareWar = %v, want ErrAlreadyAtWar", err)
	}
}

func TestClanWar_EndWar(t *testing.T) {
	attacker := New(1, "Attackers", 100)
	defender := New(2, "Defenders", 200)

	if err := attacker.DeclareWar(defender.ID()); err != nil {
		t.Fatalf("DeclareWar: %v", err)
	}
	defender.AcceptWar(attacker.ID())

	// Завершаем войну с обеих сторон
	attacker.EndWar(defender.ID())
	defender.EndWar(attacker.ID())

	// Проверяем, что война завершена
	if attacker.IsAtWarWith(defender.ID()) {
		t.Error("attacker still at war after EndWar")
	}
	if defender.IsUnderAttack(attacker.ID()) {
		t.Error("defender still under attack after EndWar")
	}

	if len(attacker.WarList()) != 0 {
		t.Errorf("attacker.WarList() = %v, want empty", attacker.WarList())
	}
	if len(defender.AttackerList()) != 0 {
		t.Errorf("defender.AttackerList() = %v, want empty", defender.AttackerList())
	}
}

func TestClanWar_ReputationCost(t *testing.T) {
	attacker := New(1, "Attackers", 100)
	attacker.SetReputation(2000)

	// Имитируем штраф за объявление войны
	result := attacker.AddReputation(-500)
	if result != 1500 {
		t.Errorf("Reputation after war declaration = %d, want 1500", result)
	}

	// Ещё один штраф за остановку войны
	result = attacker.AddReputation(-500)
	if result != 1000 {
		t.Errorf("Reputation after war stop = %d, want 1000", result)
	}
}

func TestClanWar_SurrenderReputation(t *testing.T) {
	attacker := New(1, "Attackers", 100)
	attacker.SetReputation(2000)

	defender := New(2, "Defenders", 200)
	defender.SetReputation(1000)

	// Объявляем и принимаем войну
	if err := attacker.DeclareWar(defender.ID()); err != nil {
		t.Fatalf("DeclareWar: %v", err)
	}
	defender.AcceptWar(attacker.ID())

	// Капитуляция: атакующий теряет 500, защитник получает 500
	attacker.AddReputation(-500)
	defender.AddReputation(500)

	attacker.EndWar(defender.ID())
	defender.EndWar(attacker.ID())

	if attacker.Reputation() != 1500 {
		t.Errorf("attacker Reputation after surrender = %d, want 1500", attacker.Reputation())
	}
	if defender.Reputation() != 1500 {
		t.Errorf("defender Reputation after surrender = %d, want 1500", defender.Reputation())
	}
}

func TestClanWar_MultipleWars(t *testing.T) {
	c1 := New(1, "Clan1", 100)
	c2 := New(2, "Clan2", 200)
	c3 := New(3, "Clan3", 300)

	// c1 объявляет войну c2 и c3
	if err := c1.DeclareWar(c2.ID()); err != nil {
		t.Fatalf("DeclareWar(c2): %v", err)
	}
	c2.AcceptWar(c1.ID())

	if err := c1.DeclareWar(c3.ID()); err != nil {
		t.Fatalf("DeclareWar(c3): %v", err)
	}
	c3.AcceptWar(c1.ID())

	wars := c1.WarList()
	if len(wars) != 2 {
		t.Errorf("c1.WarList() has %d entries, want 2", len(wars))
	}

	// Завершаем войну только с c2
	c1.EndWar(c2.ID())
	c2.EndWar(c1.ID())

	wars = c1.WarList()
	if len(wars) != 1 {
		t.Errorf("c1.WarList() has %d entries after ending war with c2, want 1", len(wars))
	}
	if !c1.IsAtWarWith(c3.ID()) {
		t.Error("c1 should still be at war with c3")
	}
}

func TestClanWar_EndNonExistentWar(t *testing.T) {
	c := New(1, "Clan1", 100)

	// EndWar на несуществующую войну не должен паниковать
	c.EndWar(999)

	if c.IsAtWarWith(999) {
		t.Error("should not be at war with non-existent clan")
	}
}

func TestClanWar_BidirectionalWar(t *testing.T) {
	// Оба клана объявляют войну друг другу
	c1 := New(1, "Clan1", 100)
	c2 := New(2, "Clan2", 200)

	// c1 объявляет c2
	if err := c1.DeclareWar(c2.ID()); err != nil {
		t.Fatalf("c1.DeclareWar(c2): %v", err)
	}
	c2.AcceptWar(c1.ID())

	// c2 объявляет c1
	if err := c2.DeclareWar(c1.ID()); err != nil {
		t.Fatalf("c2.DeclareWar(c1): %v", err)
	}
	c1.AcceptWar(c2.ID())

	// Оба клана должны быть и в atWarWith, и в atWarAttackers
	if !c1.IsAtWarWith(c2.ID()) {
		t.Error("c1.IsAtWarWith(c2) = false")
	}
	if !c1.IsUnderAttack(c2.ID()) {
		t.Error("c1.IsUnderAttack(c2) = false")
	}
	if !c2.IsAtWarWith(c1.ID()) {
		t.Error("c2.IsAtWarWith(c1) = false")
	}
	if !c2.IsUnderAttack(c1.ID()) {
		t.Error("c2.IsUnderAttack(c1) = false")
	}

	// EndWar очищает обе стороны у одного клана
	c1.EndWar(c2.ID())

	// c1 больше не воюет
	if c1.IsAtWarWith(c2.ID()) {
		t.Error("c1 should not be at war after EndWar")
	}
	if c1.IsUnderAttack(c2.ID()) {
		t.Error("c1 should not be under attack after EndWar")
	}

	// c2 всё ещё считает, что объявил войну c1 (пока c2 сам не вызовет EndWar)
	if !c2.IsAtWarWith(c1.ID()) {
		t.Error("c2 should still be at war until its own EndWar")
	}
}

func TestClanWar_MemberTitleSetting(t *testing.T) {
	c := New(1, "TestClan", 100)
	c.SetLevel(5)

	leader := NewMember(100, "Leader", 80, 0, PledgeMain, 1)
	if err := c.AddMember(leader); err != nil {
		t.Fatalf("AddMember(leader): %v", err)
	}

	member := NewMember(200, "Member", 60, 0, PledgeMain, 5)
	if err := c.AddMember(member); err != nil {
		t.Fatalf("AddMember(member): %v", err)
	}

	// Лидер устанавливает титул мемберу
	target := c.MemberByName("Member")
	if target == nil {
		t.Fatal("MemberByName(Member) = nil")
	}

	target.SetTitle("Knight Commander")

	if target.Title() != "Knight Commander" {
		t.Errorf("Title = %q, want %q", target.Title(), "Knight Commander")
	}
}

func TestClanWar_LeaderPrivilegeCheck(t *testing.T) {
	c := New(1, "TestClan", 100)

	leader := NewMember(100, "Leader", 80, 0, PledgeMain, 1)
	if err := c.AddMember(leader); err != nil {
		t.Fatalf("AddMember(leader): %v", err)
	}

	officer := NewMember(200, "Officer", 60, 0, PledgeMain, 2)
	if err := c.AddMember(officer); err != nil {
		t.Fatalf("AddMember(officer): %v", err)
	}

	regular := NewMember(300, "Regular", 40, 0, PledgeMain, 5)
	if err := c.AddMember(regular); err != nil {
		t.Fatalf("AddMember(regular): %v", err)
	}

	// Лидер имеет все привилегии
	if !leader.HasPrivilege(PrivAll) {
		t.Error("leader should have PrivAll")
	}

	// Офицер (grade 2) имеет право на войну
	if !officer.HasPrivilege(PrivCLPledgeWar) {
		t.Error("officer should have PrivCLPledgeWar")
	}

	// Обычный мембер (grade 5) не имеет права на войну
	if regular.HasPrivilege(PrivCLPledgeWar) {
		t.Error("regular member should not have PrivCLPledgeWar")
	}

	// Лидер имеет право выдавать титулы
	if !leader.HasPrivilege(PrivCLGiveTitles) {
		t.Error("leader should have PrivCLGiveTitles")
	}

	// Офицер (grade 2) имеет право на титулы
	if !officer.HasPrivilege(PrivCLGiveTitles) {
		t.Error("officer should have PrivCLGiveTitles")
	}

	// Обычный мембер (grade 5) не имеет права на титулы
	if regular.HasPrivilege(PrivCLGiveTitles) {
		t.Error("regular member should not have PrivCLGiveTitles")
	}
}
