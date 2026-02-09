package ai

import (
	"log/slog"
	"sync/atomic"

	"github.com/udisondev/la2go/internal/model"
)

// BasicNpcAI implements basic NPC AI (MVP: IDLE → ACTIVE state machine)
type BasicNpcAI struct {
	npc       *model.Npc
	isRunning atomic.Bool
	tickCount atomic.Int32
}

// NewBasicNpcAI creates new basic NPC AI
func NewBasicNpcAI(npc *model.Npc) *BasicNpcAI {
	return &BasicNpcAI{
		npc: npc,
	}
}

// Start starts AI controller
func (ai *BasicNpcAI) Start() {
	ai.isRunning.Store(true)
	ai.SetIntention(model.IntentionActive)
	slog.Debug("basic AI started",
		"npc", ai.npc.Name(),
		"objectID", ai.npc.ObjectID(),
		"intention", model.IntentionActive)
}

// Stop stops AI controller
func (ai *BasicNpcAI) Stop() {
	ai.isRunning.Store(false)
	ai.SetIntention(model.IntentionIdle)
	slog.Debug("basic AI stopped",
		"npc", ai.npc.Name(),
		"objectID", ai.npc.ObjectID())
}

// SetIntention sets AI intention
func (ai *BasicNpcAI) SetIntention(intention model.Intention) {
	oldIntention := ai.npc.Intention()
	ai.npc.SetIntention(intention)

	if oldIntention != intention && IsDebugEnabled() {
		slog.Debug("AI intention changed",
			"npc", ai.npc.Name(),
			"objectID", ai.npc.ObjectID(),
			"from", oldIntention,
			"to", intention)
	}
}

// CurrentIntention returns current AI intention
func (ai *BasicNpcAI) CurrentIntention() model.Intention {
	return ai.npc.Intention()
}

// Tick performs AI tick (called every second)
// MVP: Simple state machine IDLE ↔ ACTIVE
func (ai *BasicNpcAI) Tick() {
	if !ai.isRunning.Load() {
		return
	}

	// Increment tick count
	ticks := ai.tickCount.Add(1)

	currentIntention := ai.CurrentIntention()

	// MVP logic: toggle between IDLE and ACTIVE every 5 ticks
	// This is just for demonstration — real AI would be more complex
	if ticks%5 == 0 {
		if currentIntention == model.IntentionActive {
			ai.SetIntention(model.IntentionIdle)
		} else {
			ai.SetIntention(model.IntentionActive)
		}
	}
}
