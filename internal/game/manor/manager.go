package manor

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/udisondev/la2go/internal/data"
)

// ManorStore persists seed production and crop procurement data.
type ManorStore interface {
	LoadProduction(ctx context.Context, castleID int32) ([]ProductionRow, error)
	LoadProcure(ctx context.Context, castleID int32) ([]ProcureRow, error)
	SaveAll(ctx context.Context, castleID int32, production []ProductionRow, procure []ProcureRow) error
	DeleteAll(ctx context.Context, castleID int32) error
}

// ProductionRow represents a DB row for seed production.
type ProductionRow struct {
	CastleID    int32
	SeedID      int32
	Amount      int32
	StartAmount int32
	Price       int64
	NextPeriod  bool
}

// ProcureRow represents a DB row for crop procurement.
type ProcureRow struct {
	CastleID    int32
	CropID      int32
	Amount      int32
	StartAmount int32
	Price       int64
	RewardType  int32
	NextPeriod  bool
}

// CastleProvider gives the manor manager access to castle treasury.
type CastleProvider interface {
	CastleIDs() []int32
	Treasury(castleID int32) int64
	AddToTreasury(castleID int32, amount int64)
}

// Manager manages the castle manor system.
// Thread-safe: all mutable state protected by mu.
type Manager struct {
	mu sync.RWMutex

	mode Mode

	// Per-castle production/procurement lists.
	// castleID → []*SeedProduction / []*CropProcure
	production     map[int32][]*SeedProduction
	productionNext map[int32][]*SeedProduction
	procure        map[int32][]*CropProcure
	procureNext    map[int32][]*CropProcure

	store    ManorStore
	castles  CastleProvider
	stopChan chan struct{}
}

// NewManager creates a new manor manager.
func NewManager(store ManorStore, castles CastleProvider) *Manager {
	return &Manager{
		mode:           ModeApproved,
		production:     make(map[int32][]*SeedProduction),
		productionNext: make(map[int32][]*SeedProduction),
		procure:        make(map[int32][]*CropProcure),
		procureNext:    make(map[int32][]*CropProcure),
		store:          store,
		castles:        castles,
		stopChan:       make(chan struct{}),
	}
}

// Init loads manor data from DB and determines the current mode.
func (m *Manager) Init(ctx context.Context) error {
	castleIDs := m.castles.CastleIDs()

	for _, castleID := range castleIDs {
		prodRows, err := m.store.LoadProduction(ctx, castleID)
		if err != nil {
			return fmt.Errorf("load production castle %d: %w", castleID, err)
		}

		var pCurrent, pNext []*SeedProduction
		seedIDs := data.GetAllSeedIDs()
		for _, row := range prodRows {
			if _, ok := seedIDs[row.SeedID]; !ok {
				slog.Warn("unknown seed in manor DB", "seedID", row.SeedID, "castleID", castleID)
				continue
			}
			sp := NewSeedProduction(row.SeedID, row.Amount, row.Price, row.StartAmount)
			if row.NextPeriod {
				pNext = append(pNext, sp)
			} else {
				pCurrent = append(pCurrent, sp)
			}
		}

		procRows, err := m.store.LoadProcure(ctx, castleID)
		if err != nil {
			return fmt.Errorf("load procure castle %d: %w", castleID, err)
		}

		var cCurrent, cNext []*CropProcure
		cropIDs := data.GetAllCropIDs()
		for _, row := range procRows {
			if _, ok := cropIDs[row.CropID]; !ok {
				slog.Warn("unknown crop in manor DB", "cropID", row.CropID, "castleID", castleID)
				continue
			}
			cp := NewCropProcure(row.CropID, row.Amount, row.RewardType, row.StartAmount, row.Price)
			if row.NextPeriod {
				cNext = append(cNext, cp)
			} else {
				cCurrent = append(cCurrent, cp)
			}
		}

		m.mu.Lock()
		m.production[castleID] = pCurrent
		m.productionNext[castleID] = pNext
		m.procure[castleID] = cCurrent
		m.procureNext[castleID] = cNext
		m.mu.Unlock()
	}

	// Determine initial mode from current time.
	// Default schedule: APPROVED 06:00→06:00, MAINTENANCE 06:00→06:03, MODIFIABLE 06:03→20:00.
	m.mode = m.determineMode(time.Now())

	slog.Info("manor system initialized",
		"mode", m.mode,
		"castles", len(castleIDs))

	return nil
}

// Mode returns the current manor mode.
func (m *Manager) Mode() Mode {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.mode
}

// IsModifiable returns true if manor settings can be changed.
func (m *Manager) IsModifiable() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.mode == ModeModifiable
}

// IsApproved returns true if manor settings are locked.
func (m *Manager) IsApproved() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.mode == ModeApproved
}

// IsUnderMaintenance returns true if manor is in maintenance mode.
func (m *Manager) IsUnderMaintenance() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.mode == ModeMaintenance
}

// SeedProduction returns the seed production list for a castle.
func (m *Manager) SeedProduction(castleID int32, nextPeriod bool) []*SeedProduction {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if nextPeriod {
		return m.productionNext[castleID]
	}
	return m.production[castleID]
}

// SeedProduct returns a specific seed production entry.
func (m *Manager) SeedProduct(castleID, seedID int32, nextPeriod bool) *SeedProduction {
	for _, sp := range m.SeedProduction(castleID, nextPeriod) {
		if sp.SeedID() == seedID {
			return sp
		}
	}
	return nil
}

// CropProcureList returns the crop procurement list for a castle.
func (m *Manager) CropProcureList(castleID int32, nextPeriod bool) []*CropProcure {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if nextPeriod {
		return m.procureNext[castleID]
	}
	return m.procure[castleID]
}

// CropProcureEntry returns a specific crop procurement entry.
func (m *Manager) CropProcureEntry(castleID, cropID int32, nextPeriod bool) *CropProcure {
	for _, cp := range m.CropProcureList(castleID, nextPeriod) {
		if cp.CropID() == cropID {
			return cp
		}
	}
	return nil
}

// SetNextSeedProduction sets the next period seed production for a castle.
// Called when a clan leader submits seed settings.
func (m *Manager) SetNextSeedProduction(castleID int32, list []*SeedProduction) {
	m.mu.Lock()
	m.productionNext[castleID] = list
	m.mu.Unlock()
}

// SetNextCropProcure sets the next period crop procurement for a castle.
// Called when a clan leader submits crop settings.
func (m *Manager) SetNextCropProcure(castleID int32, list []*CropProcure) {
	m.mu.Lock()
	m.procureNext[castleID] = list
	m.mu.Unlock()
}

// ManorCost calculates total cost of manor for a castle in the given period.
func (m *Manager) ManorCost(castleID int32, nextPeriod bool) int64 {
	production := m.SeedProduction(castleID, nextPeriod)
	procure := m.CropProcureList(castleID, nextPeriod)

	var total int64
	for _, sp := range production {
		refPrice := data.SeedReferencePrice(sp.SeedID())
		total += refPrice * int64(sp.StartAmount())
	}
	for _, cp := range procure {
		total += cp.Price() * int64(cp.StartAmount())
	}
	return total
}

// ChangeMode transitions the manor to the next mode.
// APPROVED → MAINTENANCE → MODIFIABLE → APPROVED.
// This is the core of the manor period cycle.
func (m *Manager) ChangeMode(ctx context.Context) {
	m.mu.Lock()
	currentMode := m.mode
	m.mu.Unlock()

	switch currentMode {
	case ModeApproved:
		m.transitionToMaintenance(ctx)
	case ModeMaintenance:
		m.transitionToModifiable()
	case ModeModifiable:
		m.transitionToApproved(ctx)
	}
}

// transitionToMaintenance: APPROVED → MAINTENANCE.
// Processes crops: adds mature items to warehouse, returns unused adena to treasury.
// Rotates next→current period.
func (m *Manager) transitionToMaintenance(ctx context.Context) {
	castleIDs := m.castles.CastleIDs()

	m.mu.Lock()
	for _, castleID := range castleIDs {
		// Возврат неиспользованных средств из текущих закупок в казну.
		for _, crop := range m.procure[castleID] {
			if crop.StartAmount() > 0 && crop.Amount() > 0 {
				refund := int64(crop.Amount()) * crop.Price()
				m.castles.AddToTreasury(castleID, refund)
			}
		}

		// Ротация: next → current.
		nextProd := m.productionNext[castleID]
		nextProc := m.procureNext[castleID]
		m.production[castleID] = nextProd
		m.procure[castleID] = nextProc

		// Подготовка next period: копия текущего с полными amount,
		// если в казне достаточно средств.
		manorCost := m.calculateManorCost(nextProd, nextProc)
		if m.castles.Treasury(castleID) < manorCost {
			m.productionNext[castleID] = nil
			m.procureNext[castleID] = nil
		} else {
			newProd := make([]*SeedProduction, len(nextProd))
			for i, sp := range nextProd {
				newProd[i] = NewSeedProduction(sp.SeedID(), sp.StartAmount(), sp.Price(), sp.StartAmount())
			}
			m.productionNext[castleID] = newProd

			newProc := make([]*CropProcure, len(nextProc))
			for i, cp := range nextProc {
				newProc[i] = NewCropProcure(cp.CropID(), cp.StartAmount(), cp.RewardType(), cp.StartAmount(), cp.Price())
			}
			m.procureNext[castleID] = newProc
		}
	}
	m.mode = ModeMaintenance
	m.mu.Unlock()

	// Сохраняем в БД.
	if err := m.saveAll(ctx); err != nil {
		slog.Error("save manor after maintenance transition", "error", err)
	}

	slog.Info("manor mode changed", "mode", ModeMaintenance)
}

// transitionToModifiable: MAINTENANCE → MODIFIABLE.
// Notifies clan leaders that they can now modify manor settings.
func (m *Manager) transitionToModifiable() {
	m.mu.Lock()
	m.mode = ModeModifiable
	m.mu.Unlock()

	slog.Info("manor mode changed", "mode", ModeModifiable)
}

// transitionToApproved: MODIFIABLE → APPROVED.
// Validates treasury, deducts manor costs.
func (m *Manager) transitionToApproved(ctx context.Context) {
	castleIDs := m.castles.CastleIDs()

	m.mu.Lock()
	for _, castleID := range castleIDs {
		manorCost := m.calculateManorCost(m.productionNext[castleID], m.procureNext[castleID])
		if m.castles.Treasury(castleID) < manorCost {
			// Недостаточно средств — очищаем next period.
			m.productionNext[castleID] = nil
			m.procureNext[castleID] = nil
			slog.Warn("manor: insufficient treasury, cleared next period",
				"castleID", castleID,
				"treasury", m.castles.Treasury(castleID),
				"cost", manorCost)
		} else {
			m.castles.AddToTreasury(castleID, -manorCost)
		}
	}
	m.mode = ModeApproved
	m.mu.Unlock()

	if err := m.saveAll(ctx); err != nil {
		slog.Error("save manor after approved transition", "error", err)
	}

	slog.Info("manor mode changed", "mode", ModeApproved)
}

// ResetManorData clears all manor data for a castle.
func (m *Manager) ResetManorData(ctx context.Context, castleID int32) error {
	m.mu.Lock()
	m.production[castleID] = nil
	m.productionNext[castleID] = nil
	m.procure[castleID] = nil
	m.procureNext[castleID] = nil
	m.mu.Unlock()

	if err := m.store.DeleteAll(ctx, castleID); err != nil {
		return fmt.Errorf("delete manor data castle %d: %w", castleID, err)
	}
	return nil
}

// Save persists all manor data to DB.
func (m *Manager) Save(ctx context.Context) error {
	return m.saveAll(ctx)
}

// RunSaveLoop periodically saves manor data.
// Goroutine exits when ctx is cancelled.
func (m *Manager) RunSaveLoop(ctx context.Context, interval time.Duration) error {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			// Финальное сохранение перед выходом.
			if err := m.saveAll(context.Background()); err != nil {
				slog.Error("final manor save", "error", err)
			}
			return ctx.Err()
		case <-ticker.C:
			if err := m.saveAll(ctx); err != nil {
				slog.Error("periodic manor save", "error", err)
			}
		}
	}
}

// RunModeLoop manages manor mode transitions on schedule.
// Default schedule: APPROVED at 20:00, MAINTENANCE at 06:00, MODIFIABLE at 06:03.
// Goroutine exits when ctx is cancelled.
func (m *Manager) RunModeLoop(ctx context.Context) error {
	for {
		nextChange := m.nextModeChangeTime(time.Now())
		delay := time.Until(nextChange)
		if delay <= 0 {
			delay = time.Second
		}

		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
			m.ChangeMode(ctx)
		}
	}
}

// Stop signals the manager to stop.
func (m *Manager) Stop() {
	close(m.stopChan)
}

// --- Internal helpers ---

// calculateManorCost computes total manor cost from raw lists (no lock needed, caller holds mu).
func (m *Manager) calculateManorCost(production []*SeedProduction, procure []*CropProcure) int64 {
	var total int64
	for _, sp := range production {
		refPrice := data.SeedReferencePrice(sp.SeedID())
		total += refPrice * int64(sp.StartAmount())
	}
	for _, cp := range procure {
		total += cp.Price() * int64(cp.StartAmount())
	}
	return total
}

func (m *Manager) saveAll(ctx context.Context) error {
	castleIDs := m.castles.CastleIDs()

	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, castleID := range castleIDs {
		var prodRows []ProductionRow
		for _, sp := range m.production[castleID] {
			prodRows = append(prodRows, ProductionRow{
				CastleID:    castleID,
				SeedID:      sp.SeedID(),
				Amount:      sp.Amount(),
				StartAmount: sp.StartAmount(),
				Price:       sp.Price(),
				NextPeriod:  false,
			})
		}
		for _, sp := range m.productionNext[castleID] {
			prodRows = append(prodRows, ProductionRow{
				CastleID:    castleID,
				SeedID:      sp.SeedID(),
				Amount:      sp.Amount(),
				StartAmount: sp.StartAmount(),
				Price:       sp.Price(),
				NextPeriod:  true,
			})
		}

		var procRows []ProcureRow
		for _, cp := range m.procure[castleID] {
			procRows = append(procRows, ProcureRow{
				CastleID:    castleID,
				CropID:      cp.CropID(),
				Amount:      cp.Amount(),
				StartAmount: cp.StartAmount(),
				Price:       cp.Price(),
				RewardType:  cp.RewardType(),
				NextPeriod:  false,
			})
		}
		for _, cp := range m.procureNext[castleID] {
			procRows = append(procRows, ProcureRow{
				CastleID:    castleID,
				CropID:      cp.CropID(),
				Amount:      cp.Amount(),
				StartAmount: cp.StartAmount(),
				Price:       cp.Price(),
				RewardType:  cp.RewardType(),
				NextPeriod:  true,
			})
		}

		if err := m.store.SaveAll(ctx, castleID, prodRows, procRows); err != nil {
			return fmt.Errorf("save manor castle %d: %w", castleID, err)
		}
	}
	return nil
}

// determineMode calculates the current manor mode based on time of day.
// Default schedule:
//
//	06:00 → MAINTENANCE (3 min)
//	06:03 → MODIFIABLE
//	20:00 → APPROVED
//
// Config defaults used here (same as Java ALT_MANOR_*).
func (m *Manager) determineMode(now time.Time) Mode {
	hour := now.Hour()
	min := now.Minute()

	const (
		refreshHour      = 6
		refreshMin       = 0
		maintenanceMin   = 3
		approveHour      = 20
		approveMin       = 0
	)

	maintenanceEnd := refreshMin + maintenanceMin

	// 06:00-06:03 → MAINTENANCE
	if hour == refreshHour && min >= refreshMin && min < maintenanceEnd {
		return ModeMaintenance
	}

	// 06:03-20:00 → MODIFIABLE
	if (hour == refreshHour && min >= maintenanceEnd) ||
		(hour > refreshHour && hour < approveHour) ||
		(hour == approveHour && min <= approveMin) {
		return ModeModifiable
	}

	// 20:00-06:00 → APPROVED
	return ModeApproved
}

// nextModeChangeTime calculates when the next mode transition should happen.
func (m *Manager) nextModeChangeTime(now time.Time) time.Time {
	m.mu.RLock()
	mode := m.mode
	m.mu.RUnlock()

	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	switch mode {
	case ModeApproved:
		// Next: MAINTENANCE at 06:00.
		next := today.Add(6 * time.Hour)
		if !next.After(now) {
			next = next.Add(24 * time.Hour)
		}
		return next

	case ModeMaintenance:
		// Next: MODIFIABLE at 06:03.
		next := today.Add(6*time.Hour + 3*time.Minute)
		if !next.After(now) {
			next = next.Add(24 * time.Hour)
		}
		return next

	case ModeModifiable:
		// Next: APPROVED at 20:00.
		next := today.Add(20 * time.Hour)
		if !next.After(now) {
			next = next.Add(24 * time.Hour)
		}
		return next

	default:
		// Disabled — never change.
		return now.Add(24 * time.Hour)
	}
}
