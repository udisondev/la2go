package data

import (
	"testing"
)

func TestGetClassInfo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		classID int32
		wantNil bool
		wantName string
		wantRace int32
	}{
		{"Human Fighter", 0, false, "Human Fighter", RaceHuman},
		{"Gladiator", 2, false, "Gladiator", RaceHuman},
		{"Temple Knight", 20, false, "Temple Knight", RaceElf},
		{"Abyss Walker", 36, false, "Abyss Walker", RaceDarkElf},
		{"Destroyer", 46, false, "Destroyer", RaceOrc},
		{"Warsmith", 57, false, "Warsmith", RaceDwarf},
		{"Duelist (3rd)", 88, false, "Duelist", RaceHuman},
		{"Maestro (3rd)", 118, false, "Maestro", RaceDwarf},
		{"unknown", 999, true, "", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			info := GetClassInfo(tt.classID)
			if tt.wantNil {
				if info != nil {
					t.Errorf("GetClassInfo(%d) = %v; want nil", tt.classID, info)
				}
				return
			}
			if info == nil {
				t.Fatalf("GetClassInfo(%d) = nil; want non-nil", tt.classID)
			}
			if info.Name != tt.wantName {
				t.Errorf("Name = %q; want %q", info.Name, tt.wantName)
			}
			if info.Race != tt.wantRace {
				t.Errorf("Race = %d; want %d", info.Race, tt.wantRace)
			}
		})
	}
}

func TestClassLevel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		classID int32
		want    int
	}{
		{"base class (Fighter)", 0, 0},
		{"1st occupation (Warrior)", 1, 1},
		{"2nd occupation (Gladiator)", 2, 2},
		{"3rd occupation (Duelist)", 88, 3},
		{"base class (Elf Mage)", 25, 0},
		{"2nd occupation (Elder)", 30, 2},
		{"3rd occupation (Eva's Saint)", 105, 3},
		{"unknown class", 999, -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := ClassLevel(tt.classID); got != tt.want {
				t.Errorf("ClassLevel(%d) = %d; want %d", tt.classID, got, tt.want)
			}
		})
	}
}

func TestClassChildOf(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		classID  int32
		parentID int32
		want     bool
	}{
		{"Gladiator child of Warrior", 2, 1, true},
		{"Gladiator child of Fighter", 2, 0, true},
		{"Duelist child of Gladiator", 88, 2, true},
		{"Duelist child of Fighter", 88, 0, true},
		{"Fighter not child of anything", 0, 1, false},
		{"Gladiator not child of self", 2, 2, false},
		{"Gladiator not child of Knight", 2, 4, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := ClassChildOf(tt.classID, tt.parentID); got != tt.want {
				t.Errorf("ClassChildOf(%d, %d) = %v; want %v", tt.classID, tt.parentID, got, tt.want)
			}
		})
	}
}

func TestClassEqualsOrChildOf(t *testing.T) {
	t.Parallel()

	if !ClassEqualsOrChildOf(2, 2) {
		t.Error("ClassEqualsOrChildOf(2, 2) = false; want true (same class)")
	}
	if !ClassEqualsOrChildOf(88, 2) {
		t.Error("ClassEqualsOrChildOf(88, 2) = false; want true (Duelist child of Gladiator)")
	}
	if ClassEqualsOrChildOf(2, 88) {
		t.Error("ClassEqualsOrChildOf(2, 88) = true; want false")
	}
}

func TestClassRootID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		classID int32
		want    int32
	}{
		{0, 0},   // Fighter → Fighter
		{2, 0},   // Gladiator → Fighter
		{88, 0},  // Duelist → Fighter
		{30, 25}, // Elder → Elven Mystic
		{118, 53}, // Maestro → Dwarven Fighter
	}

	for _, tt := range tests {
		if got := ClassRootID(tt.classID); got != tt.want {
			t.Errorf("ClassRootID(%d) = %d; want %d", tt.classID, got, tt.want)
		}
	}
}

func TestGet2ndClassID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		classID int32
		want    int32
	}{
		{2, 2},    // Gladiator (2nd) → 2
		{88, 2},   // Duelist (3rd) → parent = Gladiator(2)
		{0, -1},   // Fighter (base) → invalid
		{1, -1},   // Warrior (1st) → invalid
		{105, 30}, // Eva's Saint (3rd) → Elder(30)
	}

	for _, tt := range tests {
		if got := Get2ndClassID(tt.classID); got != tt.want {
			t.Errorf("Get2ndClassID(%d) = %d; want %d", tt.classID, got, tt.want)
		}
	}
}

func TestIsSubclassEligible(t *testing.T) {
	t.Parallel()

	if !IsSubclassEligible(2) {
		t.Error("IsSubclassEligible(2) = false; want true (Gladiator, 2nd)")
	}
	if !IsSubclassEligible(88) {
		t.Error("IsSubclassEligible(88) = false; want true (Duelist, 3rd)")
	}
	if IsSubclassEligible(0) {
		t.Error("IsSubclassEligible(0) = true; want false (Fighter, base)")
	}
	if IsSubclassEligible(1) {
		t.Error("IsSubclassEligible(1) = true; want false (Warrior, 1st)")
	}
}

func TestGetAvailableSubClasses_Basic(t *testing.T) {
	t.Parallel()

	// Gladiator (Human, classID=2) with no existing subclasses
	available := GetAvailableSubClasses(2, RaceHuman, nil)

	if len(available) == 0 {
		t.Fatal("GetAvailableSubClasses(2, Human, nil) returned empty list")
	}

	// Should NOT contain: own class (2), neverSubclassed (51, 57)
	avSet := make(map[int32]struct{}, len(available))
	for _, id := range available {
		avSet[id] = struct{}{}
	}

	if _, ok := avSet[2]; ok {
		t.Error("available contains own class (Gladiator=2)")
	}
	if _, ok := avSet[51]; ok {
		t.Error("available contains Overlord (51) — neverSubclassed")
	}
	if _, ok := avSet[57]; ok {
		t.Error("available contains Warsmith (57) — neverSubclassed")
	}
}

func TestGetAvailableSubClasses_ElfDarkElfRestriction(t *testing.T) {
	t.Parallel()

	// Elf Temple Knight (classID=20) should not see Dark Elf classes
	available := GetAvailableSubClasses(20, RaceElf, nil)

	for _, id := range available {
		info := GetClassInfo(id)
		if info != nil && info.Race == RaceDarkElf {
			t.Errorf("Elf player can see Dark Elf class %d (%s)", id, info.Name)
		}
	}

	// Dark Elf Abyss Walker (classID=36) should not see Elf classes
	available = GetAvailableSubClasses(36, RaceDarkElf, nil)

	for _, id := range available {
		info := GetClassInfo(id)
		if info != nil && info.Race == RaceElf {
			t.Errorf("Dark Elf player can see Elf class %d (%s)", id, info.Name)
		}
	}
}

func TestGetAvailableSubClasses_MutualExclusion(t *testing.T) {
	t.Parallel()

	// Paladin (5) is in tank group {5, 6, 20, 33}
	// Available subclasses should NOT contain other tanks
	available := GetAvailableSubClasses(5, RaceHuman, nil)
	avSet := make(map[int32]struct{}, len(available))
	for _, id := range available {
		avSet[id] = struct{}{}
	}

	tanks := []int32{5, 6, 20, 33}
	for _, id := range tanks {
		if _, ok := avSet[id]; ok {
			info := GetClassInfo(id)
			t.Errorf("Paladin can see mutual exclusion class %d (%s)", id, info.Name)
		}
	}
}

func TestGetAvailableSubClasses_ExistingSubclass(t *testing.T) {
	t.Parallel()

	// Gladiator (2) has existing subclass Sorcerer (12)
	// Available should NOT contain: 12 (Sorcerer), 27 (Spellsinger), 40 (Spellhowler) — nuker group
	available := GetAvailableSubClasses(2, RaceHuman, []int32{12})
	avSet := make(map[int32]struct{}, len(available))
	for _, id := range available {
		avSet[id] = struct{}{}
	}

	nukers := []int32{12, 27, 40}
	for _, id := range nukers {
		if _, ok := avSet[id]; ok {
			t.Errorf("available contains nuker class %d despite existing subclass Sorcerer(12)", id)
		}
	}
}

func TestGetAvailableSubClasses_3rdClass(t *testing.T) {
	t.Parallel()

	// Duelist (88, 3rd class, parent=Gladiator(2))
	// Should work the same as base Gladiator
	av3rd := GetAvailableSubClasses(88, RaceHuman, nil)
	av2nd := GetAvailableSubClasses(2, RaceHuman, nil)

	if len(av3rd) != len(av2nd) {
		t.Errorf("3rd class available (%d) != 2nd class available (%d)", len(av3rd), len(av2nd))
	}
}

func TestIsValidSubClass(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		classID   int32
		baseClass int32
		race      int32
		existing  []int32
		want      bool
	}{
		{
			name: "valid subclass",
			classID: 16, baseClass: 2, race: RaceHuman, // Bishop for Gladiator
			want: true,
		},
		{
			name: "own class",
			classID: 2, baseClass: 2, race: RaceHuman, // Gladiator for Gladiator
			want: false,
		},
		{
			name: "neverSubclassed",
			classID: 51, baseClass: 2, race: RaceHuman, // Overlord
			want: false,
		},
		{
			name: "Elf trying Dark Elf",
			classID: 33, baseClass: 20, race: RaceElf, // Shillien Knight for Temple Knight
			want: false,
		},
		{
			name: "mutual exclusion",
			classID: 6, baseClass: 5, race: RaceHuman, // Dark Avenger for Paladin (same tank group)
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := IsValidSubClass(tt.classID, tt.baseClass, tt.race, tt.existing); got != tt.want {
				t.Errorf("IsValidSubClass(%d, %d, %d, %v) = %v; want %v",
					tt.classID, tt.baseClass, tt.race, tt.existing, got, tt.want)
			}
		})
	}
}

func TestClassInfo_IsMage(t *testing.T) {
	t.Parallel()

	// Fighter classes
	for _, id := range []int32{0, 2, 5, 8, 20, 36, 46, 55, 88} {
		info := GetClassInfo(id)
		if info == nil {
			t.Fatalf("GetClassInfo(%d) = nil", id)
		}
		if info.IsMage {
			t.Errorf("class %d (%s) IsMage = true; want false", id, info.Name)
		}
	}

	// Mage classes
	for _, id := range []int32{10, 12, 16, 27, 40, 51, 94, 105} {
		info := GetClassInfo(id)
		if info == nil {
			t.Fatalf("GetClassInfo(%d) = nil", id)
		}
		if !info.IsMage {
			t.Errorf("class %d (%s) IsMage = false; want true", id, info.Name)
		}
	}
}

func TestClassInfo_IsSummoner(t *testing.T) {
	t.Parallel()

	summoners := []int32{14, 28, 41, 96, 104, 111}
	for _, id := range summoners {
		info := GetClassInfo(id)
		if info == nil {
			t.Fatalf("GetClassInfo(%d) = nil", id)
		}
		if !info.IsSummoner {
			t.Errorf("class %d (%s) IsSummoner = false; want true", id, info.Name)
		}
	}

	nonSummoners := []int32{12, 27, 40, 94, 103}
	for _, id := range nonSummoners {
		info := GetClassInfo(id)
		if info == nil {
			t.Fatalf("GetClassInfo(%d) = nil", id)
		}
		if info.IsSummoner {
			t.Errorf("class %d (%s) IsSummoner = true; want false", id, info.Name)
		}
	}
}

func TestAllClassesHaveValidParent(t *testing.T) {
	t.Parallel()

	for id, info := range classTable {
		if info.ParentID >= 0 {
			parent := GetClassInfo(info.ParentID)
			if parent == nil {
				t.Errorf("class %d (%s) has parentID %d but parent not found", id, info.Name, info.ParentID)
			}
		}
	}
}

func TestThirdClassGroupCount(t *testing.T) {
	t.Parallel()

	// THIRD_CLASS_GROUP should have 31 classes (all 2nd occupation)
	if len(thirdClassGroup) != 31 {
		t.Errorf("thirdClassGroup has %d entries; want 31", len(thirdClassGroup))
	}

	// All should be level 2
	for id := range thirdClassGroup {
		if level := ClassLevel(id); level != 2 {
			t.Errorf("thirdClassGroup class %d has level %d; want 2", id, level)
		}
	}
}
