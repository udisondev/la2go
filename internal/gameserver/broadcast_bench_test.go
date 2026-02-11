package gameserver

import (
	"testing"
	"time"

	"github.com/udisondev/la2go/internal/model"
	"github.com/udisondev/la2go/internal/testutil"
	"github.com/udisondev/la2go/internal/world"
)

// BenchmarkBroadcast_ToAll measures broadcast to ALL clients (worst case).
// This is the SLOW path — sends to 100% of clients regardless of visibility.
func BenchmarkBroadcast_ToAll(b *testing.B) {
	sizes := []int{10, 100, 1000}

	for _, size := range sizes {
		b.Run("clients="+itoa(size), func(b *testing.B) {
			cm := setupClientManager(size)
			payload := []byte{0x01, 0x02, 0x03, 0x04}

			b.ResetTimer()
			b.ReportAllocs()

			for range b.N {
				cm.BroadcastToAll(payload, len(payload))
			}
		})
	}
}

// BenchmarkBroadcast_ToVisible measures broadcast to VISIBLE clients (fast path).
// Uses visibility cache (Phase 4.5 PR3) to filter ~5% visible players.
// Expected: -95% broadcast cost compared to BroadcastToAll.
func BenchmarkBroadcast_ToVisible(b *testing.B) {
	sizes := []int{10, 100, 1000}

	for _, size := range sizes {
		b.Run("clients="+itoa(size), func(b *testing.B) {
			cm, sourcePlayer := setupClientManagerWithPlayers(size)
			payload := []byte{0x01, 0x02, 0x03, 0x04}

			b.ResetTimer()
			b.ReportAllocs()

			for range b.N {
				cm.BroadcastToVisible(sourcePlayer, payload, len(payload))
			}
		})
	}
}

// BenchmarkBroadcast_ToVisibleExcept measures broadcast excluding one player.
// Typical for broadcasting player's own actions to others.
func BenchmarkBroadcast_ToVisibleExcept(b *testing.B) {
	sizes := []int{10, 100, 1000}

	for _, size := range sizes {
		b.Run("clients="+itoa(size), func(b *testing.B) {
			cm, sourcePlayer := setupClientManagerWithPlayers(size)
			excludePlayer := sourcePlayer // exclude source from broadcast
			payload := []byte{0x01, 0x02, 0x03, 0x04}

			b.ResetTimer()
			b.ReportAllocs()

			for range b.N {
				cm.BroadcastToVisibleExcept(sourcePlayer, excludePlayer, payload, len(payload))
			}
		})
	}
}

// BenchmarkBroadcast_ToRegion measures broadcast to specific region.
// Useful for area-of-effect announcements (castle siege, boss spawn).
func BenchmarkBroadcast_ToRegion(b *testing.B) {
	sizes := []int{10, 100, 1000}

	for _, size := range sizes {
		b.Run("clients="+itoa(size), func(b *testing.B) {
			cm := setupClientManagerWithPlayersInRegion(size, 0, 0)
			payload := []byte{0x01, 0x02, 0x03, 0x04}

			b.ResetTimer()
			b.ReportAllocs()

			for range b.N {
				cm.BroadcastToRegion(0, 0, payload, len(payload))
			}
		})
	}
}

// setupClientManager creates ClientManager with N authenticated clients.
// Phase 7.0: Uses BytePool + SetWritePool for broadcast encryption.
func setupClientManager(n int) *ClientManager {
	cm := NewClientManager()
	pool := NewBytePool(128)
	cm.SetWritePool(pool)

	for i := range n {
		conn := testutil.NewMockConn()
		client, _ := NewGameClient(conn, make([]byte, 16), pool, 16, 0)
		accountName := "account" + itoa(i)
		client.SetAccountName(accountName)
		client.SetState(ClientStateAuthenticated)
		cm.Register(accountName, client)
	}

	return cm
}

// setupClientManagerWithPlayers creates ClientManager with N players in IN_GAME state.
// Returns ClientManager and a source player for visibility tests.
// Phase 7.0: Uses BytePool + SetWritePool for broadcast encryption.
func setupClientManagerWithPlayers(n int) (*ClientManager, *model.Player) {
	cm := NewClientManager()
	pool := NewBytePool(128)
	cm.SetWritePool(pool)

	var sourcePlayer *model.Player

	for i := range n {
		conn := testutil.NewMockConn()
		client, _ := NewGameClient(conn, make([]byte, 16), pool, 16, 0)
		accountName := "account" + itoa(i)
		client.SetAccountName(accountName)
		client.SetState(ClientStateInGame)

		// Create player at coordinates (10000 + i*100, 20000 + i*100)
		// This spreads players across different regions
		player, _ := model.NewPlayer(uint32(i+1), int64(i+1), 1, "Player"+itoa(i), 10, 0, 1)
		player.SetLocation(model.Location{X: int32(10000 + i*100), Y: int32(20000 + i*100), Z: 0, Heading: 0})

		cm.Register(accountName, client)
		cm.RegisterPlayer(player, client)

		if i == 0 {
			sourcePlayer = player
		}
	}

	return cm, sourcePlayer
}

// setupClientManagerWithPlayersInRegion creates N players in specific region.
// Phase 7.0: Uses BytePool + SetWritePool for broadcast encryption.
func setupClientManagerWithPlayersInRegion(n int, regionX, regionY int32) *ClientManager {
	cm := NewClientManager()
	pool := NewBytePool(128)
	cm.SetWritePool(pool)

	// Calculate base coordinates for region
	baseX := regionX * 16384
	baseY := regionY * 16384

	for i := range n {
		conn := testutil.NewMockConn()
		client, _ := NewGameClient(conn, make([]byte, 16), pool, 16, 0)
		accountName := "account" + itoa(i)
		client.SetAccountName(accountName)
		client.SetState(ClientStateInGame)

		// Create player within target region
		player, _ := model.NewPlayer(uint32(i+1), int64(i+1), 1, "Player"+itoa(i), 10, 0, 1)
		player.SetLocation(model.Location{X: baseX + int32(i*10), Y: baseY + int32(i*10), Z: 0, Heading: 0})

		cm.Register(accountName, client)
		cm.RegisterPlayer(player, client)
	}

	return cm
}

// BenchmarkBroadcast_ToVisibleNear measures broadcast using LODNear filtering.
// Phase 4.13: Expected -89% packet reduction vs BroadcastToVisible (50 vs 450 objects).
// NOTE: Requires full world + visibility cache setup to see real impact.
// Current benchmark shows API overhead only (visibility cache empty in unit tests).
func BenchmarkBroadcast_ToVisibleNear(b *testing.B) {
	sizes := []int{10, 100, 1000}

	for _, size := range sizes {
		b.Run("clients="+itoa(size), func(b *testing.B) {
			cm, sourcePlayer := setupClientManagerWithPlayers(size)
			payload := []byte{0x01, 0x02, 0x03, 0x04}

			b.ResetTimer()
			b.ReportAllocs()

			for range b.N {
				cm.BroadcastToVisibleNear(sourcePlayer, payload, len(payload))
			}
		})
	}
}

// BenchmarkBroadcast_ToVisibleMedium measures broadcast using LODMedium filtering.
// Phase 4.13: Expected -56% packet reduction vs BroadcastToVisible (200 vs 450 objects).
func BenchmarkBroadcast_ToVisibleMedium(b *testing.B) {
	sizes := []int{10, 100, 1000}

	for _, size := range sizes {
		b.Run("clients="+itoa(size), func(b *testing.B) {
			cm, sourcePlayer := setupClientManagerWithPlayers(size)
			payload := []byte{0x01, 0x02, 0x03, 0x04}

			b.ResetTimer()
			b.ReportAllocs()

			for range b.N {
				cm.BroadcastToVisibleMedium(sourcePlayer, payload, len(payload))
			}
		})
	}
}

// BenchmarkBroadcast_ToVisibleNearExcept measures broadcast using LODNear with exclusion.
// Phase 4.13: Most common pattern for player movement broadcasts.
func BenchmarkBroadcast_ToVisibleNearExcept(b *testing.B) {
	sizes := []int{10, 100, 1000}

	for _, size := range sizes {
		b.Run("clients="+itoa(size), func(b *testing.B) {
			cm, sourcePlayer := setupClientManagerWithPlayers(size)
			excludePlayer := sourcePlayer
			payload := []byte{0x01, 0x02, 0x03, 0x04}

			b.ResetTimer()
			b.ReportAllocs()

			for range b.N {
				cm.BroadcastToVisibleNearExcept(sourcePlayer, excludePlayer, payload, len(payload))
			}
		})
	}
}

// itoa converts int to string (simple helper for benchmark names).
func itoa(n int) string {
	if n == 0 {
		return "0"
	}

	negative := false
	if n < 0 {
		negative = true
		n = -n
	}

	buf := make([]byte, 0, 12)
	for n > 0 {
		buf = append(buf, byte('0'+n%10))
		n /= 10
	}

	// Reverse
	for i, j := 0, len(buf)-1; i < j; i, j = i+1, j-1 {
		buf[i], buf[j] = buf[j], buf[i]
	}

	if negative {
		return "-" + string(buf)
	}
	return string(buf)
}

// BenchmarkBroadcast_ReverseCache measures reverse cache performance improvement.
// Phase 4.18 Optimization 1: Expected -99.999% (100,000× faster) for large player counts.
//
// Measures O(N×M) → O(M) improvement:
// - Before: Iterate ALL players (N=100K) × check visibility (M=100) = 10M operations
// - After: Lookup observers (M=100) via reverse cache
//
// Expected results:
// - 100 players: ~1µs (O(M) already fast)
// - 1K players: ~10µs (O(M) still fast)
// - 10K players: ~100µs (O(M) still fast, but O(N×M) would be 1s)
// - 100K players: ~1ms (O(M) still fast, but O(N×M) would be 100s)
func BenchmarkBroadcast_ReverseCache(b *testing.B) {
	sizes := []int{100, 1000, 10000}

	for _, size := range sizes {
		b.Run("players="+itoa(size), func(b *testing.B) {
			// Setup: create ClientManager + VisibilityManager with reverse cache
			cm, sourcePlayer, visibilityMgr := setupClientManagerWithReverseCacheBench(size)

			// Trigger batch update to build reverse cache
			visibilityMgr.UpdateAll()

			payload := []byte{0x01, 0x02, 0x03, 0x04}

			b.ResetTimer()
			b.ReportAllocs()

			for range b.N {
				// This call uses GetObservers() reverse cache (Phase 4.18 Optimization 1)
				// Expected: O(M) = ~100 lookups, NOT O(N×M) = 100K × 100 = 10M operations
				cm.BroadcastToVisibleNear(sourcePlayer, payload, len(payload))
			}
		})
	}
}

// setupClientManagerWithReverseCacheBench creates ClientManager + VisibilityManager with reverse cache.
// Returns ClientManager, source player, and VisibilityManager for triggering batch update.
//
// Players are placed in clusters of 10 in same region to create realistic visibility patterns.
// Expected: sourcePlayer sees ~100 nearby players (10 in same region + 90 in adjacent regions).
// Phase 7.0: Uses BytePool + SetWritePool for broadcast encryption.
func setupClientManagerWithReverseCacheBench(n int) (*ClientManager, *model.Player, *world.VisibilityManager) {
	cm := NewClientManager()
	pool := NewBytePool(128)
	cm.SetWritePool(pool)

	// Create world and visibility manager
	worldInstance := world.Instance()
	visibilityMgr := world.NewVisibilityManager(worldInstance, 100*time.Millisecond, 200*time.Millisecond)

	// Link ClientManager to VisibilityManager (Phase 4.18 Optimization 1)
	cm.SetVisibilityManager(visibilityMgr)

	var sourcePlayer *model.Player

	// Create players in clusters (10 players per region)
	// Spreads across multiple regions to create realistic visibility
	clusterSize := 10
	regionSize := int32(16384) // world.RegionSize

	for i := range n {
		conn := testutil.NewMockConn()
		client, _ := NewGameClient(conn, make([]byte, 16), pool, 16, 0)
		accountName := "account" + itoa(i)
		client.SetAccountName(accountName)
		client.SetState(ClientStateInGame)

		// Calculate region coordinates (cluster index determines region)
		clusterIndex := i / clusterSize
		regionX := int32(clusterIndex % 10) // 10 regions wide
		regionY := int32(clusterIndex / 10) // 10+ regions tall

		// Calculate position within region (spread 10 players across region)
		localIndex := i % clusterSize
		offsetX := int32(localIndex * 1000) // 1000 units apart
		offsetY := int32(localIndex * 1000)

		x := regionX*regionSize + offsetX
		y := regionY*regionSize + offsetY

		// Create player
		player, _ := model.NewPlayer(uint32(i+1), int64(i+1), 1, "Player"+itoa(i), 10, 0, 1)
		player.SetLocation(model.Location{X: x, Y: y, Z: 0, Heading: 0})

		// Add to world grid (required for visibility queries)
		worldObj := model.NewWorldObject(player.ObjectID(), player.Name(), player.Location())
		if err := worldInstance.AddObject(worldObj); err != nil {
			continue // Skip if add fails
		}

		// Register with ClientManager
		cm.Register(accountName, client)
		cm.RegisterPlayer(player, client)

		// Register with VisibilityManager (required for batch updates)
		visibilityMgr.RegisterPlayer(player)

		if i == 0 {
			sourcePlayer = player
		}
	}

	return cm, sourcePlayer, visibilityMgr
}
