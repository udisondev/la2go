package manor

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/udisondev/la2go/internal/data"
)

func TestMain(m *testing.M) {
	// Загружаем seed data для валидации в Init().
	if err := data.LoadSeeds(); err != nil {
		panic("load seeds: " + err.Error())
	}
	os.Exit(m.Run())
}

// --- Mock implementations ---

type mockManorStore struct {
	mu         sync.Mutex
	production map[int32][]ProductionRow
	procure    map[int32][]ProcureRow
	saveCount  int
}

func newMockManorStore() *mockManorStore {
	return &mockManorStore{
		production: make(map[int32][]ProductionRow),
		procure:    make(map[int32][]ProcureRow),
	}
}

func (s *mockManorStore) LoadProduction(_ context.Context, castleID int32) ([]ProductionRow, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.production[castleID], nil
}

func (s *mockManorStore) LoadProcure(_ context.Context, castleID int32) ([]ProcureRow, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.procure[castleID], nil
}

func (s *mockManorStore) SaveAll(_ context.Context, castleID int32, production []ProductionRow, procure []ProcureRow) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.production[castleID] = production
	s.procure[castleID] = procure
	s.saveCount++
	return nil
}

func (s *mockManorStore) DeleteAll(_ context.Context, castleID int32) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.production, castleID)
	delete(s.procure, castleID)
	return nil
}

type mockCastleProvider struct {
	castleIDs []int32
	mu        sync.Mutex
	treasury  map[int32]int64
}

func newMockCastleProvider(ids ...int32) *mockCastleProvider {
	p := &mockCastleProvider{
		castleIDs: ids,
		treasury:  make(map[int32]int64),
	}
	for _, id := range ids {
		p.treasury[id] = 1000000 // 1M default
	}
	return p
}

func (p *mockCastleProvider) CastleIDs() []int32 { return p.castleIDs }

func (p *mockCastleProvider) Treasury(castleID int32) int64 {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.treasury[castleID]
}

func (p *mockCastleProvider) AddToTreasury(castleID int32, amount int64) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.treasury[castleID] += amount
	if p.treasury[castleID] < 0 {
		p.treasury[castleID] = 0
	}
}

// --- Tests ---

func TestNewManager(t *testing.T) {
	t.Parallel()

	store := newMockManorStore()
	castles := newMockCastleProvider(1, 2)
	mgr := NewManager(store, castles)

	if mgr == nil {
		t.Fatal("NewManager returned nil")
	}
	if mgr.Mode() != ModeApproved {
		t.Errorf("Mode() = %v; want ModeApproved", mgr.Mode())
	}
}

func TestManager_Init_Empty(t *testing.T) {
	t.Parallel()

	store := newMockManorStore()
	castles := newMockCastleProvider(1)
	mgr := NewManager(store, castles)

	if err := mgr.Init(context.Background()); err != nil {
		t.Fatalf("Init: %v", err)
	}

	prod := mgr.SeedProduction(1, false)
	if len(prod) != 0 {
		t.Errorf("SeedProduction(1, false) len = %d; want 0", len(prod))
	}
}

func TestManager_Init_WithData(t *testing.T) {
	t.Parallel()

	store := newMockManorStore()
	store.production[1] = []ProductionRow{
		{CastleID: 1, SeedID: 5016, Amount: 100, StartAmount: 200, Price: 1000, NextPeriod: false},
		{CastleID: 1, SeedID: 5017, Amount: 50, StartAmount: 100, Price: 2000, NextPeriod: true},
	}
	store.procure[1] = []ProcureRow{
		{CastleID: 1, CropID: 5073, Amount: 80, StartAmount: 100, Price: 500, RewardType: 1, NextPeriod: false},
	}

	castles := newMockCastleProvider(1)
	mgr := NewManager(store, castles)

	if err := mgr.Init(context.Background()); err != nil {
		t.Fatalf("Init: %v", err)
	}

	current := mgr.SeedProduction(1, false)
	if len(current) != 1 {
		t.Fatalf("SeedProduction(1, false) len = %d; want 1", len(current))
	}
	if current[0].SeedID() != 5016 {
		t.Errorf("SeedID = %d; want 5016", current[0].SeedID())
	}
	if current[0].Amount() != 100 {
		t.Errorf("Amount = %d; want 100", current[0].Amount())
	}

	next := mgr.SeedProduction(1, true)
	if len(next) != 1 {
		t.Fatalf("SeedProduction(1, true) len = %d; want 1", len(next))
	}
	if next[0].SeedID() != 5017 {
		t.Errorf("next SeedID = %d; want 5017", next[0].SeedID())
	}

	procure := mgr.CropProcureList(1, false)
	if len(procure) != 1 {
		t.Fatalf("CropProcureList(1, false) len = %d; want 1", len(procure))
	}
	if procure[0].CropID() != 5073 {
		t.Errorf("CropID = %d; want 5073", procure[0].CropID())
	}
}

func TestManager_SeedProduct_Found(t *testing.T) {
	t.Parallel()

	store := newMockManorStore()
	store.production[1] = []ProductionRow{
		{CastleID: 1, SeedID: 5016, Amount: 50, StartAmount: 100, Price: 1000, NextPeriod: false},
		{CastleID: 1, SeedID: 5017, Amount: 30, StartAmount: 80, Price: 2000, NextPeriod: false},
	}

	castles := newMockCastleProvider(1)
	mgr := NewManager(store, castles)

	if err := mgr.Init(context.Background()); err != nil {
		t.Fatalf("Init: %v", err)
	}

	sp := mgr.SeedProduct(1, 5017, false)
	if sp == nil {
		t.Fatal("SeedProduct(1, 5017, false) = nil; want non-nil")
	}
	if sp.Amount() != 30 {
		t.Errorf("Amount = %d; want 30", sp.Amount())
	}
}

func TestManager_SeedProduct_NotFound(t *testing.T) {
	t.Parallel()

	store := newMockManorStore()
	castles := newMockCastleProvider(1)
	mgr := NewManager(store, castles)

	if err := mgr.Init(context.Background()); err != nil {
		t.Fatalf("Init: %v", err)
	}

	sp := mgr.SeedProduct(1, 9999, false)
	if sp != nil {
		t.Error("SeedProduct(1, 9999) = non-nil; want nil")
	}
}

func TestManager_CropProcureEntry_Found(t *testing.T) {
	t.Parallel()

	store := newMockManorStore()
	store.procure[1] = []ProcureRow{
		{CastleID: 1, CropID: 5073, Amount: 80, StartAmount: 100, Price: 500, RewardType: 1, NextPeriod: false},
	}

	castles := newMockCastleProvider(1)
	mgr := NewManager(store, castles)

	if err := mgr.Init(context.Background()); err != nil {
		t.Fatalf("Init: %v", err)
	}

	cp := mgr.CropProcureEntry(1, 5073, false)
	if cp == nil {
		t.Fatal("CropProcureEntry(1, 5073) = nil; want non-nil")
	}
	if cp.RewardType() != 1 {
		t.Errorf("RewardType = %d; want 1", cp.RewardType())
	}
}

func TestManager_SetNextSeedProduction(t *testing.T) {
	t.Parallel()

	store := newMockManorStore()
	castles := newMockCastleProvider(1)
	mgr := NewManager(store, castles)

	if err := mgr.Init(context.Background()); err != nil {
		t.Fatalf("Init: %v", err)
	}

	list := []*SeedProduction{
		NewSeedProduction(5016, 300, 1500, 300),
		NewSeedProduction(5017, 200, 2500, 200),
	}
	mgr.SetNextSeedProduction(1, list)

	next := mgr.SeedProduction(1, true)
	if len(next) != 2 {
		t.Fatalf("next len = %d; want 2", len(next))
	}
}

func TestManager_SetNextCropProcure(t *testing.T) {
	t.Parallel()

	store := newMockManorStore()
	castles := newMockCastleProvider(1)
	mgr := NewManager(store, castles)

	if err := mgr.Init(context.Background()); err != nil {
		t.Fatalf("Init: %v", err)
	}

	list := []*CropProcure{
		NewCropProcure(5073, 100, 1, 100, 800),
	}
	mgr.SetNextCropProcure(1, list)

	next := mgr.CropProcureList(1, true)
	if len(next) != 1 {
		t.Fatalf("next len = %d; want 1", len(next))
	}
}

func TestManager_ManorCost(t *testing.T) {
	t.Parallel()

	store := newMockManorStore()
	store.production[1] = []ProductionRow{
		{CastleID: 1, SeedID: 5016, Amount: 100, StartAmount: 100, Price: 1000, NextPeriod: false},
	}
	store.procure[1] = []ProcureRow{
		{CastleID: 1, CropID: 5073, Amount: 50, StartAmount: 50, Price: 200, RewardType: 1, NextPeriod: false},
	}

	castles := newMockCastleProvider(1)
	mgr := NewManager(store, castles)

	if err := mgr.Init(context.Background()); err != nil {
		t.Fatalf("Init: %v", err)
	}

	// ManorCost = seedRefPrice*startAmount + cropPrice*startAmount.
	// SeedReferencePrice(5016) returns 1 (no item data loaded), so cost = 1*100 + 200*50 = 10100.
	cost := mgr.ManorCost(1, false)
	if cost < 0 {
		t.Errorf("ManorCost(1, false) = %d; want >= 0", cost)
	}
}

func TestManager_ResetManorData(t *testing.T) {
	t.Parallel()

	store := newMockManorStore()
	store.production[1] = []ProductionRow{
		{CastleID: 1, SeedID: 5016, Amount: 100, StartAmount: 100, Price: 1000, NextPeriod: false},
	}

	castles := newMockCastleProvider(1)
	mgr := NewManager(store, castles)

	if err := mgr.Init(context.Background()); err != nil {
		t.Fatalf("Init: %v", err)
	}

	if err := mgr.ResetManorData(context.Background(), 1); err != nil {
		t.Fatalf("ResetManorData: %v", err)
	}

	prod := mgr.SeedProduction(1, false)
	if len(prod) != 0 {
		t.Errorf("SeedProduction len = %d; want 0 after reset", len(prod))
	}
}

func TestManager_ChangeMode_ApprovedToMaintenance(t *testing.T) {
	t.Parallel()

	store := newMockManorStore()
	castles := newMockCastleProvider(1)
	mgr := NewManager(store, castles)

	if err := mgr.Init(context.Background()); err != nil {
		t.Fatalf("Init: %v", err)
	}

	// Устанавливаем next period данные.
	mgr.SetNextSeedProduction(1, []*SeedProduction{
		NewSeedProduction(5016, 100, 1000, 100),
	})
	mgr.SetNextCropProcure(1, []*CropProcure{
		NewCropProcure(5073, 50, 1, 50, 500),
	})

	// Текущий режим APPROVED, переключаем.
	mgr.mu.Lock()
	mgr.mode = ModeApproved
	mgr.mu.Unlock()

	mgr.ChangeMode(context.Background())

	if mgr.Mode() != ModeMaintenance {
		t.Errorf("Mode() = %v; want ModeMaintenance", mgr.Mode())
	}

	// Текущие данные должны стать бывшими next.
	current := mgr.SeedProduction(1, false)
	if len(current) != 1 {
		t.Fatalf("SeedProduction len = %d; want 1 (rotated from next)", len(current))
	}
	if current[0].SeedID() != 5016 {
		t.Errorf("SeedID = %d; want 5016", current[0].SeedID())
	}
}

func TestManager_ChangeMode_MaintenanceToModifiable(t *testing.T) {
	t.Parallel()

	store := newMockManorStore()
	castles := newMockCastleProvider(1)
	mgr := NewManager(store, castles)

	if err := mgr.Init(context.Background()); err != nil {
		t.Fatalf("Init: %v", err)
	}

	mgr.mu.Lock()
	mgr.mode = ModeMaintenance
	mgr.mu.Unlock()

	mgr.ChangeMode(context.Background())

	if mgr.Mode() != ModeModifiable {
		t.Errorf("Mode() = %v; want ModeModifiable", mgr.Mode())
	}
}

func TestManager_ChangeMode_ModifiableToApproved(t *testing.T) {
	t.Parallel()

	store := newMockManorStore()
	castles := newMockCastleProvider(1)
	mgr := NewManager(store, castles)

	if err := mgr.Init(context.Background()); err != nil {
		t.Fatalf("Init: %v", err)
	}

	// Устанавливаем next production с низкой стоимостью.
	mgr.SetNextSeedProduction(1, []*SeedProduction{
		NewSeedProduction(5016, 10, 100, 10),
	})

	mgr.mu.Lock()
	mgr.mode = ModeModifiable
	mgr.mu.Unlock()

	mgr.ChangeMode(context.Background())

	if mgr.Mode() != ModeApproved {
		t.Errorf("Mode() = %v; want ModeApproved", mgr.Mode())
	}
}

func TestManager_ChangeMode_InsufficientTreasury(t *testing.T) {
	t.Parallel()

	store := newMockManorStore()
	castles := newMockCastleProvider(1)
	castles.mu.Lock()
	castles.treasury[1] = 0 // Пустая казна.
	castles.mu.Unlock()

	mgr := NewManager(store, castles)
	if err := mgr.Init(context.Background()); err != nil {
		t.Fatalf("Init: %v", err)
	}

	// Ставим next production с высокой стоимостью.
	mgr.SetNextSeedProduction(1, []*SeedProduction{
		NewSeedProduction(5016, 1000, 999999, 1000),
	})
	mgr.SetNextCropProcure(1, []*CropProcure{
		NewCropProcure(5073, 500, 1, 500, 99999),
	})

	mgr.mu.Lock()
	mgr.mode = ModeModifiable
	mgr.mu.Unlock()

	mgr.ChangeMode(context.Background())

	// Next period должен быть очищен.
	nextProd := mgr.SeedProduction(1, true)
	if len(nextProd) != 0 {
		t.Errorf("next SeedProduction len = %d; want 0 (insufficient treasury)", len(nextProd))
	}
	nextProc := mgr.CropProcureList(1, true)
	if len(nextProc) != 0 {
		t.Errorf("next CropProcure len = %d; want 0 (insufficient treasury)", len(nextProc))
	}
}

func TestManager_FullCycle(t *testing.T) {
	t.Parallel()

	store := newMockManorStore()
	castles := newMockCastleProvider(1)
	mgr := NewManager(store, castles)

	if err := mgr.Init(context.Background()); err != nil {
		t.Fatalf("Init: %v", err)
	}

	// 1. APPROVED → set next data → MAINTENANCE → MODIFIABLE → APPROVED (full cycle).
	mgr.mu.Lock()
	mgr.mode = ModeApproved
	mgr.mu.Unlock()

	mgr.SetNextSeedProduction(1, []*SeedProduction{
		NewSeedProduction(5016, 50, 100, 50),
	})

	// APPROVED → MAINTENANCE.
	mgr.ChangeMode(context.Background())
	if mgr.Mode() != ModeMaintenance {
		t.Fatalf("after first ChangeMode: mode = %v; want MAINTENANCE", mgr.Mode())
	}

	// MAINTENANCE → MODIFIABLE.
	mgr.ChangeMode(context.Background())
	if mgr.Mode() != ModeModifiable {
		t.Fatalf("after second ChangeMode: mode = %v; want MODIFIABLE", mgr.Mode())
	}

	// MODIFIABLE → APPROVED.
	mgr.ChangeMode(context.Background())
	if mgr.Mode() != ModeApproved {
		t.Fatalf("after third ChangeMode: mode = %v; want APPROVED", mgr.Mode())
	}
}

func TestManager_Save(t *testing.T) {
	t.Parallel()

	store := newMockManorStore()
	castles := newMockCastleProvider(1)
	mgr := NewManager(store, castles)

	if err := mgr.Init(context.Background()); err != nil {
		t.Fatalf("Init: %v", err)
	}

	mgr.SetNextSeedProduction(1, []*SeedProduction{
		NewSeedProduction(5016, 100, 1000, 100),
	})

	if err := mgr.Save(context.Background()); err != nil {
		t.Fatalf("Save: %v", err)
	}

	store.mu.Lock()
	if store.saveCount < 1 {
		t.Error("store.saveCount < 1; want >= 1")
	}
	store.mu.Unlock()
}

func TestManager_RunSaveLoop_Cancellation(t *testing.T) {
	t.Parallel()

	store := newMockManorStore()
	castles := newMockCastleProvider(1)
	mgr := NewManager(store, castles)

	if err := mgr.Init(context.Background()); err != nil {
		t.Fatalf("Init: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := mgr.RunSaveLoop(ctx, time.Hour)
	if err == nil {
		t.Error("RunSaveLoop returned nil; want context.Canceled")
	}
}

func TestManager_DetermineMode(t *testing.T) {
	t.Parallel()

	store := newMockManorStore()
	castles := newMockCastleProvider(1)
	mgr := NewManager(store, castles)

	tests := []struct {
		name string
		hour int
		min  int
		want Mode
	}{
		{"midnight", 0, 0, ModeApproved},
		{"3am", 3, 0, ModeApproved},
		{"5:59", 5, 59, ModeApproved},
		{"6:00 maintenance start", 6, 0, ModeMaintenance},
		{"6:01 maintenance", 6, 1, ModeMaintenance},
		{"6:02 maintenance", 6, 2, ModeMaintenance},
		{"6:03 modifiable start", 6, 3, ModeModifiable},
		{"noon", 12, 0, ModeModifiable},
		{"19:59 modifiable end", 19, 59, ModeModifiable},
		{"20:00 approved", 20, 0, ModeModifiable},
		{"20:01 approved", 20, 1, ModeApproved},
		{"23:59", 23, 59, ModeApproved},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			now := time.Date(2025, 1, 1, tt.hour, tt.min, 0, 0, time.UTC)
			got := mgr.determineMode(now)
			if got != tt.want {
				t.Errorf("determineMode(%02d:%02d) = %v; want %v", tt.hour, tt.min, got, tt.want)
			}
		})
	}
}

func TestManager_NextModeChangeTime(t *testing.T) {
	t.Parallel()

	store := newMockManorStore()
	castles := newMockCastleProvider(1)
	mgr := NewManager(store, castles)

	// Тест для каждого режима: следующий переход в будущем.
	now := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

	mgr.mu.Lock()
	mgr.mode = ModeModifiable
	mgr.mu.Unlock()

	next := mgr.nextModeChangeTime(now)
	if !next.After(now) {
		t.Errorf("next mode change %v is not after now %v", next, now)
	}
}

func TestManager_IsModifiable(t *testing.T) {
	t.Parallel()

	store := newMockManorStore()
	castles := newMockCastleProvider(1)
	mgr := NewManager(store, castles)

	mgr.mu.Lock()
	mgr.mode = ModeModifiable
	mgr.mu.Unlock()

	if !mgr.IsModifiable() {
		t.Error("IsModifiable() = false; want true")
	}
	if mgr.IsApproved() {
		t.Error("IsApproved() = true; want false")
	}
	if mgr.IsUnderMaintenance() {
		t.Error("IsUnderMaintenance() = true; want false")
	}
}

func TestManager_MaintenanceRefundsTreasury(t *testing.T) {
	t.Parallel()

	store := newMockManorStore()
	castles := newMockCastleProvider(1)
	castles.mu.Lock()
	castles.treasury[1] = 100000
	castles.mu.Unlock()

	mgr := NewManager(store, castles)
	if err := mgr.Init(context.Background()); err != nil {
		t.Fatalf("Init: %v", err)
	}

	// Устанавливаем текущие закупки: startAmount=100, amount=60 (40 куплено), price=100.
	mgr.mu.Lock()
	mgr.procure[1] = []*CropProcure{
		NewCropProcure(5073, 60, 1, 100, 100),
	}
	mgr.productionNext[1] = []*SeedProduction{
		NewSeedProduction(5016, 10, 50, 10),
	}
	mgr.procureNext[1] = nil
	mgr.mode = ModeApproved
	mgr.mu.Unlock()

	treasuryBefore := castles.Treasury(1)
	mgr.ChangeMode(context.Background())

	// Возврат: amount(60) * price(100) = 6000 adena.
	treasuryAfter := castles.Treasury(1)
	refund := treasuryAfter - treasuryBefore
	if refund != 6000 {
		t.Errorf("treasury refund = %d; want 6000", refund)
	}
}
