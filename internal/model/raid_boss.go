package model

import "sync/atomic"

// RaidBossStatus represents the current status of a raid boss.
type RaidBossStatus int32

const (
	RaidStatusAlive    RaidBossStatus = 0
	RaidStatusDead     RaidBossStatus = 1
	RaidStatusFighting RaidBossStatus = 2
)

// RaidBoss represents a raid boss NPC. Extends Monster with raid-specific behavior:
// isRaid=true, lethalable=false, respawn tracked in DB.
//
// Phase 23: Raid Boss System.
// Java reference: RaidBoss.java.
type RaidBoss struct {
	*Monster

	status      atomic.Int32 // RaidBossStatus
	useRaidCurse atomic.Bool  // whether to apply raid curse on low-level attackers
}

// NewRaidBoss creates a new RaidBoss wrapping a Monster.
// Sets isRaid=true and initializes status to ALIVE.
func NewRaidBoss(objectID uint32, templateID int32, template *NpcTemplate) *RaidBoss {
	monster := NewMonster(objectID, templateID, template)

	rb := &RaidBoss{
		Monster: monster,
	}

	rb.status.Store(int32(RaidStatusAlive))
	rb.useRaidCurse.Store(true)

	// Override WorldObject.Data to point to RaidBoss (not Monster)
	// Enables type assertion: target.Data.(*RaidBoss)
	monster.Npc.WorldObject.Data = rb

	return rb
}

// IsRaid always returns true for RaidBoss.
func (rb *RaidBoss) IsRaid() bool {
	return true
}

// IsLethalable returns false — raid bosses are immune to lethal attacks.
func (rb *RaidBoss) IsLethalable() bool {
	return false
}

// Status returns current raid boss status (atomic read).
func (rb *RaidBoss) Status() RaidBossStatus {
	return RaidBossStatus(rb.status.Load())
}

// SetStatus sets raid boss status (atomic write).
func (rb *RaidBoss) SetStatus(s RaidBossStatus) {
	rb.status.Store(int32(s))
}

// UseRaidCurse returns whether raid curse applies to low-level attackers.
func (rb *RaidBoss) UseRaidCurse() bool {
	return rb.useRaidCurse.Load()
}

// SetUseRaidCurse sets raid curse flag.
func (rb *RaidBoss) SetUseRaidCurse(use bool) {
	rb.useRaidCurse.Store(use)
}
