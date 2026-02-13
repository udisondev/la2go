package quest

import (
	"maps"
	"sync"
)

// State constants matching L2J QuestState.
const (
	StateCreated   byte = 0
	StateStarted   byte = 1
	StateCompleted byte = 2
)

// ReservedVarCond is the EAV variable name for quest cond (progress step).
const ReservedVarCond = "<state>"

// QuestState tracks a single player's progress in a specific quest.
// EAV model: variables stored as map[string]string, persisted to character_quests.
// Thread-safe via mutex.
type QuestState struct {
	mu sync.RWMutex

	questID   int32
	questName string
	charID    int64
	state     byte              // StateCreated / StateStarted / StateCompleted
	vars      map[string]string // EAV variables
	changed   bool              // dirty flag for persistence
}

// NewQuestState creates a quest state for the given player and quest.
func NewQuestState(questID int32, questName string, charID int64, state byte) *QuestState {
	return &QuestState{
		questID:   questID,
		questName: questName,
		charID:    charID,
		state:     state,
		vars:      make(map[string]string, 4),
	}
}

// QuestID returns the quest identifier.
func (qs *QuestState) QuestID() int32 {
	return qs.questID
}

// QuestName returns the quest name.
func (qs *QuestState) QuestName() string {
	return qs.questName
}

// CharID returns the character ID.
func (qs *QuestState) CharID() int64 {
	return qs.charID
}

// State returns the current quest state.
func (qs *QuestState) State() byte {
	qs.mu.RLock()
	defer qs.mu.RUnlock()
	return qs.state
}

// SetState updates the quest state and marks as changed.
func (qs *QuestState) SetState(state byte) {
	qs.mu.Lock()
	defer qs.mu.Unlock()
	qs.state = state
	qs.changed = true
}

// IsStarted returns true if quest is in progress.
func (qs *QuestState) IsStarted() bool {
	qs.mu.RLock()
	defer qs.mu.RUnlock()
	return qs.state == StateStarted
}

// IsCompleted returns true if quest is finished.
func (qs *QuestState) IsCompleted() bool {
	qs.mu.RLock()
	defer qs.mu.RUnlock()
	return qs.state == StateCompleted
}

// GetCond returns the quest progress step (cond variable).
func (qs *QuestState) GetCond() int {
	qs.mu.RLock()
	defer qs.mu.RUnlock()
	val, ok := qs.vars[ReservedVarCond]
	if !ok {
		return 0
	}
	var n int
	for _, c := range val {
		if c < '0' || c > '9' {
			return 0
		}
		n = n*10 + int(c-'0')
	}
	return n
}

// SetCond sets the quest progress step.
func (qs *QuestState) SetCond(cond int) {
	qs.mu.Lock()
	defer qs.mu.Unlock()
	qs.vars[ReservedVarCond] = intToString(cond)
	qs.changed = true
}

// Get returns a quest variable value.
func (qs *QuestState) Get(key string) string {
	qs.mu.RLock()
	defer qs.mu.RUnlock()
	return qs.vars[key]
}

// Set sets a quest variable.
func (qs *QuestState) Set(key, value string) {
	qs.mu.Lock()
	defer qs.mu.Unlock()
	qs.vars[key] = value
	qs.changed = true
}

// Unset removes a quest variable.
func (qs *QuestState) Unset(key string) {
	qs.mu.Lock()
	defer qs.mu.Unlock()
	delete(qs.vars, key)
	qs.changed = true
}

// Vars returns a snapshot of all variables (copy).
func (qs *QuestState) Vars() map[string]string {
	qs.mu.RLock()
	defer qs.mu.RUnlock()
	snapshot := make(map[string]string, len(qs.vars))
	maps.Copy(snapshot, qs.vars)
	return snapshot
}

// IsChanged returns true if state was modified since last save.
func (qs *QuestState) IsChanged() bool {
	qs.mu.RLock()
	defer qs.mu.RUnlock()
	return qs.changed
}

// ClearChanged resets the dirty flag after successful save.
func (qs *QuestState) ClearChanged() {
	qs.mu.Lock()
	defer qs.mu.Unlock()
	qs.changed = false
}

// intToString converts int to string without fmt.
func intToString(n int) string {
	if n == 0 {
		return "0"
	}
	neg := false
	if n < 0 {
		neg = true
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
