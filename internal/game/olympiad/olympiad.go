package olympiad

import (
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

// Olympiad управляет полным жизненным циклом олимпиады:
// Competition (period=0) → Validation (period=1) → новый цикл.
// L2J reference: Olympiad.java
//
// ⚠️ КРИТИЧНО: все shared флаги используют atomic.Bool/Int32
// вместо static boolean (fix Java race condition).
type Olympiad struct {
	// Period state — atomic для thread safety
	period       atomic.Int32 // 0 = Competition, 1 = Validation
	currentCycle atomic.Int32
	olympiadEnd  atomic.Int64 // Unix ms
	validationEnd atomic.Int64
	nextWeeklyChange atomic.Int64
	compEnd      atomic.Int64

	// ⚠️ atomic.Bool — fix Java race condition
	inCompPeriod atomic.Bool

	mu       sync.RWMutex
	compStart time.Time

	manager   *Manager
	heroTable *HeroTable

	// Ранги за последний месяц
	ranksMu sync.RWMutex
	ranks   map[int64]Rank
}

// NewOlympiad создаёт новый экземпляр олимпиады.
func NewOlympiad() *Olympiad {
	nobles := NewNobleTable()
	return &Olympiad{
		manager:   NewManager(nobles),
		heroTable: NewHeroTable(),
		ranks:     make(map[int64]Rank),
	}
}

// Manager returns the olympiad manager (matchmaking).
func (o *Olympiad) Manager() *Manager { return o.manager }

// HeroTable returns the hero table.
func (o *Olympiad) HeroTable() *HeroTable { return o.heroTable }

// Nobles returns the noble table.
func (o *Olympiad) Nobles() *NobleTable { return o.manager.Nobles() }

// Period returns the current olympiad period.
func (o *Olympiad) Period() Period { return Period(o.period.Load()) }

// CurrentCycle returns the current monthly cycle number.
func (o *Olympiad) CurrentCycle() int32 { return o.currentCycle.Load() }

// InCompPeriod reports whether competitions are currently active.
func (o *Olympiad) InCompPeriod() bool { return o.inCompPeriod.Load() }

// IsOlympiadEnd reports whether the olympiad is in validation (period != 0).
func (o *Olympiad) IsOlympiadEnd() bool { return o.period.Load() != 0 }

// OlympiadEnd returns the timestamp (Unix ms) when the month ends.
func (o *Olympiad) OlympiadEnd() int64 { return o.olympiadEnd.Load() }

// RemainingTimeToEnd returns time until month end.
func (o *Olympiad) RemainingTimeToEnd() time.Duration {
	end := time.UnixMilli(o.olympiadEnd.Load())
	remaining := time.Until(end)
	if remaining < 0 {
		return 0
	}
	return remaining
}

// CompStart returns the start time of today's competitions.
func (o *Olympiad) CompStart() time.Time {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return o.compStart
}

// CompEnd returns the timestamp (Unix ms) when today's comp period ends.
func (o *Olympiad) CompEnd() int64 { return o.compEnd.Load() }

// SetPeriod устанавливает период (для загрузки из БД).
func (o *Olympiad) SetPeriod(p Period) { o.period.Store(int32(p)) }

// SetCurrentCycle устанавливает номер цикла (для загрузки из БД).
func (o *Olympiad) SetCurrentCycle(c int32) { o.currentCycle.Store(c) }

// SetOlympiadEnd устанавливает конец месяца (для загрузки из БД).
func (o *Olympiad) SetOlympiadEnd(ms int64) { o.olympiadEnd.Store(ms) }

// SetValidationEnd устанавливает конец валидации (для загрузки из БД).
func (o *Olympiad) SetValidationEnd(ms int64) { o.validationEnd.Store(ms) }

// SetNextWeeklyChange устанавливает время следующего weekly grant.
func (o *Olympiad) SetNextWeeklyChange(ms int64) { o.nextWeeklyChange.Store(ms) }

// StartCompPeriod запускает период соревнований.
func (o *Olympiad) StartCompPeriod() {
	o.inCompPeriod.Store(true)
}

// EndCompPeriod останавливает период соревнований.
func (o *Olympiad) EndCompPeriod() {
	o.inCompPeriod.Store(false)
}

// SetCompSchedule устанавливает время comp start/end для текущего дня.
func (o *Olympiad) SetCompSchedule(start time.Time) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.compStart = start
	o.compEnd.Store(start.Add(CompPeriodDuration).UnixMilli())
}

// EndMonth переводит олимпиаду в период валидации.
// Выбирает героев, рассчитывает ранги.
func (o *Olympiad) EndMonth() []*HeroCandidate {
	// Перейти в validation
	o.period.Store(int32(PeriodValidation))
	o.validationEnd.Store(time.Now().Add(ValidationPeriod).UnixMilli())

	// Выбрать героев
	candidates := SelectHeroes(o.manager.Nobles())

	// Установить новых героев
	o.heroTable.ComputeNewHeroes(candidates)

	// Рассчитать ранги
	allNobles := o.manager.Nobles().All()
	newRanks := CalculateRanks(allNobles)

	o.ranksMu.Lock()
	o.ranks = newRanks
	o.ranksMu.Unlock()

	return candidates
}

// EndValidation завершает период валидации и начинает новый цикл.
func (o *Olympiad) EndValidation() {
	o.currentCycle.Add(1)
	o.period.Store(int32(PeriodCompetition))

	// Очистить ранги
	o.ranksMu.Lock()
	o.ranks = make(map[int64]Rank)
	o.ranksMu.Unlock()
}

// SetNewOlympiadEnd устанавливает конец через +1 месяц.
func (o *Olympiad) SetNewOlympiadEnd() {
	now := time.Now()
	nextMonth := now.AddDate(0, 1, 0)
	// Первое число следующего месяца, 12:00
	end := time.Date(nextMonth.Year(), nextMonth.Month(), 1, 12, 0, 0, 0, now.Location())
	o.olympiadEnd.Store(end.UnixMilli())
	o.nextWeeklyChange.Store(now.Add(WeeklyPeriod).UnixMilli())
}

// GrantWeeklyPoints начисляет еженедельные очки всем nobles.
func (o *Olympiad) GrantWeeklyPoints() {
	if o.Period() == PeriodValidation {
		return
	}
	o.manager.Nobles().GrantAllWeeklyPoints()
	o.nextWeeklyChange.Store(time.Now().Add(WeeklyPeriod).UnixMilli())
}

// GetRank returns the rank for a noble (0 if not ranked).
func (o *Olympiad) GetRank(charID int64) Rank {
	o.ranksMu.RLock()
	defer o.ranksMu.RUnlock()
	return o.ranks[charID]
}

// RegisterNoble регистрирует благородного в олимпиаду.
// Returns error reason string (empty = success).
func (o *Olympiad) RegisterNoble(charID int64, classID int32, classBased bool) string {
	if !o.InCompPeriod() {
		return "competitions not active"
	}

	if o.Period() == PeriodValidation {
		return "validation period"
	}

	// Уже зарегистрирован?
	noble := o.manager.Nobles().Get(charID)
	if noble != nil && o.manager.IsRegistered(uint32(charID)) {
		return "already registered"
	}

	// Зарегистрировать noble если ещё нет
	if noble == nil {
		noble = o.manager.Nobles().Register(charID, classID)
	}

	// Проверить минимум очков
	points := noble.Points()
	if classBased && points < 3 {
		return "not enough points for classed (min 3)"
	}
	if !classBased && points < 5 {
		return "not enough points for non-classed (min 5)"
	}

	return ""
}

// GetNoblePoints returns olympiad points for a noble (0 if not registered).
func (o *Olympiad) GetNoblePoints(charID int64) int32 {
	noble := o.manager.Nobles().Get(charID)
	if noble == nil {
		return 0
	}
	return noble.Points()
}

// GetNobleStats returns stats snapshot for a noble.
func (o *Olympiad) GetNobleStats(charID int64) (NobleStats, bool) {
	noble := o.manager.Nobles().Get(charID)
	if noble == nil {
		return NobleStats{}, false
	}
	return noble.Stats(), true
}

// GetClassLeaderboard returns top nobles for a class, sorted by points.
func (o *Olympiad) GetClassLeaderboard(classID int32) []NobleStats {
	classNobles := o.manager.Nobles().ByClassID(classID)
	if len(classNobles) == 0 {
		return nil
	}

	stats := make([]NobleStats, 0, len(classNobles))
	for _, n := range classNobles {
		stats = append(stats, n.Stats())
	}

	// Сортировка: points DESC → compDone DESC → compWon DESC
	sort.Slice(stats, func(i, j int) bool {
		if stats[i].Points != stats[j].Points {
			return stats[i].Points > stats[j].Points
		}
		if stats[i].CompDone != stats[j].CompDone {
			return stats[i].CompDone > stats[j].CompDone
		}
		return stats[i].CompWon > stats[j].CompWon
	})

	return stats
}
