package model

import (
	"testing"
)

func TestNewSpawn(t *testing.T) {
	spawn := NewSpawn(
		1,       // spawnID
		1000,    // templateID
		17000,   // x
		170000,  // y
		-3500,   // z
		0,       // heading
		3,       // maximumCount
		true,    // doRespawn
	)

	if spawn.SpawnID() != 1 {
		t.Errorf("SpawnID() = %d, want 1", spawn.SpawnID())
	}
	if spawn.TemplateID() != 1000 {
		t.Errorf("TemplateID() = %d, want 1000", spawn.TemplateID())
	}
	if spawn.Location().X != 17000 {
		t.Errorf("Location().X = %d, want 17000", spawn.Location().X)
	}
	if spawn.Location().Y != 170000 {
		t.Errorf("Location().Y = %d, want 170000", spawn.Location().Y)
	}
	if spawn.Location().Z != -3500 {
		t.Errorf("Location().Z = %d, want -3500", spawn.Location().Z)
	}
	if spawn.Heading() != 0 {
		t.Errorf("Heading() = %d, want 0", spawn.Heading())
	}
	if spawn.MaximumCount() != 3 {
		t.Errorf("MaximumCount() = %d, want 3", spawn.MaximumCount())
	}
	if !spawn.DoRespawn() {
		t.Error("DoRespawn() = false, want true")
	}
	if spawn.CurrentCount() != 0 {
		t.Errorf("CurrentCount() = %d, want 0", spawn.CurrentCount())
	}
}

func TestSpawn_CountOperations(t *testing.T) {
	spawn := NewSpawn(1, 1000, 0, 0, 0, 0, 5, true)

	if spawn.CurrentCount() != 0 {
		t.Errorf("initial CurrentCount() = %d, want 0", spawn.CurrentCount())
	}

	spawn.IncreaseCount()
	if spawn.CurrentCount() != 1 {
		t.Errorf("after IncreaseCount() CurrentCount() = %d, want 1", spawn.CurrentCount())
	}

	spawn.IncreaseCount()
	spawn.IncreaseCount()
	if spawn.CurrentCount() != 3 {
		t.Errorf("after 3x IncreaseCount() CurrentCount() = %d, want 3", spawn.CurrentCount())
	}

	spawn.DecreaseCount()
	if spawn.CurrentCount() != 2 {
		t.Errorf("after DecreaseCount() CurrentCount() = %d, want 2", spawn.CurrentCount())
	}
}

func TestSpawn_NPCList(t *testing.T) {
	spawn := NewSpawn(1, 1000, 0, 0, 0, 0, 3, true)
	template := NewNpcTemplate(1000, "Wolf", "", 1, 1000, 500, 100, 50, 80, 40, 0, 120, 253, 30, 60)

	npc1 := NewNpc(100, 1000, template)
	npc2 := NewNpc(101, 1000, template)

	spawn.AddNpc(npc1)
	spawn.AddNpc(npc2)

	npcs := spawn.NPCs()
	if len(npcs) != 2 {
		t.Errorf("NPCs() length = %d, want 2", len(npcs))
	}

	spawn.RemoveNpc(npc1)
	npcs = spawn.NPCs()
	if len(npcs) != 1 {
		t.Errorf("after RemoveNpc() NPCs() length = %d, want 1", len(npcs))
	}
	if npcs[0] != npc2 {
		t.Error("remaining NPC should be npc2")
	}
}
