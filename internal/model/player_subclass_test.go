package model

import (
	"errors"
	"testing"

	"github.com/udisondev/la2go/internal/data"
)

// newTestPlayerForSubclass creates a player suitable for subclass tests.
// Gladiator (classID=2), level 75, Human.
func newTestPlayerForSubclass(t *testing.T) *Player {
	t.Helper()
	p, err := NewPlayer(1, 100, 1, "TestSub", 75, data.RaceHuman, 2) // Gladiator, Human
	if err != nil {
		t.Fatalf("NewPlayer() error: %v", err)
	}
	return p
}

func TestBaseClassID(t *testing.T) {
	t.Parallel()

	p := newTestPlayerForSubclass(t)
	if got := p.BaseClassID(); got != 2 {
		t.Errorf("BaseClassID() = %d; want 2", got)
	}
}

func TestActiveClassIndex_Default(t *testing.T) {
	t.Parallel()

	p := newTestPlayerForSubclass(t)
	if got := p.ActiveClassIndex(); got != 0 {
		t.Errorf("ActiveClassIndex() = %d; want 0 (base class)", got)
	}
	if p.IsSubClassActive() {
		t.Error("IsSubClassActive() = true; want false for new player")
	}
}

func TestAddSubClass(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		classID   int32
		index     int32
		wantErr   error
	}{
		{"valid: Bishop for Gladiator", 16, 1, nil},
		{"valid: Elder for Gladiator", 30, 2, nil},
		{"valid: Tyrant for Gladiator", 48, 3, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			p := newTestPlayerForSubclass(t)
			sub, err := p.AddSubClass(tt.classID, tt.index)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("AddSubClass() error = %v; want %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("AddSubClass() error: %v", err)
			}

			if sub.ClassID != tt.classID {
				t.Errorf("SubClass.ClassID = %d; want %d", sub.ClassID, tt.classID)
			}
			if sub.ClassIndex != tt.index {
				t.Errorf("SubClass.ClassIndex = %d; want %d", sub.ClassIndex, tt.index)
			}
			if sub.Level != data.BaseSubclassLevel {
				t.Errorf("SubClass.Level = %d; want %d", sub.Level, data.BaseSubclassLevel)
			}
		})
	}
}

func TestAddSubClass_Errors(t *testing.T) {
	t.Parallel()

	t.Run("invalid index 0", func(t *testing.T) {
		t.Parallel()
		p := newTestPlayerForSubclass(t)
		_, err := p.AddSubClass(16, 0)
		if !errors.Is(err, ErrInvalidClassIndex) {
			t.Errorf("error = %v; want %v", err, ErrInvalidClassIndex)
		}
	})

	t.Run("invalid index 4", func(t *testing.T) {
		t.Parallel()
		p := newTestPlayerForSubclass(t)
		_, err := p.AddSubClass(16, 4)
		if !errors.Is(err, ErrInvalidClassIndex) {
			t.Errorf("error = %v; want %v", err, ErrInvalidClassIndex)
		}
	})

	t.Run("slot already occupied", func(t *testing.T) {
		t.Parallel()
		p := newTestPlayerForSubclass(t)
		if _, err := p.AddSubClass(16, 1); err != nil {
			t.Fatal(err)
		}
		_, err := p.AddSubClass(30, 1)
		if !errors.Is(err, ErrSubclassExists) {
			t.Errorf("error = %v; want %v", err, ErrSubclassExists)
		}
	})

	t.Run("max subclasses reached", func(t *testing.T) {
		t.Parallel()
		p := newTestPlayerForSubclass(t)
		// Fill all 3 slots via RestoreSubClass (bypasses level check)
		for i, classID := range []int32{16, 30, 48} {
			p.RestoreSubClass(&SubClass{ClassID: classID, ClassIndex: int32(i + 1), Level: 75})
		}
		// 4th should fail — max subclasses reached
		_, err := p.AddSubClass(46, 1) // slot 1 occupied
		if !errors.Is(err, ErrMaxSubclasses) {
			t.Errorf("error = %v; want %v", err, ErrMaxSubclasses)
		}
	})

	t.Run("level too low", func(t *testing.T) {
		t.Parallel()
		p, err := NewPlayer(1, 100, 1, "LowLvl", 40, data.RaceHuman, 2)
		if err != nil {
			t.Fatal(err)
		}
		_, addErr := p.AddSubClass(16, 1)
		if !errors.Is(addErr, ErrLevelTooLow) {
			t.Errorf("error = %v; want %v", addErr, ErrLevelTooLow)
		}
	})

	t.Run("invalid class choice", func(t *testing.T) {
		t.Parallel()
		p := newTestPlayerForSubclass(t)
		// Overlord (51) is neverSubclassed
		_, err := p.AddSubClass(51, 1)
		if !errors.Is(err, ErrInvalidSubclass) {
			t.Errorf("error = %v; want %v", err, ErrInvalidSubclass)
		}
	})

	t.Run("own class as subclass", func(t *testing.T) {
		t.Parallel()
		p := newTestPlayerForSubclass(t)
		// Gladiator (2) — own class
		_, err := p.AddSubClass(2, 1)
		if !errors.Is(err, ErrInvalidSubclass) {
			t.Errorf("error = %v; want %v", err, ErrInvalidSubclass)
		}
	})
}

func TestSubClassCount(t *testing.T) {
	t.Parallel()

	p := newTestPlayerForSubclass(t)
	if got := p.SubClassCount(); got != 0 {
		t.Errorf("SubClassCount() = %d; want 0", got)
	}

	if _, err := p.AddSubClass(16, 1); err != nil {
		t.Fatal(err)
	}
	if got := p.SubClassCount(); got != 1 {
		t.Errorf("SubClassCount() = %d; want 1", got)
	}
}

func TestGetSubClass(t *testing.T) {
	t.Parallel()

	p := newTestPlayerForSubclass(t)

	if p.GetSubClass(1) != nil {
		t.Error("GetSubClass(1) != nil for empty player")
	}

	if _, err := p.AddSubClass(16, 1); err != nil {
		t.Fatal(err)
	}

	sub := p.GetSubClass(1)
	if sub == nil {
		t.Fatal("GetSubClass(1) = nil after AddSubClass")
	}
	if sub.ClassID != 16 {
		t.Errorf("ClassID = %d; want 16", sub.ClassID)
	}
}

func TestSetActiveClass(t *testing.T) {
	t.Parallel()

	p := newTestPlayerForSubclass(t)

	// Add Bishop as subclass
	if _, err := p.AddSubClass(16, 1); err != nil {
		t.Fatal(err)
	}

	// Switch to subclass
	if err := p.SetActiveClass(1); err != nil {
		t.Fatalf("SetActiveClass(1) error: %v", err)
	}

	if got := p.ActiveClassIndex(); got != 1 {
		t.Errorf("ActiveClassIndex() = %d; want 1", got)
	}
	if !p.IsSubClassActive() {
		t.Error("IsSubClassActive() = false; want true")
	}
	if got := p.ClassID(); got != 16 {
		t.Errorf("ClassID() = %d; want 16 (Bishop)", got)
	}
	if got := p.Level(); got != data.BaseSubclassLevel {
		t.Errorf("Level() = %d; want %d", got, data.BaseSubclassLevel)
	}

	// Switch back to base
	if err := p.SetActiveClass(0); err != nil {
		t.Fatalf("SetActiveClass(0) error: %v", err)
	}

	if p.IsSubClassActive() {
		t.Error("IsSubClassActive() = true; want false after switching back")
	}
	if got := p.ClassID(); got != 2 {
		t.Errorf("ClassID() = %d; want 2 (Gladiator)", got)
	}
}

func TestSetActiveClass_SavesState(t *testing.T) {
	t.Parallel()

	p := newTestPlayerForSubclass(t)

	// Add Bishop subclass
	if _, err := p.AddSubClass(16, 1); err != nil {
		t.Fatal(err)
	}

	// Switch to subclass, modify level
	if err := p.SetActiveClass(1); err != nil {
		t.Fatal(err)
	}
	if err := p.SetLevel(50); err != nil {
		t.Fatal(err)
	}
	p.SetExperience(12345)

	// Switch back to base — subclass state should be saved
	if err := p.SetActiveClass(0); err != nil {
		t.Fatal(err)
	}

	// Check subclass saved level/exp
	sub := p.GetSubClass(1)
	if sub == nil {
		t.Fatal("GetSubClass(1) = nil")
	}
	if sub.Level != 50 {
		t.Errorf("SubClass.Level = %d; want 50", sub.Level)
	}
	if sub.Exp != 12345 {
		t.Errorf("SubClass.Exp = %d; want 12345", sub.Exp)
	}
}

func TestSetActiveClass_Errors(t *testing.T) {
	t.Parallel()

	t.Run("already active", func(t *testing.T) {
		t.Parallel()
		p := newTestPlayerForSubclass(t)
		err := p.SetActiveClass(0) // already at base
		if !errors.Is(err, ErrAlreadyActiveClass) {
			t.Errorf("error = %v; want %v", err, ErrAlreadyActiveClass)
		}
	})

	t.Run("invalid index", func(t *testing.T) {
		t.Parallel()
		p := newTestPlayerForSubclass(t)
		err := p.SetActiveClass(4)
		if !errors.Is(err, ErrInvalidClassIndex) {
			t.Errorf("error = %v; want %v", err, ErrInvalidClassIndex)
		}
	})

	t.Run("empty slot", func(t *testing.T) {
		t.Parallel()
		p := newTestPlayerForSubclass(t)
		err := p.SetActiveClass(1)
		if !errors.Is(err, ErrSubclassNotFound) {
			t.Errorf("error = %v; want %v", err, ErrSubclassNotFound)
		}
	})
}

func TestModifySubClass(t *testing.T) {
	t.Parallel()

	p := newTestPlayerForSubclass(t)

	// Add Bishop (16) at slot 1
	if _, err := p.AddSubClass(16, 1); err != nil {
		t.Fatal(err)
	}

	// Modify to Elder (30)
	sub, err := p.ModifySubClass(1, 30)
	if err != nil {
		t.Fatalf("ModifySubClass() error: %v", err)
	}

	if sub.ClassID != 30 {
		t.Errorf("ClassID = %d; want 30 (Elder)", sub.ClassID)
	}
	if sub.Level != data.BaseSubclassLevel {
		t.Errorf("Level = %d; want %d (reset)", sub.Level, data.BaseSubclassLevel)
	}

	// Verify in player
	stored := p.GetSubClass(1)
	if stored == nil || stored.ClassID != 30 {
		t.Errorf("stored subclass ClassID = %v; want 30", stored)
	}
}

func TestModifySubClass_Errors(t *testing.T) {
	t.Parallel()

	t.Run("empty slot", func(t *testing.T) {
		t.Parallel()
		p := newTestPlayerForSubclass(t)
		_, err := p.ModifySubClass(1, 30)
		if !errors.Is(err, ErrSubclassNotFound) {
			t.Errorf("error = %v; want %v", err, ErrSubclassNotFound)
		}
	})

	t.Run("invalid class", func(t *testing.T) {
		t.Parallel()
		p := newTestPlayerForSubclass(t)
		if _, err := p.AddSubClass(16, 1); err != nil {
			t.Fatal(err)
		}
		// Overlord (51) — neverSubclassed
		_, err := p.ModifySubClass(1, 51)
		if !errors.Is(err, ErrInvalidSubclass) {
			t.Errorf("error = %v; want %v", err, ErrInvalidSubclass)
		}
	})
}

func TestRemoveSubClass(t *testing.T) {
	t.Parallel()

	p := newTestPlayerForSubclass(t)

	if _, err := p.AddSubClass(16, 1); err != nil {
		t.Fatal(err)
	}

	removed, err := p.RemoveSubClass(1)
	if err != nil {
		t.Fatalf("RemoveSubClass() error: %v", err)
	}
	if removed.ClassID != 16 {
		t.Errorf("removed ClassID = %d; want 16", removed.ClassID)
	}

	if p.GetSubClass(1) != nil {
		t.Error("slot 1 should be nil after RemoveSubClass")
	}
	if got := p.SubClassCount(); got != 0 {
		t.Errorf("SubClassCount() = %d; want 0", got)
	}
}

func TestRemoveSubClass_Errors(t *testing.T) {
	t.Parallel()

	p := newTestPlayerForSubclass(t)

	_, err := p.RemoveSubClass(1)
	if !errors.Is(err, ErrSubclassNotFound) {
		t.Errorf("error = %v; want %v", err, ErrSubclassNotFound)
	}

	_, err = p.RemoveSubClass(0)
	if !errors.Is(err, ErrInvalidClassIndex) {
		t.Errorf("error = %v; want %v", err, ErrInvalidClassIndex)
	}
}

func TestRestoreSubClass(t *testing.T) {
	t.Parallel()

	p := newTestPlayerForSubclass(t)

	sub := &SubClass{
		ClassID:    30,
		ClassIndex: 2,
		Level:      60,
		Exp:        88511413,
		SP:         5000,
	}

	p.RestoreSubClass(sub)

	if got := p.SubClassCount(); got != 1 {
		t.Errorf("SubClassCount() = %d; want 1", got)
	}

	restored := p.GetSubClass(2)
	if restored == nil {
		t.Fatal("GetSubClass(2) = nil after RestoreSubClass")
	}
	if restored.ClassID != 30 {
		t.Errorf("ClassID = %d; want 30", restored.ClassID)
	}
	if restored.Level != 60 {
		t.Errorf("Level = %d; want 60", restored.Level)
	}
}

func TestSaveActiveSubClassState(t *testing.T) {
	t.Parallel()

	p := newTestPlayerForSubclass(t)

	if _, err := p.AddSubClass(16, 1); err != nil {
		t.Fatal(err)
	}
	if err := p.SetActiveClass(1); err != nil {
		t.Fatal(err)
	}

	// Modify player state directly
	if err := p.SetLevel(55); err != nil {
		t.Fatal(err)
	}
	p.SetExperience(99999)

	// Save state
	p.SaveActiveSubClassState()

	sub := p.GetSubClass(1)
	if sub.Level != 55 {
		t.Errorf("Level = %d; want 55", sub.Level)
	}
	if sub.Exp != 99999 {
		t.Errorf("Exp = %d; want 99999", sub.Exp)
	}
}

func TestSaveActiveSubClassState_BaseClass(t *testing.T) {
	t.Parallel()

	p := newTestPlayerForSubclass(t)

	// Should be no-op when base class is active
	p.SaveActiveSubClassState() // no panic
}

func TestExistingSubClassIDs(t *testing.T) {
	t.Parallel()

	p := newTestPlayerForSubclass(t)

	ids := p.ExistingSubClassIDs()
	if len(ids) != 0 {
		t.Errorf("ExistingSubClassIDs() = %v; want empty", ids)
	}

	// Use RestoreSubClass to bypass level-check on existing subs
	p.RestoreSubClass(&SubClass{ClassID: 16, ClassIndex: 1, Level: 75})
	p.RestoreSubClass(&SubClass{ClassID: 30, ClassIndex: 2, Level: 75})

	ids = p.ExistingSubClassIDs()
	if len(ids) != 2 {
		t.Errorf("len(ExistingSubClassIDs()) = %d; want 2", len(ids))
	}
}

func TestSubClasses_ReturnsCopy(t *testing.T) {
	t.Parallel()

	p := newTestPlayerForSubclass(t)
	if _, err := p.AddSubClass(16, 1); err != nil {
		t.Fatal(err)
	}

	subs := p.SubClasses()
	// Modifying the copy should not affect the player
	subs[1].Level = 80

	original := p.GetSubClass(1)
	if original.Level == 80 {
		t.Error("SubClasses() should return a copy, not a reference")
	}
}
