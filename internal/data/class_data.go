package data

import "slices"

// Race IDs for Lineage 2 Interlude.
const (
	RaceHuman   int32 = 0
	RaceElf     int32 = 1
	RaceDarkElf int32 = 2
	RaceOrc     int32 = 3
	RaceDwarf   int32 = 4
)

// Subclass system constants.
const (
	BaseSubclassLevel int32 = 40 // New subclass starts at level 40
	MaxSubclasses           = 3  // Max number of subclasses (indices 1-3)
	MinSubclassLevel  int32 = 75 // Min level to add a subclass
	ClassIndexBase          = 0  // Base class always at index 0
)

// ClassInfo holds metadata for a single player class.
type ClassInfo struct {
	ID         int32
	Name       string
	IsMage     bool
	IsSummoner bool
	Race       int32
	ParentID   int32 // -1 = no parent (base class)
}

// classTable stores all 88 Interlude classes (IDs 0-57, 88-118).
// Indexed by classID.
var classTable map[int32]*ClassInfo

func init() {
	classTable = map[int32]*ClassInfo{
		// --- Human Fighter line ---
		0:  {ID: 0, Name: "Human Fighter", Race: RaceHuman, ParentID: -1},
		1:  {ID: 1, Name: "Warrior", Race: RaceHuman, ParentID: 0},
		2:  {ID: 2, Name: "Gladiator", Race: RaceHuman, ParentID: 1},
		3:  {ID: 3, Name: "Warlord", Race: RaceHuman, ParentID: 1},
		4:  {ID: 4, Name: "Knight", Race: RaceHuman, ParentID: 0},
		5:  {ID: 5, Name: "Paladin", Race: RaceHuman, ParentID: 4},
		6:  {ID: 6, Name: "Dark Avenger", Race: RaceHuman, ParentID: 4},
		7:  {ID: 7, Name: "Rogue", Race: RaceHuman, ParentID: 0},
		8:  {ID: 8, Name: "Treasure Hunter", Race: RaceHuman, ParentID: 7},
		9:  {ID: 9, Name: "Hawkeye", Race: RaceHuman, ParentID: 7},

		// --- Human Mage line ---
		10: {ID: 10, Name: "Human Mystic", IsMage: true, Race: RaceHuman, ParentID: -1},
		11: {ID: 11, Name: "Human Wizard", IsMage: true, Race: RaceHuman, ParentID: 10},
		12: {ID: 12, Name: "Sorcerer", IsMage: true, Race: RaceHuman, ParentID: 11},
		13: {ID: 13, Name: "Necromancer", IsMage: true, Race: RaceHuman, ParentID: 11},
		14: {ID: 14, Name: "Warlock", IsMage: true, IsSummoner: true, Race: RaceHuman, ParentID: 11},
		15: {ID: 15, Name: "Cleric", IsMage: true, Race: RaceHuman, ParentID: 10},
		16: {ID: 16, Name: "Bishop", IsMage: true, Race: RaceHuman, ParentID: 15},
		17: {ID: 17, Name: "Prophet", IsMage: true, Race: RaceHuman, ParentID: 15},

		// --- Elf Fighter line ---
		18: {ID: 18, Name: "Elven Fighter", Race: RaceElf, ParentID: -1},
		19: {ID: 19, Name: "Elven Knight", Race: RaceElf, ParentID: 18},
		20: {ID: 20, Name: "Temple Knight", Race: RaceElf, ParentID: 19},
		21: {ID: 21, Name: "Swordsinger", Race: RaceElf, ParentID: 19},
		22: {ID: 22, Name: "Elven Scout", Race: RaceElf, ParentID: 18},
		23: {ID: 23, Name: "Plains Walker", Race: RaceElf, ParentID: 22},
		24: {ID: 24, Name: "Silver Ranger", Race: RaceElf, ParentID: 22},

		// --- Elf Mage line ---
		25: {ID: 25, Name: "Elven Mystic", IsMage: true, Race: RaceElf, ParentID: -1},
		26: {ID: 26, Name: "Elven Wizard", IsMage: true, Race: RaceElf, ParentID: 25},
		27: {ID: 27, Name: "Spellsinger", IsMage: true, Race: RaceElf, ParentID: 26},
		28: {ID: 28, Name: "Elemental Summoner", IsMage: true, IsSummoner: true, Race: RaceElf, ParentID: 26},
		29: {ID: 29, Name: "Elven Oracle", IsMage: true, Race: RaceElf, ParentID: 25},
		30: {ID: 30, Name: "Elder", IsMage: true, Race: RaceElf, ParentID: 29},

		// --- Dark Elf Fighter line ---
		31: {ID: 31, Name: "Dark Fighter", Race: RaceDarkElf, ParentID: -1},
		32: {ID: 32, Name: "Palus Knight", Race: RaceDarkElf, ParentID: 31},
		33: {ID: 33, Name: "Shillien Knight", Race: RaceDarkElf, ParentID: 32},
		34: {ID: 34, Name: "Bladedancer", Race: RaceDarkElf, ParentID: 32},
		35: {ID: 35, Name: "Assassin", Race: RaceDarkElf, ParentID: 31},
		36: {ID: 36, Name: "Abyss Walker", Race: RaceDarkElf, ParentID: 35},
		37: {ID: 37, Name: "Phantom Ranger", Race: RaceDarkElf, ParentID: 35},

		// --- Dark Elf Mage line ---
		38: {ID: 38, Name: "Dark Mystic", IsMage: true, Race: RaceDarkElf, ParentID: -1},
		39: {ID: 39, Name: "Dark Wizard", IsMage: true, Race: RaceDarkElf, ParentID: 38},
		40: {ID: 40, Name: "Spellhowler", IsMage: true, Race: RaceDarkElf, ParentID: 39},
		41: {ID: 41, Name: "Phantom Summoner", IsMage: true, IsSummoner: true, Race: RaceDarkElf, ParentID: 39},
		42: {ID: 42, Name: "Shillien Oracle", IsMage: true, Race: RaceDarkElf, ParentID: 38},
		43: {ID: 43, Name: "Shillien Elder", IsMage: true, Race: RaceDarkElf, ParentID: 42},

		// --- Orc Fighter line ---
		44: {ID: 44, Name: "Orc Fighter", Race: RaceOrc, ParentID: -1},
		45: {ID: 45, Name: "Orc Raider", Race: RaceOrc, ParentID: 44},
		46: {ID: 46, Name: "Destroyer", Race: RaceOrc, ParentID: 45},
		47: {ID: 47, Name: "Orc Monk", Race: RaceOrc, ParentID: 44},
		48: {ID: 48, Name: "Tyrant", Race: RaceOrc, ParentID: 47},

		// --- Orc Mage line ---
		49: {ID: 49, Name: "Orc Mystic", IsMage: true, Race: RaceOrc, ParentID: -1},
		50: {ID: 50, Name: "Orc Shaman", IsMage: true, Race: RaceOrc, ParentID: 49},
		51: {ID: 51, Name: "Overlord", IsMage: true, Race: RaceOrc, ParentID: 50},
		52: {ID: 52, Name: "Warcryer", IsMage: true, Race: RaceOrc, ParentID: 50},

		// --- Dwarf Fighter line ---
		53: {ID: 53, Name: "Dwarven Fighter", Race: RaceDwarf, ParentID: -1},
		54: {ID: 54, Name: "Scavenger", Race: RaceDwarf, ParentID: 53},
		55: {ID: 55, Name: "Bounty Hunter", Race: RaceDwarf, ParentID: 54},
		56: {ID: 56, Name: "Artisan", Race: RaceDwarf, ParentID: 53},
		57: {ID: 57, Name: "Warsmith", Race: RaceDwarf, ParentID: 56},

		// --- 3rd Class: Human ---
		88:  {ID: 88, Name: "Duelist", Race: RaceHuman, ParentID: 2},
		89:  {ID: 89, Name: "Dreadnought", Race: RaceHuman, ParentID: 3},
		90:  {ID: 90, Name: "Phoenix Knight", Race: RaceHuman, ParentID: 5},
		91:  {ID: 91, Name: "Hell Knight", Race: RaceHuman, ParentID: 6},
		92:  {ID: 92, Name: "Sagittarius", Race: RaceHuman, ParentID: 9},
		93:  {ID: 93, Name: "Adventurer", Race: RaceHuman, ParentID: 8},
		94:  {ID: 94, Name: "Archmage", IsMage: true, Race: RaceHuman, ParentID: 12},
		95:  {ID: 95, Name: "Soultaker", IsMage: true, Race: RaceHuman, ParentID: 13},
		96:  {ID: 96, Name: "Arcana Lord", IsMage: true, IsSummoner: true, Race: RaceHuman, ParentID: 14},
		97:  {ID: 97, Name: "Cardinal", IsMage: true, Race: RaceHuman, ParentID: 16},
		98:  {ID: 98, Name: "Hierophant", IsMage: true, Race: RaceHuman, ParentID: 17},

		// --- 3rd Class: Elf ---
		99:  {ID: 99, Name: "Eva's Templar", Race: RaceElf, ParentID: 20},
		100: {ID: 100, Name: "Sword Muse", Race: RaceElf, ParentID: 21},
		101: {ID: 101, Name: "Wind Rider", Race: RaceElf, ParentID: 23},
		102: {ID: 102, Name: "Moonlight Sentinel", Race: RaceElf, ParentID: 24},
		103: {ID: 103, Name: "Mystic Muse", IsMage: true, Race: RaceElf, ParentID: 27},
		104: {ID: 104, Name: "Elemental Master", IsMage: true, IsSummoner: true, Race: RaceElf, ParentID: 28},
		105: {ID: 105, Name: "Eva's Saint", IsMage: true, Race: RaceElf, ParentID: 30},

		// --- 3rd Class: Dark Elf ---
		106: {ID: 106, Name: "Shillien Templar", Race: RaceDarkElf, ParentID: 33},
		107: {ID: 107, Name: "Spectral Dancer", Race: RaceDarkElf, ParentID: 34},
		108: {ID: 108, Name: "Ghost Hunter", Race: RaceDarkElf, ParentID: 36},
		109: {ID: 109, Name: "Ghost Sentinel", Race: RaceDarkElf, ParentID: 37},
		110: {ID: 110, Name: "Storm Screamer", IsMage: true, Race: RaceDarkElf, ParentID: 40},
		111: {ID: 111, Name: "Spectral Master", IsMage: true, IsSummoner: true, Race: RaceDarkElf, ParentID: 41},
		112: {ID: 112, Name: "Shillien Saint", IsMage: true, Race: RaceDarkElf, ParentID: 43},

		// --- 3rd Class: Orc ---
		113: {ID: 113, Name: "Titan", Race: RaceOrc, ParentID: 46},
		114: {ID: 114, Name: "Grand Khavatari", Race: RaceOrc, ParentID: 48},
		115: {ID: 115, Name: "Dominator", IsMage: true, Race: RaceOrc, ParentID: 51},
		116: {ID: 116, Name: "Doomcryer", IsMage: true, Race: RaceOrc, ParentID: 52},

		// --- 3rd Class: Dwarf ---
		117: {ID: 117, Name: "Fortune Seeker", Race: RaceDwarf, ParentID: 55},
		118: {ID: 118, Name: "Maestro", Race: RaceDwarf, ParentID: 57},
	}
}

// thirdClassGroup contains 2nd occupation class IDs eligible for subclass.
// Java: CategoryData.xml → THIRD_CLASS_GROUP.
var thirdClassGroup = map[int32]struct{}{
	2: {}, 3: {}, 5: {}, 6: {}, 8: {}, 9: {},
	12: {}, 13: {}, 14: {}, 16: {}, 17: {},
	20: {}, 21: {}, 23: {}, 24: {},
	27: {}, 28: {}, 30: {},
	33: {}, 34: {}, 36: {}, 37: {},
	40: {}, 41: {}, 43: {},
	46: {}, 48: {}, 51: {}, 52: {},
	55: {}, 57: {},
}

// neverSubclassed contains classes that can NEVER be taken as a subclass.
var neverSubclassed = map[int32]struct{}{
	51: {}, // Overlord
	57: {}, // Warsmith
}

// mutualExclusionGroups — mutual exclusion groups for subclasses.
// If a player has any class from a group, they cannot take another from the same group.
// Java: VillageMaster.java subclasseSet1-5.
var mutualExclusionGroups = [][]int32{
	{5, 6, 20, 33},   // Tanks: Paladin, Dark Avenger, Temple Knight, Shillien Knight
	{8, 36, 23},       // Daggers: Treasure Hunter, Abyss Walker, Plains Walker
	{9, 24, 37},       // Archers: Hawkeye, Silver Ranger, Phantom Ranger
	{14, 28, 41},      // Summoners: Warlock, Elemental Summoner, Phantom Summoner
	{12, 27, 40},      // Nukers: Sorcerer, Spellsinger, Spellhowler
}

// classToExclusionGroup maps each class in a mutual exclusion group to its group index.
// Built from mutualExclusionGroups at init.
var classToExclusionGroup map[int32]int

func init() {
	classToExclusionGroup = make(map[int32]int, 32)
	for i, group := range mutualExclusionGroups {
		for _, classID := range group {
			classToExclusionGroup[classID] = i
		}
	}
}

// GetClassInfo returns class info by ID. Returns nil if unknown.
func GetClassInfo(classID int32) *ClassInfo {
	return classTable[classID]
}

// ClassLevel returns occupation level (0=base, 1=1st, 2=2nd, 3=3rd).
// Returns -1 for unknown classes.
func ClassLevel(classID int32) int {
	info := classTable[classID]
	if info == nil {
		return -1
	}
	if info.ParentID < 0 {
		return 0
	}
	return 1 + ClassLevel(info.ParentID)
}

// ClassParent returns parent class ID. Returns -1 if base class or unknown.
func ClassParent(classID int32) int32 {
	info := classTable[classID]
	if info == nil {
		return -1
	}
	return info.ParentID
}

// ClassRootID returns the root (base occupation) class ID.
func ClassRootID(classID int32) int32 {
	info := classTable[classID]
	if info == nil {
		return classID
	}
	if info.ParentID < 0 {
		return classID
	}
	return ClassRootID(info.ParentID)
}

// ClassChildOf returns true if classID is a child (direct or transitive) of parentID.
func ClassChildOf(classID, parentID int32) bool {
	info := classTable[classID]
	if info == nil || info.ParentID < 0 {
		return false
	}
	if info.ParentID == parentID {
		return true
	}
	return ClassChildOf(info.ParentID, parentID)
}

// ClassEqualsOrChildOf returns true if classID == targetID or classID is a child of targetID.
func ClassEqualsOrChildOf(classID, targetID int32) bool {
	if classID == targetID {
		return true
	}
	return ClassChildOf(classID, targetID)
}

// Get2ndClassID returns the 2nd occupation class ID for a given class.
// If classID is already 2nd occupation (level=2), returns it unchanged.
// If classID is 3rd occupation (level=3), returns its parent.
// Returns -1 for classes below 2nd occupation.
func Get2ndClassID(classID int32) int32 {
	level := ClassLevel(classID)
	switch level {
	case 2:
		return classID
	case 3:
		return ClassParent(classID)
	default:
		return -1
	}
}

// IsSubclassEligible checks whether a player's base class is eligible for subclass system.
// Only 2nd+ occupation classes (level >= 2) qualify.
func IsSubclassEligible(baseClassID int32) bool {
	return ClassLevel(baseClassID) >= 2
}

// GetAvailableSubClasses returns a list of class IDs available as subclasses
// for a player with the given base class, race, and existing subclasses.
//
// Rules (Java: VillageMaster.getSubclasses + getAvailableSubClasses):
//  1. Start with THIRD_CLASS_GROUP
//  2. Remove neverSubclassed (Overlord, Warsmith)
//  3. Remove own class (2nd class equivalent)
//  4. Elf ⇄ Dark Elf restriction
//  5. Remove mutual exclusion group of base class
//  6. Remove existing subclasses and their parent/child classes
func GetAvailableSubClasses(baseClassID, playerRace int32, existingSubClassIDs []int32) []int32 {
	// Resolve base class to 2nd occupation for comparison
	base2nd := Get2ndClassID(baseClassID)
	if base2nd < 0 {
		return nil
	}

	// Start with copy of thirdClassGroup
	available := make(map[int32]struct{}, len(thirdClassGroup))
	for id := range thirdClassGroup {
		available[id] = struct{}{}
	}

	// Remove neverSubclassed
	for id := range neverSubclassed {
		delete(available, id)
	}

	// Remove own 2nd class
	delete(available, base2nd)

	// Elf ⇄ Dark Elf restriction
	switch playerRace {
	case RaceElf:
		for id := range available {
			info := classTable[id]
			if info != nil && info.Race == RaceDarkElf {
				delete(available, id)
			}
		}
	case RaceDarkElf:
		for id := range available {
			info := classTable[id]
			if info != nil && info.Race == RaceElf {
				delete(available, id)
			}
		}
	}

	// Remove mutual exclusion group of base class
	if groupIdx, ok := classToExclusionGroup[base2nd]; ok {
		for _, id := range mutualExclusionGroups[groupIdx] {
			delete(available, id)
		}
	}

	// Remove existing subclasses and their parent/child classes
	for _, subID := range existingSubClassIDs {
		sub2nd := Get2ndClassID(subID)
		if sub2nd < 0 {
			sub2nd = subID
		}

		for id := range available {
			if id == sub2nd || ClassEqualsOrChildOf(id, sub2nd) || ClassEqualsOrChildOf(sub2nd, id) {
				delete(available, id)
			}
		}

		// Also remove mutual exclusion group of existing subclass
		if groupIdx, ok := classToExclusionGroup[sub2nd]; ok {
			for _, id := range mutualExclusionGroups[groupIdx] {
				delete(available, id)
			}
		}
	}

	result := make([]int32, 0, len(available))
	for id := range available {
		result = append(result, id)
	}
	return result
}

// IsValidSubClass checks if classID is a valid subclass choice for the given player state.
func IsValidSubClass(classID, baseClassID, playerRace int32, existingSubClassIDs []int32) bool {
	available := GetAvailableSubClasses(baseClassID, playerRace, existingSubClassIDs)
	return slices.Contains(available, classID)
}
