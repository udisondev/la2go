package clan

import (
	"sync"
	"testing"
	"time"
)

// --- Alliance Penalty ---

func TestClan_AllyPenalty(t *testing.T) {
	t.Parallel()

	c := New(1, "TestClan", 100)

	if c.HasAllyPenalty() {
		t.Error("new clan should not have penalty")
	}

	expiry := time.Now().Add(time.Hour).UnixMilli()
	c.SetAllyPenalty(expiry, AllyPenaltyClanLeaved)

	if !c.HasAllyPenalty() {
		t.Error("clan should have penalty after SetAllyPenalty")
	}
	if c.AllyPenaltyType() != AllyPenaltyClanLeaved {
		t.Errorf("AllyPenaltyType() = %d; want %d", c.AllyPenaltyType(), AllyPenaltyClanLeaved)
	}
	if c.AllyPenaltyExpiryTime() != expiry {
		t.Errorf("AllyPenaltyExpiryTime() = %d; want %d", c.AllyPenaltyExpiryTime(), expiry)
	}
}

func TestClan_AllyPenalty_AllTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		penaltyType int32
	}{
		{"none", AllyPenaltyNone},
		{"clan_leaved", AllyPenaltyClanLeaved},
		{"clan_dismissed", AllyPenaltyClanDismissed},
		{"dismiss_clan", AllyPenaltyDismissClan},
		{"dissolve_ally", AllyPenaltyDissolveAlly},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := New(1, "TestClan", 100)

			c.SetAllyPenalty(999, tt.penaltyType)
			if c.AllyPenaltyType() != tt.penaltyType {
				t.Errorf("AllyPenaltyType() = %d; want %d", c.AllyPenaltyType(), tt.penaltyType)
			}
		})
	}
}

func TestClan_AllyPenalty_Clear(t *testing.T) {
	t.Parallel()

	c := New(1, "TestClan", 100)
	c.SetAllyPenalty(time.Now().Add(time.Hour).UnixMilli(), AllyPenaltyClanLeaved)

	if !c.HasAllyPenalty() {
		t.Fatal("expected penalty to be active")
	}

	// Сброс штрафа
	c.SetAllyPenalty(0, AllyPenaltyNone)

	if c.HasAllyPenalty() {
		t.Error("clan should not have penalty after clearing")
	}
	if c.AllyPenaltyType() != AllyPenaltyNone {
		t.Errorf("AllyPenaltyType() = %d; want %d", c.AllyPenaltyType(), AllyPenaltyNone)
	}
}

// --- ClearAlly ---

func TestClan_ClearAlly(t *testing.T) {
	t.Parallel()

	c := New(1, "TestClan", 100)
	c.SetAllyID(1)
	c.SetAllyName("TestAlliance")
	c.SetAllyCrestID(42)

	c.ClearAlly()

	if c.AllyID() != 0 {
		t.Errorf("AllyID() = %d; want 0", c.AllyID())
	}
	if c.AllyName() != "" {
		t.Errorf("AllyName() = %q; want empty", c.AllyName())
	}
	if c.AllyCrestID() != 0 {
		t.Errorf("AllyCrestID() = %d; want 0", c.AllyCrestID())
	}
}

func TestClan_ClearAlly_AlreadyClear(t *testing.T) {
	t.Parallel()

	c := New(1, "TestClan", 100)

	// ClearAlly на клане без альянса не должно паниковать
	c.ClearAlly()

	if c.AllyID() != 0 {
		t.Errorf("AllyID() = %d; want 0", c.AllyID())
	}
}

// --- IsAllyLeader ---

func TestClan_IsAllyLeader(t *testing.T) {
	t.Parallel()

	c := New(5, "LeaderClan", 100)

	if c.IsAllyLeader() {
		t.Error("clan with allyID=0 should not be ally leader")
	}

	c.SetAllyID(5) // allyID == clanID — лидер
	if !c.IsAllyLeader() {
		t.Error("clan with allyID==clanID should be ally leader")
	}

	c.SetAllyID(99) // allyID != clanID — участник
	if c.IsAllyLeader() {
		t.Error("clan with allyID!=clanID should not be ally leader")
	}
}

// --- AllyID / AllyName / AllyCrestID ---

func TestClan_AllyFields(t *testing.T) {
	t.Parallel()

	c := New(1, "TestClan", 100)

	// Начальные значения
	if c.AllyID() != 0 {
		t.Errorf("initial AllyID() = %d; want 0", c.AllyID())
	}
	if c.AllyName() != "" {
		t.Errorf("initial AllyName() = %q; want empty", c.AllyName())
	}
	if c.AllyCrestID() != 0 {
		t.Errorf("initial AllyCrestID() = %d; want 0", c.AllyCrestID())
	}

	c.SetAllyID(10)
	c.SetAllyName("Alliance")
	c.SetAllyCrestID(777)

	if c.AllyID() != 10 {
		t.Errorf("AllyID() = %d; want 10", c.AllyID())
	}
	if c.AllyName() != "Alliance" {
		t.Errorf("AllyName() = %q; want %q", c.AllyName(), "Alliance")
	}
	if c.AllyCrestID() != 777 {
		t.Errorf("AllyCrestID() = %d; want 777", c.AllyCrestID())
	}
}

// --- Table.AllyExists ---

func TestTable_AllyExists(t *testing.T) {
	t.Parallel()

	tbl := NewTable()

	c1, err := tbl.Create("Clan1", 100)
	if err != nil {
		t.Fatalf("Create Clan1: %v", err)
	}
	c1.SetAllyID(c1.ID())
	c1.SetAllyName("TestAlliance")

	if _, err := tbl.Create("Clan2", 200); err != nil {
		t.Fatalf("Create Clan2: %v", err)
	}

	if !tbl.AllyExists("TestAlliance") {
		t.Error("AllyExists should return true for existing alliance")
	}
	if !tbl.AllyExists("testalliance") {
		t.Error("AllyExists should be case-insensitive")
	}
	if !tbl.AllyExists("TESTALLIANCE") {
		t.Error("AllyExists should be case-insensitive (upper)")
	}
	if tbl.AllyExists("NonExistent") {
		t.Error("AllyExists should return false for non-existing alliance")
	}
}

func TestTable_AllyExists_MemberNotLeader(t *testing.T) {
	t.Parallel()

	tbl := NewTable()

	// Клан-участник альянса (allyID != clanID) не должен считаться лидером
	c1, err := tbl.Create("Clan1", 100)
	if err != nil {
		t.Fatalf("Create Clan1: %v", err)
	}
	c1.SetAllyID(999) // allyID != clanID — просто участник
	c1.SetAllyName("SomeAlly")

	// Альянс не считается существующим, т.к. нет лидера
	if tbl.AllyExists("SomeAlly") {
		t.Error("AllyExists should return false when no clan is the ally leader")
	}
}

// --- Table.ClanAllies ---

func TestTable_ClanAllies(t *testing.T) {
	t.Parallel()

	tbl := NewTable()

	c1, err := tbl.Create("LeaderClan", 100)
	if err != nil {
		t.Fatalf("Create LeaderClan: %v", err)
	}
	c1.SetAllyID(c1.ID())
	c1.SetAllyName("Alliance1")

	c2, err := tbl.Create("MemberClan", 200)
	if err != nil {
		t.Fatalf("Create MemberClan: %v", err)
	}
	c2.SetAllyID(c1.ID())
	c2.SetAllyName("Alliance1")

	if _, err := tbl.Create("OtherClan", 300); err != nil {
		t.Fatalf("Create OtherClan: %v", err)
	}

	allies := tbl.ClanAllies(c1.ID())
	if len(allies) != 2 {
		t.Fatalf("len(ClanAllies) = %d; want 2", len(allies))
	}

	// Проверяем, что оба клана присутствуют
	foundLeader, foundMember := false, false
	for _, a := range allies {
		if a.Name() == "LeaderClan" {
			foundLeader = true
		}
		if a.Name() == "MemberClan" {
			foundMember = true
		}
	}
	if !foundLeader {
		t.Error("ClanAllies should contain LeaderClan")
	}
	if !foundMember {
		t.Error("ClanAllies should contain MemberClan")
	}
}

func TestTable_ClanAllies_Empty(t *testing.T) {
	t.Parallel()

	tbl := NewTable()

	allies := tbl.ClanAllies(999)
	if len(allies) != 0 {
		t.Errorf("len(ClanAllies(999)) = %d; want 0", len(allies))
	}
}

// --- Table.ClanAllyCount ---

func TestTable_ClanAllyCount(t *testing.T) {
	t.Parallel()

	tbl := NewTable()

	c1, err := tbl.Create("Leader", 100)
	if err != nil {
		t.Fatalf("Create Leader: %v", err)
	}
	c1.SetAllyID(c1.ID())

	c2, err := tbl.Create("Member1", 200)
	if err != nil {
		t.Fatalf("Create Member1: %v", err)
	}
	c2.SetAllyID(c1.ID())

	c3, err := tbl.Create("Member2", 300)
	if err != nil {
		t.Fatalf("Create Member2: %v", err)
	}
	c3.SetAllyID(c1.ID())

	count := tbl.ClanAllyCount(c1.ID())
	if count != 3 {
		t.Errorf("ClanAllyCount = %d; want 3", count)
	}

	// Не в альянсе
	count = tbl.ClanAllyCount(999)
	if count != 0 {
		t.Errorf("ClanAllyCount(999) = %d; want 0", count)
	}
}

// --- Alliance Penalty Constants ---

func TestAlliancePenaltyConstants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		got  int32
		want int32
	}{
		{"AllyPenaltyNone", AllyPenaltyNone, 0},
		{"AllyPenaltyClanLeaved", AllyPenaltyClanLeaved, 1},
		{"AllyPenaltyClanDismissed", AllyPenaltyClanDismissed, 2},
		{"AllyPenaltyDismissClan", AllyPenaltyDismissClan, 3},
		{"AllyPenaltyDissolveAlly", AllyPenaltyDissolveAlly, 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.got != tt.want {
				t.Errorf("%s = %d; want %d", tt.name, tt.got, tt.want)
			}
		})
	}
}

// --- Alliance Limits ---

func TestAllianceLimits(t *testing.T) {
	t.Parallel()

	if MaxClansInAlly != 3 {
		t.Errorf("MaxClansInAlly = %d; want 3", MaxClansInAlly)
	}
	if MaxAllyNameLen != 16 {
		t.Errorf("MaxAllyNameLen = %d; want 16", MaxAllyNameLen)
	}
	if MinAllyNameLen != 2 {
		t.Errorf("MinAllyNameLen = %d; want 2", MinAllyNameLen)
	}
}

// --- Alliance Errors ---

func TestAllianceErrors(t *testing.T) {
	t.Parallel()

	// Проверяем, что sentinel errors не nil
	errors := []struct {
		name string
		err  error
	}{
		{"ErrAlreadyInAlly", ErrAlreadyInAlly},
		{"ErrNotInAlly", ErrNotInAlly},
		{"ErrNotAllyLeader", ErrNotAllyLeader},
		{"ErrAllyFull", ErrAllyFull},
		{"ErrAllyNameTaken", ErrAllyNameTaken},
		{"ErrAllyNameInvalid", ErrAllyNameInvalid},
		{"ErrAllyPenalty", ErrAllyPenalty},
	}

	for _, tt := range errors {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.err == nil {
				t.Errorf("%s is nil", tt.name)
			}
			if tt.err.Error() == "" {
				t.Errorf("%s.Error() is empty", tt.name)
			}
		})
	}
}

// --- Concurrent Alliance Operations ---

func TestClan_ConcurrentAllyAccess(t *testing.T) {
	t.Parallel()

	c := New(1, "ConcurrentClan", 100)

	var wg sync.WaitGroup

	// Конкурентная запись/чтение полей альянса
	for range 20 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.SetAllyID(10)
			_ = c.AllyID()
			c.SetAllyName("Concurrent")
			_ = c.AllyName()
			c.SetAllyCrestID(99)
			_ = c.AllyCrestID()
			_ = c.IsAllyLeader()
		}()
	}

	// Конкурентная работа со штрафами
	for range 10 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.SetAllyPenalty(time.Now().UnixMilli(), AllyPenaltyClanLeaved)
			_ = c.HasAllyPenalty()
			_ = c.AllyPenaltyType()
			_ = c.AllyPenaltyExpiryTime()
		}()
	}

	// Конкурентный ClearAlly
	for range 5 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.ClearAlly()
		}()
	}

	wg.Wait()
	// Тест проходит, если нет data race (go test -race)
}

func TestTable_ConcurrentAllyExists(t *testing.T) {
	t.Parallel()

	tbl := NewTable()

	c, err := tbl.Create("LeaderClan", 100)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	c.SetAllyID(c.ID())
	c.SetAllyName("TestAlly")

	var wg sync.WaitGroup

	for range 20 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = tbl.AllyExists("TestAlly")
			_ = tbl.ClanAllies(c.ID())
			_ = tbl.ClanAllyCount(c.ID())
		}()
	}

	wg.Wait()
}
