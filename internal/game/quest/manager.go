package quest

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/udisondev/la2go/internal/model"
)

// QuestRepository defines the interface for quest data persistence.
// Implemented in the db package.
type QuestRepository interface {
	LoadByCharacterID(ctx context.Context, charID int64) ([]QuestVar, error)
	SaveQuestState(ctx context.Context, charID int64, questName string, vars map[string]string) error
	DeleteQuest(ctx context.Context, charID int64, questName string) error
}

// QuestVar represents a single quest variable row from the database.
type QuestVar struct {
	QuestName string
	Variable  string
	Value     string
}

// Manager manages quest registration, event dispatch, and player quest states.
// Thread-safe for concurrent access.
type Manager struct {
	mu sync.RWMutex

	// Registered quests
	questsByID   map[int32]*Quest  // questID → Quest
	questsByName map[string]*Quest // questName → Quest

	// NPC event index: eventType → npcTemplateID → []*Quest
	npcIndex map[EventType]map[int32][]*Quest

	// Player quest states: charID → questName → *QuestState
	playerStates map[int64]map[string]*QuestState

	repo   QuestRepository
	timers *TimerManager
}

// NewManager creates a new quest manager.
func NewManager(repo QuestRepository) *Manager {
	return &Manager{
		questsByID:   make(map[int32]*Quest, 64),
		questsByName: make(map[string]*Quest, 64),
		npcIndex:     make(map[EventType]map[int32][]*Quest),
		playerStates: make(map[int64]map[string]*QuestState, 256),
		repo:         repo,
		timers:       NewTimerManager(),
	}
}

// RegisterQuest adds a quest to the manager and builds event indexes.
func (m *Manager) RegisterQuest(q *Quest) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.questsByID[q.id]; exists {
		return fmt.Errorf("quest ID %d already registered", q.id)
	}
	if _, exists := m.questsByName[q.name]; exists {
		return fmt.Errorf("quest %q already registered", q.name)
	}

	m.questsByID[q.id] = q
	m.questsByName[q.name] = q

	// Строим индекс NPC→quest для каждого типа события
	eventTypes := []EventType{EventTalk, EventFirstTalk, EventKill, EventAttack, EventSpawn, EventSkillSee, EventAggro}
	for _, et := range eventTypes {
		npcIDs := q.RegisteredNPCs(et)
		if len(npcIDs) == 0 {
			continue
		}
		if m.npcIndex[et] == nil {
			m.npcIndex[et] = make(map[int32][]*Quest, 16)
		}
		for _, npcID := range npcIDs {
			m.npcIndex[et][npcID] = append(m.npcIndex[et][npcID], q)
		}
	}

	slog.Debug("quest registered",
		"questID", q.id,
		"questName", q.name)

	return nil
}

// GetQuest returns a quest by ID.
func (m *Manager) GetQuest(questID int32) *Quest {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.questsByID[questID]
}

// GetQuestByName returns a quest by name.
func (m *Manager) GetQuestByName(name string) *Quest {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.questsByName[name]
}

// QuestCount returns the number of registered quests.
func (m *Manager) QuestCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.questsByID)
}

// LoadPlayerQuests loads quest states for a player from the database.
// Called on player login.
func (m *Manager) LoadPlayerQuests(ctx context.Context, charID int64) error {
	if m.repo == nil {
		return nil
	}

	vars, err := m.repo.LoadByCharacterID(ctx, charID)
	if err != nil {
		return fmt.Errorf("loading quests for character %d: %w", charID, err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	states := make(map[string]*QuestState, len(vars)/2+1)

	for _, v := range vars {
		q, ok := m.questsByName[v.QuestName]
		if !ok {
			slog.Warn("loaded quest state for unregistered quest",
				"questName", v.QuestName,
				"characterID", charID)
			continue
		}

		qs, exists := states[v.QuestName]
		if !exists {
			qs = NewQuestState(q.id, v.QuestName, charID, StateCreated)
			states[v.QuestName] = qs
		}

		qs.vars[v.Variable] = v.Value

		// Восстанавливаем состояние из переменной <state>
		if v.Variable == ReservedVarCond {
			switch v.Value {
			case "0":
				qs.state = StateCreated
			case "2":
				qs.state = StateCompleted
			default:
				qs.state = StateStarted
			}
		}
	}

	m.playerStates[charID] = states

	slog.Debug("loaded player quests",
		"characterID", charID,
		"questCount", len(states))

	return nil
}

// GetQuestState returns a player's state for a specific quest.
func (m *Manager) GetQuestState(charID int64, questName string) *QuestState {
	m.mu.RLock()
	defer m.mu.RUnlock()
	states := m.playerStates[charID]
	if states == nil {
		return nil
	}
	return states[questName]
}

// GetActiveQuests returns all active (started) quest states for a player.
func (m *Manager) GetActiveQuests(charID int64) []*QuestState {
	m.mu.RLock()
	defer m.mu.RUnlock()

	states := m.playerStates[charID]
	if len(states) == 0 {
		return nil
	}

	active := make([]*QuestState, 0, len(states))
	for _, qs := range states {
		if qs.State() == StateStarted {
			active = append(active, qs)
		}
	}
	return active
}

// StartQuest creates a new quest state and sets it to STARTED.
func (m *Manager) StartQuest(charID int64, questID int32) (*QuestState, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	q, ok := m.questsByID[questID]
	if !ok {
		return nil, fmt.Errorf("quest %d not registered", questID)
	}

	states := m.playerStates[charID]
	if states == nil {
		states = make(map[string]*QuestState, 4)
		m.playerStates[charID] = states
	}

	if existing, ok := states[q.name]; ok && existing.State() == StateStarted {
		return existing, nil // уже начат
	}

	qs := NewQuestState(questID, q.name, charID, StateStarted)
	qs.SetCond(1)
	states[q.name] = qs

	return qs, nil
}

// ExitQuest completes or abandons a quest, removes quest items.
func (m *Manager) ExitQuest(charID int64, questName string, completed bool) error {
	m.mu.Lock()
	states := m.playerStates[charID]
	if states == nil {
		m.mu.Unlock()
		return nil
	}
	qs, ok := states[questName]
	if !ok {
		m.mu.Unlock()
		return nil
	}

	if completed {
		qs.SetState(StateCompleted)
	} else {
		delete(states, questName)
	}
	m.mu.Unlock()

	// Удаляем данные из БД при отмене
	if !completed && m.repo != nil {
		if err := m.repo.DeleteQuest(context.Background(), charID, questName); err != nil {
			return fmt.Errorf("deleting quest %q for character %d: %w", questName, charID, err)
		}
	}

	// Отменяем таймеры этого квеста для игрока
	m.timers.CancelAllForQuest(questName)

	return nil
}

// SavePlayerQuests saves all changed quest states for a player.
func (m *Manager) SavePlayerQuests(ctx context.Context, charID int64) error {
	if m.repo == nil {
		return nil
	}

	m.mu.RLock()
	states := m.playerStates[charID]
	// Собираем изменённые квесты
	var toSave []*QuestState
	for _, qs := range states {
		if qs.IsChanged() {
			toSave = append(toSave, qs)
		}
	}
	m.mu.RUnlock()

	for _, qs := range toSave {
		vars := qs.Vars()
		if err := m.repo.SaveQuestState(ctx, charID, qs.QuestName(), vars); err != nil {
			return fmt.Errorf("saving quest %q for character %d: %w", qs.QuestName(), charID, err)
		}
		qs.ClearChanged()
	}

	return nil
}

// UnloadPlayer removes a player's quest states from memory.
// Called on player logout after saving.
func (m *Manager) UnloadPlayer(charID int64) {
	m.mu.Lock()
	delete(m.playerStates, charID)
	m.mu.Unlock()
}

// DispatchEvent fires quest hooks for a game event.
// Returns the HTML string from the first matching quest that returns non-empty.
func (m *Manager) DispatchEvent(event *Event) string {
	if event.Player == nil {
		return ""
	}
	charID := event.Player.CharacterID()

	m.mu.RLock()

	// Получаем квесты, зарегистрированные для этого NPC и типа события
	var quests []*Quest
	if npcMap, ok := m.npcIndex[event.Type]; ok {
		quests = npcMap[event.NpcID]
	}

	// Для глобальных событий (ItemUse, EnterZone, ExitZone) проверяем все квесты
	if len(quests) == 0 && (event.Type == EventItemUse || event.Type == EventEnterZone || event.Type == EventExitZone) {
		for _, q := range m.questsByID {
			if q.GetHook(event.Type, event.NpcID) != nil {
				quests = append(quests, q)
			}
		}
	}

	// Получаем состояния квестов игрока
	playerStates := m.playerStates[charID]

	m.mu.RUnlock()

	if len(quests) == 0 {
		return ""
	}

	for _, q := range quests {
		hook := q.GetHook(event.Type, event.NpcID)
		if hook == nil {
			continue
		}

		// Получаем или создаём QuestState
		var qs *QuestState
		if playerStates != nil {
			qs = playerStates[q.name]
		}
		if qs == nil {
			qs = NewQuestState(q.id, q.name, charID, StateCreated)
		}

		result := hook(event, qs)

		// Автоматически сохраняем QuestState, если hook его изменил
		// (например, перевёл квест в STARTED через SetState/SetCond)
		if qs.IsChanged() {
			m.mu.Lock()
			if m.playerStates[charID] == nil {
				m.playerStates[charID] = make(map[string]*QuestState, 4)
			}
			m.playerStates[charID][q.name] = qs
			m.mu.Unlock()
		}

		if result != "" {
			return result
		}
	}

	return ""
}

// GetQuestsForNPC returns quests that have hooks for an NPC (for quest marks).
func (m *Manager) GetQuestsForNPC(npcID int32) []*Quest {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Проверяем onTalk, т.к. это основной тип взаимодействия
	if npcMap, ok := m.npcIndex[EventTalk]; ok {
		return npcMap[npcID]
	}
	return nil
}

// RemoveQuestItems removes all quest-specific items from a player's inventory.
func (m *Manager) RemoveQuestItems(player *model.Player, questName string) {
	m.mu.RLock()
	q := m.questsByName[questName]
	m.mu.RUnlock()

	if q == nil {
		return
	}

	inv := player.Inventory()
	for _, itemID := range q.questItems {
		count := inv.CountItemsByID(itemID)
		if count > 0 {
			inv.RemoveItemsByID(itemID, count)
		}
	}
}

// TimerManager returns the internal timer manager.
func (m *Manager) TimerManager() *TimerManager {
	return m.timers
}

// Shutdown stops all timers and cleans up resources.
func (m *Manager) Shutdown() {
	m.timers.Shutdown()
}
