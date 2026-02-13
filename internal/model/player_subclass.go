package model

import (
	"errors"
	"fmt"
	"sync"

	"github.com/udisondev/la2go/internal/data"
)

// Subclass-related errors.
var (
	ErrSubclassLocked     = errors.New("subclass operation in progress")
	ErrMaxSubclasses      = errors.New("maximum number of subclasses reached")
	ErrInvalidClassIndex  = errors.New("invalid class index")
	ErrSubclassExists     = errors.New("subclass slot already occupied")
	ErrInvalidSubclass    = errors.New("class not allowed as subclass")
	ErrLevelTooLow        = errors.New("level too low for subclass")
	ErrSubclassNotFound   = errors.New("subclass not found at index")
	ErrAlreadyActiveClass = errors.New("class already active")
)

// subclassFields holds all subclass-related state for Player.
// Extracted to keep Player struct manageable.
type subclassFields struct {
	subclassMu   sync.Mutex             // Dedicated mutex for subclass operations (not playerMu!)
	baseClassID  int32                  // Original base class (never changes)
	activeIndex  int32                  // Currently active class index (0=base, 1-3=sub)
	subClasses   map[int32]*SubClass    // key=classIndex (1-3)
}

// initSubclassFields initializes subclass state for a new player.
func (p *Player) initSubclassFields() {
	p.baseClassID = p.classID
	p.activeIndex = data.ClassIndexBase
	p.subClasses = make(map[int32]*SubClass, data.MaxSubclasses)
}

// BaseClassID returns the player's original base class ID (immutable after creation).
func (p *Player) BaseClassID() int32 {
	return p.baseClassID
}

// SetBaseClassID sets the base class ID (for DB restore only).
func (p *Player) SetBaseClassID(classID int32) {
	p.baseClassID = classID
}

// ActiveClassIndex returns the currently active class index (0=base, 1-3=sub).
func (p *Player) ActiveClassIndex() int32 {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()
	return p.activeIndex
}

// IsSubClassActive returns true if the player's active class is a subclass (not base).
func (p *Player) IsSubClassActive() bool {
	return p.ActiveClassIndex() > 0
}

// SubClasses returns a copy of all subclasses.
func (p *Player) SubClasses() map[int32]*SubClass {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()

	result := make(map[int32]*SubClass, len(p.subClasses))
	for k, v := range p.subClasses {
		cp := *v
		result[k] = &cp
	}
	return result
}

// SubClassCount returns the number of subclasses.
func (p *Player) SubClassCount() int {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()
	return len(p.subClasses)
}

// GetSubClass returns the subclass at the given index (1-3). Returns nil if empty.
func (p *Player) GetSubClass(classIndex int32) *SubClass {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()
	return p.subClasses[classIndex]
}

// AddSubClass adds a new subclass to the player.
// Validates: max subclasses, slot availability, class eligibility, level requirements.
// Returns the populated SubClass on success.
//
// Phase 14: Subclass System.
// Java reference: Player.addSubClass(), VillageMaster case 4.
func (p *Player) AddSubClass(classID, classIndex int32) (*SubClass, error) {
	if !p.subclassMu.TryLock() {
		return nil, ErrSubclassLocked
	}
	defer p.subclassMu.Unlock()

	p.playerMu.Lock()
	defer p.playerMu.Unlock()

	// Validate class index (1-3)
	if classIndex < 1 || classIndex > int32(data.MaxSubclasses) {
		return nil, ErrInvalidClassIndex
	}

	// Check max subclasses
	if len(p.subClasses) >= data.MaxSubclasses {
		return nil, ErrMaxSubclasses
	}

	// Check slot not occupied
	if _, exists := p.subClasses[classIndex]; exists {
		return nil, ErrSubclassExists
	}

	// Level check: player base level >= 75, all existing subs >= 75
	if p.level < data.MinSubclassLevel {
		return nil, fmt.Errorf("%w: base level %d, need %d", ErrLevelTooLow, p.level, data.MinSubclassLevel)
	}
	for _, sub := range p.subClasses {
		if sub.Level < data.MinSubclassLevel {
			return nil, fmt.Errorf("%w: subclass index %d level %d, need %d",
				ErrLevelTooLow, sub.ClassIndex, sub.Level, data.MinSubclassLevel)
		}
	}

	// Validate class choice against rules
	existingIDs := p.existingSubClassIDsLocked()
	if !data.IsValidSubClass(classID, p.baseClassID, p.raceID, existingIDs) {
		return nil, ErrInvalidSubclass
	}

	sub := NewSubClass(classID, classIndex)
	p.subClasses[classIndex] = sub

	return sub, nil
}

// ModifySubClass replaces an existing subclass at classIndex with a new class.
// The old subclass data is completely removed (hennas, skills, effects should be
// cleared by the caller before invoking this method).
// Returns the new SubClass on success.
//
// Phase 14: Subclass System.
// Java reference: Player.modifySubClass().
func (p *Player) ModifySubClass(classIndex, newClassID int32) (*SubClass, error) {
	if !p.subclassMu.TryLock() {
		return nil, ErrSubclassLocked
	}
	defer p.subclassMu.Unlock()

	p.playerMu.Lock()
	defer p.playerMu.Unlock()

	if classIndex < 1 || classIndex > int32(data.MaxSubclasses) {
		return nil, ErrInvalidClassIndex
	}

	old := p.subClasses[classIndex]
	if old == nil {
		return nil, ErrSubclassNotFound
	}

	// Build existing IDs excluding the slot being modified
	existingIDs := make([]int32, 0, len(p.subClasses)-1)
	for idx, sub := range p.subClasses {
		if idx != classIndex {
			existingIDs = append(existingIDs, sub.ClassID)
		}
	}

	if !data.IsValidSubClass(newClassID, p.baseClassID, p.raceID, existingIDs) {
		return nil, ErrInvalidSubclass
	}

	// Replace with fresh subclass
	sub := NewSubClass(newClassID, classIndex)
	p.subClasses[classIndex] = sub

	return sub, nil
}

// SetActiveClass switches the player's active class to the given index.
// Index 0 restores the base class, 1-3 activates a subclass.
//
// Before calling: the caller (handler) must save current class state.
// After calling: the caller must restore skills/effects/hennas for the new class.
//
// Phase 14: Subclass System.
// Java reference: Player.setActiveClass().
func (p *Player) SetActiveClass(classIndex int32) error {
	if !p.subclassMu.TryLock() {
		return ErrSubclassLocked
	}
	defer p.subclassMu.Unlock()

	p.playerMu.Lock()
	defer p.playerMu.Unlock()

	if p.activeIndex == classIndex {
		return ErrAlreadyActiveClass
	}

	if classIndex != data.ClassIndexBase && (classIndex < 1 || classIndex > int32(data.MaxSubclasses)) {
		return ErrInvalidClassIndex
	}

	// Validate target subclass exists (skip for base class, index 0)
	var targetSub *SubClass
	if classIndex > 0 {
		targetSub = p.subClasses[classIndex]
		if targetSub == nil {
			return ErrSubclassNotFound
		}
	}

	// Save current active class state into subclass before switching
	if p.activeIndex > 0 {
		if current := p.subClasses[p.activeIndex]; current != nil {
			current.Level = p.level
			current.Exp = p.experience
			current.SP = p.sp
			current.ClassID = p.classID
		}
	}

	// Switch to target class
	if classIndex == data.ClassIndexBase {
		p.classID = p.baseClassID
		p.activeIndex = data.ClassIndexBase
	} else {
		p.classID = targetSub.ClassID
		p.level = targetSub.Level
		p.experience = targetSub.Exp
		p.sp = targetSub.SP
		p.activeIndex = classIndex
	}

	return nil
}

// SaveActiveSubClassState persists current level/exp/sp into the active subclass slot.
// No-op if base class is active (index 0).
// Called before logout/disconnect to ensure subclass state is up-to-date.
func (p *Player) SaveActiveSubClassState() {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()

	if p.activeIndex == 0 {
		return
	}
	if sub := p.subClasses[p.activeIndex]; sub != nil {
		sub.Level = p.level
		sub.Exp = p.experience
		sub.SP = p.sp
	}
}

// RestoreSubClass restores a subclass from DB data (no validation).
// Used during player load.
func (p *Player) RestoreSubClass(sub *SubClass) {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()

	if p.subClasses == nil {
		p.subClasses = make(map[int32]*SubClass, data.MaxSubclasses)
	}
	p.subClasses[sub.ClassIndex] = sub
}

// RemoveSubClass removes a subclass from the given slot.
// Returns the removed SubClass or error.
func (p *Player) RemoveSubClass(classIndex int32) (*SubClass, error) {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()

	if classIndex < 1 || classIndex > int32(data.MaxSubclasses) {
		return nil, ErrInvalidClassIndex
	}

	sub := p.subClasses[classIndex]
	if sub == nil {
		return nil, ErrSubclassNotFound
	}

	delete(p.subClasses, classIndex)
	return sub, nil
}

// existingSubClassIDsLocked returns class IDs of all subclasses.
// Must be called with playerMu held.
func (p *Player) existingSubClassIDsLocked() []int32 {
	ids := make([]int32, 0, len(p.subClasses))
	for _, sub := range p.subClasses {
		ids = append(ids, sub.ClassID)
	}
	return ids
}

// ExistingSubClassIDs returns class IDs of all current subclasses. Thread-safe.
func (p *Player) ExistingSubClassIDs() []int32 {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()
	return p.existingSubClassIDsLocked()
}
