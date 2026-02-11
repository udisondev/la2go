package model

import (
	"sync"
	"sync/atomic"
)

// AggroInfo tracks hate and damage from a single attacker.
// Phase 5.7: NPC Aggro & Basic AI.
// Java reference: AggroInfo.java
type AggroInfo struct {
	hate   atomic.Int64
	damage atomic.Int64
}

// Hate returns current hate value (atomic read).
func (a *AggroInfo) Hate() int64 {
	return a.hate.Load()
}

// AddHate adds hate value (atomic).
func (a *AggroInfo) AddHate(amount int64) {
	a.hate.Add(amount)
}

// Damage returns total damage dealt (atomic read).
func (a *AggroInfo) Damage() int64 {
	return a.damage.Load()
}

// AddDamage adds damage value (atomic).
func (a *AggroInfo) AddDamage(amount int64) {
	a.damage.Add(amount)
}

// AggroList manages hate for an NPC against multiple attackers.
// Thread-safe via sync.Map.
// Phase 5.7: NPC Aggro & Basic AI.
// Java reference: Attackable.addDamageHate(), AggroInfo
type AggroList struct {
	entries sync.Map // map[uint32]*AggroInfo — objectID -> AggroInfo
}

// NewAggroList creates a new empty AggroList.
func NewAggroList() *AggroList {
	return &AggroList{}
}

// AddHate adds hate for an attacker. Creates entry if not exists.
// Hate formula from Java: hateValue = (damage * 100) / (npcLevel + 7)
// Caller should compute hate value before calling this.
func (l *AggroList) AddHate(objectID uint32, hate int64) {
	info := l.getOrCreate(objectID)
	info.AddHate(hate)
}

// AddDamage records damage from an attacker. Creates entry if not exists.
func (l *AggroList) AddDamage(objectID uint32, damage int64) {
	info := l.getOrCreate(objectID)
	info.AddDamage(damage)
}

// GetMostHated returns objectID of the attacker with highest hate.
// Returns 0 if list is empty.
func (l *AggroList) GetMostHated() uint32 {
	var maxHate int64
	var mostHatedID uint32

	l.entries.Range(func(key, value any) bool {
		objectID := key.(uint32)
		info := value.(*AggroInfo)
		hate := info.Hate()

		if hate > maxHate || mostHatedID == 0 {
			maxHate = hate
			mostHatedID = objectID
		}
		return true
	})

	return mostHatedID
}

// Get returns AggroInfo for a specific attacker.
// Returns nil if not found.
func (l *AggroList) Get(objectID uint32) *AggroInfo {
	value, ok := l.entries.Load(objectID)
	if !ok {
		return nil
	}
	return value.(*AggroInfo)
}

// Remove removes an attacker from the hate list.
func (l *AggroList) Remove(objectID uint32) {
	l.entries.Delete(objectID)
}

// Clear removes all entries from the hate list.
func (l *AggroList) Clear() {
	l.entries.Range(func(key, _ any) bool {
		l.entries.Delete(key)
		return true
	})
}

// IsEmpty returns true if hate list has no entries.
func (l *AggroList) IsEmpty() bool {
	empty := true
	l.entries.Range(func(_, _ any) bool {
		empty = false
		return false // stop iteration
	})
	return empty
}

// getOrCreate returns existing AggroInfo or creates a new one.
// Fast path: Load() first to avoid allocating &AggroInfo{} on every call.
func (l *AggroList) getOrCreate(objectID uint32) *AggroInfo {
	if v, ok := l.entries.Load(objectID); ok {
		return v.(*AggroInfo)
	}
	v, _ := l.entries.LoadOrStore(objectID, &AggroInfo{})
	return v.(*AggroInfo)
}

// CalcHateValue calculates hate from damage using Java formula.
// Formula: (damage * 100) / (npcLevel + 7)
func CalcHateValue(damage int32, npcLevel int32) int64 {
	if npcLevel < 1 {
		npcLevel = 1
	}
	return (int64(damage) * 100) / int64(npcLevel+7)
}
