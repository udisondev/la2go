package olympiad

import "sync"

// Noble хранит олимпийскую статистику одного благородного персонажа.
// Потокобезопасный: используется sync.RWMutex.
type Noble struct {
	mu sync.RWMutex

	charID  int64
	classID int32
	points  int32
	compDone int32
	compWon  int32
	compLost int32
	compDrawn int32
}

// NewNoble создаёт запись Noble с начальными очками.
func NewNoble(charID int64, classID int32) *Noble {
	return &Noble{
		charID:  charID,
		classID: classID,
		points:  StartPoints,
	}
}

// CharID returns the character DB ID.
func (n *Noble) CharID() int64 {
	return n.charID
}

// ClassID returns the character's class ID.
func (n *Noble) ClassID() int32 {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.classID
}

// SetClassID updates the noble's class ID (after class transfer).
func (n *Noble) SetClassID(id int32) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.classID = id
}

// Points returns current olympiad points.
func (n *Noble) Points() int32 {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.points
}

// SetPoints sets olympiad points (used for DB load).
func (n *Noble) SetPoints(pts int32) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.points = pts
}

// AddPoints adjusts points by delta (can be negative).
func (n *Noble) AddPoints(delta int32) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.points += delta
	if n.points < 0 {
		n.points = 0
	}
}

// CompDone returns total completed matches.
func (n *Noble) CompDone() int32 {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.compDone
}

// CompWon returns total won matches.
func (n *Noble) CompWon() int32 {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.compWon
}

// CompLost returns total lost matches.
func (n *Noble) CompLost() int32 {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.compLost
}

// CompDrawn returns total drawn matches.
func (n *Noble) CompDrawn() int32 {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.compDrawn
}

// Stats returns a snapshot of all stats.
func (n *Noble) Stats() NobleStats {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return NobleStats{
		CharID:    n.charID,
		ClassID:   n.classID,
		Points:    n.points,
		CompDone:  n.compDone,
		CompWon:   n.compWon,
		CompLost:  n.compLost,
		CompDrawn: n.compDrawn,
	}
}

// RecordWin записывает победу.
func (n *Noble) RecordWin(pointsGain int32) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.compDone++
	n.compWon++
	n.points += pointsGain
}

// RecordLoss записывает поражение.
func (n *Noble) RecordLoss(pointsLoss int32) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.compDone++
	n.compLost++
	n.points -= pointsLoss
	if n.points < 0 {
		n.points = 0
	}
}

// RecordDraw записывает ничью.
func (n *Noble) RecordDraw(pointsLoss int32) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.compDone++
	n.compDrawn++
	n.points -= pointsLoss
	if n.points < 0 {
		n.points = 0
	}
}

// GrantWeeklyPoints начисляет еженедельные очки с учётом лимита.
// Лимит: (compDone * 10) + 12
func (n *Noble) GrantWeeklyPoints() {
	n.mu.Lock()
	defer n.mu.Unlock()
	cap := n.compDone*10 + 12
	if n.points < cap {
		n.points += WeeklyPoints
		if n.points > cap {
			n.points = cap
		}
	}
}

// LoadStats loads stats from DB (bypasses validation).
func (n *Noble) LoadStats(stats NobleStats) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.classID = stats.ClassID
	n.points = stats.Points
	n.compDone = stats.CompDone
	n.compWon = stats.CompWon
	n.compLost = stats.CompLost
	n.compDrawn = stats.CompDrawn
}

// NobleStats — иммутабельный снимок статистики Noble.
type NobleStats struct {
	CharID    int64
	ClassID   int32
	Points    int32
	CompDone  int32
	CompWon   int32
	CompLost  int32
	CompDrawn int32
}

// NobleTable хранит всех зарегистрированных nobles.
// Потокобезопасна через sync.RWMutex.
type NobleTable struct {
	mu     sync.RWMutex
	nobles map[int64]*Noble // charID → Noble
}

// NewNobleTable creates an empty NobleTable.
func NewNobleTable() *NobleTable {
	return &NobleTable{
		nobles: make(map[int64]*Noble),
	}
}

// Register добавляет нового noble или возвращает существующего.
func (t *NobleTable) Register(charID int64, classID int32) *Noble {
	t.mu.Lock()
	defer t.mu.Unlock()
	if n, ok := t.nobles[charID]; ok {
		return n
	}
	n := NewNoble(charID, classID)
	t.nobles[charID] = n
	return n
}

// Get возвращает Noble по charID (nil если не зарегистрирован).
func (t *NobleTable) Get(charID int64) *Noble {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.nobles[charID]
}

// Remove удаляет Noble (для тестов/admin).
func (t *NobleTable) Remove(charID int64) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.nobles, charID)
}

// All возвращает snapshot всех nobles.
func (t *NobleTable) All() []NobleStats {
	t.mu.RLock()
	defer t.mu.RUnlock()
	result := make([]NobleStats, 0, len(t.nobles))
	for _, n := range t.nobles {
		result = append(result, n.Stats())
	}
	return result
}

// Count returns the number of registered nobles.
func (t *NobleTable) Count() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return len(t.nobles)
}

// ByClassID возвращает список nobles для указанного класса.
func (t *NobleTable) ByClassID(classID int32) []*Noble {
	t.mu.RLock()
	defer t.mu.RUnlock()
	var result []*Noble
	for _, n := range t.nobles {
		if n.ClassID() == classID {
			result = append(result, n)
		}
	}
	return result
}

// GrantAllWeeklyPoints начисляет еженедельные очки всем nobles.
func (t *NobleTable) GrantAllWeeklyPoints() {
	t.mu.RLock()
	defer t.mu.RUnlock()
	for _, n := range t.nobles {
		n.GrantWeeklyPoints()
	}
}
