package model

import "sync/atomic"

// Monster represents an aggressive NPC (mob)
type Monster struct {
	*Npc // embedding Npc

	isAggressive atomic.Bool
	aggroRange   int32
}

// NewMonster creates a new Monster instance
func NewMonster(objectID uint32, templateID int32, template *NpcTemplate) *Monster {
	npc := NewNpc(objectID, templateID, template)

	monster := &Monster{
		Npc:        npc,
		aggroRange: template.AggroRange(),
	}

	// Set aggressive flag based on aggro range
	monster.isAggressive.Store(template.AggroRange() > 0)

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
