package model

import "sync/atomic"

// GrandBossStatus represents the state of a grand boss.
// Managed externally by GrandBossManager (not by the boss itself).
type GrandBossStatus int32

const (
	GrandBossAlive    GrandBossStatus = 0
	GrandBossDead     GrandBossStatus = 1
	GrandBossFighting GrandBossStatus = 2
	// Additional statuses used by specific bosses (Antharas wait, etc.)
	GrandBossWaiting GrandBossStatus = 3
)

// GrandBoss represents a grand boss NPC (Antharas, Valakas, Baium, Zaken, etc.).
// Extends Monster with grand-boss-specific behavior:
// isRaid=true, lethalable=false, status managed by GrandBossManager.
//
// Phase 23: Raid Boss System.
// Java reference: GrandBoss.java.
type GrandBoss struct {
	*Monster

	status atomic.Int32 // GrandBossStatus
	bossID int32        // unique boss identifier (different from npcID/templateID)
}

// NewGrandBoss creates a new GrandBoss wrapping a Monster.
// bossID is the unique grand boss identifier used by GrandBossManager.
func NewGrandBoss(objectID uint32, templateID int32, template *NpcTemplate, bossID int32) *GrandBoss {
	monster := NewMonster(objectID, templateID, template)

	gb := &GrandBoss{
		Monster: monster,
		bossID:  bossID,
	}

	gb.status.Store(int32(GrandBossAlive))

	// Override WorldObject.Data to point to GrandBoss (not Monster)
	monster.Npc.WorldObject.Data = gb

	return gb
}

// IsRaid always returns true for GrandBoss.
func (gb *GrandBoss) IsRaid() bool {
	return true
}

// IsLethalable returns false — grand bosses are immune to lethal attacks.
func (gb *GrandBoss) IsLethalable() bool {
	return false
}

// Status returns current grand boss status (atomic read).
func (gb *GrandBoss) Status() GrandBossStatus {
	return GrandBossStatus(gb.status.Load())
}

// SetStatus sets grand boss status (atomic write).
func (gb *GrandBoss) SetStatus(s GrandBossStatus) {
	gb.status.Store(int32(s))
}

// BossID returns the unique grand boss identifier.
func (gb *GrandBoss) BossID() int32 {
	return gb.bossID
}
