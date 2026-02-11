package model

import "sync/atomic"

// Monster represents an aggressive NPC (mob).
// Phase 5.7: Added AggroList and target tracking.
type Monster struct {
	*Npc // embedding Npc

	isAggressive atomic.Bool
	aggroRange   int32
	aggroList    *AggroList    // Phase 5.7: hate tracking
	target       atomic.Uint32 // Phase 5.7: current target objectID
}

// NewMonster creates a new Monster instance.
// Phase 5.7: initializes AggroList.
func NewMonster(objectID uint32, templateID int32, template *NpcTemplate) *Monster {
	npc := NewNpc(objectID, templateID, template)

	monster := &Monster{
		Npc:        npc,
		aggroRange: template.AggroRange(),
		aggroList:  NewAggroList(),
	}

	// Set aggressive flag based on aggro range
	monster.isAggressive.Store(template.AggroRange() > 0)

	// Phase 5.6: override WorldObject.Data to point to Monster (not Npc)
	// This enables type assertion in CombatManager: target.Data.(*Monster)
	npc.WorldObject.Data = monster

	return monster
}

// IsAggressive returns whether monster is aggressive (atomic read)
func (m *Monster) IsAggressive() bool {
	return m.isAggressive.Load()
}

// SetAggressive sets aggressive flag (atomic write)
func (m *Monster) SetAggressive(aggressive bool) {
	m.isAggressive.Store(aggressive)
}

// AggroRange returns aggro range
func (m *Monster) AggroRange() int32 {
	return m.aggroRange
}

// AggroList returns the hate list for this monster.
// Phase 5.7: NPC Aggro & Basic AI.
func (m *Monster) AggroList() *AggroList {
	return m.aggroList
}

// Target returns current target objectID (0 if no target).
// Phase 5.7: NPC Aggro & Basic AI.
func (m *Monster) Target() uint32 {
	return m.target.Load()
}

// SetTarget sets current target objectID.
// Phase 5.7: NPC Aggro & Basic AI.
func (m *Monster) SetTarget(objectID uint32) {
	m.target.Store(objectID)
}

// ClearTarget clears current target.
// Phase 5.7: NPC Aggro & Basic AI.
func (m *Monster) ClearTarget() {
	m.target.Store(0)
}
