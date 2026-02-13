package data

// ClassRace returns the race ID for a given class ID.
// 0=Human, 1=Elf, 2=DarkElf, 3=Orc, 4=Dwarf
//
// Java reference: PlayerClass.getRace()
func ClassRace(classID int32) int32 {
	switch {
	// Human classes
	case classID >= 0 && classID <= 9:
		return 0 // Human
	case classID >= 10 && classID <= 17:
		return 0 // Human
	// 2nd class Human
	case classID >= 88 && classID <= 101:
		return 0 // Human

	// Elf classes
	case classID >= 18 && classID <= 24:
		return 1 // Elf
	case classID >= 25 && classID <= 30:
		return 1 // Elf
	// 2nd class Elf
	case classID >= 102 && classID <= 107:
		return 1 // Elf

	// Dark Elf classes
	case classID >= 31 && classID <= 37:
		return 2 // Dark Elf
	case classID >= 38 && classID <= 43:
		return 2 // Dark Elf
	// 2nd class Dark Elf
	case classID >= 108 && classID <= 113:
		return 2 // Dark Elf

	// Orc classes
	case classID >= 44 && classID <= 48:
		return 3 // Orc
	case classID >= 49 && classID <= 52:
		return 3 // Orc
	// 2nd class Orc
	case classID >= 114 && classID <= 117:
		return 3 // Orc

	// Dwarf classes
	case classID >= 53 && classID <= 57:
		return 4 // Dwarf
	// 2nd class Dwarf
	case classID == 118:
		return 4 // Dwarf

	default:
		return 0 // Fallback to Human
	}
}

// IsBaseClass returns true if the classID is a base (starting) class.
func IsBaseClass(classID int32) bool {
	return ClassLevel(classID) == 0
}
