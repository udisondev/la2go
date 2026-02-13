package sevensigns

import (
	"math"
	"sync"
)

// Manager manages the Seven Signs system state.
//
// Thread-safe: all public methods acquire appropriate locks.
type Manager struct {
	mu      sync.RWMutex
	status  Status
	players map[int64]*PlayerData // charID → data

	// Festival results indexed by cycle → festivalID*2+cabalIndex.
	festivals map[int32]map[int32]*FestivalResult // cycle → compositeKey → result
}

// NewManager creates a new Seven Signs manager with default initial state.
func NewManager() *Manager {
	return &Manager{
		players:   make(map[int64]*PlayerData, 256),
		festivals: make(map[int32]map[int32]*FestivalResult, 4),
		status: Status{
			CurrentCycle: 1,
			ActivePeriod: PeriodRecruitment,
		},
	}
}

// CurrentPeriod returns the active period.
func (m *Manager) CurrentPeriod() Period {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.status.ActivePeriod
}

// CurrentCycle returns the current cycle number.
func (m *Manager) CurrentCycle() int32 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.status.CurrentCycle
}

// PreviousWinner returns the cabal that won last cycle.
func (m *Manager) PreviousWinner() Cabal {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.status.PreviousWinner
}

// SealOwner returns which cabal controls a seal.
func (m *Manager) SealOwner(seal Seal) Cabal {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.status.SealOwner(seal)
}

// IsRecruitmentPeriod returns true during the recruitment phase.
func (m *Manager) IsRecruitmentPeriod() bool {
	return m.CurrentPeriod() == PeriodRecruitment
}

// IsCompetitionPeriod returns true during the competition phase.
func (m *Manager) IsCompetitionPeriod() bool {
	return m.CurrentPeriod() == PeriodCompetition
}

// IsCompResultsPeriod returns true during the results phase.
func (m *Manager) IsCompResultsPeriod() bool {
	return m.CurrentPeriod() == PeriodResults
}

// IsSealValidationPeriod returns true during seal validation.
func (m *Manager) IsSealValidationPeriod() bool {
	return m.CurrentPeriod() == PeriodSealValidation
}

// PlayerCabal returns the cabal a character belongs to.
func (m *Manager) PlayerCabal(charID int64) Cabal {
	m.mu.RLock()
	defer m.mu.RUnlock()
	pd := m.players[charID]
	if pd == nil {
		return CabalNull
	}
	return pd.Cabal
}

// PlayerSeal returns the seal a character is contributing to.
func (m *Manager) PlayerSeal(charID int64) Seal {
	m.mu.RLock()
	defer m.mu.RUnlock()
	pd := m.players[charID]
	if pd == nil {
		return SealNull
	}
	return pd.Seal
}

// PlayerData returns a copy of the character's Seven Signs data.
// Returns nil if the character has no data.
func (m *Manager) PlayerData(charID int64) *PlayerData {
	m.mu.RLock()
	defer m.mu.RUnlock()
	pd := m.players[charID]
	if pd == nil {
		return nil
	}
	cp := *pd
	return &cp
}

// Status returns a copy of the global status.
func (m *Manager) Status() Status {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.status
}

// CabalHighestScore returns the cabal with the highest normalized score,
// or CabalNull if tied.
// Java reference: SevenSigns.getCabalHighestScore()
func (m *Manager) CabalHighestScore() Cabal {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.cabalHighestScoreLocked()
}

// cabalHighestScoreLocked returns the winning cabal. Must be called under lock.
func (m *Manager) cabalHighestScoreLocked() Cabal {
	dawnScore := m.currentScore(CabalDawn)
	duskScore := m.currentScore(CabalDusk)

	if dawnScore > duskScore {
		return CabalDawn
	}
	if duskScore > dawnScore {
		return CabalDusk
	}
	return CabalNull
}

// currentScore returns the normalized score for a cabal.
// Formula: round(stoneScore / totalStoneScore * 500) + festivalScore
// Java reference: SevenSigns.getCurrentScore()
func (m *Manager) currentScore(cabal Cabal) int64 {
	totalStoneScore := float64(m.status.DawnStoneScore + m.status.DuskStoneScore)
	if totalStoneScore == 0 {
		totalStoneScore = 1
	}

	switch cabal {
	case CabalDawn:
		return int64(math.Round(float64(m.status.DawnStoneScore)/totalStoneScore*500)) + int64(m.status.DawnFestivalScore)
	case CabalDusk:
		return int64(math.Round(float64(m.status.DuskStoneScore)/totalStoneScore*500)) + int64(m.status.DuskFestivalScore)
	default:
		return 0
	}
}

// DawnScore returns the normalized Dawn faction score.
// Java reference: SevenSigns.getCurrentScore(CABAL_DAWN)
func (m *Manager) DawnScore() int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.currentScore(CabalDawn)
}

// DuskScore returns the normalized Dusk faction score.
// Java reference: SevenSigns.getCurrentScore(CABAL_DUSK)
func (m *Manager) DuskScore() int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.currentScore(CabalDusk)
}

// JoinCabal registers a character to a cabal.
// Only allowed during recruitment period.
// Returns false if already joined or not in recruitment period.
func (m *Manager) JoinCabal(charID int64, cabal Cabal) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.status.ActivePeriod != PeriodRecruitment {
		return false
	}
	if cabal != CabalDawn && cabal != CabalDusk {
		return false
	}

	pd := m.players[charID]
	if pd != nil && pd.Cabal != CabalNull {
		return false
	}

	if pd == nil {
		pd = &PlayerData{CharID: charID}
		m.players[charID] = pd
	}
	pd.Cabal = cabal
	return true
}

// ChooseSeal sets a character's seal choice.
// Requires the character to have joined a cabal.
func (m *Manager) ChooseSeal(charID int64, seal Seal) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	pd := m.players[charID]
	if pd == nil || pd.Cabal == CabalNull {
		return false
	}
	if seal < SealAvarice || seal > SealStrife {
		return false
	}
	pd.Seal = seal
	return true
}

// ContributeStones converts seal stones from a player's inventory into
// contribution points. Returns the number of contribution points earned.
func (m *Manager) ContributeStones(charID int64, blue, green, red int32) int64 {
	m.mu.Lock()
	defer m.mu.Unlock()

	pd := m.players[charID]
	if pd == nil || pd.Cabal == CabalNull || pd.Seal == SealNull {
		return 0
	}

	contrib := int64(blue)*BlueContribPoints +
		int64(green)*GreenContribPoints +
		int64(red)*RedContribPoints

	pd.BlueStones += blue
	pd.GreenStones += green
	pd.RedStones += red
	pd.ContributionScore += contrib

	// Обновить глобальные score.
	dawn, dusk := m.status.SealScore(pd.Seal)
	if pd.Cabal == CabalDawn {
		m.status.DawnStoneScore += contrib
		m.setSealScore(pd.Seal, dawn+int32(contrib), dusk)
	} else {
		m.status.DuskStoneScore += contrib
		m.setSealScore(pd.Seal, dawn, dusk+int32(contrib))
	}

	return contrib
}

func (m *Manager) setSealScore(seal Seal, dawn, dusk int32) {
	switch seal {
	case SealAvarice:
		m.status.AvariceDawnScore = dawn
		m.status.AvariceDuskScore = dusk
	case SealGnosis:
		m.status.GnosisDawnScore = dawn
		m.status.GnosisDuskScore = dusk
	case SealStrife:
		m.status.StrifeDawnScore = dawn
		m.status.StrifeDuskScore = dusk
	}
}

// AddAncientAdena adds ancient adena to a player's Seven Signs balance.
func (m *Manager) AddAncientAdena(charID int64, amount int64) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	pd := m.players[charID]
	if pd == nil {
		return false
	}
	pd.AncientAdena += amount
	return true
}

// DeductAncientAdena deducts ancient adena. Returns false if insufficient.
func (m *Manager) DeductAncientAdena(charID int64, amount int64) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	pd := m.players[charID]
	if pd == nil || pd.AncientAdena < amount {
		return false
	}
	pd.AncientAdena -= amount
	return true
}

// TransitionPeriod advances to the next period.
// Returns the new period.
func (m *Manager) TransitionPeriod() Period {
	m.mu.Lock()
	defer m.mu.Unlock()

	switch m.status.ActivePeriod {
	case PeriodRecruitment:
		m.status.ActivePeriod = PeriodCompetition
	case PeriodCompetition:
		m.status.ActivePeriod = PeriodResults
		m.computeResults()
	case PeriodResults:
		m.status.ActivePeriod = PeriodSealValidation
	case PeriodSealValidation:
		m.status.ActivePeriod = PeriodRecruitment
		m.startNewCycle()
	}
	return m.status.ActivePeriod
}

// computeResults determines seal owners based on 35%/10% threshold rules.
// Java reference: SevenSigns.calcNewSealOwners()
// Called under write lock.
func (m *Manager) computeResults() {
	winner := m.cabalHighestScoreLocked()
	m.status.PreviousWinner = winner

	for _, seal := range []Seal{SealAvarice, SealGnosis, SealStrife} {
		prevOwner := m.status.SealOwner(seal)
		newOwner := CabalNull

		dawnProportion := m.getSealProportion(seal, CabalDawn)
		totalDawn := m.getTotalMembers(CabalDawn)
		if totalDawn == 0 {
			totalDawn = 1
		}
		dawnPercent := int(math.Round(float64(dawnProportion) / float64(totalDawn) * 100))

		duskProportion := m.getSealProportion(seal, CabalDusk)
		totalDusk := m.getTotalMembers(CabalDusk)
		if totalDusk == 0 {
			totalDusk = 1
		}
		duskPercent := int(math.Round(float64(duskProportion) / float64(totalDusk) * 100))

		switch prevOwner {
		case CabalNull:
			switch winner {
			case CabalNull:
				newOwner = CabalNull
			case CabalDawn:
				if dawnPercent >= 35 {
					newOwner = CabalDawn
				}
			case CabalDusk:
				if duskPercent >= 35 {
					newOwner = CabalDusk
				}
			}

		case CabalDawn:
			switch winner {
			case CabalNull:
				if dawnPercent >= 10 {
					newOwner = CabalDawn
				}
			case CabalDawn:
				if dawnPercent >= 10 {
					newOwner = CabalDawn
				}
			case CabalDusk:
				if duskPercent >= 35 {
					newOwner = CabalDusk
				} else if dawnPercent >= 10 {
					newOwner = CabalDawn
				}
			}

		case CabalDusk:
			switch winner {
			case CabalNull:
				if duskPercent >= 10 {
					newOwner = CabalDusk
				}
			case CabalDawn:
				if dawnPercent >= 35 {
					newOwner = CabalDawn
				} else if duskPercent >= 10 {
					newOwner = CabalDusk
				}
			case CabalDusk:
				if duskPercent >= 10 {
					newOwner = CabalDusk
				}
			}
		}

		m.status.SetSealOwner(seal, newOwner)
	}
}

// getTotalMembers counts all players belonging to a cabal.
// Java reference: SevenSigns.getTotalMembers()
// Called under lock.
func (m *Manager) getTotalMembers(cabal Cabal) int {
	count := 0
	for _, pd := range m.players {
		if pd.Cabal == cabal {
			count++
		}
	}
	return count
}

// getSealProportion counts players of a cabal who chose a specific seal.
// Java reference: SevenSigns.getSealProportion()
// Called under lock.
func (m *Manager) getSealProportion(seal Seal, cabal Cabal) int {
	count := 0
	for _, pd := range m.players {
		if pd.Cabal == cabal && pd.Seal == seal {
			count++
		}
	}
	return count
}

// startNewCycle resets per-cycle data and increments the cycle counter.
// Called under write lock.
func (m *Manager) startNewCycle() {
	m.status.CurrentCycle++
	m.status.DawnStoneScore = 0
	m.status.DawnFestivalScore = 0
	m.status.DuskStoneScore = 0
	m.status.DuskFestivalScore = 0
	m.status.AvariceDawnScore = 0
	m.status.AvariceDuskScore = 0
	m.status.GnosisDawnScore = 0
	m.status.GnosisDuskScore = 0
	m.status.StrifeDawnScore = 0
	m.status.StrifeDuskScore = 0

	// Очистить участие игроков.
	for _, pd := range m.players {
		pd.Cabal = CabalNull
		pd.Seal = SealNull
		pd.BlueStones = 0
		pd.GreenStones = 0
		pd.RedStones = 0
		pd.ContributionScore = 0
	}
}

// LoadStatus overwrites the global status from persisted data.
func (m *Manager) LoadStatus(s Status) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.status = s
}

// LoadPlayerData inserts or replaces a player's Seven Signs data.
func (m *Manager) LoadPlayerData(pd *PlayerData) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.players[pd.CharID] = pd
}

// AllPlayerData returns a snapshot of all player data for persistence.
func (m *Manager) AllPlayerData() []*PlayerData {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*PlayerData, 0, len(m.players))
	for _, pd := range m.players {
		cp := *pd
		result = append(result, &cp)
	}
	return result
}

// AddFestivalScore adds festival points to a cabal's score.
func (m *Manager) AddFestivalScore(cabal Cabal, score int32) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if cabal == CabalDawn {
		m.status.DawnFestivalScore += score
	} else if cabal == CabalDusk {
		m.status.DuskFestivalScore += score
	}
}

// SaveFestivalResult records a festival result.
func (m *Manager) SaveFestivalResult(result *FestivalResult) {
	m.mu.Lock()
	defer m.mu.Unlock()

	cycleMap := m.festivals[result.Cycle]
	if cycleMap == nil {
		cycleMap = make(map[int32]*FestivalResult, FestivalCount*2)
		m.festivals[result.Cycle] = cycleMap
	}

	key := result.FestivalID*2 + int32(result.Cabal)
	cycleMap[key] = result
}

// FestivalResults returns all festival results for a cycle.
func (m *Manager) FestivalResults(cycle int32) []*FestivalResult {
	m.mu.RLock()
	defer m.mu.RUnlock()

	cycleMap := m.festivals[cycle]
	if cycleMap == nil {
		return nil
	}

	results := make([]*FestivalResult, 0, len(cycleMap))
	for _, r := range cycleMap {
		cp := *r
		results = append(results, &cp)
	}
	return results
}
