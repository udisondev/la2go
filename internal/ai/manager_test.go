package ai

import (
	"context"
	"testing"
	"time"

	"github.com/udisondev/la2go/internal/model"
)

func TestTickManager_RegisterUnregister(t *testing.T) {
	mgr := NewTickManager()

	template := model.NewNpcTemplate(1000, "Wolf", "", 1, 1000, 500, 0, 0, 0, 0, 0, 80, 253, 30, 60)
	npc := model.NewNpc(1, 1000, template)
	ai := NewBasicNpcAI(npc)

	// Register AI
	mgr.Register(1, ai)

	// Verify count
	if mgr.Count() != 1 {
		t.Errorf("Count() after Register() = %d, want 1", mgr.Count())
	}

	// Verify controller can be retrieved
	controller, err := mgr.GetController(1)
	if err != nil {
		t.Fatalf("GetController() error = %v", err)
	}

	if controller.CurrentIntention() != model.IntentionActive {
		t.Errorf("controller.CurrentIntention() = %v, want ACTIVE", controller.CurrentIntention())
	}

	// Unregister AI
	mgr.Unregister(1)

	// Verify count
	if mgr.Count() != 0 {
		t.Errorf("Count() after Unregister() = %d, want 0", mgr.Count())
	}

	// Verify controller is removed
	_, err = mgr.GetController(1)
	if err == nil {
		t.Error("GetController() after Unregister() should return error")
	}
}

func TestTickManager_Start(t *testing.T) {
	mgr := NewTickManager()

	template := model.NewNpcTemplate(1000, "Wolf", "", 1, 1000, 500, 0, 0, 0, 0, 0, 80, 253, 30, 60)
	npc := model.NewNpc(1, 1000, template)
	ai := NewBasicNpcAI(npc)

	// Register AI
	mgr.Register(1, ai)

	// Start manager with timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Start manager in goroutine
	done := make(chan error, 1)
	go func() {
		done <- mgr.Start(ctx)
	}()

	// Wait for at least 1 tick
	time.Sleep(1100 * time.Millisecond)

	// Cancel context to stop manager
	cancel()

	// Wait for manager to stop
	select {
	case err := <-done:
		if err != context.Canceled && err != context.DeadlineExceeded {
			t.Errorf("Start() error = %v, want context.Canceled or DeadlineExceeded", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("Start() did not stop after context cancel")
	}
}

func TestTickManager_MultipleControllers(t *testing.T) {
	mgr := NewTickManager()

	template := model.NewNpcTemplate(1000, "Wolf", "", 1, 1000, 500, 0, 0, 0, 0, 0, 80, 253, 30, 60)

	// Register 10 NPCs
	for i := range 10 {
		npc := model.NewNpc(uint32(i+1), 1000, template)
		ai := NewBasicNpcAI(npc)
		mgr.Register(uint32(i+1), ai)
	}

	// Verify count
	if mgr.Count() != 10 {
		t.Errorf("Count() after registering 10 controllers = %d, want 10", mgr.Count())
	}

	// Tick all
	mgr.tickAll()

	// Unregister all
	for i := range 10 {
		mgr.Unregister(uint32(i + 1))
	}

	// Verify count
	if mgr.Count() != 0 {
		t.Errorf("Count() after unregistering all = %d, want 0", mgr.Count())
	}
}
