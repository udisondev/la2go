package clan

import (
	"sync"
	"testing"
)

func TestTable_Create(t *testing.T) {
	tbl := NewTable()

	c, err := tbl.Create("TestClan", 100)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if c.ID() != 1 {
		t.Errorf("Clan ID = %d, want 1", c.ID())
	}
	if c.Name() != "TestClan" {
		t.Errorf("Clan Name = %q, want %q", c.Name(), "TestClan")
	}
	if c.LeaderID() != 100 {
		t.Errorf("Clan LeaderID = %d, want 100", c.LeaderID())
	}
	if tbl.Count() != 1 {
		t.Errorf("Count = %d, want 1", tbl.Count())
	}
}

func TestTable_Create_NameTaken(t *testing.T) {
	tbl := NewTable()
	tbl.Create("Alpha", 100) //nolint:errcheck

	_, err := tbl.Create("alpha", 200) // case-insensitive
	if err != ErrClanNameTaken {
		t.Errorf("Create duplicate = %v, want ErrClanNameTaken", err)
	}
}

func TestTable_Create_InvalidName(t *testing.T) {
	tbl := NewTable()

	tests := []struct {
		name string
	}{
		{"A"},                           // too short
		{"ThisIsAVeryLongClanNameTest"}, // too long (>16)
		{"Bad Name"},                    // space
		{"Bad!Name"},                    // special char
	}
	for _, tt := range tests {
		_, err := tbl.Create(tt.name, 100)
		if err == nil {
			t.Errorf("Create(%q) should return error", tt.name)
		}
	}
}

func TestTable_Create_ValidNames(t *testing.T) {
	tbl := NewTable()

	validNames := []string{"AB", "TestClan", "Clan123", "ABCDEFGHIJKLMNOP"}
	for _, name := range validNames {
		_, err := tbl.Create(name, 100)
		if err != nil {
			t.Errorf("Create(%q) = %v, want nil", name, err)
		}
	}
}

func TestTable_Disband(t *testing.T) {
	tbl := NewTable()
	c, _ := tbl.Create("TestClan", 100)

	if err := tbl.Disband(c.ID()); err != nil {
		t.Fatalf("Disband: %v", err)
	}
	if tbl.Count() != 0 {
		t.Errorf("Count = %d after disband, want 0", tbl.Count())
	}

	// Disband non-existent.
	if err := tbl.Disband(999); err != ErrClanNotFound {
		t.Errorf("Disband(999) = %v, want ErrClanNotFound", err)
	}
}

func TestTable_ClanByName(t *testing.T) {
	tbl := NewTable()
	tbl.Create("MyClan", 100) //nolint:errcheck

	// Case-insensitive lookup.
	c := tbl.ClanByName("myclan")
	if c == nil {
		t.Fatal("ClanByName(myclan) = nil, want clan")
	}
	if c.Name() != "MyClan" {
		t.Errorf("ClanByName Name = %q, want %q", c.Name(), "MyClan")
	}

	// Not found.
	if tbl.ClanByName("Nope") != nil {
		t.Error("ClanByName(Nope) should be nil")
	}
}

func TestTable_Clan(t *testing.T) {
	tbl := NewTable()
	created, _ := tbl.Create("TestClan", 100)

	got := tbl.Clan(created.ID())
	if got == nil {
		t.Fatal("Clan by ID = nil")
	}
	if got.Name() != "TestClan" {
		t.Errorf("Clan Name = %q, want %q", got.Name(), "TestClan")
	}

	if tbl.Clan(999) != nil {
		t.Error("Clan(999) should be nil")
	}
}

func TestTable_Register(t *testing.T) {
	tbl := NewTable()

	c := New(42, "Preloaded", 100)
	if err := tbl.Register(c); err != nil {
		t.Fatalf("Register: %v", err)
	}
	if tbl.Count() != 1 {
		t.Errorf("Count = %d, want 1", tbl.Count())
	}

	// Duplicate.
	c2 := New(43, "Preloaded", 200)
	if err := tbl.Register(c2); err == nil {
		t.Error("Register duplicate name should return error")
	}
}

func TestTable_ForEach(t *testing.T) {
	tbl := NewTable()
	tbl.Create("AA", 1) //nolint:errcheck
	tbl.Create("BB", 2) //nolint:errcheck
	tbl.Create("CC", 3) //nolint:errcheck

	count := 0
	tbl.ForEach(func(c *Clan) bool {
		count++
		return true
	})
	if count != 3 {
		t.Errorf("ForEach visited %d clans, want 3", count)
	}
}

func TestTable_ForEach_EarlyStop(t *testing.T) {
	tbl := NewTable()
	tbl.Create("AA", 1) //nolint:errcheck
	tbl.Create("BB", 2) //nolint:errcheck
	tbl.Create("CC", 3) //nolint:errcheck

	count := 0
	tbl.ForEach(func(c *Clan) bool {
		count++
		return count < 2
	})
	if count != 2 {
		t.Errorf("ForEach stopped at %d, want 2", count)
	}
}

func TestTable_SetNextID(t *testing.T) {
	tbl := NewTable()
	tbl.SetNextID(100)

	c, _ := tbl.Create("TestClan", 1)
	if c.ID() != 101 {
		t.Errorf("Clan ID after SetNextID(100) = %d, want 101", c.ID())
	}
}

func TestTable_ConcurrentCreateDisband(t *testing.T) {
	tbl := NewTable()

	var wg sync.WaitGroup
	for i := range 20 {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			name := string(rune('A'+idx/2)) + string(rune('a'+idx))
			c, _ := tbl.Create(name, int64(idx))
			if c != nil {
				tbl.Disband(c.ID()) //nolint:errcheck
			}
		}(i)
	}
	wg.Wait()
	// No race = pass.
}

func TestTable_Disband_RemovesNameIndex(t *testing.T) {
	tbl := NewTable()
	c, _ := tbl.Create("TestClan", 100)
	tbl.Disband(c.ID()) //nolint:errcheck

	// Same name should be available again.
	c2, err := tbl.Create("TestClan", 200)
	if err != nil {
		t.Fatalf("Re-create after disband: %v", err)
	}
	if c2 == nil {
		t.Error("Re-create should return clan")
	}
}
