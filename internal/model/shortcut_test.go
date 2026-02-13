package model

import "testing"

func TestShortcutKey(t *testing.T) {
	tests := []struct {
		name string
		slot int8
		page int8
		want int32
	}{
		{"slot 0, page 0", 0, 0, 0},
		{"slot 11, page 0", 11, 0, 11},
		{"slot 0, page 1", 0, 1, 12},
		{"slot 5, page 3", 5, 3, 41},
		{"slot 11, page 9", 11, 9, 119},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shortcutKey(tt.slot, tt.page)
			if got != tt.want {
				t.Errorf("shortcutKey(%d, %d) = %d; want %d", tt.slot, tt.page, got, tt.want)
			}
		})
	}
}

func TestPlayerRegisterShortcut(t *testing.T) {
	p, err := NewPlayer(1, 100, 1, "TestPlayer", 1, 0, 0)
	if err != nil {
		t.Fatal(err)
	}

	sc := &Shortcut{
		Slot:  3,
		Page:  0,
		Type:  ShortcutTypeSkill,
		ID:    1001,
		Level: 5,
	}

	p.RegisterShortcut(sc)

	shortcuts := p.GetShortcuts()
	if len(shortcuts) != 1 {
		t.Fatalf("GetShortcuts() len = %d; want 1", len(shortcuts))
	}
	if shortcuts[0].ID != 1001 {
		t.Errorf("Shortcut.ID = %d; want 1001", shortcuts[0].ID)
	}
	if shortcuts[0].Level != 5 {
		t.Errorf("Shortcut.Level = %d; want 5", shortcuts[0].Level)
	}
}

func TestPlayerDeleteShortcut(t *testing.T) {
	p, err := NewPlayer(1, 100, 1, "TestPlayer", 1, 0, 0)
	if err != nil {
		t.Fatal(err)
	}

	// Register two shortcuts
	p.RegisterShortcut(&Shortcut{Slot: 0, Page: 0, Type: ShortcutTypeAction, ID: 1})
	p.RegisterShortcut(&Shortcut{Slot: 1, Page: 0, Type: ShortcutTypeAction, ID: 2})

	if len(p.GetShortcuts()) != 2 {
		t.Fatalf("expected 2 shortcuts, got %d", len(p.GetShortcuts()))
	}

	// Delete first shortcut
	p.DeleteShortcut(0, 0)

	shortcuts := p.GetShortcuts()
	if len(shortcuts) != 1 {
		t.Fatalf("expected 1 shortcut after delete, got %d", len(shortcuts))
	}
	if shortcuts[0].ID != 2 {
		t.Errorf("remaining shortcut ID = %d; want 2", shortcuts[0].ID)
	}
}

func TestPlayerDeleteShortcutNonexistent(t *testing.T) {
	p, err := NewPlayer(1, 100, 1, "TestPlayer", 1, 0, 0)
	if err != nil {
		t.Fatal(err)
	}

	// Should not panic on deleting nonexistent shortcut
	p.DeleteShortcut(5, 3)

	if len(p.GetShortcuts()) != 0 {
		t.Errorf("expected 0 shortcuts, got %d", len(p.GetShortcuts()))
	}
}

func TestPlayerRegisterShortcutOverwrite(t *testing.T) {
	p, err := NewPlayer(1, 100, 1, "TestPlayer", 1, 0, 0)
	if err != nil {
		t.Fatal(err)
	}

	// Register first shortcut at slot 0, page 0
	p.RegisterShortcut(&Shortcut{Slot: 0, Page: 0, Type: ShortcutTypeSkill, ID: 100, Level: 1})

	// Overwrite with a different shortcut at the same slot/page
	p.RegisterShortcut(&Shortcut{Slot: 0, Page: 0, Type: ShortcutTypeItem, ID: 200, Level: 0})

	shortcuts := p.GetShortcuts()
	if len(shortcuts) != 1 {
		t.Fatalf("expected 1 shortcut after overwrite, got %d", len(shortcuts))
	}
	if shortcuts[0].Type != ShortcutTypeItem {
		t.Errorf("overwritten shortcut type = %d; want %d", shortcuts[0].Type, ShortcutTypeItem)
	}
	if shortcuts[0].ID != 200 {
		t.Errorf("overwritten shortcut ID = %d; want 200", shortcuts[0].ID)
	}
}

func TestPlayerSetShortcuts(t *testing.T) {
	p, err := NewPlayer(1, 100, 1, "TestPlayer", 1, 0, 0)
	if err != nil {
		t.Fatal(err)
	}

	// Register a shortcut first
	p.RegisterShortcut(&Shortcut{Slot: 0, Page: 0, Type: ShortcutTypeAction, ID: 1})

	// Replace all with a new set
	newShortcuts := []*Shortcut{
		{Slot: 5, Page: 2, Type: ShortcutTypeSkill, ID: 500, Level: 3},
		{Slot: 11, Page: 9, Type: ShortcutTypeMacro, ID: 600},
	}
	p.SetShortcuts(newShortcuts)

	shortcuts := p.GetShortcuts()
	if len(shortcuts) != 2 {
		t.Fatalf("expected 2 shortcuts after SetShortcuts, got %d", len(shortcuts))
	}

	// Verify old shortcut is gone
	for _, sc := range shortcuts {
		if sc.Slot == 0 && sc.Page == 0 {
			t.Error("old shortcut (slot=0, page=0) should have been replaced")
		}
	}
}

func TestPlayerGetShortcutsEmpty(t *testing.T) {
	p, err := NewPlayer(1, 100, 1, "TestPlayer", 1, 0, 0)
	if err != nil {
		t.Fatal(err)
	}

	shortcuts := p.GetShortcuts()
	if shortcuts != nil {
		t.Errorf("expected nil for empty shortcuts, got %v", shortcuts)
	}
}

func TestPlayerGetSkillLevel(t *testing.T) {
	p, err := NewPlayer(1, 100, 1, "TestPlayer", 1, 0, 0)
	if err != nil {
		t.Fatal(err)
	}

	// Not learned
	if lvl := p.GetSkillLevel(999); lvl != 0 {
		t.Errorf("GetSkillLevel(999) = %d; want 0 for unlearned skill", lvl)
	}

	// Learn a skill
	p.AddSkill(100, 3, false)

	if lvl := p.GetSkillLevel(100); lvl != 3 {
		t.Errorf("GetSkillLevel(100) = %d; want 3", lvl)
	}
}

func TestPlayerDifferentPages(t *testing.T) {
	p, err := NewPlayer(1, 100, 1, "TestPlayer", 1, 0, 0)
	if err != nil {
		t.Fatal(err)
	}

	// Same slot, different pages — must be independent
	p.RegisterShortcut(&Shortcut{Slot: 0, Page: 0, Type: ShortcutTypeAction, ID: 1})
	p.RegisterShortcut(&Shortcut{Slot: 0, Page: 1, Type: ShortcutTypeAction, ID: 2})
	p.RegisterShortcut(&Shortcut{Slot: 0, Page: 9, Type: ShortcutTypeAction, ID: 3})

	if len(p.GetShortcuts()) != 3 {
		t.Fatalf("expected 3 shortcuts across different pages, got %d", len(p.GetShortcuts()))
	}

	// Delete from page 1 only
	p.DeleteShortcut(0, 1)

	if len(p.GetShortcuts()) != 2 {
		t.Fatalf("expected 2 shortcuts after deleting from page 1, got %d", len(p.GetShortcuts()))
	}
}
