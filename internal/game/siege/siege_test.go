package siege

import (
	"sync"
	"testing"
	"time"
)

func TestNewCastle(t *testing.T) {
	t.Parallel()
	c := NewCastle(1, "Gludio", 30)

	if c.ID() != 1 {
		t.Errorf("ID() = %d; want 1", c.ID())
	}
	if c.Name() != "Gludio" {
		t.Errorf("Name() = %q; want %q", c.Name(), "Gludio")
	}
	if c.OwnerClanID() != 0 {
		t.Errorf("OwnerClanID() = %d; want 0", c.OwnerClanID())
	}
	if c.HasOwner() {
		t.Error("HasOwner() = true; want false")
	}
}

func TestCastle_TaxRate(t *testing.T) {
	t.Parallel()
	c := NewCastle(1, "Gludio", 30)

	c.SetTaxRate(15)
	if c.TaxRate() != 15 {
		t.Errorf("TaxRate() = %d; want 15", c.TaxRate())
	}

	// Clamp to max.
	c.SetTaxRate(50)
	if c.TaxRate() != MaxTaxRate {
		t.Errorf("TaxRate() = %d; want %d", c.TaxRate(), MaxTaxRate)
	}

	// Clamp to min.
	c.SetTaxRate(-5)
	if c.TaxRate() != MinTaxRate {
		t.Errorf("TaxRate() = %d; want %d", c.TaxRate(), MinTaxRate)
	}
}

func TestCastle_Treasury(t *testing.T) {
	t.Parallel()
	c := NewCastle(1, "Gludio", 30)

	c.AddToTreasury(1000)
	if c.Treasury() != 1000 {
		t.Errorf("Treasury() = %d; want 1000", c.Treasury())
	}

	c.AddToTreasury(-2000)
	if c.Treasury() != 0 {
		t.Errorf("Treasury() = %d; want 0 (clamped)", c.Treasury())
	}
}

func TestCastle_TaxAmount(t *testing.T) {
	t.Parallel()
	c := NewCastle(1, "Gludio", 30)

	c.SetTaxRate(10)
	if got := c.TaxAmount(1000); got != 100 {
		t.Errorf("TaxAmount(1000) = %d; want 100", got)
	}

	c.SetTaxRate(0)
	if got := c.TaxAmount(1000); got != 0 {
		t.Errorf("TaxAmount(1000) = %d; want 0", got)
	}
}

func TestCastle_SiegeDate(t *testing.T) {
	t.Parallel()
	c := NewCastle(1, "Gludio", 30)

	now := time.Now()
	c.SetSiegeDate(now)
	if !c.SiegeDate().Equal(now) {
		t.Errorf("SiegeDate() mismatch")
	}
}

func TestCastle_HasOwner(t *testing.T) {
	t.Parallel()
	c := NewCastle(1, "Gludio", 30)

	if c.HasOwner() {
		t.Error("HasOwner() = true on new castle; want false")
	}
	c.SetOwnerClanID(42)
	if !c.HasOwner() {
		t.Error("HasOwner() = false after SetOwnerClanID(42); want true")
	}
	c.SetOwnerClanID(0)
	if c.HasOwner() {
		t.Error("HasOwner() = true after reset; want false")
	}
}

func TestCastle_SetTreasury(t *testing.T) {
	t.Parallel()
	c := NewCastle(1, "Gludio", 30)

	c.SetTreasury(5000)
	if c.Treasury() != 5000 {
		t.Errorf("Treasury() = %d; want 5000", c.Treasury())
	}
}

func TestCastle_MaxMercenaries(t *testing.T) {
	t.Parallel()
	c := NewCastle(5, "Aden", 36)
	if c.MaxMercenaries() != 36 {
		t.Errorf("MaxMercenaries() = %d; want 36", c.MaxMercenaries())
	}
}

func TestCastle_TimeRegistrationOver(t *testing.T) {
	t.Parallel()
	c := NewCastle(1, "Gludio", 30)

	if c.IsTimeRegistrationOver() {
		t.Error("new castle: IsTimeRegistrationOver() = true; want false")
	}
	c.SetTimeRegistrationOver(true)
	if !c.IsTimeRegistrationOver() {
		t.Error("after set true: IsTimeRegistrationOver() = false; want true")
	}
}

func TestCastle_SiegePointer(t *testing.T) {
	t.Parallel()
	c := NewCastle(1, "Gludio", 30)

	if c.Siege() != nil {
		t.Error("Siege() != nil on new castle")
	}
	s := NewSiege(c)
	c.SetSiege(s)
	if c.Siege() != s {
		t.Error("Siege() != expected after SetSiege")
	}
	c.SetSiege(nil)
	if c.Siege() != nil {
		t.Error("Siege() != nil after SetSiege(nil)")
	}
}

func TestCastle_ConcurrentAccess(t *testing.T) {
	t.Parallel()
	c := NewCastle(1, "Gludio", 30)

	var wg sync.WaitGroup
	for range 100 {
		wg.Add(3)
		go func() {
			defer wg.Done()
			c.SetTaxRate(10)
			c.TaxRate()
		}()
		go func() {
			defer wg.Done()
			c.AddToTreasury(100)
			c.Treasury()
		}()
		go func() {
			defer wg.Done()
			c.SetOwnerClanID(42)
			c.OwnerClanID()
			c.HasOwner()
		}()
	}
	wg.Wait()
}

func TestSiegeClan_Types(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		clanType   SiegeClanType
		isAttacker bool
		isDefender bool
		isPending  bool
	}{
		{"owner", ClanTypeOwner, false, true, false},
		{"defender", ClanTypeDefender, false, true, false},
		{"attacker", ClanTypeAttacker, true, false, false},
		{"pending", ClanTypeDefenderNotApproved, false, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sc := NewSiegeClan(1, "TestClan", tt.clanType)
			if sc.IsAttacker() != tt.isAttacker {
				t.Errorf("IsAttacker() = %v; want %v", sc.IsAttacker(), tt.isAttacker)
			}
			if sc.IsDefender() != tt.isDefender {
				t.Errorf("IsDefender() = %v; want %v", sc.IsDefender(), tt.isDefender)
			}
			if sc.IsPending() != tt.isPending {
				t.Errorf("IsPending() = %v; want %v", sc.IsPending(), tt.isPending)
			}
		})
	}
}

func TestSiege_RegisterAttacker(t *testing.T) {
	t.Parallel()
	c := NewCastle(1, "Gludio", 30)
	s := NewSiege(c)

	sc := NewSiegeClan(100, "AttackerClan", ClanTypeAttacker)
	s.RegisterAttacker(sc)

	if !s.IsAttacker(100) {
		t.Error("IsAttacker(100) = false; want true")
	}
	if s.AttackerCount() != 1 {
		t.Errorf("AttackerCount() = %d; want 1", s.AttackerCount())
	}
	if s.IsClanRegistered(100) != true {
		t.Error("IsClanRegistered(100) = false; want true")
	}
}

func TestSiege_RegisterDefender(t *testing.T) {
	t.Parallel()
	c := NewCastle(1, "Gludio", 30)
	s := NewSiege(c)

	sc := NewSiegeClan(200, "DefenderClan", ClanTypeDefender)
	s.RegisterDefender(sc)

	if !s.IsDefender(200) {
		t.Error("IsDefender(200) = false; want true")
	}
	if s.DefenderCount() != 1 {
		t.Errorf("DefenderCount() = %d; want 1", s.DefenderCount())
	}
}

func TestSiege_PendingDefender(t *testing.T) {
	t.Parallel()
	c := NewCastle(1, "Gludio", 30)
	s := NewSiege(c)

	sc := NewSiegeClan(300, "PendingClan", ClanTypeDefenderNotApproved)
	s.RegisterPendingDefender(sc)

	if s.IsClanRegistered(300) != true {
		t.Error("IsClanRegistered(300) = false; want true")
	}
	if len(s.PendingClans()) != 1 {
		t.Errorf("PendingClans() count = %d; want 1", len(s.PendingClans()))
	}

	// Approve
	ok := s.ApprovePendingDefender(300)
	if !ok {
		t.Error("ApprovePendingDefender(300) = false; want true")
	}
	if !s.IsDefender(300) {
		t.Error("IsDefender(300) = false after approval; want true")
	}
	if len(s.PendingClans()) != 0 {
		t.Errorf("PendingClans() count = %d after approval; want 0", len(s.PendingClans()))
	}
}

func TestSiege_RemoveClan(t *testing.T) {
	t.Parallel()
	c := NewCastle(1, "Gludio", 30)
	s := NewSiege(c)

	s.RegisterAttacker(NewSiegeClan(100, "Clan1", ClanTypeAttacker))
	s.RegisterDefender(NewSiegeClan(200, "Clan2", ClanTypeDefender))

	s.RemoveClan(100)
	if s.IsAttacker(100) {
		t.Error("IsAttacker(100) = true after remove; want false")
	}
	if s.AttackerCount() != 0 {
		t.Errorf("AttackerCount() = %d; want 0", s.AttackerCount())
	}
}

func TestSiege_StartAndEnd(t *testing.T) {
	t.Parallel()
	c := NewCastle(1, "Gludio", 30)
	c.SetOwnerClanID(500) // Castle has owner
	s := NewSiege(c)

	s.RegisterAttacker(NewSiegeClan(100, "Attackers", ClanTypeAttacker))
	s.RegisterPendingDefender(NewSiegeClan(200, "PendingDef", ClanTypeDefenderNotApproved))

	s.StartSiege()

	if s.State() != StateRunning {
		t.Errorf("State() = %d; want %d (running)", s.State(), StateRunning)
	}
	if !s.IsInProgress() {
		t.Error("IsInProgress() = false; want true")
	}

	// Owner auto-added as defender.
	if !s.IsDefender(500) {
		t.Error("Owner clan 500 not auto-added as defender")
	}

	// Pending defender moved to defenders.
	if !s.IsDefender(200) {
		t.Error("Pending clan 200 not moved to defenders")
	}
	if len(s.PendingClans()) != 0 {
		t.Errorf("PendingClans() = %d; want 0 after start", len(s.PendingClans()))
	}

	s.EndSiege()

	if s.State() != StateInactive {
		t.Errorf("State() = %d; want %d (inactive)", s.State(), StateInactive)
	}
}

func TestSiege_MidVictory(t *testing.T) {
	t.Parallel()
	c := NewCastle(1, "Gludio", 30)
	c.SetOwnerClanID(500)
	s := NewSiege(c)

	s.RegisterAttacker(NewSiegeClan(100, "Attackers", ClanTypeAttacker))
	s.RegisterAttacker(NewSiegeClan(101, "Attackers2", ClanTypeAttacker))
	s.StartSiege()

	// Clan 100 captures the castle.
	s.MidVictory(100)

	if c.OwnerClanID() != 100 {
		t.Errorf("OwnerClanID() = %d; want 100", c.OwnerClanID())
	}

	// Clan 100 is now a defender (owner).
	if !s.IsDefender(100) {
		t.Error("new owner (100) not in defenders")
	}
	if s.IsAttacker(100) {
		t.Error("new owner (100) still in attackers")
	}
}

func TestSiege_ControlTowers(t *testing.T) {
	t.Parallel()
	c := NewCastle(1, "Gludio", 30)
	s := NewSiege(c)

	towers := []*TowerSpawn{
		{X: 100, Y: 200, Z: 300, NpcID: 35001},
		{X: 400, Y: 500, Z: 600, NpcID: 35002},
		{X: 700, Y: 800, Z: 900, NpcID: 35003},
	}
	s.SetControlTowers(towers)

	if s.ControlTowersAlive() != 3 {
		t.Errorf("ControlTowersAlive() = %d; want 3", s.ControlTowersAlive())
	}

	allDestroyed := s.ControlTowerDestroyed()
	if allDestroyed {
		t.Error("ControlTowerDestroyed() = true; want false (2 remaining)")
	}
	if s.ControlTowersAlive() != 2 {
		t.Errorf("ControlTowersAlive() = %d; want 2", s.ControlTowersAlive())
	}

	s.ControlTowerDestroyed()
	allDestroyed = s.ControlTowerDestroyed()
	if !allDestroyed {
		t.Error("ControlTowerDestroyed() = false; want true (all destroyed)")
	}
}

func TestSiege_ClearClans(t *testing.T) {
	t.Parallel()
	c := NewCastle(1, "Gludio", 30)
	s := NewSiege(c)

	s.RegisterAttacker(NewSiegeClan(1, "A", ClanTypeAttacker))
	s.RegisterDefender(NewSiegeClan(2, "D", ClanTypeDefender))
	s.RegisterPendingDefender(NewSiegeClan(3, "P", ClanTypeDefenderNotApproved))

	s.ClearClans()

	if s.AttackerCount() != 0 {
		t.Errorf("AttackerCount() = %d; want 0", s.AttackerCount())
	}
	if s.DefenderCount() != 0 {
		t.Errorf("DefenderCount() = %d; want 0", s.DefenderCount())
	}
	if len(s.PendingClans()) != 0 {
		t.Errorf("PendingClans() = %d; want 0", len(s.PendingClans()))
	}
}

func TestSiege_Timer(t *testing.T) {
	t.Parallel()
	c := NewCastle(1, "Gludio", 30)
	s := NewSiege(c)

	timer := time.NewTimer(1 * time.Hour)
	defer timer.Stop()

	s.SetTimer(timer)
	s.StopTimer()
	// Should not panic on double stop.
	s.StopTimer()
}

func TestSiege_StateTransitions(t *testing.T) {
	t.Parallel()
	c := NewCastle(1, "Gludio", 30)
	s := NewSiege(c)

	if s.State() != StateInactive {
		t.Errorf("initial State() = %d; want %d (inactive)", s.State(), StateInactive)
	}

	s.SetState(StateRegistration)
	if s.State() != StateRegistration {
		t.Errorf("State() = %d; want %d (registration)", s.State(), StateRegistration)
	}
	if !s.IsRegistration() {
		t.Error("IsRegistration() = false; want true in StateRegistration")
	}

	s.SetState(StateCountdown)
	if s.State() != StateCountdown {
		t.Errorf("State() = %d; want %d (countdown)", s.State(), StateCountdown)
	}
	if s.IsRegistration() {
		t.Error("IsRegistration() = true in StateCountdown; want false")
	}
}

func TestSiege_IsRegistration_Inactive(t *testing.T) {
	t.Parallel()
	c := NewCastle(1, "Gludio", 30)
	s := NewSiege(c)
	// StateInactive также разрешает регистрацию.
	if !s.IsRegistration() {
		t.Error("IsRegistration() = false in StateInactive; want true")
	}
}

func TestSiege_AttackerClansSnapshot(t *testing.T) {
	t.Parallel()
	c := NewCastle(1, "Gludio", 30)
	s := NewSiege(c)

	s.RegisterAttacker(NewSiegeClan(10, "A1", ClanTypeAttacker))
	s.RegisterAttacker(NewSiegeClan(20, "A2", ClanTypeAttacker))

	clans := s.AttackerClans()
	if len(clans) != 2 {
		t.Errorf("AttackerClans() len = %d; want 2", len(clans))
	}

	// Получение конкретного клана.
	ac := s.AttackerClan(10)
	if ac == nil || ac.ClanID != 10 {
		t.Errorf("AttackerClan(10) = %v; want clan 10", ac)
	}
	if s.AttackerClan(999) != nil {
		t.Error("AttackerClan(999) != nil; want nil")
	}
}

func TestSiege_DefenderClansSnapshot(t *testing.T) {
	t.Parallel()
	c := NewCastle(1, "Gludio", 30)
	s := NewSiege(c)

	s.RegisterDefender(NewSiegeClan(10, "D1", ClanTypeDefender))

	clans := s.DefenderClans()
	if len(clans) != 1 {
		t.Errorf("DefenderClans() len = %d; want 1", len(clans))
	}

	dc := s.DefenderClan(10)
	if dc == nil || dc.ClanID != 10 {
		t.Errorf("DefenderClan(10) = %v; want clan 10", dc)
	}
	if s.DefenderClan(999) != nil {
		t.Error("DefenderClan(999) != nil; want nil")
	}
}

func TestSiege_FlameTowers(t *testing.T) {
	t.Parallel()
	c := NewCastle(1, "Gludio", 30)
	s := NewSiege(c)

	flames := []*TowerSpawn{
		{X: 10, Y: 20, Z: 30, NpcID: 13004, ZoneIDs: []int32{1, 2}},
	}
	s.SetFlameTowers(flames)

	got := s.FlameTowers()
	if len(got) != 1 {
		t.Errorf("FlameTowers() len = %d; want 1", len(got))
	}
	if got[0].NpcID != 13004 {
		t.Errorf("FlameTowers()[0].NpcID = %d; want 13004", got[0].NpcID)
	}
}

func TestSiege_StartEndTime(t *testing.T) {
	t.Parallel()
	c := NewCastle(1, "Gludio", 30)
	s := NewSiege(c)

	before := time.Now()
	s.StartSiege()
	after := time.Now()

	start := s.StartTime()
	if start.Before(before) || start.After(after) {
		t.Errorf("StartTime() = %v; want between %v and %v", start, before, after)
	}

	s.EndSiege()
	end := s.EndTime()
	if end.Before(start) {
		t.Errorf("EndTime() %v before StartTime() %v", end, start)
	}
}

func TestSiege_ApprovePendingDefender_NotInList(t *testing.T) {
	t.Parallel()
	c := NewCastle(1, "Gludio", 30)
	s := NewSiege(c)

	if s.ApprovePendingDefender(999) {
		t.Error("ApprovePendingDefender(999) = true; want false for non-existent")
	}
}

func TestSiege_ControlTowerDestroyed_BelowZero(t *testing.T) {
	t.Parallel()
	c := NewCastle(1, "Gludio", 30)
	s := NewSiege(c)

	// Нет башен — destroy не уходит ниже нуля.
	s.ControlTowerDestroyed()
	if s.ControlTowersAlive() != 0 {
		t.Errorf("ControlTowersAlive() = %d; want 0 (clamped)", s.ControlTowersAlive())
	}
}

func TestSiege_Castle(t *testing.T) {
	t.Parallel()
	c := NewCastle(1, "Gludio", 30)
	s := NewSiege(c)

	if s.Castle() != c {
		t.Error("Castle() does not return expected castle")
	}
}

func TestSiege_ControlTowersSnapshot(t *testing.T) {
	t.Parallel()
	c := NewCastle(1, "Gludio", 30)
	s := NewSiege(c)

	towers := []*TowerSpawn{
		{X: 100, Y: 200, Z: 300, NpcID: 13002},
	}
	s.SetControlTowers(towers)

	got := s.ControlTowers()
	if len(got) != 1 {
		t.Errorf("ControlTowers() len = %d; want 1", len(got))
	}
}

func TestSiege_ConcurrentRegistration(t *testing.T) {
	t.Parallel()
	c := NewCastle(1, "Gludio", 30)
	s := NewSiege(c)

	var wg sync.WaitGroup
	for i := range 50 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.RegisterAttacker(NewSiegeClan(int32(i), "A", ClanTypeAttacker))
		}()
	}
	wg.Wait()

	if s.AttackerCount() != 50 {
		t.Errorf("AttackerCount() = %d; want 50", s.AttackerCount())
	}
}

func TestSiege_SetTimerReplaces(t *testing.T) {
	t.Parallel()
	c := NewCastle(1, "Gludio", 30)
	s := NewSiege(c)

	t1 := time.NewTimer(time.Hour)
	t2 := time.NewTimer(time.Hour)
	defer t1.Stop()
	defer t2.Stop()

	s.SetTimer(t1)
	s.SetTimer(t2) // Должен остановить t1 и заменить.
	s.StopTimer()
}

func TestSiege_StartSiege_NoOwner(t *testing.T) {
	t.Parallel()
	c := NewCastle(1, "Gludio", 30) // Нет владельца.
	s := NewSiege(c)

	s.RegisterAttacker(NewSiegeClan(100, "Att", ClanTypeAttacker))
	s.StartSiege()

	if s.State() != StateRunning {
		t.Errorf("State() = %d; want %d", s.State(), StateRunning)
	}
	// Без владельца не должно быть автоматического защитника.
	if s.DefenderCount() != 0 {
		t.Errorf("DefenderCount() = %d; want 0 (no owner)", s.DefenderCount())
	}
}
