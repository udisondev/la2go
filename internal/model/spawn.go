package model

import (
	"sync"
	"sync/atomic"
)

// Spawn represents a spawn point for NPCs
type Spawn struct {
	spawnID      int64
	templateID   int32
	location     Location
	maximumCount int32
	doRespawn    bool

	mu           sync.RWMutex
	currentCount atomic.Int32
	npcList      []*Npc // currently spawned NPCs (forward reference, resolved after npc.go compiled)
}

// NewSpawn creates a new spawn point
func NewSpawn(
	spawnID int64,
	templateID int32,
	x, y, z int32,
	heading uint16,
	maximumCount int32,
	doRespawn bool,
) *Spawn {
	s := &Spawn{
		spawnID:      spawnID,
		templateID:   templateID,
		location:     NewLocation(x, y, z, heading),
		maximumCount: maximumCount,
		doRespawn:    doRespawn,
		npcList:      make([]*Npc, 0, maximumCount),
	}
	s.currentCount.Store(0)
	return s
}

// SpawnID returns spawn ID
func (s *Spawn) SpawnID() int64 {
	return s.spawnID
}

// TemplateID returns template ID
func (s *Spawn) TemplateID() int32 {
	return s.templateID
}

// Location returns spawn location
func (s *Spawn) Location() Location {
	return s.location
}

// Heading returns spawn heading
func (s *Spawn) Heading() uint16 {
	return s.location.Heading
}

// MaximumCount returns maximum number of NPCs that can spawn
func (s *Spawn) MaximumCount() int32 {
	return s.maximumCount
}

// DoRespawn returns whether NPCs should respawn after death
func (s *Spawn) DoRespawn() bool {
	return s.doRespawn
}

// CurrentCount returns current spawned count (atomic read)
func (s *Spawn) CurrentCount() int32 {
	return s.currentCount.Load()
}

// IncreaseCount increases spawned count by 1 (atomic)
func (s *Spawn) IncreaseCount() {
	s.currentCount.Add(1)
}

// DecreaseCount decreases spawned count by 1 (atomic)
func (s *Spawn) DecreaseCount() {
	s.currentCount.Add(-1)
}

// AddNpc adds NPC to spawn's NPC list
func (s *Spawn) AddNpc(npc *Npc) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.npcList = append(s.npcList, npc)
}

// RemoveNpc removes NPC from spawn's NPC list
func (s *Spawn) RemoveNpc(npc *Npc) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, n := range s.npcList {
		if n == npc {
			s.npcList = append(s.npcList[:i], s.npcList[i+1:]...)
			break
		}
	}
}

// NPCs returns copy of spawned NPCs list
func (s *Spawn) NPCs() []*Npc {
	s.mu.RLock()
	defer s.mu.RUnlock()
	npcs := make([]*Npc, len(s.npcList))
	copy(npcs, s.npcList)
	return npcs
}
