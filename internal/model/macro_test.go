package model

import "testing"

func TestPlayerAutoSoulShot(t *testing.T) {
	t.Parallel()

	p, err := NewPlayer(1, 100, 1, "TestSS", 1, 0, 0)
	if err != nil {
		t.Fatal(err)
	}

	// Initially empty
	if shots := p.AutoSoulShots(); len(shots) != 0 {
		t.Fatalf("AutoSoulShots() len = %d; want 0", len(shots))
	}
	if p.HasAutoSoulShot(3947) {
		t.Error("HasAutoSoulShot(3947) = true; want false")
	}

	// Add
	p.AddAutoSoulShot(3947) // Blessed SoulShot D
	p.AddAutoSoulShot(3948) // Blessed SoulShot C

	if !p.HasAutoSoulShot(3947) {
		t.Error("HasAutoSoulShot(3947) = false; want true")
	}
	if !p.HasAutoSoulShot(3948) {
		t.Error("HasAutoSoulShot(3948) = false; want true")
	}
	if shots := p.AutoSoulShots(); len(shots) != 2 {
		t.Errorf("AutoSoulShots() len = %d; want 2", len(shots))
	}

	// Remove
	p.RemoveAutoSoulShot(3947)
	if p.HasAutoSoulShot(3947) {
		t.Error("after remove: HasAutoSoulShot(3947) = true; want false")
	}
	if shots := p.AutoSoulShots(); len(shots) != 1 {
		t.Errorf("AutoSoulShots() len = %d; want 1", len(shots))
	}

	// Remove non-existent — no panic
	p.RemoveAutoSoulShot(9999)
}

func TestPlayerAutoSoulShotDuplicate(t *testing.T) {
	t.Parallel()

	p, err := NewPlayer(1, 100, 1, "TestDup", 1, 0, 0)
	if err != nil {
		t.Fatal(err)
	}

	p.AddAutoSoulShot(3947)
	p.AddAutoSoulShot(3947)
	if shots := p.AutoSoulShots(); len(shots) != 1 {
		t.Errorf("AutoSoulShots() len = %d; want 1 (no duplicates)", len(shots))
	}
}

func TestPlayerSetAutoSoulShots(t *testing.T) {
	t.Parallel()

	p, err := NewPlayer(1, 100, 1, "TestSet", 1, 0, 0)
	if err != nil {
		t.Fatal(err)
	}

	p.AddAutoSoulShot(1000)
	p.SetAutoSoulShots([]int32{2000, 3000})

	if p.HasAutoSoulShot(1000) {
		t.Error("HasAutoSoulShot(1000) = true; want false after SetAutoSoulShots")
	}
	if !p.HasAutoSoulShot(2000) || !p.HasAutoSoulShot(3000) {
		t.Error("missing items from SetAutoSoulShots")
	}
}

func TestPlayerRegisterMacro(t *testing.T) {
	t.Parallel()

	p, err := NewPlayer(1, 100, 1, "TestMac", 1, 0, 0)
	if err != nil {
		t.Fatal(err)
	}

	m := &Macro{
		ID:   1,
		Name: "Attack",
		Desc: "Auto attack macro",
		Icon: 2,
		Commands: []MacroCmd{
			{Entry: 0, Type: MacroCmdSkill, D1: 101, Command: "/use 101"},
		},
	}

	if err := p.RegisterMacro(m); err != nil {
		t.Fatalf("RegisterMacro() error = %v", err)
	}

	if p.MacroCount() != 1 {
		t.Errorf("MacroCount() = %d; want 1", p.MacroCount())
	}
	if p.MacroRevision() != 1 {
		t.Errorf("MacroRevision() = %d; want 1", p.MacroRevision())
	}

	got := p.GetMacro(1)
	if got == nil {
		t.Fatal("GetMacro(1) = nil; want macro")
	}
	if got.Name != "Attack" {
		t.Errorf("Macro.Name = %q; want %q", got.Name, "Attack")
	}
}

func TestPlayerDeleteMacro(t *testing.T) {
	t.Parallel()

	p, err := NewPlayer(1, 100, 1, "TestDel", 1, 0, 0)
	if err != nil {
		t.Fatal(err)
	}

	m := &Macro{ID: 5, Name: "Heal"}
	if err := p.RegisterMacro(m); err != nil {
		t.Fatal(err)
	}

	p.DeleteMacro(5)
	if p.MacroCount() != 0 {
		t.Errorf("MacroCount() = %d; want 0", p.MacroCount())
	}
	if got := p.GetMacro(5); got != nil {
		t.Error("GetMacro(5) should be nil after delete")
	}
	// Both register and delete increment revision
	if p.MacroRevision() != 2 {
		t.Errorf("MacroRevision() = %d; want 2", p.MacroRevision())
	}
}

func TestPlayerDeleteMacroNonexistent(t *testing.T) {
	t.Parallel()

	p, err := NewPlayer(1, 100, 1, "TestDN", 1, 0, 0)
	if err != nil {
		t.Fatal(err)
	}

	// Should not panic
	p.DeleteMacro(999)
}

func TestPlayerMacroUpdateExisting(t *testing.T) {
	t.Parallel()

	p, err := NewPlayer(1, 100, 1, "TestUpd", 1, 0, 0)
	if err != nil {
		t.Fatal(err)
	}

	m := &Macro{ID: 1, Name: "v1"}
	if err := p.RegisterMacro(m); err != nil {
		t.Fatal(err)
	}

	m2 := &Macro{ID: 1, Name: "v2"}
	if err := p.RegisterMacro(m2); err != nil {
		t.Fatal(err)
	}

	if p.MacroCount() != 1 {
		t.Errorf("MacroCount() = %d; want 1 (update, not add)", p.MacroCount())
	}
	if got := p.GetMacro(1); got.Name != "v2" {
		t.Errorf("Macro.Name = %q; want %q", got.Name, "v2")
	}
}

func TestPlayerMacroLimit(t *testing.T) {
	t.Parallel()

	p, err := NewPlayer(1, 100, 1, "TestLim", 1, 0, 0)
	if err != nil {
		t.Fatal(err)
	}

	for i := range MaxMacros {
		m := &Macro{ID: int32(i + 1), Name: "macro"}
		if err := p.RegisterMacro(m); err != nil {
			t.Fatalf("RegisterMacro(%d) unexpected error: %v", i+1, err)
		}
	}

	if p.MacroCount() != MaxMacros {
		t.Fatalf("MacroCount() = %d; want %d", p.MacroCount(), MaxMacros)
	}

	// 25th macro should fail
	m25 := &Macro{ID: int32(MaxMacros + 1), Name: "excess"}
	if err := p.RegisterMacro(m25); err == nil {
		t.Error("RegisterMacro() expected error for exceeding limit")
	}
}

func TestPlayerGetMacrosEmpty(t *testing.T) {
	t.Parallel()

	p, err := NewPlayer(1, 100, 1, "TestEmp", 1, 0, 0)
	if err != nil {
		t.Fatal(err)
	}

	macros := p.GetMacros()
	if len(macros) != 0 {
		t.Errorf("GetMacros() len = %d; want 0", len(macros))
	}
}

func TestPlayerSetMacros(t *testing.T) {
	t.Parallel()

	p, err := NewPlayer(1, 100, 1, "TestSet", 1, 0, 0)
	if err != nil {
		t.Fatal(err)
	}

	// First register one manually
	if err := p.RegisterMacro(&Macro{ID: 99, Name: "old"}); err != nil {
		t.Fatal(err)
	}

	// Then bulk-set (simulating login restore)
	p.SetMacros([]*Macro{
		{ID: 1, Name: "m1"},
		{ID: 2, Name: "m2"},
	})

	if p.MacroCount() != 2 {
		t.Errorf("MacroCount() = %d; want 2", p.MacroCount())
	}
	if got := p.GetMacro(99); got != nil {
		t.Error("GetMacro(99) should be nil after SetMacros (replaced)")
	}
	if got := p.GetMacro(1); got == nil || got.Name != "m1" {
		t.Error("GetMacro(1) expected m1")
	}
}
