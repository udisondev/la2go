package model

import "sync"

// PlayerMovement tracks client and server positions separately for desync detection.
// This structure is used to maintain two independent position states:
//   - Client position: reported by ValidatePosition packets (~200ms interval)
//   - Server position: validated and updated by MoveToLocation handlers
//
// Thread-safe: uses RWMutex for concurrent access.
type PlayerMovement struct {
	mu sync.RWMutex

	// Client-reported position (from ValidatePosition packets)
	clientX       int32
	clientY       int32
	clientZ       int32
	clientHeading int32

	// Server-validated position (from MoveToLocation packets)
	lastServerX int32
	lastServerY int32
	lastServerZ int32
}

// NewPlayerMovement creates a new PlayerMovement with initial coordinates.
// Both client and server positions are initialized to the same values.
func NewPlayerMovement(x, y, z int32) *PlayerMovement {
	return &PlayerMovement{
		clientX:     x,
		clientY:     y,
		clientZ:     z,
		lastServerX: x,
		lastServerY: y,
		lastServerZ: z,
	}
}

// ClientPosition returns the client-reported position.
// Thread-safe: acquires read lock.
func (pm *PlayerMovement) ClientPosition() (x, y, z, heading int32) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.clientX, pm.clientY, pm.clientZ, pm.clientHeading
}

// SetClientPosition updates the client-reported position.
// Called when processing ValidatePosition packets.
// Thread-safe: acquires write lock.
func (pm *PlayerMovement) SetClientPosition(x, y, z, heading int32) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.clientX = x
	pm.clientY = y
	pm.clientZ = z
	pm.clientHeading = heading
}

// LastServerPosition returns the last server-validated position.
// Thread-safe: acquires read lock.
func (pm *PlayerMovement) LastServerPosition() (x, y, z int32) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.lastServerX, pm.lastServerY, pm.lastServerZ
}

// SetLastServerPosition updates the last server-validated position.
// Called when processing MoveToLocation packets after validation succeeds.
// Thread-safe: acquires write lock.
func (pm *PlayerMovement) SetLastServerPosition(x, y, z int32) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.lastServerX = x
	pm.lastServerY = y
	pm.lastServerZ = z
}
