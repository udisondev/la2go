package raid

import (
	"context"
	"sync"
	"testing"
	"time"
)

// mockRaidSpawnStore implements RaidSpawnStore for testing.
type mockRaidSpawnStore struct {
	mu   sync.Mutex
	rows map[int32]RaidSpawnRow
}

func newMockRaidSpawnStore() *mockRaidSpawnStore {
	return &mockRaidSpawnStore{rows: make(map[int32]RaidSpawnRow)}
}

func (s *mockRaidSpawnStore) LoadAllRaidSpawns(_ context.Context) ([]RaidSpawnRow, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := make([]RaidSpawnRow, 0, len(s.rows))
	for _, row := range s.rows {
		result = append(result, row)
	}
	return result, nil
}

func (s *mockRaidSpawnStore) SaveRaidSpawn(_ context.Context, row RaidSpawnRow) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.rows[row.BossID] = row
	return nil
}

func (s *mockRaidSpawnStore) DeleteRaidSpawn(_ context.Context, bossID int32) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.rows, bossID)
	return nil
}

func (s *mockRaidSpawnStore) getRow(bossID int32) (RaidSpawnRow, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	row, ok := s.rows[bossID]
	return row, ok
}

func TestSpawnManager_Init_SpawnsAliveBosse(t *testing.T) {
	t.Parallel()

	store := newMockRaidSpawnStore()
	store.rows[25001] = RaidSpawnRow{
		BossID:      25001,
		RespawnTime: 0, // no respawn time → spawn immediately
		CurrentHP:   50000,
		CurrentMP:   10000,
	}
	store.rows[25002] = RaidSpawnRow{
		BossID:      25002,
		RespawnTime: time.Now().Add(-1 * time.Hour).Unix(), // past → spawn immediately
		CurrentHP:   30000,
		CurrentMP:   5000,
	}

	var spawned []int32
	var mu sync.Mutex
	respawnFn := func(bossID int32) error {
		mu.Lock()
		spawned = append(spawned, bossID)
		mu.Unlock()
		return nil
	}

	mgr := NewSpawnManager(store, respawnFn)

	if err := mgr.Init(context.Background()); err != nil {
		t.Fatalf("Init: %v", err)
	}

	mu.Lock()
	if len(spawned) != 2 {
		t.Errorf("spawned count = %d; want 2", len(spawned))
	}
	mu.Unlock()

	if mgr.EntryCount() != 2 {
		t.Errorf("EntryCount() = %d; want 2", mgr.EntryCount())
	}
}

func TestSpawnManager_Init_SchedulesFutureRespawn(t *testing.T) {
	t.Parallel()

	store := newMockRaidSpawnStore()
	store.rows[25003] = RaidSpawnRow{
		BossID:      25003,
		RespawnTime: time.Now().Add(1 * time.Hour).Unix(), // future → don't spawn yet
		CurrentHP:   40000,
		CurrentMP:   8000,
	}

	spawnCalled := false
	respawnFn := func(_ int32) error {
		spawnCalled = true
		return nil
	}

	mgr := NewSpawnManager(store, respawnFn)

	if err := mgr.Init(context.Background()); err != nil {
		t.Fatalf("Init: %v", err)
	}

	if spawnCalled {
		t.Error("respawnFn was called for future respawn; want not called")
	}

	if mgr.IsAlive(25003) {
		t.Error("boss 25003 IsAlive = true; want false (respawn in future)")
	}

	respawnTime := mgr.GetRespawnTime(25003)
	if respawnTime.IsZero() {
		t.Error("GetRespawnTime(25003) is zero; want non-zero")
	}
}

func TestSpawnManager_OnRaidBossDeath(t *testing.T) {
	t.Parallel()

	store := newMockRaidSpawnStore()
	mgr := NewSpawnManager(store, nil)
	mgr.SetRespawnDelays(3600, 7200) // 1-2h for fast test

	// Mark boss as alive first
	mgr.OnRaidBossSpawned(25001)
	if !mgr.IsAlive(25001) {
		t.Fatal("boss should be alive after OnRaidBossSpawned")
	}

	// Kill boss
	if err := mgr.OnRaidBossDeath(context.Background(), 25001, 0, 0); err != nil {
		t.Fatalf("OnRaidBossDeath: %v", err)
	}

	if mgr.IsAlive(25001) {
		t.Error("boss IsAlive = true after death; want false")
	}

	// Check DB was updated
	row, ok := store.getRow(25001)
	if !ok {
		t.Fatal("raid spawn not saved to store")
	}
	if row.RespawnTime <= 0 {
		t.Error("RespawnTime = 0; want >0")
	}

	// Respawn time should be 1-2h in the future
	respawnTime := time.Unix(row.RespawnTime, 0)
	minRespawn := time.Now().Add(3599 * time.Second) // ~1h
	maxRespawn := time.Now().Add(7201 * time.Second) // ~2h
	if respawnTime.Before(minRespawn) || respawnTime.After(maxRespawn) {
		t.Errorf("respawn time %v outside expected range [%v, %v]",
			respawnTime, minRespawn, maxRespawn)
	}
}

func TestSpawnManager_CheckPendingRespawns(t *testing.T) {
	t.Parallel()

	store := newMockRaidSpawnStore()
	mgr := NewSpawnManager(store, nil)

	// Add two dead bosses: one past respawn, one future
	mgr.mu.Lock()
	mgr.entries[25001] = &raidSpawnEntry{
		BossID:      25001,
		RespawnTime: time.Now().Add(-10 * time.Minute), // past
		IsAlive:     false,
	}
	mgr.entries[25002] = &raidSpawnEntry{
		BossID:      25002,
		RespawnTime: time.Now().Add(1 * time.Hour), // future
		IsAlive:     false,
	}
	mgr.entries[25003] = &raidSpawnEntry{
		BossID:  25003,
		IsAlive: true, // already alive
	}
	mgr.mu.Unlock()

	ready := mgr.CheckPendingRespawns()
	if len(ready) != 1 {
		t.Fatalf("CheckPendingRespawns returned %d; want 1", len(ready))
	}
	if ready[0] != 25001 {
		t.Errorf("ready[0] = %d; want 25001", ready[0])
	}
}

func TestSpawnManager_OnRaidBossSpawned(t *testing.T) {
	t.Parallel()

	store := newMockRaidSpawnStore()
	mgr := NewSpawnManager(store, nil)

	if mgr.IsAlive(25001) {
		t.Error("boss should not be alive before spawned")
	}

	mgr.OnRaidBossSpawned(25001)
	if !mgr.IsAlive(25001) {
		t.Error("boss should be alive after OnRaidBossSpawned")
	}
}

func TestSpawnManager_CalculateRespawnDelay(t *testing.T) {
	t.Parallel()

	store := newMockRaidSpawnStore()
	mgr := NewSpawnManager(store, nil)
	mgr.SetRespawnDelays(100, 200)

	// Run multiple times to ensure range
	for range 100 {
		delay := mgr.calculateRespawnDelay()
		if delay < 100 || delay > 200 {
			t.Fatalf("calculateRespawnDelay() = %d; want [100, 200]", delay)
		}
	}
}

func TestSpawnManager_CalculateRespawnDelay_EqualMinMax(t *testing.T) {
	t.Parallel()

	store := newMockRaidSpawnStore()
	mgr := NewSpawnManager(store, nil)
	mgr.SetRespawnDelays(300, 300)

	delay := mgr.calculateRespawnDelay()
	if delay != 300 {
		t.Errorf("calculateRespawnDelay() = %d; want 300 (min=max)", delay)
	}
}

func TestSpawnManager_RunRespawnLoop_Cancellation(t *testing.T) {
	t.Parallel()

	store := newMockRaidSpawnStore()
	spawnCalled := false
	mgr := NewSpawnManager(store, func(_ int32) error {
		spawnCalled = true
		return nil
	})

	// Add a dead boss ready to respawn
	mgr.mu.Lock()
	mgr.entries[25001] = &raidSpawnEntry{
		BossID:      25001,
		RespawnTime: time.Now().Add(-1 * time.Minute),
		IsAlive:     false,
	}
	mgr.mu.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := mgr.RunRespawnLoop(ctx)
	if err == nil {
		t.Error("RunRespawnLoop returned nil; want context.Canceled")
	}

	_ = spawnCalled
}

func TestSpawnManager_GetRespawnTime_AliveReturnsZero(t *testing.T) {
	t.Parallel()

	store := newMockRaidSpawnStore()
	mgr := NewSpawnManager(store, nil)

	mgr.OnRaidBossSpawned(25001)

	rt := mgr.GetRespawnTime(25001)
	if !rt.IsZero() {
		t.Errorf("GetRespawnTime for alive boss = %v; want zero", rt)
	}
}

func TestSpawnManager_GetRespawnTime_Unknown(t *testing.T) {
	t.Parallel()

	store := newMockRaidSpawnStore()
	mgr := NewSpawnManager(store, nil)

	rt := mgr.GetRespawnTime(99999)
	if !rt.IsZero() {
		t.Errorf("GetRespawnTime for unknown boss = %v; want zero", rt)
	}
}

func TestSpawnManager_IsAlive_Unknown(t *testing.T) {
	t.Parallel()

	store := newMockRaidSpawnStore()
	mgr := NewSpawnManager(store, nil)

	if mgr.IsAlive(99999) {
		t.Error("IsAlive(unknown) = true; want false")
	}
}

func TestSpawnManager_EntryCount(t *testing.T) {
	t.Parallel()

	store := newMockRaidSpawnStore()
	mgr := NewSpawnManager(store, nil)

	if mgr.EntryCount() != 0 {
		t.Errorf("EntryCount() = %d; want 0", mgr.EntryCount())
	}

	mgr.OnRaidBossSpawned(25001)
	mgr.OnRaidBossSpawned(25002)

	if mgr.EntryCount() != 2 {
		t.Errorf("EntryCount() = %d; want 2", mgr.EntryCount())
	}
}
