package world

import "sync/atomic"

// ObjectIDGenerator generates unique object IDs for all world entities.
// Phase 4.15: Centralized ID generation to prevent collisions between players, NPCs, items.
//
// ID ranges (convention):
//   0x00000000 - 0x0FFFFFFF: Reserved (0 = invalid/mock objects)
//   0x10000000 - 0x1FFFFFFF: Players (268M IDs)
//   0x20000000 - 0x2FFFFFFF: NPCs (268M IDs)
//   0x30000000 - 0x3FFFFFFF: Items on ground (268M IDs)
//   0x40000000 - 0xFFFFFFFF: Reserved for future use
type ObjectIDGenerator struct {
	nextPlayerID atomic.Uint32
	nextNpcID    atomic.Uint32
	nextItemID   atomic.Uint32
}

// NewObjectIDGenerator creates a new ID generator.
func NewObjectIDGenerator() *ObjectIDGenerator {
	gen := &ObjectIDGenerator{}
	gen.nextPlayerID.Store(0x10000000) // Start at 268M (player range)
	gen.nextNpcID.Store(0x20000000)    // Start at 536M (NPC range)
	gen.nextItemID.Store(0x30000000)   // Start at 805M (item range)
	return gen
}

// NextPlayerID generates next unique player object ID.
// Thread-safe via atomic increment.
func (g *ObjectIDGenerator) NextPlayerID() uint32 {
	return g.nextPlayerID.Add(1)
}

// NextNpcID generates next unique NPC object ID.
// Thread-safe via atomic increment.
func (g *ObjectIDGenerator) NextNpcID() uint32 {
	return g.nextNpcID.Add(1)
}

// NextItemID generates next unique item object ID.
// Thread-safe via atomic increment.
func (g *ObjectIDGenerator) NextItemID() uint32 {
	return g.nextItemID.Add(1)
}

// Global ID generator (singleton pattern).
// Initialized on first access via sync.Once in Instance().
var globalIDGenerator = NewObjectIDGenerator()

// IDGenerator returns global object ID generator.
// Thread-safe singleton.
func IDGenerator() *ObjectIDGenerator {
	return globalIDGenerator
}
