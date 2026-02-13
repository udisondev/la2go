// Package duel implements the Lineage 2 duel system (1v1 and party duels).
// Manages duel lifecycle: request → countdown → fighting → result.
package duel

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/udisondev/la2go/internal/model"
)

// DuelState represents a participant's state in a duel.
type DuelState int32

const (
	StateNoDuel     DuelState = 0 // Not in a duel
	StateDuelling   DuelState = 1 // Fighting
	StateDead       DuelState = 2 // Defeated
	StateWinner     DuelState = 3 // Won
	StateInterrupted DuelState = 4 // Duel interrupted
)

// Result represents the outcome of a duel.
type Result int

const (
	ResultContinue      Result = iota // Duel continues
	ResultTeam1Win                    // Challenger wins
	ResultTeam2Win                    // Opponent wins
	ResultTeam1Surrender              // Challenger surrendered
	ResultTeam2Surrender              // Opponent surrendered
	ResultCanceled                    // Duel canceled (distance, zone, etc.)
	ResultTimeout                     // Time expired (draw)
)

// Duel duration limits.
const (
	PlayerDuelDuration = 120 * time.Second // 2 minutes for 1v1
	PartyDuelDuration  = 300 * time.Second // 5 minutes for party
	CountdownStart     = 5                  // Countdown before fight
	MaxDistance1v1     = 1600                // Max distance for 1v1 (cancel if exceeded)
)

// PlayerCondition stores a participant's state before the duel for restoration after.
type PlayerCondition struct {
	ObjectID uint32
	HP       int32
	MP       int32
	CP       int32
	X, Y, Z  int32
}

// Duel represents an active duel between two players (or parties).
type Duel struct {
	mu sync.RWMutex

	id          int32
	playerA     *model.Player // challenger
	playerB     *model.Player // opponent
	partyDuel   bool
	endTime     time.Time
	countdown   atomic.Int32
	finished    atomic.Bool
	surrenderReq atomic.Int32 // 0=none, 1=team1, 2=team2

	// Participant states saved before duel, restored after.
	conditions map[uint32]*PlayerCondition

	// Participant duel states (objectID → DuelState).
	states map[uint32]*atomic.Int32

	// cancelFunc stops the duel goroutine.
	cancelCh chan struct{}
}

// NewDuel creates a duel between two players.
func NewDuel(id int32, playerA, playerB *model.Player, partyDuel bool) *Duel {
	duration := PlayerDuelDuration
	if partyDuel {
		duration = PartyDuelDuration
	}

	d := &Duel{
		id:         id,
		playerA:    playerA,
		playerB:    playerB,
		partyDuel:  partyDuel,
		endTime:    time.Now().Add(duration),
		conditions: make(map[uint32]*PlayerCondition, 4),
		states:     make(map[uint32]*atomic.Int32, 4),
		cancelCh:   make(chan struct{}),
	}
	d.countdown.Store(CountdownStart)
	return d
}

// ID returns the duel identifier.
func (d *Duel) ID() int32 { return d.id }

// PlayerA returns the challenger.
func (d *Duel) PlayerA() *model.Player { return d.playerA }

// PlayerB returns the opponent.
func (d *Duel) PlayerB() *model.Player { return d.playerB }

// IsPartyDuel returns true if this is a party duel.
func (d *Duel) IsPartyDuel() bool { return d.partyDuel }

// IsFinished returns true if the duel has ended.
func (d *Duel) IsFinished() bool { return d.finished.Load() }

// Countdown returns the current countdown value.
func (d *Duel) Countdown() int32 { return d.countdown.Load() }

// RemainingTime returns time until duel timeout.
func (d *Duel) RemainingTime() time.Duration {
	return time.Until(d.endTime)
}

// SaveCondition saves a player's state for restoration after duel.
func (d *Duel) SaveCondition(p *model.Player) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.conditions[p.ObjectID()] = &PlayerCondition{
		ObjectID: p.ObjectID(),
		HP:       p.CurrentHP(),
		MP:       p.CurrentMP(),
		CP:       p.CurrentCP(),
		X:        p.Location().X,
		Y:        p.Location().Y,
		Z:        p.Location().Z,
	}
}

// SaveAllConditions saves conditions for all duel participants.
func (d *Duel) SaveAllConditions() {
	d.SaveCondition(d.playerA)
	d.SaveCondition(d.playerB)

	if d.partyDuel {
		if pa := d.playerA.GetParty(); pa != nil {
			for _, m := range pa.Members() {
				if m.ObjectID() != d.playerA.ObjectID() {
					d.SaveCondition(m)
				}
			}
		}
		if pb := d.playerB.GetParty(); pb != nil {
			for _, m := range pb.Members() {
				if m.ObjectID() != d.playerB.ObjectID() {
					d.SaveCondition(m)
				}
			}
		}
	}
}

// RestoreConditions restores all participants' HP/MP/CP.
// If abnormalEnd is true, stats are NOT restored (canceled duels).
func (d *Duel) RestoreConditions(abnormalEnd bool) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if abnormalEnd {
		return
	}

	for _, cond := range d.conditions {
		// Условия будут применены через менеджер,
		// который имеет доступ к реальным объектам Player
		_ = cond
	}
}

// SetParticipantState sets the duel state for a participant.
func (d *Duel) SetParticipantState(objectID uint32, state DuelState) {
	d.mu.Lock()
	defer d.mu.Unlock()

	s, ok := d.states[objectID]
	if !ok {
		s = &atomic.Int32{}
		d.states[objectID] = s
	}
	s.Store(int32(state))
}

// ParticipantState returns the duel state for a participant.
func (d *Duel) ParticipantState(objectID uint32) DuelState {
	d.mu.RLock()
	defer d.mu.RUnlock()

	s, ok := d.states[objectID]
	if !ok {
		return StateNoDuel
	}
	return DuelState(s.Load())
}

// InitParticipants sets all participants to StateDuelling.
func (d *Duel) InitParticipants() {
	d.SetParticipantState(d.playerA.ObjectID(), StateDuelling)
	d.SetParticipantState(d.playerB.ObjectID(), StateDuelling)

	if d.partyDuel {
		if pa := d.playerA.GetParty(); pa != nil {
			for _, m := range pa.Members() {
				d.SetParticipantState(m.ObjectID(), StateDuelling)
			}
		}
		if pb := d.playerB.GetParty(); pb != nil {
			for _, m := range pb.Members() {
				d.SetParticipantState(m.ObjectID(), StateDuelling)
			}
		}
	}
}

// OnPlayerDefeat marks a player as defeated and checks win condition.
func (d *Duel) OnPlayerDefeat(objectID uint32) Result {
	d.SetParticipantState(objectID, StateDead)

	if !d.partyDuel {
		// 1v1: opponent wins immediately
		if objectID == d.playerA.ObjectID() {
			d.SetParticipantState(d.playerB.ObjectID(), StateWinner)
			return ResultTeam2Win
		}
		d.SetParticipantState(d.playerA.ObjectID(), StateWinner)
		return ResultTeam1Win
	}

	// Party duel: check if entire team is dead
	if d.isTeamDead(d.playerA) {
		d.setTeamState(d.playerB, StateWinner)
		return ResultTeam2Win
	}
	if d.isTeamDead(d.playerB) {
		d.setTeamState(d.playerA, StateWinner)
		return ResultTeam1Win
	}

	return ResultContinue
}

// Surrender handles a player's surrender.
func (d *Duel) Surrender(objectID uint32) Result {
	isTeam1 := d.isTeam1(objectID)
	if isTeam1 {
		if !d.surrenderReq.CompareAndSwap(0, 1) {
			return ResultContinue
		}
		d.setTeamState(d.playerA, StateDead)
		d.setTeamState(d.playerB, StateWinner)
		return ResultTeam1Surrender
	}
	if !d.surrenderReq.CompareAndSwap(0, 2) {
		return ResultContinue
	}
	d.setTeamState(d.playerB, StateDead)
	d.setTeamState(d.playerA, StateWinner)
	return ResultTeam2Surrender
}

// CheckEndCondition checks whether the duel should end.
func (d *Duel) CheckEndCondition() Result {
	if d.finished.Load() {
		return ResultContinue
	}

	// Surrender check
	switch d.surrenderReq.Load() {
	case 1:
		return ResultTeam1Surrender
	case 2:
		return ResultTeam2Surrender
	}

	// Timeout
	if time.Now().After(d.endTime) {
		return ResultTimeout
	}

	// Winner check
	if d.ParticipantState(d.playerA.ObjectID()) == StateWinner {
		return ResultTeam1Win
	}
	if d.ParticipantState(d.playerB.ObjectID()) == StateWinner {
		return ResultTeam2Win
	}

	// 1v1 specific checks
	if !d.partyDuel {
		// Distance check
		locA := d.playerA.Location()
		locB := d.playerB.Location()
		dx := int64(locA.X - locB.X)
		dy := int64(locA.Y - locB.Y)
		if dx*dx+dy*dy > MaxDistance1v1*MaxDistance1v1 {
			return ResultCanceled
		}

		// Interrupted check
		if d.ParticipantState(d.playerA.ObjectID()) == StateInterrupted ||
			d.ParticipantState(d.playerB.ObjectID()) == StateInterrupted {
			return ResultCanceled
		}
	}

	return ResultContinue
}

// Finish marks the duel as finished.
func (d *Duel) Finish() {
	d.finished.Store(true)
	close(d.cancelCh)
}

// CancelCh returns the channel that closes when duel is finished.
func (d *Duel) CancelCh() <-chan struct{} {
	return d.cancelCh
}

// GetCondition returns the saved condition for a player.
func (d *Duel) GetCondition(objectID uint32) *PlayerCondition {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.conditions[objectID]
}

// Participants returns all player objectIDs in the duel.
func (d *Duel) Participants() []uint32 {
	d.mu.RLock()
	defer d.mu.RUnlock()

	ids := make([]uint32, 0, len(d.states))
	for id := range d.states {
		ids = append(ids, id)
	}
	return ids
}

// isTeam1 checks if objectID belongs to team 1 (playerA's team).
func (d *Duel) isTeam1(objectID uint32) bool {
	if objectID == d.playerA.ObjectID() {
		return true
	}
	if d.partyDuel {
		if pa := d.playerA.GetParty(); pa != nil {
			for _, m := range pa.Members() {
				if m.ObjectID() == objectID {
					return true
				}
			}
		}
	}
	return false
}

// isTeamDead checks if all members of a player's team are dead.
func (d *Duel) isTeamDead(leader *model.Player) bool {
	if d.ParticipantState(leader.ObjectID()) != StateDead {
		return false
	}
	if !d.partyDuel {
		return true
	}
	if pa := leader.GetParty(); pa != nil {
		for _, m := range pa.Members() {
			if d.ParticipantState(m.ObjectID()) != StateDead {
				return false
			}
		}
	}
	return true
}

// setTeamState sets state for all members of a player's team.
func (d *Duel) setTeamState(leader *model.Player, state DuelState) {
	d.SetParticipantState(leader.ObjectID(), state)
	if d.partyDuel {
		if pa := leader.GetParty(); pa != nil {
			for _, m := range pa.Members() {
				d.SetParticipantState(m.ObjectID(), state)
			}
		}
	}
}
