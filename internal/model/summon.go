package model

import (
	"sync"
	"sync/atomic"
)

// SummonType represents the kind of summoned creature.
type SummonType int32

const (
	// SummonTypePet — item-summoned pet (has inventory, feed, exp).
	SummonTypePet SummonType = 1
	// SummonTypeServitor — skill-summoned creature (time-limited).
	SummonTypeServitor SummonType = 2
)

// Summon represents a summoned creature (Pet or Servitor) in the game world.
// Embeds Character for HP/MP/stats and holds NPC template for display.
// Phase 19: Pets/Summons System.
type Summon struct {
	*Character // HP, MP, location, objectID

	mu         sync.RWMutex
	ownerID    uint32
	summonType SummonType
	templateID int32        // NPC template ID for display
	template   *NpcTemplate
	follow     atomic.Bool
	intention  atomic.Int32
	isDecayed  atomic.Bool

	// Combat
	target atomic.Uint32 // current target objectID

	// Stats from pet level data (override template stats)
	pAtk      int32
	pDef      int32
	mAtk      int32
	mDef      int32
	moveSpeed int32
	atkSpeed  int32
}

// NewSummon creates a new summoned creature with given stats.
// Phase 19: Pets/Summons System.
func NewSummon(
	objectID uint32,
	ownerID uint32,
	summonType SummonType,
	templateID int32,
	template *NpcTemplate,
	name string,
	level int32,
	maxHP, maxMP int32,
	pAtk, pDef, mAtk, mDef int32,
) *Summon {
	character := NewCharacter(objectID, name, Location{}, level, maxHP, maxMP, 0)

	s := &Summon{
		Character:  character,
		ownerID:    ownerID,
		summonType: summonType,
		templateID: templateID,
		template:   template,
		pAtk:       pAtk,
		pDef:       pDef,
		mAtk:       mAtk,
		mDef:       mDef,
		moveSpeed:  template.MoveSpeed(),
		atkSpeed:   template.AtkSpeed(),
	}

	s.follow.Store(true)
	s.intention.Store(int32(IntentionIdle))
	s.WorldObject.Data = s

	return s
}

// OwnerID returns the owner player's objectID.
func (s *Summon) OwnerID() uint32 {
	return s.ownerID
}

// Type returns summon type (Pet or Servitor).
func (s *Summon) Type() SummonType {
	return s.summonType
}

// TemplateID returns NPC template ID for display.
func (s *Summon) TemplateID() int32 {
	return s.templateID
}

// Template returns NPC template.
func (s *Summon) Template() *NpcTemplate {
	return s.template
}

// IsFollowing returns whether summon follows owner.
func (s *Summon) IsFollowing() bool {
	return s.follow.Load()
}

// SetFollow sets follow mode.
func (s *Summon) SetFollow(follow bool) {
	s.follow.Store(follow)
}

// Intention returns current AI intention.
func (s *Summon) Intention() Intention {
	return Intention(s.intention.Load())
}

// SetIntention sets AI intention.
func (s *Summon) SetIntention(intention Intention) {
	s.intention.Store(int32(intention))
}

// IsDecayed returns whether summon corpse has disappeared.
func (s *Summon) IsDecayed() bool {
	return s.isDecayed.Load()
}

// SetDecayed sets decayed state.
func (s *Summon) SetDecayed(decayed bool) {
	s.isDecayed.Store(decayed)
}

// Target returns current attack target objectID (0 if no target).
func (s *Summon) Target() uint32 {
	return s.target.Load()
}

// SetTarget sets current attack target.
func (s *Summon) SetTarget(objectID uint32) {
	s.target.Store(objectID)
}

// ClearTarget clears current target.
func (s *Summon) ClearTarget() {
	s.target.Store(0)
}

// PAtk returns physical attack.
func (s *Summon) PAtk() int32 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.pAtk
}

// PDef returns physical defense.
func (s *Summon) PDef() int32 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.pDef
}

// MAtk returns magical attack.
func (s *Summon) MAtk() int32 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.mAtk
}

// MDef returns magical defense.
func (s *Summon) MDef() int32 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.mDef
}

// MoveSpeed returns movement speed.
func (s *Summon) MoveSpeed() int32 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.moveSpeed
}

// AtkSpeed returns attack speed.
func (s *Summon) AtkSpeed() int32 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.atkSpeed
}

// UpdateStats updates combat stats (called on level change).
func (s *Summon) UpdateStats(maxHP, maxMP, pAtk, pDef, mAtk, mDef int32) {
	s.mu.Lock()
	s.pAtk = pAtk
	s.pDef = pDef
	s.mAtk = mAtk
	s.mDef = mDef
	s.mu.Unlock()

	s.SetMaxHP(maxHP)
	s.SetMaxMP(maxMP)
}

// IsPet returns true if this is a pet (not servitor).
func (s *Summon) IsPet() bool {
	return s.summonType == SummonTypePet
}

// IsServitor returns true if this is a servitor (not pet).
func (s *Summon) IsServitor() bool {
	return s.summonType == SummonTypeServitor
}
