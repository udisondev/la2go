package model

// ShortcutType represents the type of shortcut in the action bar.
type ShortcutType int8

const (
	ShortcutTypeNone   ShortcutType = 0
	ShortcutTypeItem   ShortcutType = 1
	ShortcutTypeSkill  ShortcutType = 2
	ShortcutTypeAction ShortcutType = 3
	ShortcutTypeMacro  ShortcutType = 4
	ShortcutTypeRecipe ShortcutType = 5
)

const (
	// MaxShortcutsPerBar is the number of slots per shortcut page (F1-F12).
	MaxShortcutsPerBar = 12
	// MaxShortcutPages is the maximum number of shortcut pages (0-9).
	MaxShortcutPages = 10
)

// Shortcut represents a single shortcut binding in the action bar.
// Each shortcut is identified by (Slot, Page) pair.
//
// Reference: L2J_Mobius Shortcut.java
type Shortcut struct {
	Slot  int8
	Page  int8
	Type  ShortcutType
	ID    int32 // itemObjectID / skillID / actionID / macroID / recipeID
	Level int32 // used only for ShortcutTypeSkill
}

// shortcutKey computes the map key for a shortcut: slot + page*12.
func shortcutKey(slot, page int8) int32 {
	return int32(slot) + int32(page)*MaxShortcutsPerBar
}
