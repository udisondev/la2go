package model

import "github.com/udisondev/la2go/internal/data"

// SubClass holds state for a single subclass slot.
// Each player can have up to 3 subclasses (indices 1-3).
// Index 0 is the base class (stored in Player.classID).
//
// Phase 14: Subclass System.
// Java reference: SubClassHolder.java
type SubClass struct {
	ClassID    int32 // PlayerClass ID (2-57, 88-118)
	ClassIndex int32 // Slot index (1, 2, or 3)
	Level      int32
	Exp        int64
	SP         int64
}

// NewSubClass creates a new subclass at default level 40 with appropriate XP.
func NewSubClass(classID, classIndex int32) *SubClass {
	return &SubClass{
		ClassID:    classID,
		ClassIndex: classIndex,
		Level:      data.BaseSubclassLevel,
		Exp:        data.GetExpForLevel(data.BaseSubclassLevel),
		SP:         0,
	}
}
