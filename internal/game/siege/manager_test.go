package siege

import (
	"testing"
)

func TestNewManager(t *testing.T) {
	t.Parallel()
	m := NewManager(DefaultManagerConfig())

	if len(m.Castles()) != 9 {
		t.Errorf("Castles() count = %d; want 9", len(m.Castles()))
	}

	for _, id := range []int32{1, 2, 3, 4, 5, 6, 7, 8, 9} {
		c := m.Castle(id)
		if c == nil {
			t.Errorf("Castle(%d) = nil; want non-nil", id)
		}
	}
}

func TestManager_CastleByName(t *testing.T) {
	t.Parallel()
	m := NewManager(DefaultManagerConfig())

	tests := []struct {
		name   string
		expect int32
	}{
		{"Gludio", 1},
		{"gludio", 1},
		{"ADEN", 5},
		{"Schuttgart", 9},
	}

	for _, tt := range tests {
		c := m.CastleByName(tt.name)
		if c == nil {
			t.Errorf("CastleByName(%q) = nil; want ID %d", tt.name, tt.expect)
			continue
		}
		if c.ID() != tt.expect {
			t.Errorf("CastleByName(%q).ID() = %d; want %d", tt.name, c.ID(), tt.expect)
		}
	}

	if c := m.CastleByName("NonExistent"); c != nil {
		t.Errorf("CastleByName(\"NonExistent\") = %v; want nil", c.Name())
	}
}

func TestManager_RegisterAttacker(t *testing.T) {
	t.Parallel()
	m := NewManager(DefaultManagerConfig())

	// Успешная регистрация атакующего.
	if err := m.RegisterAttacker(1, 100, "TestClan", 5); err != nil {
		t.Fatalf("RegisterAttacker() = %v; want nil", err)
	}

	// Нельзя зарегистрировать дважды.
	if err := m.RegisterAttacker(1, 100, "TestClan", 5); err != ErrClanAlreadyRegistered {
		t.Errorf("RegisterAttacker() = %v; want %v", err, ErrClanAlreadyRegistered)
	}

	// Проверяем что клан виден в регистрации.
	castleID, registered := m.IsClanRegistered(100)
	if !registered || castleID != 1 {
		t.Errorf("IsClanRegistered(100) = (%d, %v); want (1, true)", castleID, registered)
	}
}

func TestManager_RegisterAttacker_InvalidCastle(t *testing.T) {
	t.Parallel()
	m := NewManager(DefaultManagerConfig())

	if err := m.RegisterAttacker(99, 100, "TestClan", 5); err != ErrCastleNotFound {
		t.Errorf("RegisterAttacker() = %v; want %v", err, ErrCastleNotFound)
	}
}

func TestManager_RegisterAttacker_LowLevel(t *testing.T) {
	t.Parallel()
	m := NewManager(DefaultManagerConfig())

	if err := m.RegisterAttacker(1, 100, "TestClan", 2); err != ErrClanLevelTooLow {
		t.Errorf("RegisterAttacker() = %v; want %v", err, ErrClanLevelTooLow)
	}
}

func TestManager_RegisterAttacker_OwnerCannotAttack(t *testing.T) {
	t.Parallel()
	m := NewManager(DefaultManagerConfig())

	castle := m.Castle(1)
	castle.SetOwnerClanID(100)

	if err := m.RegisterAttacker(1, 100, "OwnerClan", 5); err != ErrOwnerCannotAttack {
		t.Errorf("RegisterAttacker() = %v; want %v", err, ErrOwnerCannotAttack)
	}
}

func TestManager_RegisterDefender(t *testing.T) {
	t.Parallel()
	m := NewManager(DefaultManagerConfig())

	// Замок без владельца — защитник регистрируется напрямую.
	if err := m.RegisterDefender(1, 200, "DefClan", 5); err != nil {
		t.Fatalf("RegisterDefender() = %v; want nil", err)
	}

	siege := m.Castle(1).Siege()
	if !siege.IsDefender(200) {
		t.Error("clan 200 should be defender")
	}
}

func TestManager_RegisterDefender_Pending(t *testing.T) {
	t.Parallel()
	m := NewManager(DefaultManagerConfig())

	// Замок с владельцем — защитник попадает в pending.
	castle := m.Castle(1)
	castle.SetOwnerClanID(500)

	if err := m.RegisterDefender(1, 200, "DefClan", 5); err != nil {
		t.Fatalf("RegisterDefender() = %v; want nil", err)
	}

	siege := castle.Siege()
	if siege.IsDefender(200) {
		t.Error("clan 200 should be pending, not defender")
	}
	if len(siege.PendingClans()) != 1 {
		t.Errorf("PendingClans() count = %d; want 1", len(siege.PendingClans()))
	}
}

func TestManager_Unregister(t *testing.T) {
	t.Parallel()
	m := NewManager(DefaultManagerConfig())

	if err := m.RegisterAttacker(1, 100, "TestClan", 5); err != nil {
		t.Fatal(err)
	}

	if err := m.Unregister(1, 100); err != nil {
		t.Fatalf("Unregister() = %v; want nil", err)
	}

	_, registered := m.IsClanRegistered(100)
	if registered {
		t.Error("IsClanRegistered(100) = true after unregister; want false")
	}
}

func TestManager_ApproveDefender(t *testing.T) {
	t.Parallel()
	m := NewManager(DefaultManagerConfig())

	castle := m.Castle(1)
	castle.SetOwnerClanID(500)

	if err := m.RegisterDefender(1, 200, "DefClan", 5); err != nil {
		t.Fatal(err)
	}

	if err := m.ApproveDefender(1, 200); err != nil {
		t.Fatalf("ApproveDefender() = %v; want nil", err)
	}

	siege := castle.Siege()
	if !siege.IsDefender(200) {
		t.Error("clan 200 should be approved defender")
	}
}

func TestManager_IsClanRegistered_None(t *testing.T) {
	t.Parallel()
	m := NewManager(DefaultManagerConfig())

	_, registered := m.IsClanRegistered(999)
	if registered {
		t.Error("IsClanRegistered(999) = true; want false")
	}
}

func TestManager_RegisterDuringSiege(t *testing.T) {
	t.Parallel()
	m := NewManager(DefaultManagerConfig())

	// Начинаем осаду.
	castle := m.Castle(1)
	siege := castle.Siege()
	siege.RegisterAttacker(NewSiegeClan(100, "A", ClanTypeAttacker))
	siege.StartSiege()

	// Нельзя зарегистрироваться когда осада идёт.
	if err := m.RegisterAttacker(1, 200, "NewClan", 5); err != ErrSiegeInProgress {
		t.Errorf("RegisterAttacker during siege = %v; want %v", err, ErrSiegeInProgress)
	}
}

func TestManager_Config(t *testing.T) {
	t.Parallel()
	cfg := DefaultManagerConfig()
	m := NewManager(cfg)

	got := m.Config()
	if got.AttackerMaxClans != DefaultAttackerMax {
		t.Errorf("Config().AttackerMaxClans = %d; want %d", got.AttackerMaxClans, DefaultAttackerMax)
	}
	if got.SiegeClanMinLevel != DefaultClanMinLevel {
		t.Errorf("Config().SiegeClanMinLevel = %d; want %d", got.SiegeClanMinLevel, DefaultClanMinLevel)
	}
}

func TestEqualFoldASCII(t *testing.T) {
	t.Parallel()

	tests := []struct {
		a, b string
		want bool
	}{
		{"Gludio", "gludio", true},
		{"ADEN", "Aden", true},
		{"abc", "ABC", true},
		{"abc", "abd", false},
		{"abc", "ab", false},
		{"", "", true},
	}

	for _, tt := range tests {
		if got := equalFoldASCII(tt.a, tt.b); got != tt.want {
			t.Errorf("equalFoldASCII(%q, %q) = %v; want %v", tt.a, tt.b, got, tt.want)
		}
	}
}
