package olympiad

import (
	"log/slog"
	"math/rand/v2"
	"sync"
	"sync/atomic"

	"github.com/udisondev/la2go/internal/model"
)

// Manager отвечает за matchmaking и управление стадионами.
// Потокобезопасен через sync.RWMutex.
// L2J reference: OlympiadManager.java
type Manager struct {
	mu sync.RWMutex

	stadiums  [StadiumCount]*Stadium
	games     map[int32]*Game // stadiumID → Game
	nextGameID atomic.Int32

	// Очереди регистрации
	nonClassBased []*model.Player              // non-class-based
	classBased    map[int32][]*model.Player     // classID → players

	nobles *NobleTable

	// ⚠️ atomic.Bool — fix Java race condition
	battleStarted atomic.Bool
}

// NewManager creates a new Olympiad manager.
func NewManager(nobles *NobleTable) *Manager {
	return &Manager{
		stadiums:      NewStadiums(),
		games:         make(map[int32]*Game),
		classBased:    make(map[int32][]*model.Player),
		nobles:        nobles,
	}
}

// Nobles returns the noble table.
func (m *Manager) Nobles() *NobleTable { return m.nobles }

// BattleStarted reports whether any battles are active.
func (m *Manager) BattleStarted() bool { return m.battleStarted.Load() }

// RegisterClassBased добавляет игрока в class-based очередь.
func (m *Manager) RegisterClassBased(player *model.Player, classID int32) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.classBased[classID] = append(m.classBased[classID], player)
}

// RegisterNonClassBased добавляет игрока в non-class-based очередь.
func (m *Manager) RegisterNonClassBased(player *model.Player) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.nonClassBased = append(m.nonClassBased, player)
}

// UnregisterPlayer удаляет игрока из всех очередей.
func (m *Manager) UnregisterPlayer(objectID uint32) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Non-class-based
	for i, p := range m.nonClassBased {
		if p.ObjectID() == objectID {
			m.nonClassBased = append(m.nonClassBased[:i], m.nonClassBased[i+1:]...)
			return true
		}
	}

	// Class-based
	for classID, players := range m.classBased {
		for i, p := range players {
			if p.ObjectID() == objectID {
				m.classBased[classID] = append(players[:i], players[i+1:]...)
				if len(m.classBased[classID]) == 0 {
					delete(m.classBased, classID)
				}
				return true
			}
		}
	}

	return false
}

// IsRegistered проверяет, зарегистрирован ли игрок.
func (m *Manager) IsRegistered(objectID uint32) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, p := range m.nonClassBased {
		if p.ObjectID() == objectID {
			return true
		}
	}
	for _, players := range m.classBased {
		for _, p := range players {
			if p.ObjectID() == objectID {
				return true
			}
		}
	}
	return false
}

// HasEnoughNonClassed проверяет, достаточно ли игроков для non-class матча.
func (m *Manager) HasEnoughNonClassed() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.nonClassBased) >= MinNonClassedParticipants
}

// HasEnoughClassed возвращает список classID с достаточным количеством игроков.
func (m *Manager) HasEnoughClassed() []int32 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []int32
	for classID, players := range m.classBased {
		if len(players) >= MinClassedParticipants {
			result = append(result, classID)
		}
	}
	return result
}

// NonClassedCount returns count of non-class-based registered players.
func (m *Manager) NonClassedCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.nonClassBased)
}

// ClassedCount returns count of class-based registered players for a class.
func (m *Manager) ClassedCount(classID int32) int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.classBased[classID])
}

// ClearRegistered очищает все очереди.
func (m *Manager) ClearRegistered() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.nonClassBased = nil
	m.classBased = make(map[int32][]*model.Player)
}

// nextOpponents выбирает 2 случайных соперников из списка.
// Удаляет выбранных из списка. Caller должен держать mu.Lock().
func nextOpponents(players []*model.Player) ([]*model.Player, []*model.Player) {
	if len(players) < 2 {
		return nil, players
	}

	// Случайный выбор первого
	firstIdx := rand.IntN(len(players))
	first := players[firstIdx]
	players = append(players[:firstIdx], players[firstIdx+1:]...)

	// Случайный выбор второго
	secondIdx := rand.IntN(len(players))
	second := players[secondIdx]
	players = append(players[:secondIdx], players[secondIdx+1:]...)

	return []*model.Player{first, second}, players
}

// CreateMatches формирует матчи и назначает стадионы.
// Возвращает созданные игры.
func (m *Manager) CreateMatches() []*Game {
	m.mu.Lock()
	defer m.mu.Unlock()

	readyClasses := m.hasEnoughClassedLocked()
	readyNonClassed := len(m.nonClassBased) >= MinNonClassedParticipants

	if len(readyClasses) == 0 && !readyNonClassed {
		return nil
	}

	var created []*Game

	for i := range StadiumCount {
		stadium := m.stadiums[i]
		if stadium.InUse() {
			continue
		}

		var game *Game

		if i <= NonClassedStadiumEnd {
			// Стадионы 0-10: приоритет NON_CLASSED
			game = m.tryCreateNonClassedLocked(stadium)
			if game == nil {
				game = m.tryCreateClassedLocked(stadium, readyClasses)
			}
		} else {
			// Стадионы 11-21: приоритет CLASSED
			game = m.tryCreateClassedLocked(stadium, readyClasses)
			if game == nil {
				game = m.tryCreateNonClassedLocked(stadium)
			}
		}

		if game != nil {
			stadium.SetInUse(true)
			m.games[stadium.ID()] = game
			created = append(created, game)
		}

		// Обновить readyClasses
		readyClasses = m.hasEnoughClassedLocked()
		readyNonClassed = len(m.nonClassBased) >= MinNonClassedParticipants
		if len(readyClasses) == 0 && !readyNonClassed {
			break
		}
	}

	if len(created) > 0 {
		m.battleStarted.Store(true)
	}

	return created
}

// tryCreateNonClassedLocked пытается создать non-classed матч.
// Caller должен держать mu.Lock().
func (m *Manager) tryCreateNonClassedLocked(stadium *Stadium) *Game {
	if len(m.nonClassBased) < 2 {
		return nil
	}

	opponents, remaining := nextOpponents(m.nonClassBased)
	m.nonClassBased = remaining
	if opponents == nil {
		return nil
	}

	return m.createGame(stadium, CompNonClassed, opponents[0], opponents[1])
}

// tryCreateClassedLocked пытается создать classed матч.
// Caller должен держать mu.Lock().
func (m *Manager) tryCreateClassedLocked(stadium *Stadium, readyClasses []int32) *Game {
	if len(readyClasses) == 0 {
		return nil
	}

	// Случайный выбор класса
	classID := readyClasses[rand.IntN(len(readyClasses))]
	players := m.classBased[classID]
	if len(players) < 2 {
		return nil
	}

	opponents, remaining := nextOpponents(players)
	m.classBased[classID] = remaining
	if len(m.classBased[classID]) == 0 {
		delete(m.classBased, classID)
	}
	if opponents == nil {
		return nil
	}

	return m.createGame(stadium, CompClassed, opponents[0], opponents[1])
}

func (m *Manager) createGame(stadium *Stadium, compType CompetitionType, p1, p2 *model.Player) *Game {
	n1 := m.nobles.Get(p1.CharacterID())
	n2 := m.nobles.Get(p2.CharacterID())
	if n1 == nil || n2 == nil {
		slog.Warn("noble not found for olympiad match",
			"p1", p1.Name(), "n1_exists", n1 != nil,
			"p2", p2.Name(), "n2_exists", n2 != nil)
		return nil
	}

	gameID := m.nextGameID.Add(1)
	part1 := NewParticipant(p1, n1)
	part2 := NewParticipant(p2, n2)

	return NewGame(gameID, stadium, compType, part1, part2)
}

// hasEnoughClassedLocked — locked version of HasEnoughClassed.
func (m *Manager) hasEnoughClassedLocked() []int32 {
	var result []int32
	for classID, players := range m.classBased {
		if len(players) >= MinClassedParticipants {
			result = append(result, classID)
		}
	}
	return result
}

// GetGame returns a game by stadium ID.
func (m *Manager) GetGame(stadiumID int32) *Game {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.games[stadiumID]
}

// RemoveGame removes a completed game and frees the stadium.
func (m *Manager) RemoveGame(stadiumID int32) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.games, stadiumID)

	if int(stadiumID) < StadiumCount {
		m.stadiums[stadiumID].SetInUse(false)
	}

	// Если нет активных игр, сбросить флаг
	if len(m.games) == 0 {
		m.battleStarted.Store(false)
	}
}

// ActiveGameCount returns the number of active games.
func (m *Manager) ActiveGameCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.games)
}

// FindGameByPlayer finds a game containing the given player.
func (m *Manager) FindGameByPlayer(objectID uint32) *Game {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, game := range m.games {
		if game.GetParticipant(objectID) != nil {
			return game
		}
	}
	return nil
}

// Stadium returns a stadium by ID.
func (m *Manager) Stadium(id int32) *Stadium {
	if id < 0 || id >= StadiumCount {
		return nil
	}
	return m.stadiums[id]
}
