package raid

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/udisondev/la2go/internal/model"
)

// mockGrandBossStore implements GrandBossStore for testing.
type mockGrandBossStore struct {
	mu   sync.Mutex
	rows map[int32]GrandBossDataRow
}

func newMockGrandBossStore() *mockGrandBossStore {
	return &mockGrandBossStore{rows: make(map[int32]GrandBossDataRow)}
}

func (s *mockGrandBossStore) LoadAllGrandBosses(_ context.Context) ([]GrandBossDataRow, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := make([]GrandBossDataRow, 0, len(s.rows))
	for _, row := range s.rows {
		result = append(result, row)
	}
	return result, nil
}

func (s *mockGrandBossStore) SaveGrandBoss(_ context.Context, row GrandBossDataRow) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.rows[row.BossID] = row
	return nil
}

func (s *mockGrandBossStore) GetGrandBoss(_ context.Context, bossID int32) (*GrandBossDataRow, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	row, ok := s.rows[bossID]
	if !ok {
		return nil, nil
	}
	return &row, nil
}

func (s *mockGrandBossStore) getRow(bossID int32) (GrandBossDataRow, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	row, ok := s.rows[bossID]
	return row, ok
}

// helper to create a test NpcTemplate for grand boss
func testGBTemplate(id int32) *model.NpcTemplate {
	return model.NewNpcTemplate(id, "TestBoss", "", 79,
		200000, 50000, 2000, 1500, 1800, 1200, 2000, 80, 250, 0, 0, 500000, 50000)
}

func TestGrandBossManager_Init_SpawnsAlive(t *testing.T) {
	t.Parallel()

	store := newMockGrandBossStore()
	store.rows[29001] = GrandBossDataRow{
		BossID:    29001,
		Status:    int16(model.GrandBossAlive),
		CurrentHP: 200000,
		CurrentMP: 50000,
	}

	var spawned []int32
	var mu sync.Mutex
	spawnFn := func(bossID int32) (*model.GrandBoss, error) {
		mu.Lock()
		spawned = append(spawned, bossID)
		mu.Unlock()
		tmpl := testGBTemplate(bossID)
		return model.NewGrandBoss(uint32(bossID+100000), bossID, tmpl, bossID), nil
	}

	mgr := NewGrandBossManager(store, spawnFn)
	if err := mgr.Init(context.Background()); err != nil {
		t.Fatalf("Init: %v", err)
	}

	mu.Lock()
	if len(spawned) != 1 {
		t.Errorf("spawned = %d; want 1", len(spawned))
	}
	mu.Unlock()

	if mgr.GetStatus(29001) != model.GrandBossAlive {
		t.Errorf("GetStatus(29001) = %d; want ALIVE", mgr.GetStatus(29001))
	}

	boss := mgr.GetBoss(29001)
	if boss == nil {
		t.Error("GetBoss(29001) = nil; want non-nil")
	}
}

func TestGrandBossManager_Init_RespawnsExpired(t *testing.T) {
	t.Parallel()

	store := newMockGrandBossStore()
	store.rows[29002] = GrandBossDataRow{
		BossID:      29002,
		Status:      int16(model.GrandBossDead),
		RespawnTime: time.Now().Add(-1 * time.Hour).Unix(), // past
	}

	spawnCalled := false
	spawnFn := func(bossID int32) (*model.GrandBoss, error) {
		spawnCalled = true
		tmpl := testGBTemplate(bossID)
		return model.NewGrandBoss(uint32(bossID+100000), bossID, tmpl, bossID), nil
	}

	mgr := NewGrandBossManager(store, spawnFn)
	if err := mgr.Init(context.Background()); err != nil {
		t.Fatalf("Init: %v", err)
	}

	if !spawnCalled {
		t.Error("spawnFn not called for expired dead boss")
	}

	if mgr.GetStatus(29002) != model.GrandBossAlive {
		t.Errorf("expired dead boss status = %d; want ALIVE", mgr.GetStatus(29002))
	}
}

func TestGrandBossManager_Init_KeepsDeadFuture(t *testing.T) {
	t.Parallel()

	store := newMockGrandBossStore()
	store.rows[29003] = GrandBossDataRow{
		BossID:      29003,
		Status:      int16(model.GrandBossDead),
		RespawnTime: time.Now().Add(2 * time.Hour).Unix(), // future
	}

	spawnCalled := false
	spawnFn := func(_ int32) (*model.GrandBoss, error) {
		spawnCalled = true
		return nil, nil
	}

	mgr := NewGrandBossManager(store, spawnFn)
	if err := mgr.Init(context.Background()); err != nil {
		t.Fatalf("Init: %v", err)
	}

	if spawnCalled {
		t.Error("spawnFn was called for future dead boss; want not called")
	}

	if mgr.GetStatus(29003) != model.GrandBossDead {
		t.Errorf("future dead boss status = %d; want DEAD", mgr.GetStatus(29003))
	}
}

func TestGrandBossManager_OnDeath(t *testing.T) {
	t.Parallel()

	store := newMockGrandBossStore()
	mgr := NewGrandBossManager(store, nil)

	// Manually add alive boss
	mgr.OnGrandBossSpawned(29001, model.NewGrandBoss(100001, 29001, testGBTemplate(29001), 29001))

	if mgr.GetStatus(29001) != model.GrandBossAlive {
		t.Fatal("boss should be alive before death")
	}

	// Kill boss (48h respawn)
	if err := mgr.OnGrandBossDeath(context.Background(), 29001, 172800); err != nil {
		t.Fatalf("OnGrandBossDeath: %v", err)
	}

	if mgr.GetStatus(29001) != model.GrandBossDead {
		t.Errorf("status after death = %d; want DEAD", mgr.GetStatus(29001))
	}

	if mgr.GetBoss(29001) != nil {
		t.Error("GetBoss after death = non-nil; want nil")
	}

	// Check DB
	row, ok := store.getRow(29001)
	if !ok {
		t.Fatal("boss not saved to store after death")
	}
	if row.Status != int16(model.GrandBossDead) {
		t.Errorf("DB status = %d; want DEAD(%d)", row.Status, model.GrandBossDead)
	}
	if row.RespawnTime <= 0 {
		t.Error("DB RespawnTime <= 0; want >0")
	}
}

func TestGrandBossManager_SetStatus(t *testing.T) {
	t.Parallel()

	store := newMockGrandBossStore()
	mgr := NewGrandBossManager(store, nil)

	if err := mgr.SetStatus(context.Background(), 29001, model.GrandBossFighting); err != nil {
		t.Fatalf("SetStatus: %v", err)
	}

	if mgr.GetStatus(29001) != model.GrandBossFighting {
		t.Errorf("GetStatus = %d; want FIGHTING", mgr.GetStatus(29001))
	}

	// Check DB
	row, ok := store.getRow(29001)
	if !ok {
		t.Fatal("boss not saved to store")
	}
	if row.Status != int16(model.GrandBossFighting) {
		t.Errorf("DB status = %d; want FIGHTING(%d)", row.Status, model.GrandBossFighting)
	}
}

func TestGrandBossManager_CheckPendingRespawns(t *testing.T) {
	t.Parallel()

	store := newMockGrandBossStore()
	mgr := NewGrandBossManager(store, nil)

	// Dead boss with past respawn
	mgr.mu.Lock()
	mgr.entries[29001] = &grandBossEntry{
		BossID:      29001,
		Status:      model.GrandBossDead,
		RespawnTime: time.Now().Add(-5 * time.Minute),
	}
	// Dead boss with future respawn
	mgr.entries[29002] = &grandBossEntry{
		BossID:      29002,
		Status:      model.GrandBossDead,
		RespawnTime: time.Now().Add(1 * time.Hour),
	}
	// Alive boss
	mgr.entries[29003] = &grandBossEntry{
		BossID: 29003,
		Status: model.GrandBossAlive,
	}
	mgr.mu.Unlock()

	ready := mgr.CheckPendingRespawns()
	if len(ready) != 1 {
		t.Fatalf("CheckPendingRespawns returned %d; want 1", len(ready))
	}
	if ready[0] != 29001 {
		t.Errorf("ready[0] = %d; want 29001", ready[0])
	}
}

func TestGrandBossManager_EntryCount(t *testing.T) {
	t.Parallel()

	store := newMockGrandBossStore()
	mgr := NewGrandBossManager(store, nil)

	if mgr.EntryCount() != 0 {
		t.Errorf("EntryCount() = %d; want 0", mgr.EntryCount())
	}

	mgr.OnGrandBossSpawned(29001, model.NewGrandBoss(100001, 29001, testGBTemplate(29001), 29001))
	if mgr.EntryCount() != 1 {
		t.Errorf("EntryCount() = %d; want 1", mgr.EntryCount())
	}

	mgr.OnGrandBossSpawned(29002, model.NewGrandBoss(100002, 29002, testGBTemplate(29002), 29002))
	if mgr.EntryCount() != 2 {
		t.Errorf("EntryCount() = %d; want 2", mgr.EntryCount())
	}
}

func TestGrandBossManager_GetBoss_Nil(t *testing.T) {
	t.Parallel()

	store := newMockGrandBossStore()
	mgr := NewGrandBossManager(store, nil)

	// Unknown boss
	if boss := mgr.GetBoss(99999); boss != nil {
		t.Errorf("GetBoss(99999) = %v; want nil", boss)
	}
}

func TestGrandBossManager_SaveEntry_WithLiveBoss(t *testing.T) {
	t.Parallel()

	store := newMockGrandBossStore()
	mgr := NewGrandBossManager(store, nil)

	tmpl := testGBTemplate(29001)
	boss := model.NewGrandBoss(100001, 29001, tmpl, 29001)
	boss.SetLocation(model.NewLocation(10000, 20000, -3000, 0))
	boss.SetCurrentHP(150000)
	boss.SetCurrentMP(40000)

	mgr.OnGrandBossSpawned(29001, boss)

	// Trigger save via SetStatus (which calls saveEntry)
	if err := mgr.SetStatus(context.Background(), 29001, model.GrandBossFighting); err != nil {
		t.Fatalf("SetStatus: %v", err)
	}

	row, ok := store.getRow(29001)
	if !ok {
		t.Fatal("boss not saved to store after SetStatus")
	}

	// Should have location from live boss
	if row.LocX != 10000 {
		t.Errorf("LocX = %d; want 10000", row.LocX)
	}
	if row.LocY != 20000 {
		t.Errorf("LocY = %d; want 20000", row.LocY)
	}
	if row.CurrentHP != 150000 {
		t.Errorf("CurrentHP = %f; want 150000", row.CurrentHP)
	}
}

func TestGrandBossManager_RunRespawnLoop_Cancellation(t *testing.T) {
	t.Parallel()

	store := newMockGrandBossStore()

	spawnCalled := false
	spawnFn := func(bossID int32) (*model.GrandBoss, error) {
		spawnCalled = true
		tmpl := testGBTemplate(bossID)
		return model.NewGrandBoss(uint32(bossID+100000), bossID, tmpl, bossID), nil
	}

	mgr := NewGrandBossManager(store, spawnFn)

	// Add a boss ready to respawn
	mgr.mu.Lock()
	mgr.entries[29001] = &grandBossEntry{
		BossID:      29001,
		Status:      model.GrandBossDead,
		RespawnTime: time.Now().Add(-1 * time.Minute),
	}
	mgr.mu.Unlock()

	// Cancel immediately — loop should exit
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := mgr.RunRespawnLoop(ctx)
	if err == nil {
		t.Error("RunRespawnLoop returned nil error; want context.Canceled")
	}

	// spawnCalled may or may not be true depending on timing
	_ = spawnCalled
}

func TestGrandBossManager_RunSaveLoop_Cancellation(t *testing.T) {
	t.Parallel()

	store := newMockGrandBossStore()
	mgr := NewGrandBossManager(store, nil)

	mgr.OnGrandBossSpawned(29001, model.NewGrandBoss(100001, 29001, testGBTemplate(29001), 29001))

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := mgr.RunSaveLoop(ctx)
	if err == nil {
		t.Error("RunSaveLoop returned nil error; want context.Canceled")
	}

	// On shutdown, saveAll should be called — verify data in store
	_, ok := store.getRow(29001)
	if !ok {
		t.Error("boss not saved on shutdown")
	}
}

func TestGrandBossManager_GetStatus_Default(t *testing.T) {
	t.Parallel()

	store := newMockGrandBossStore()
	mgr := NewGrandBossManager(store, nil)

	// Unknown boss — default to ALIVE
	if status := mgr.GetStatus(99999); status != model.GrandBossAlive {
		t.Errorf("GetStatus(unknown) = %d; want ALIVE(%d)", status, model.GrandBossAlive)
	}
}
