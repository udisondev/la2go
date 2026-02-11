package ai

import (
	"testing"

	"github.com/udisondev/la2go/internal/model"
)

func TestBasicNpcAI_StartStop(t *testing.T) {
	template := model.NewNpcTemplate(1000, "Wolf", "", 1, 1000, 500, 0, 0, 0, 0, 0, 80, 253, 30, 60, 0, 0)
	npc := model.NewNpc(1, 1000, template)

	ai := NewBasicNpcAI(npc)

	// Initial state should be IDLE (from Npc constructor)
	if ai.CurrentIntention() != model.IntentionIdle {
		t.Errorf("initial CurrentIntention() = %v, want IDLE", ai.CurrentIntention())
	}

	// Start AI
	ai.Start()

	// After start, intention should be ACTIVE
	if ai.CurrentIntention() != model.IntentionActive {
		t.Errorf("after Start() CurrentIntention() = %v, want ACTIVE", ai.CurrentIntention())
	}

	// Stop AI
	ai.Stop()

	// After stop, intention should be IDLE
	if ai.CurrentIntention() != model.IntentionIdle {
		t.Errorf("after Stop() CurrentIntention() = %v, want IDLE", ai.CurrentIntention())
	}
}

func TestBasicNpcAI_Tick(t *testing.T) {
	template := model.NewNpcTemplate(1000, "Wolf", "", 1, 1000, 500, 0, 0, 0, 0, 0, 80, 253, 30, 60, 0, 0)
	npc := model.NewNpc(1, 1000, template)

	ai := NewBasicNpcAI(npc)
	ai.Start()

	// Initial state: ACTIVE
	if ai.CurrentIntention() != model.IntentionActive {
		t.Fatalf("initial intention = %v, want ACTIVE", ai.CurrentIntention())
	}

	// Tick 4 times (should not change intention yet)
	for range 4 {
		ai.Tick()
	}

	if ai.CurrentIntention() != model.IntentionActive {
		t.Errorf("after 4 ticks intention = %v, want ACTIVE", ai.CurrentIntention())
	}

	// Tick 5th time (should change to IDLE)
	ai.Tick()

	if ai.CurrentIntention() != model.IntentionIdle {
		t.Errorf("after 5 ticks intention = %v, want IDLE", ai.CurrentIntention())
	}

	// Tick another 5 times (should change back to ACTIVE)
	for range 5 {
		ai.Tick()
	}

	if ai.CurrentIntention() != model.IntentionActive {
		t.Errorf("after 10 ticks intention = %v, want ACTIVE", ai.CurrentIntention())
	}
}

func TestBasicNpcAI_Tick_WhenStopped(t *testing.T) {
	template := model.NewNpcTemplate(1000, "Wolf", "", 1, 1000, 500, 0, 0, 0, 0, 0, 80, 253, 30, 60, 0, 0)
	npc := model.NewNpc(1, 1000, template)

	ai := NewBasicNpcAI(npc)
	// Don't start AI

	// Tick should do nothing when AI is not running
	for range 10 {
		ai.Tick()
	}

	// Intention should remain IDLE
	if ai.CurrentIntention() != model.IntentionIdle {
		t.Errorf("intention after ticks without Start() = %v, want IDLE", ai.CurrentIntention())
	}
}
