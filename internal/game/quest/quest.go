// Package quest implements the quest system framework for the Lineage 2 server.
// Provides Quest registration, QuestState tracking, event dispatch, and timer management.
// Quests register hook functions that fire on game events (NPC talk, kill, attack, etc.).
package quest

import (
	"github.com/udisondev/la2go/internal/model"
)

// EventType identifies the kind of quest event.
type EventType int

const (
	EventTalk       EventType = iota // NPC dialog interaction
	EventFirstTalk                   // First time talking to NPC (before dialog)
	EventKill                        // NPC killed
	EventAttack                      // NPC attacked
	EventSpawn                       // NPC spawned
	EventSkillSee                    // Skill used nearby
	EventAggro                       // NPC gained aggro on player
	EventItemUse                     // Quest item used
	EventEnterZone                   // Player entered zone
	EventExitZone                    // Player exited zone
)

// Event carries quest event data to hook functions.
type Event struct {
	Type     EventType
	Player   *model.Player
	NpcID    int32          // Template ID of the NPC involved (0 if none)
	TargetID uint32         // Object ID of target (NPC/player) (0 if none)
	SkillID  int32          // Skill used (for EventSkillSee)
	IsPet    bool           // True if action from pet
	Params   map[string]any // Additional parameters
}

// HookFunc is the callback signature for quest event handlers.
// Returns the HTML response to show to the player (empty string = no response).
type HookFunc func(event *Event, qs *QuestState) string

// Quest defines a quest with event hooks.
// Each quest has a unique ID and name, plus sets of NPC IDs it reacts to.
type Quest struct {
	id   int32
	name string

	// Event hooks: NPC template ID → handler function
	onTalk      map[int32]HookFunc
	onFirstTalk map[int32]HookFunc
	onKill      map[int32]HookFunc
	onAttack    map[int32]HookFunc
	onSpawn     map[int32]HookFunc
	onSkillSee  map[int32]HookFunc
	onAggro     map[int32]HookFunc

	// Global hooks (no NPC filter)
	onItemUse   HookFunc
	onEnterZone HookFunc
	onExitZone  HookFunc

	// Quest item IDs that should be removed on quest exit
	questItems []int32
}

// NewQuest creates a new quest definition.
func NewQuest(id int32, name string) *Quest {
	return &Quest{
		id:          id,
		name:        name,
		onTalk:      make(map[int32]HookFunc, 4),
		onFirstTalk: make(map[int32]HookFunc, 2),
		onKill:      make(map[int32]HookFunc, 4),
		onAttack:    make(map[int32]HookFunc, 2),
		onSpawn:     make(map[int32]HookFunc, 2),
		onSkillSee:  make(map[int32]HookFunc, 2),
		onAggro:     make(map[int32]HookFunc, 2),
	}
}

// ID returns the quest identifier.
func (q *Quest) ID() int32 { return q.id }

// Name returns the quest name.
func (q *Quest) Name() string { return q.name }

// AddTalkID registers an onTalk hook for an NPC template ID.
func (q *Quest) AddTalkID(npcID int32, fn HookFunc) {
	q.onTalk[npcID] = fn
}

// AddFirstTalkID registers an onFirstTalk hook for an NPC template ID.
func (q *Quest) AddFirstTalkID(npcID int32, fn HookFunc) {
	q.onFirstTalk[npcID] = fn
}

// AddKillID registers an onKill hook for an NPC template ID.
func (q *Quest) AddKillID(npcID int32, fn HookFunc) {
	q.onKill[npcID] = fn
}

// AddAttackID registers an onAttack hook for an NPC template ID.
func (q *Quest) AddAttackID(npcID int32, fn HookFunc) {
	q.onAttack[npcID] = fn
}

// AddSpawnID registers an onSpawn hook for an NPC template ID.
func (q *Quest) AddSpawnID(npcID int32, fn HookFunc) {
	q.onSpawn[npcID] = fn
}

// AddSkillSeeID registers an onSkillSee hook for an NPC template ID.
func (q *Quest) AddSkillSeeID(npcID int32, fn HookFunc) {
	q.onSkillSee[npcID] = fn
}

// AddAggroID registers an onAggro hook for an NPC template ID.
func (q *Quest) AddAggroID(npcID int32, fn HookFunc) {
	q.onAggro[npcID] = fn
}

// SetOnItemUse sets a global item use handler.
func (q *Quest) SetOnItemUse(fn HookFunc) {
	q.onItemUse = fn
}

// SetOnEnterZone sets a global zone enter handler.
func (q *Quest) SetOnEnterZone(fn HookFunc) {
	q.onEnterZone = fn
}

// SetOnExitZone sets a global zone exit handler.
func (q *Quest) SetOnExitZone(fn HookFunc) {
	q.onExitZone = fn
}

// AddQuestItem adds an item ID that should be removed when quest exits.
func (q *Quest) AddQuestItem(itemID int32) {
	q.questItems = append(q.questItems, itemID)
}

// QuestItems returns the list of item IDs to remove on quest exit.
func (q *Quest) QuestItems() []int32 {
	return q.questItems
}

// GetHook returns the appropriate hook function for the given event type and NPC.
// Returns nil if no hook is registered.
func (q *Quest) GetHook(eventType EventType, npcID int32) HookFunc {
	switch eventType {
	case EventTalk:
		return q.onTalk[npcID]
	case EventFirstTalk:
		return q.onFirstTalk[npcID]
	case EventKill:
		return q.onKill[npcID]
	case EventAttack:
		return q.onAttack[npcID]
	case EventSpawn:
		return q.onSpawn[npcID]
	case EventSkillSee:
		return q.onSkillSee[npcID]
	case EventAggro:
		return q.onAggro[npcID]
	case EventItemUse:
		return q.onItemUse
	case EventEnterZone:
		return q.onEnterZone
	case EventExitZone:
		return q.onExitZone
	default:
		return nil
	}
}

// HasHook returns true if the quest has a hook for the given event type and NPC.
func (q *Quest) HasHook(eventType EventType, npcID int32) bool {
	return q.GetHook(eventType, npcID) != nil
}

// RegisteredNPCs returns all NPC IDs that have hooks for the given event type.
func (q *Quest) RegisteredNPCs(eventType EventType) []int32 {
	var m map[int32]HookFunc
	switch eventType {
	case EventTalk:
		m = q.onTalk
	case EventFirstTalk:
		m = q.onFirstTalk
	case EventKill:
		m = q.onKill
	case EventAttack:
		m = q.onAttack
	case EventSpawn:
		m = q.onSpawn
	case EventSkillSee:
		m = q.onSkillSee
	case EventAggro:
		m = q.onAggro
	default:
		return nil
	}

	ids := make([]int32, 0, len(m))
	for id := range m {
		ids = append(ids, id)
	}
	return ids
}
