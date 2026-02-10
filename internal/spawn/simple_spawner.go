package spawn

import (
	"context"
	"fmt"

	"github.com/udisondev/la2go/internal/ai"
	"github.com/udisondev/la2go/internal/model"
)

// SimpleSpawner provides hardcoded test spawns for MVP demonstration
// Not for production use — real spawns should be loaded from database
type SimpleSpawner struct {
	manager *Manager
}

// NewSimpleSpawner creates new simple spawner
func NewSimpleSpawner(manager *Manager) *SimpleSpawner {
	return &SimpleSpawner{
		manager: manager,
	}
}

// SpawnTestNpc spawns a single test NPC at Talking Island coordinates
// Returns spawned NPC or error
func (s *SimpleSpawner) SpawnTestNpc(ctx context.Context) (*model.Npc, error) {
	// Hardcoded test template (Wolf, level 5)
	template := model.NewNpcTemplate(
		1000,              // templateID
		"Wolf",            // name
		"Wild Beast",      // title
		5,                 // level
		1500,              // maxHP
		800,               // maxMP
		100, 50,           // pAtk, pDef
		80, 40,            // mAtk, mDef
		300,               // aggroRange
		120,               // moveSpeed
		253,               // atkSpeed
		30, 60,            // respawnMin, respawnMax
	)

	// Hardcoded test spawn (Talking Island coordinates)
	spawn := model.NewSpawn(
		999999,  // spawnID (fake ID for test)
		1000,    // templateID
		17000,   // x
		170000,  // y
		-3500,   // z
		0,       // heading
		1,       // maximumCount
		true,    // doRespawn
	)

	// Store template in memory (bypass database for test)
	// In real implementation, template would be in database
	// For now, we'll create NPC directly

	// Generate unique objectID
	objectID := s.manager.objectIDCounter.Add(1)

	// Create NPC
	npc := model.NewNpc(objectID, template.TemplateID(), template)
	npc.SetSpawn(spawn)
	npc.SetLocation(spawn.Location())

	// Add to world (Phase 4.10 Part 2: use AddNpc for NPC tracking)
	if err := s.manager.world.AddNpc(npc); err != nil {
		return nil, fmt.Errorf("adding test NPC to world: %w", err)
	}

	// Create and register AI
	npcAI := ai.NewBasicNpcAI(npc)
	s.manager.aiManager.Register(objectID, npcAI)

	return npc, nil
}

// SpawnTestNpcAt spawns test NPC at specific coordinates
func (s *SimpleSpawner) SpawnTestNpcAt(ctx context.Context, x, y, z int32) (*model.Npc, error) {
	template := model.NewNpcTemplate(
		1001, "Test Orc", "", 10, 2000, 1000,
		150, 75, 100, 50, 0, 100, 273, 60, 120,
	)

	spawn := model.NewSpawn(
		999998, 1001, x, y, z, 0, 1, true,
	)

	objectID := s.manager.objectIDCounter.Add(1)
	npc := model.NewNpc(objectID, template.TemplateID(), template)
	npc.SetSpawn(spawn)
	npc.SetLocation(spawn.Location())

	// Phase 4.10 Part 2: use AddNpc for NPC tracking
	if err := s.manager.world.AddNpc(npc); err != nil {
		return nil, fmt.Errorf("adding test NPC to world: %w", err)
	}

	npcAI := ai.NewBasicNpcAI(npc)
	s.manager.aiManager.Register(objectID, npcAI)

	return npc, nil
}
