package zone

import (
	"strconv"
	"sync"
	"time"
)

// BossZone represents a raid/grand boss area with player whitelist.
// Java reference: BossZone.java — does NOT set any ZoneId flags.
// Players not in whitelist are teleported out.
type BossZone struct {
	*BaseZone
	timeInvade time.Duration
	oustX      int32
	oustY      int32
	oustZ      int32

	mu                     sync.RWMutex
	playersAllowed         map[uint32]struct{}        // objectID → present
	playerAllowedReEntry   map[uint32]int64           // objectID → expiration (unix ms)
}

// NewBossZone creates a BossZone.
func NewBossZone(base *BaseZone) *BossZone {
	z := &BossZone{
		BaseZone:             base,
		playersAllowed:       make(map[uint32]struct{}),
		playerAllowedReEntry: make(map[uint32]int64),
	}
	if v, ok := base.params["InvadeTime"]; ok {
		if n, err := strconv.Atoi(v); err == nil {
			z.timeInvade = time.Duration(n) * time.Millisecond
		}
	}
	if v, ok := base.params["oustX"]; ok {
		if n, err := strconv.Atoi(v); err == nil {
			z.oustX = int32(n)
		}
	}
	if v, ok := base.params["oustY"]; ok {
		if n, err := strconv.Atoi(v); err == nil {
			z.oustY = int32(n)
		}
	}
	if v, ok := base.params["oustZ"]; ok {
		if n, err := strconv.Atoi(v); err == nil {
			z.oustZ = int32(n)
		}
	}
	return z
}

// IsPeace returns false.
func (z *BossZone) IsPeace() bool { return false }

// AllowsPvP returns true.
func (z *BossZone) AllowsPvP() bool { return true }

// AllowPlayerEntry adds a player to the whitelist with a duration.
func (z *BossZone) AllowPlayerEntry(objectID uint32, duration time.Duration) {
	z.mu.Lock()
	z.playersAllowed[objectID] = struct{}{}
	z.playerAllowedReEntry[objectID] = time.Now().Add(duration).UnixMilli()
	z.mu.Unlock()
}

// RemovePlayer removes a player from the whitelist.
func (z *BossZone) RemovePlayer(objectID uint32) {
	z.mu.Lock()
	delete(z.playersAllowed, objectID)
	delete(z.playerAllowedReEntry, objectID)
	z.mu.Unlock()
}

// IsPlayerAllowed checks if a player is in the whitelist.
func (z *BossZone) IsPlayerAllowed(objectID uint32) bool {
	z.mu.RLock()
	defer z.mu.RUnlock()
	_, ok := z.playersAllowed[objectID]
	return ok
}

// OustLocation returns the teleport-out coordinates.
func (z *BossZone) OustLocation() (int32, int32, int32) {
	return z.oustX, z.oustY, z.oustZ
}

// SetAllowedPlayers sets the full whitelist.
func (z *BossZone) SetAllowedPlayers(objectIDs []uint32) {
	z.mu.Lock()
	z.playersAllowed = make(map[uint32]struct{}, len(objectIDs))
	for _, id := range objectIDs {
		z.playersAllowed[id] = struct{}{}
	}
	z.mu.Unlock()
}

// ClearAllowed clears all whitelisted players.
func (z *BossZone) ClearAllowed() {
	z.mu.Lock()
	z.playersAllowed = make(map[uint32]struct{})
	z.playerAllowedReEntry = make(map[uint32]int64)
	z.mu.Unlock()
}
