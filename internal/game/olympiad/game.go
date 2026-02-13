package olympiad

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/udisondev/la2go/internal/model"
)

// GameResult определяет результат олимпийского боя.
type GameResult int32

const (
	ResultContinue  GameResult = 0
	ResultPlayer1   GameResult = 1
	ResultPlayer2   GameResult = 2
	ResultDraw      GameResult = 3
	ResultPlayer1DC GameResult = 4 // player1 disconnected
	ResultPlayer2DC GameResult = 5 // player2 disconnected
)

// Participant хранит данные об участнике матча.
type Participant struct {
	Player     *model.Player
	Noble      *Noble
	Name       string
	ObjectID   uint32
	ClassID    int32
	DamageDealt int64
	HPBefore   int32

	// Состояние перед боем (для восстановления)
	savedHP int32
	savedMP int32
	savedCP int32
}

// NewParticipant создаёт участника матча.
func NewParticipant(player *model.Player, noble *Noble) *Participant {
	return &Participant{
		Player:   player,
		Noble:    noble,
		Name:     player.Name(),
		ObjectID: player.ObjectID(),
		ClassID:  noble.ClassID(),
	}
}

// SaveCondition сохраняет текущее состояние для восстановления после боя.
func (p *Participant) SaveCondition() {
	p.savedHP = p.Player.CurrentHP()
	p.savedMP = p.Player.CurrentMP()
	p.savedCP = p.Player.CurrentCP()
}

// RestoreCondition восстанавливает состояние до боя.
func (p *Participant) RestoreCondition() {
	p.Player.SetCurrentHP(p.savedHP)
	p.Player.SetCurrentMP(p.savedMP)
	p.Player.SetCurrentCP(p.savedCP)
}

// Game представляет один олимпийский бой между двумя участниками.
// КРИТИЧНО: battleStarted и gameIsStarted используют atomic.Bool
// (fix Java race condition — static boolean без volatile).
type Game struct {
	id       int32
	stadium  *Stadium
	compType CompetitionType

	player1 *Participant
	player2 *Participant

	// ⚠️ atomic.Bool — fix Java race condition (Olympiad.java lines 86-87)
	battleStarted atomic.Bool
	gameIsStarted atomic.Bool
	gameFinished  atomic.Bool

	startTime time.Time
	endTime   time.Time

	mu       sync.RWMutex
	cancelCh chan struct{}
}

// NewGame создаёт новый олимпийский бой.
func NewGame(id int32, stadium *Stadium, compType CompetitionType, p1, p2 *Participant) *Game {
	return &Game{
		id:       id,
		stadium:  stadium,
		compType: compType,
		player1:  p1,
		player2:  p2,
		cancelCh: make(chan struct{}),
	}
}

// ID returns the game identifier.
func (g *Game) ID() int32 { return g.id }

// Stadium returns the arena where the match takes place.
func (g *Game) Stadium() *Stadium { return g.stadium }

// CompType returns the competition type.
func (g *Game) CompType() CompetitionType { return g.compType }

// Player1 returns the first participant.
func (g *Game) Player1() *Participant { return g.player1 }

// Player2 returns the second participant.
func (g *Game) Player2() *Participant { return g.player2 }

// CancelCh returns the cancellation channel.
func (g *Game) CancelCh() <-chan struct{} { return g.cancelCh }

// BattleStarted reports whether the battle has begun (atomic, thread-safe).
func (g *Game) BattleStarted() bool { return g.battleStarted.Load() }

// SetBattleStarted marks the battle as started (atomic, thread-safe).
func (g *Game) SetBattleStarted() { g.battleStarted.Store(true) }

// GameIsStarted reports whether the game setup is done (atomic, thread-safe).
func (g *Game) GameIsStarted() bool { return g.gameIsStarted.Load() }

// SetGameIsStarted marks the game setup as done (atomic, thread-safe).
func (g *Game) SetGameIsStarted() { g.gameIsStarted.Store(true) }

// IsFinished reports whether the game is over (atomic, thread-safe).
func (g *Game) IsFinished() bool { return g.gameFinished.Load() }

// StartTime returns when the battle started.
func (g *Game) StartTime() time.Time {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.startTime
}

// RemainingTime returns time until battle end.
func (g *Game) RemainingTime() time.Duration {
	g.mu.RLock()
	defer g.mu.RUnlock()
	if g.endTime.IsZero() {
		return BattleDuration
	}
	remaining := time.Until(g.endTime)
	if remaining < 0 {
		return 0
	}
	return remaining
}

// Start запускает бой: сохраняет состояния, запоминает время.
func (g *Game) Start() {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.player1.SaveCondition()
	g.player2.SaveCondition()
	g.player1.HPBefore = g.player1.Player.CurrentHP()
	g.player2.HPBefore = g.player2.Player.CurrentHP()
	g.startTime = time.Now()
	g.endTime = g.startTime.Add(BattleDuration)
	g.battleStarted.Store(true)
}

// CalculateResult определяет результат боя.
// Приоритет: disconnect > HP=0 > damage dealt > draw.
func (g *Game) CalculateResult() GameResult {
	p1Dead := g.player1.Player.CurrentHP() <= 0
	p2Dead := g.player2.Player.CurrentHP() <= 0

	switch {
	case p1Dead && p2Dead:
		return ResultDraw
	case p1Dead:
		return ResultPlayer2
	case p2Dead:
		return ResultPlayer1
	default:
		// По урону
		if g.player1.DamageDealt > g.player2.DamageDealt {
			return ResultPlayer1
		}
		if g.player2.DamageDealt > g.player1.DamageDealt {
			return ResultPlayer2
		}
		return ResultDraw
	}
}

// CheckTimeout проверяет истечение времени боя.
func (g *Game) CheckTimeout() bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	if g.endTime.IsZero() {
		return false
	}
	return time.Now().After(g.endTime)
}

// CalculatePointDiff вычисляет разницу очков для победителя/проигравшего.
// Formula: min(min(p1Points, p2Points) / div, MaxPoints)
func (g *Game) CalculatePointDiff() int32 {
	p1Points := g.player1.Noble.Points()
	p2Points := g.player2.Noble.Points()

	minPoints := p1Points
	if p2Points < minPoints {
		minPoints = p2Points
	}

	div := int32(ClassedScoreDiv)
	if g.compType == CompNonClassed {
		div = NonClassedScoreDiv
	}

	diff := minPoints / div
	if diff > MaxPoints {
		diff = MaxPoints
	}
	if diff < 1 {
		diff = 1
	}
	return diff
}

// CalculateDrawPenalty вычисляет штраф при ничьей.
// Formula: min(points/5, MaxPoints)
func CalculateDrawPenalty(points int32) int32 {
	penalty := points / DrawPenaltyDiv
	if penalty > MaxPoints {
		return MaxPoints
	}
	if penalty < 1 {
		return 1
	}
	return penalty
}

// CalculateDefaultPenalty вычисляет штраф при неявке.
// Formula: min(points/3, MaxPoints)
func CalculateDefaultPenalty(points int32) int32 {
	penalty := points / DefaultPenaltyDiv
	if penalty > MaxPoints {
		return MaxPoints
	}
	if penalty < 1 {
		return 1
	}
	return penalty
}

// ApplyResult применяет результат боя к статистике Noble.
func (g *Game) ApplyResult(result GameResult) {
	diff := g.CalculatePointDiff()

	switch result {
	case ResultPlayer1:
		g.player1.Noble.RecordWin(diff)
		g.player2.Noble.RecordLoss(diff)
	case ResultPlayer2:
		g.player2.Noble.RecordWin(diff)
		g.player1.Noble.RecordLoss(diff)
	case ResultDraw:
		p1Penalty := CalculateDrawPenalty(g.player1.Noble.Points())
		p2Penalty := CalculateDrawPenalty(g.player2.Noble.Points())
		g.player1.Noble.RecordDraw(p1Penalty)
		g.player2.Noble.RecordDraw(p2Penalty)
	case ResultPlayer1DC:
		// Player1 disconnected — player2 wins
		penalty := CalculateDefaultPenalty(g.player1.Noble.Points())
		g.player1.Noble.RecordLoss(penalty)
		g.player2.Noble.RecordWin(diff)
	case ResultPlayer2DC:
		// Player2 disconnected — player1 wins
		penalty := CalculateDefaultPenalty(g.player2.Noble.Points())
		g.player2.Noble.RecordLoss(penalty)
		g.player1.Noble.RecordWin(diff)
	}
}

// Finish завершает бой и освобождает стадион.
func (g *Game) Finish() {
	if !g.gameFinished.CompareAndSwap(false, true) {
		return // уже завершён
	}
	close(g.cancelCh)
	g.stadium.SetInUse(false)
}

// RestoreConditions восстанавливает HP/MP/CP участников.
func (g *Game) RestoreConditions() {
	g.player1.RestoreCondition()
	g.player2.RestoreCondition()
}

// RecordDamage добавляет урон к статистике участника.
func (g *Game) RecordDamage(attackerObjID uint32, damage int64) {
	switch attackerObjID {
	case g.player1.ObjectID:
		g.mu.Lock()
		g.player1.DamageDealt += damage
		g.mu.Unlock()
	case g.player2.ObjectID:
		g.mu.Lock()
		g.player2.DamageDealt += damage
		g.mu.Unlock()
	}
}

// GetOpponent returns the other participant.
func (g *Game) GetOpponent(objectID uint32) *Participant {
	if g.player1.ObjectID == objectID {
		return g.player2
	}
	if g.player2.ObjectID == objectID {
		return g.player1
	}
	return nil
}

// GetParticipant returns the participant by objectID.
func (g *Game) GetParticipant(objectID uint32) *Participant {
	if g.player1.ObjectID == objectID {
		return g.player1
	}
	if g.player2.ObjectID == objectID {
		return g.player2
	}
	return nil
}
