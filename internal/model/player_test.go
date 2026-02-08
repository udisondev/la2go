package model

import (
	"sync"
	"testing"
	"time"
)

func TestNewPlayer(t *testing.T) {
	tests := []struct {
		name        string
		characterID int64
		accountID   int64
		playerName  string
		level       int32
		raceID      int32
		classID     int32
		wantErr     bool
	}{
		{
			name:        "valid player",
			characterID: 1,
			accountID:   100,
			playerName:  "TestHero",
			level:       1,
			raceID:      0,
			classID:     0,
			wantErr:     false,
		},
		{
			name:        "max level",
			characterID: 2,
			accountID:   100,
			playerName:  "MaxLevel",
			level:       80,
			raceID:      0,
			classID:     0,
			wantErr:     false,
		},
		{
			name:        "name too short",
			characterID: 3,
			accountID:   100,
			playerName:  "A",
			level:       1,
			raceID:      0,
			classID:     0,
			wantErr:     true,
		},
		{
			name:        "empty name",
			characterID: 4,
			accountID:   100,
			playerName:  "",
			level:       1,
			raceID:      0,
			classID:     0,
			wantErr:     true,
		},
		{
			name:        "level too low",
			characterID: 5,
			accountID:   100,
			playerName:  "TestHero",
			level:       0,
			raceID:      0,
			classID:     0,
			wantErr:     true,
		},
		{
			name:        "level too high",
			characterID: 6,
			accountID:   100,
			playerName:  "TestHero",
			level:       81,
			raceID:      0,
			classID:     0,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			player, err := NewPlayer(tt.characterID, tt.accountID, tt.playerName, tt.level, tt.raceID, tt.classID)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewPlayer() error = nil, wantErr = true")
				}
				return
			}

			if err != nil {
				t.Errorf("NewPlayer() unexpected error = %v", err)
				return
			}

			if player == nil {
				t.Fatal("NewPlayer() returned nil")
			}

			// Проверяем поля
			if player.CharacterID() != tt.characterID {
				t.Errorf("CharacterID() = %d, want %d", player.CharacterID(), tt.characterID)
			}
			if player.AccountID() != tt.accountID {
				t.Errorf("AccountID() = %d, want %d", player.AccountID(), tt.accountID)
			}
			if player.Name() != tt.playerName {
				t.Errorf("Name() = %q, want %q", player.Name(), tt.playerName)
			}
			if player.Level() != tt.level {
				t.Errorf("Level() = %d, want %d", player.Level(), tt.level)
			}
			if player.RaceID() != tt.raceID {
				t.Errorf("RaceID() = %d, want %d", player.RaceID(), tt.raceID)
			}
			if player.ClassID() != tt.classID {
				t.Errorf("ClassID() = %d, want %d", player.ClassID(), tt.classID)
			}

			// Experience должно быть 0
			if player.Experience() != 0 {
				t.Errorf("Experience() = %d, want 0", player.Experience())
			}

			// CreatedAt должно быть недавно
			if time.Since(player.CreatedAt()) > time.Second {
				t.Errorf("CreatedAt() = %v, want recent time", player.CreatedAt())
			}

			// Stats должны быть > 0
			if player.MaxHP() <= 0 {
				t.Errorf("MaxHP() = %d, want > 0", player.MaxHP())
			}
			if player.MaxMP() <= 0 {
				t.Errorf("MaxMP() = %d, want > 0", player.MaxMP())
			}
			if player.MaxCP() <= 0 {
				t.Errorf("MaxCP() = %d, want > 0", player.MaxCP())
			}
		})
	}
}

func TestPlayer_ImmutableFields(t *testing.T) {
	player, err := NewPlayer(123, 456, "TestHero", 1, 0, 0)
	if err != nil {
		t.Fatalf("NewPlayer() error = %v", err)
	}

	// CharacterID и AccountID должны быть immutable
	id1 := player.CharacterID()
	id2 := player.CharacterID()
	if id1 != id2 {
		t.Errorf("CharacterID changed: %d != %d", id1, id2)
	}

	accID1 := player.AccountID()
	accID2 := player.AccountID()
	if accID1 != accID2 {
		t.Errorf("AccountID changed: %d != %d", accID1, accID2)
	}
}

func TestPlayer_SetLevel(t *testing.T) {
	player, _ := NewPlayer(1, 100, "TestHero", 1, 0, 0)

	tests := []struct {
		name    string
		level   int32
		wantErr bool
	}{
		{"valid level 10", 10, false},
		{"valid level 80", 80, false},
		{"invalid level 0", 0, true},
		{"invalid level 81", 81, true},
		{"invalid level -1", -1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := player.SetLevel(tt.level)

			if tt.wantErr {
				if err == nil {
					t.Errorf("SetLevel(%d) error = nil, wantErr = true", tt.level)
				}
			} else {
				if err != nil {
					t.Errorf("SetLevel(%d) unexpected error = %v", tt.level, err)
				}
				if player.Level() != tt.level {
					t.Errorf("After SetLevel(%d), Level() = %d", tt.level, player.Level())
				}
			}
		})
	}
}

func TestPlayer_Experience(t *testing.T) {
	player, _ := NewPlayer(1, 100, "TestHero", 1, 0, 0)

	// Initial experience = 0
	if player.Experience() != 0 {
		t.Errorf("Initial Experience() = %d, want 0", player.Experience())
	}

	// AddExperience positive
	player.AddExperience(1000)
	if player.Experience() != 1000 {
		t.Errorf("After AddExperience(1000), Experience() = %d, want 1000", player.Experience())
	}

	// AddExperience more
	player.AddExperience(500)
	if player.Experience() != 1500 {
		t.Errorf("After AddExperience(500), Experience() = %d, want 1500", player.Experience())
	}

	// AddExperience negative (penalty)
	player.AddExperience(-200)
	if player.Experience() != 1300 {
		t.Errorf("After AddExperience(-200), Experience() = %d, want 1300", player.Experience())
	}

	// AddExperience negative below 0 — должно clamp к 0
	player.AddExperience(-2000)
	if player.Experience() != 0 {
		t.Errorf("After AddExperience(-2000), Experience() = %d, want 0 (clamped)", player.Experience())
	}

	// SetExperience
	player.SetExperience(50000)
	if player.Experience() != 50000 {
		t.Errorf("After SetExperience(50000), Experience() = %d", player.Experience())
	}

	// SetExperience negative — должно clamp к 0
	player.SetExperience(-100)
	if player.Experience() != 0 {
		t.Errorf("After SetExperience(-100), Experience() = %d, want 0 (clamped)", player.Experience())
	}
}

func TestPlayer_LastLogin(t *testing.T) {
	player, _ := NewPlayer(1, 100, "TestHero", 1, 0, 0)

	// Initial LastLogin — zero value
	if !player.LastLogin().IsZero() {
		t.Errorf("Initial LastLogin() = %v, want zero time", player.LastLogin())
	}

	// UpdateLastLogin
	before := time.Now()
	player.UpdateLastLogin()
	after := time.Now()

	lastLogin := player.LastLogin()
	if lastLogin.Before(before) || lastLogin.After(after) {
		t.Errorf("UpdateLastLogin() = %v, want between %v and %v", lastLogin, before, after)
	}

	// SetLastLogin (для загрузки из DB)
	customTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	player.SetLastLogin(customTime)

	if player.LastLogin() != customTime {
		t.Errorf("After SetLastLogin, LastLogin() = %v, want %v", player.LastLogin(), customTime)
	}
}

func TestPlayer_CreatedAt(t *testing.T) {
	player, _ := NewPlayer(1, 100, "TestHero", 1, 0, 0)

	// CreatedAt должно быть недавно
	if time.Since(player.CreatedAt()) > time.Second {
		t.Errorf("CreatedAt() = %v, want recent time", player.CreatedAt())
	}

	// SetCreatedAt (для загрузки из DB)
	customTime := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)
	player.SetCreatedAt(customTime)

	if player.CreatedAt() != customTime {
		t.Errorf("After SetCreatedAt, CreatedAt() = %v, want %v", player.CreatedAt(), customTime)
	}
}

func TestPlayer_SetCharacterID(t *testing.T) {
	player, _ := NewPlayer(0, 100, "TestHero", 1, 0, 0)

	if player.CharacterID() != 0 {
		t.Errorf("Initial CharacterID() = %d, want 0", player.CharacterID())
	}

	// SetCharacterID (для repository.Create)
	player.SetCharacterID(999)

	if player.CharacterID() != 999 {
		t.Errorf("After SetCharacterID, CharacterID() = %d, want 999", player.CharacterID())
	}
}

func TestPlayer_RaceAndClass(t *testing.T) {
	player, _ := NewPlayer(1, 100, "TestHero", 1, 5, 10)

	if player.RaceID() != 5 {
		t.Errorf("RaceID() = %d, want 5", player.RaceID())
	}
	if player.ClassID() != 10 {
		t.Errorf("ClassID() = %d, want 10", player.ClassID())
	}

	// SetRaceID
	player.SetRaceID(7)
	if player.RaceID() != 7 {
		t.Errorf("After SetRaceID, RaceID() = %d, want 7", player.RaceID())
	}

	// SetClassID (изменение профессии)
	player.SetClassID(20)
	if player.ClassID() != 20 {
		t.Errorf("After SetClassID, ClassID() = %d, want 20", player.ClassID())
	}
}

func TestPlayer_InheritedCharacter(t *testing.T) {
	player, _ := NewPlayer(1, 100, "TestHero", 1, 0, 0)

	// Проверяем что Character методы работают
	player.SetCurrentHP(500)
	if player.CurrentHP() != 500 {
		t.Errorf("CurrentHP() = %d, want 500", player.CurrentHP())
	}

	if player.IsDead() {
		t.Error("IsDead() = true, want false")
	}

	player.SetCurrentHP(0)
	if !player.IsDead() {
		t.Error("IsDead() = false, want true (HP=0)")
	}
}

func TestPlayer_InheritedWorldObject(t *testing.T) {
	player, _ := NewPlayer(1, 100, "TestHero", 1, 0, 0)

	// Проверяем что WorldObject методы работают
	newLoc := NewLocation(100, 200, 300, 1000)
	player.SetLocation(newLoc)

	loc := player.Location()
	if loc != newLoc {
		t.Errorf("Location() = %+v, want %+v", loc, newLoc)
	}

	if player.X() != 100 {
		t.Errorf("X() = %d, want 100", player.X())
	}
}

func TestPlayer_ConcurrentLevelUpdates(t *testing.T) {
	player, _ := NewPlayer(1, 100, "TestHero", 1, 0, 0)

	const numUpdaters = 50
	var wg sync.WaitGroup
	wg.Add(numUpdaters)

	// Concurrent level updates (1-80)
	for i := range numUpdaters {
		go func(id int) {
			defer wg.Done()

			for j := range 100 {
				level := int32((j % 80) + 1)
				if err := player.SetLevel(level); err != nil {
					t.Errorf("SetLevel(%d) error = %v", level, err)
				}
			}
		}(i)
	}

	wg.Wait()

	// Финальный level должен быть в валидных пределах
	level := player.Level()
	if level < 1 || level > 80 {
		t.Errorf("Invalid level after concurrent updates: %d", level)
	}
}

func TestPlayer_ConcurrentExperienceUpdates(t *testing.T) {
	player, _ := NewPlayer(1, 100, "TestHero", 1, 0, 0)

	const numUpdaters = 50
	var wg sync.WaitGroup
	wg.Add(numUpdaters)

	// Concurrent experience updates
	for range numUpdaters {
		go func() {
			defer wg.Done()

			for range 100 {
				player.AddExperience(100)
			}
		}()
	}

	wg.Wait()

	// Experience должно быть >= 0
	exp := player.Experience()
	if exp < 0 {
		t.Errorf("Invalid experience after concurrent updates: %d", exp)
	}
}

func TestPlayer_MixedConcurrentAccess(t *testing.T) {
	player, _ := NewPlayer(1, 100, "TestHero", 10, 0, 0)

	const numReaders = 50
	const numWriters = 10
	var wg sync.WaitGroup
	wg.Add(numReaders + numWriters)

	// Readers
	for range numReaders {
		go func() {
			defer wg.Done()

			for range 500 {
				_ = player.Level()
				_ = player.Experience()
				_ = player.RaceID()
				_ = player.ClassID()
				_ = player.LastLogin()
			}
		}()
	}

	// Writers
	for i := range numWriters {
		go func(id int) {
			defer wg.Done()

			for j := range 100 {
				level := int32((j % 80) + 1)
				_ = player.SetLevel(level)
				player.AddExperience(int64(id * 100))
				player.UpdateLastLogin()
			}
		}(i)
	}

	wg.Wait()

	// Финальные значения должны быть консистентными
	if player.Level() < 1 || player.Level() > 80 {
		t.Errorf("Invalid level: %d", player.Level())
	}
	if player.Experience() < 0 {
		t.Errorf("Invalid experience: %d", player.Experience())
	}
}

// Benchmark для hot path methods
func BenchmarkPlayer_Level(b *testing.B) {
	player, _ := NewPlayer(1, 100, "TestHero", 10, 0, 0)

	b.ResetTimer()
	for b.Loop() {
		_ = player.Level()
	}
}

func BenchmarkPlayer_AddExperience(b *testing.B) {
	player, _ := NewPlayer(1, 100, "TestHero", 10, 0, 0)

	b.ResetTimer()
	for b.Loop() {
		player.AddExperience(100)
	}
}

func BenchmarkPlayer_UpdateLastLogin(b *testing.B) {
	player, _ := NewPlayer(1, 100, "TestHero", 10, 0, 0)

	b.ResetTimer()
	for b.Loop() {
		player.UpdateLastLogin()
	}
}
