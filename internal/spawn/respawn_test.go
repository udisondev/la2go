package spawn

import (
	"context"
	"testing"
	"time"

	"github.com/udisondev/la2go/internal/ai"
	"github.com/udisondev/la2go/internal/model"
	"github.com/udisondev/la2go/internal/world"
)

func TestRespawnTaskManager_ScheduleRespawn(t *testing.T) {
	// Setup
	npcRepo := newMockNpcRepository()
	spawnRepo := newMockSpawnRepository()
	w := world.Instance()
	aiMgr := ai.NewTickManager()

	spawnMgr := NewManager(
		npcRepo,
		spawnRepo,
		w,
		aiMgr,
		nil,
	)

	respawnMgr := NewRespawnTaskManager(spawnMgr)

	// Create test spawn
	spawn := model.NewSpawn(100, 1000, 17000, 170000, -3500, 0, 1, true)

	// Schedule respawn
	respawnMgr.ScheduleRespawn(spawn, 5) // 5 seconds

	// Verify task is scheduled
	if respawnMgr.TaskCount() != 1 {
		t.Errorf("TaskCount() after schedule = %d, want 1", respawnMgr.TaskCount())
	}

	// Get task
	task, ok := respawnMgr.GetTask(100)
	if !ok {
		t.Fatal("GetTask() returned false after schedule")
	}

	if task.Spawn != spawn {
		t.Error("task.Spawn != expected spawn")
	}

	// Verify respawn time is ~5 seconds from now
	expectedTime := time.Now().Add(5 * time.Second)
	diff := task.RespawnTime.Sub(expectedTime).Abs()
	if diff > 100*time.Millisecond {
		t.Errorf("task.RespawnTime diff = %v, want < 100ms", diff)
	}
}

func TestRespawnTaskManager_CancelRespawn(t *testing.T) {
	spawnMgr := NewManager(nil, nil, world.Instance(), ai.NewTickManager(), nil)
	respawnMgr := NewRespawnTaskManager(spawnMgr)

	spawn := model.NewSpawn(101, 1000, 0, 0, 0, 0, 1, true)

	// Schedule respawn
	respawnMgr.ScheduleRespawn(spawn, 10)

	if respawnMgr.TaskCount() != 1 {
		t.Fatalf("TaskCount() after schedule = %d, want 1", respawnMgr.TaskCount())
	}

	// Cancel respawn
	respawnMgr.CancelRespawn(101)

	// Verify task is removed
	if respawnMgr.TaskCount() != 0 {
		t.Errorf("TaskCount() after cancel = %d, want 0", respawnMgr.TaskCount())
	}

	_, ok := respawnMgr.GetTask(101)
	if ok {
		t.Error("GetTask() returned true after cancel")
	}
}

func TestRespawnTaskManager_Start(t *testing.T) {
	// Setup
	npcRepo := newMockNpcRepository()
	spawnRepo := newMockSpawnRepository()
	w := world.Instance()
	aiMgr := ai.NewTickManager()

	spawnMgr := NewManager(
		npcRepo,
		spawnRepo,
		w,
		aiMgr,
		nil,
	)

	// Add test template
	template := model.NewNpcTemplate(
		2000, "RespawnTest", "", 1, 1000, 500,
		0, 0, 0, 0, 0, 80, 253, 1, 1, 0, 0,
	)
	npcRepo.AddTemplate(template)

	respawnMgr := NewRespawnTaskManager(spawnMgr)

	// Create test spawn
	spawn := model.NewSpawn(200, 2000, 17000, 170000, -3500, 0, 1, true)

	// Schedule very short respawn (1 second)
	respawnMgr.ScheduleRespawn(spawn, 1)

	// Start manager with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- respawnMgr.Start(ctx)
	}()

	// Wait for respawn to trigger
	time.Sleep(1500 * time.Millisecond)

	// Task should be executed and removed
	if respawnMgr.TaskCount() != 0 {
		t.Errorf("TaskCount() after respawn = %d, want 0", respawnMgr.TaskCount())
	}

	// Stop manager
	cancel()

	select {
	case <-done:
		// OK
	case <-time.After(2 * time.Second):
		t.Fatal("respawn manager did not stop")
	}
}

func TestRespawnTaskManager_MultipleTasksprocessing(t *testing.T) {
	npcRepo := newMockNpcRepository()
	spawnRepo := newMockSpawnRepository()
	w := world.Instance()
	aiMgr := ai.NewTickManager()

	spawnMgr := NewManager(
		npcRepo,
		spawnRepo,
		w,
		aiMgr,
		nil,
	)

	template := model.NewNpcTemplate(
		2001, "MultiTest", "", 1, 1000, 500,
		0, 0, 0, 0, 0, 80, 253, 1, 1, 0, 0,
	)
	npcRepo.AddTemplate(template)

	respawnMgr := NewRespawnTaskManager(spawnMgr)

	// Schedule 5 respawns
	for i := range 5 {
		spawn := model.NewSpawn(int64(300+i), 2001, 17000+int32(i)*100, 170000, -3500, 0, 1, true)
		respawnMgr.ScheduleRespawn(spawn, 1)
	}

	if respawnMgr.TaskCount() != 5 {
		t.Fatalf("TaskCount() after scheduling 5 = %d, want 5", respawnMgr.TaskCount())
	}

	// Start manager briefly
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	go respawnMgr.Start(ctx)

	// Wait for tasks to execute
	time.Sleep(1500 * time.Millisecond)

	// All tasks should be executed
	if respawnMgr.TaskCount() != 0 {
		t.Errorf("TaskCount() after execution = %d, want 0", respawnMgr.TaskCount())
	}

	cancel()
}
