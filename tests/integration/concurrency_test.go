package integration

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/udisondev/la2go/internal/gameserver"
	"github.com/udisondev/la2go/internal/login"
)

// TestSessionManagerConcurrency тестирует concurrent operations на SessionManager.
func TestSessionManagerConcurrency(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration tests in short mode")
	}
	t.Parallel()

	sm := login.NewSessionManager()
	const numGoroutines = 50

	var wg sync.WaitGroup

	// Concurrent Store
	for i := range numGoroutines {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			account := fmt.Sprintf("user_%d", id)
			sk := login.SessionKey{
				LoginOkID1: int32(id * 4),
				LoginOkID2: int32(id*4 + 1),
				PlayOkID1:  int32(id*4 + 2),
				PlayOkID2:  int32(id*4 + 3),
			}
			sm.Store(account, sk, nil)
		}(i)
	}

	wg.Wait()

	// Verify count
	count := sm.Count()
	assert.Equal(t, numGoroutines, count, "all sessions should be stored")

	// Concurrent Validate + Remove (simulating real handler behavior)
	validCount := 0
	var mu sync.Mutex

	for i := range numGoroutines {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			account := fmt.Sprintf("user_%d", id)
			sk := login.SessionKey{
				LoginOkID1: int32(id * 4),
				LoginOkID2: int32(id*4 + 1),
				PlayOkID1:  int32(id*4 + 2),
				PlayOkID2:  int32(id*4 + 3),
			}
			if sm.Validate(account, sk, false) {
				sm.Remove(account) // Handlers call Remove after successful Validate
				mu.Lock()
				validCount++
				mu.Unlock()
			}
		}(i)
	}

	wg.Wait()

	assert.Equal(t, numGoroutines, validCount, "all sessions should validate")

	// Verify all removed after validation
	count = sm.Count()
	assert.Equal(t, 0, count, "all sessions should be removed after validation")
}

// TestGameServerTableConcurrency тестирует concurrent operations на GameServerTable.
func TestGameServerTableConcurrency(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration tests in short mode")
	}
	t.Parallel()

	gsTable := gameserver.NewGameServerTable(nil)
	const numServers = 30

	var wg sync.WaitGroup

	// Concurrent Register
	for i := range numServers {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			serverID := 100 + id
			hexID := []byte(fmt.Sprintf("hex_id_%d", id))
			info := gameserver.NewGameServerInfo(serverID, hexID)
			gsTable.Register(serverID, info)
		}(i)
	}

	wg.Wait()

	// Concurrent GetByID
	successCount := 0
	var mu sync.Mutex

	for i := range numServers {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			serverID := 100 + id
			info, exists := gsTable.GetByID(serverID)
			if exists && info != nil {
				mu.Lock()
				successCount++
				mu.Unlock()
			}
		}(i)
	}

	wg.Wait()

	assert.Equal(t, numServers, successCount, "all servers should be retrievable")

	// Concurrent SetAuthed
	for i := range numServers {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			serverID := 100 + id
			if info, exists := gsTable.GetByID(serverID); exists {
				info.SetAuthed(true)
			}
		}(i)
	}

	wg.Wait()

	// Verify all authed
	for i := range numServers {
		serverID := 100 + i
		info, exists := gsTable.GetByID(serverID)
		assert.True(t, exists, "server %d should exist", serverID)
		assert.True(t, info.IsAuthed(), "server %d should be authed", serverID)
	}
}
