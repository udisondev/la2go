package gameserver

import (
	"testing"

	"github.com/udisondev/la2go/internal/constants"
	"github.com/udisondev/la2go/internal/model"
	"github.com/udisondev/la2go/internal/testutil"
)

// preparePacketBufferBench creates a proper packet buffer for benchmarks.
// Uses DefaultSendBufSize to ensure enough space for first packet encryption.
func preparePacketBufferBench(payload []byte) []byte {
	buf := make([]byte, constants.DefaultSendBufSize)
	copy(buf[constants.PacketHeaderSize:], payload)
	return buf
}

// BenchmarkBroadcast_ToAll measures broadcast to ALL clients (worst case).
// This is the SLOW path — sends to 100% of clients regardless of visibility.
func BenchmarkBroadcast_ToAll(b *testing.B) {
	sizes := []int{10, 100, 1000}

	for _, size := range sizes {
		b.Run("clients="+itoa(size), func(b *testing.B) {
			cm := setupClientManager(size)
			payload := []byte{0x01, 0x02, 0x03, 0x04}
			packetData := preparePacketBufferBench(payload)

			b.ResetTimer()
			b.ReportAllocs()

			for range b.N {
				cm.BroadcastToAll(packetData, len(payload))
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
			packetData := preparePacketBufferBench(payload)

			b.ResetTimer()
			b.ReportAllocs()

			for range b.N {
				cm.BroadcastToVisible(sourcePlayer, packetData, len(payload))
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
			packetData := preparePacketBufferBench(payload)

			b.ResetTimer()
			b.ReportAllocs()

			for range b.N {
				cm.BroadcastToVisibleExcept(sourcePlayer, excludePlayer, packetData, len(payload))
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
			packetData := preparePacketBufferBench(payload)

			b.ResetTimer()
			b.ReportAllocs()

			for range b.N {
				cm.BroadcastToRegion(0, 0, packetData, len(payload))
			}
		})
	}
}

// setupClientManager creates ClientManager with N authenticated clients.
func setupClientManager(n int) *ClientManager {
	cm := NewClientManager()

	for i := range n {
		conn := testutil.NewMockConn()
		client, _ := NewGameClient(conn, make([]byte, 16))
		accountName := "account" + itoa(i)
		client.SetAccountName(accountName)
		client.SetState(ClientStateAuthenticated)
		cm.Register(accountName, client)
	}

	return cm
}

// setupClientManagerWithPlayers creates ClientManager with N players in IN_GAME state.
// Returns ClientManager and a source player for visibility tests.
func setupClientManagerWithPlayers(n int) (*ClientManager, *model.Player) {
	cm := NewClientManager()
	var sourcePlayer *model.Player

	for i := range n {
		conn := testutil.NewMockConn()
		client, _ := NewGameClient(conn, make([]byte, 16))
		accountName := "account" + itoa(i)
		client.SetAccountName(accountName)
		client.SetState(ClientStateInGame)

		// Create player at coordinates (10000 + i*100, 20000 + i*100)
		// This spreads players across different regions
		player, _ := model.NewPlayer(int64(i+1), 1, "Player"+itoa(i), 10, 0, 1)
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
func setupClientManagerWithPlayersInRegion(n int, regionX, regionY int32) *ClientManager {
	cm := NewClientManager()

	// Calculate base coordinates for region
	baseX := regionX * 16384
	baseY := regionY * 16384

	for i := range n {
		conn := testutil.NewMockConn()
		client, _ := NewGameClient(conn, make([]byte, 16))
		accountName := "account" + itoa(i)
		client.SetAccountName(accountName)
		client.SetState(ClientStateInGame)

		// Create player within target region
		player, _ := model.NewPlayer(int64(i+1), 1, "Player"+itoa(i), 10, 0, 1)
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
			packetData := preparePacketBufferBench(payload)

			b.ResetTimer()
			b.ReportAllocs()

			for range b.N {
				cm.BroadcastToVisibleNear(sourcePlayer, packetData, len(payload))
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
			packetData := preparePacketBufferBench(payload)

			b.ResetTimer()
			b.ReportAllocs()

			for range b.N {
				cm.BroadcastToVisibleMedium(sourcePlayer, packetData, len(payload))
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
			packetData := preparePacketBufferBench(payload)

			b.ResetTimer()
			b.ReportAllocs()

			for range b.N {
				cm.BroadcastToVisibleNearExcept(sourcePlayer, excludePlayer, packetData, len(payload))
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
