package gameserver

import (
	"context"
	"testing"
	"time"

	"github.com/udisondev/la2go/internal/db"
	"github.com/udisondev/la2go/internal/gameserver/clientpackets"
	"github.com/udisondev/la2go/internal/gameserver/packet"
	"github.com/udisondev/la2go/internal/login"
	"github.com/udisondev/la2go/internal/model"
	"github.com/udisondev/la2go/internal/testutil"
	"github.com/udisondev/la2go/internal/world"
)

// mockCharacterRepository is a mock implementation of CharacterRepository for benchmarks.
type mockCharacterRepository struct{}

func (m *mockCharacterRepository) LoadByAccountName(ctx context.Context, accountName string) ([]*model.Player, error) {
	// Return empty slice for benchmarks (character loading not tested here)
	return []*model.Player{}, nil
}

// mockPlayerPersister is a no-op implementation of PlayerPersister for benchmarks.
type mockPlayerPersister struct{}

func (m *mockPlayerPersister) SavePlayer(ctx context.Context, player *model.Player) error {
	return nil
}

func (m *mockPlayerPersister) LoadPlayerData(ctx context.Context, charID int64) ([]db.ItemRow, []*model.SkillInfo, error) {
	return nil, nil, nil
}

// BenchmarkHandler_HandlePacket_ProtocolVersion measures full packet flow for ProtocolVersion (simplest packet).
func BenchmarkHandler_HandlePacket_ProtocolVersion(b *testing.B) {
	sessionManager := login.NewSessionManager()
	clientManager := NewClientManager()
	charRepo := &mockCharacterRepository{}
	handler := NewHandler(sessionManager, clientManager, charRepo, &mockPlayerPersister{})

	conn := testutil.NewMockConn()
	key := make([]byte, 16)
	for i := range key {
		key[i] = byte(i + 1)
	}
	client, _ := NewGameClient(conn, key, nil, 0, 0)
	client.SetState(ClientStateConnected)

	// Prepare ProtocolVersion packet (opcode 0x0E + int32 0x0106)
	data := prepareProtocolVersionPacket(0x0106)
	buf := make([]byte, 4096)
	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()

	for range b.N {
		_, _, err := handler.HandlePacket(ctx, client, data, buf)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkHandler_HandlePacket_AuthLogin measures full packet flow for AuthLogin (complex packet with SessionKey validation).
// This is the most realistic e2e benchmark for incoming packets.
func BenchmarkHandler_HandlePacket_AuthLogin(b *testing.B) {
	// Setup SessionManager with valid session
	sessionManager := login.NewSessionManager()
	testSessionKey := login.SessionKey{
		PlayOkID1:  0x12345678,
		PlayOkID2:  -0x6543210F - 1, // 0x9ABCDEF0 as negative literal (overflow)
		LoginOkID1: 0x11111111,
		LoginOkID2: 0x22222222,
	}

	// Create mock Client for SessionManager.Store
	mockClient := &login.Client{}
	sessionManager.Store("testaccount", testSessionKey, mockClient)

	clientManager := NewClientManager()
	charRepo := &mockCharacterRepository{}
	handler := NewHandler(sessionManager, clientManager, charRepo, &mockPlayerPersister{})

	conn := testutil.NewMockConn()
	key := make([]byte, 16)
	for i := range key {
		key[i] = byte(i + 1)
	}
	client, _ := NewGameClient(conn, key, nil, 0, 0)
	// AuthLogin is handled in AUTHENTICATED state (current handler logic)
	client.SetState(ClientStateAuthenticated)

	// Prepare AuthLogin packet
	data := prepareAuthLoginPacket("testaccount", testSessionKey)
	buf := make([]byte, 4096)
	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()

	for range b.N {
		_, _, err := handler.HandlePacket(ctx, client, data, buf)
		if err != nil {
			b.Fatal(err)
		}
		// Reset state for next iteration (AuthLogin transitions to AUTHENTICATED, but we need to reset for benchmark)
		client.SetState(ClientStateAuthenticated)
	}
}

// BenchmarkHandler_Dispatch_Only measures ONLY dispatch overhead (without calling handler).
// Tests nested switch overhead for different State×Opcode combinations.
func BenchmarkHandler_Dispatch_Only(b *testing.B) {
	testCases := []struct {
		state  ClientConnectionState
		opcode byte
	}{
		{ClientStateConnected, clientpackets.OpcodeProtocolVersion},
		{ClientStateConnected, 0xFF}, // Invalid opcode (default case)
		{ClientStateAuthenticated, clientpackets.OpcodeAuthLogin},
		{ClientStateAuthenticated, 0xFF}, // Invalid opcode
		{ClientStateEntering, clientpackets.OpcodeAuthLogin},
		{ClientStateInGame, clientpackets.OpcodeAuthLogin},
	}

	sessionManager := login.NewSessionManager()
	clientManager := NewClientManager()
	charRepo := &mockCharacterRepository{}
	handler := NewHandler(sessionManager, clientManager, charRepo, &mockPlayerPersister{})

	conn := testutil.NewMockConn()
	key := make([]byte, 16)
	for i := range key {
		key[i] = byte(i + 1)
	}

	for _, tc := range testCases {
		name := tc.state.String() + "_" + opcodeString(tc.opcode)
		b.Run(name, func(b *testing.B) {
			client, _ := NewGameClient(conn, key, nil, 0, 0)
			client.SetState(tc.state)

			// Minimal packet: opcode only (handler will fail, but dispatch happens)
			data := []byte{tc.opcode}
			buf := make([]byte, 4096)
			ctx := context.Background()

			b.ResetTimer()
			b.ReportAllocs()

			for range b.N {
				// Ignore error — we only care about dispatch overhead
				_, _, _ = handler.HandlePacket(ctx, client, data, buf)
			}
		})
	}
}

// BenchmarkHandler_Dispatch_Concurrent measures parallel dispatch to detect mutex contention on client.State().
func BenchmarkHandler_Dispatch_Concurrent(b *testing.B) {
	sessionManager := login.NewSessionManager()
	clientManager := NewClientManager()
	charRepo := &mockCharacterRepository{}
	handler := NewHandler(sessionManager, clientManager, charRepo, &mockPlayerPersister{})

	conn := testutil.NewMockConn()
	key := make([]byte, 16)
	for i := range key {
		key[i] = byte(i + 1)
	}
	client, _ := NewGameClient(conn, key, nil, 0, 0)
	client.SetState(ClientStateConnected)

	// Prepare ProtocolVersion packet
	data := prepareProtocolVersionPacket(0x0106)
	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		buf := make([]byte, 4096)
		for pb.Next() {
			_, _, _ = handler.HandlePacket(ctx, client, data, buf)
		}
	})
}

// prepareProtocolVersionPacket creates binary representation of ProtocolVersion packet.
func prepareProtocolVersionPacket(revision int32) []byte {
	w := packet.NewWriter(8)
	w.WriteByte(clientpackets.OpcodeProtocolVersion)
	w.WriteInt(revision)
	return w.Bytes()
}

// prepareAuthLoginPacket creates binary representation of AuthLogin packet.
func prepareAuthLoginPacket(account string, key login.SessionKey) []byte {
	w := packet.NewWriter(256)
	w.WriteByte(clientpackets.OpcodeAuthLogin)
	w.WriteString(account)
	w.WriteInt(key.PlayOkID1)
	w.WriteInt(key.PlayOkID2)
	w.WriteInt(key.LoginOkID1)
	w.WriteInt(key.LoginOkID2)
	// 4 unknown int32 fields (required by protocol)
	for range 4 {
		w.WriteInt(0)
	}
	return w.Bytes()
}

// opcodeString returns human-readable opcode name for benchmarks.
func opcodeString(opcode byte) string {
	switch opcode {
	case clientpackets.OpcodeProtocolVersion:
		return "ProtocolVersion"
	case clientpackets.OpcodeAuthLogin:
		return "AuthLogin"
	default:
		return "Unknown"
	}
}

// BenchmarkHandler_SendVisibleObjectsInfo measures parallel encryption for sendVisibleObjectsInfo.
// Simulates EnterWorld scenario with visible objects (players + NPCs + items).
// Phase 4.19: Expected -92.9% latency (22.5ms → 1.6ms) for 450 packets.
func BenchmarkHandler_SendVisibleObjectsInfo(b *testing.B) {
	benchmarks := []struct {
		name        string
		numPlayers  int
	}{
		{"10_players", 10},
		{"50_players", 50},
		{"150_players", 150},
		{"450_players", 450},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			// Setup ClientManager with players + VisibilityManager
			sessionManager := login.NewSessionManager()
			clientManager := NewClientManager()
			charRepo := &mockCharacterRepository{}
			handler := NewHandler(sessionManager, clientManager, charRepo, &mockPlayerPersister{})

			// Create world and visibility manager
			worldInstance := world.Instance()
			visibilityMgr := world.NewVisibilityManager(worldInstance, 100*time.Millisecond, 200*time.Millisecond)
			clientManager.SetVisibilityManager(visibilityMgr)

			var testPlayer *model.Player

			// Create players in same region for maximum visibility
			for i := range bm.numPlayers {
				conn := testutil.NewMockConn()
				key := make([]byte, 16)
				for j := range key {
					key[j] = byte(j + 1)
				}
				mockClient, _ := NewGameClient(conn, key, nil, 0, 0)
				mockClient.SetAccountName("account" + itoa(i))
				mockClient.SetState(ClientStateInGame)

				// Skip firstPacket encryption (authentication)
				dummyBuf := make([]byte, 1024)
				_, _ = mockClient.Encryption().EncryptPacket(dummyBuf, 2, 8)

				// Create player (spread across region)
				offsetX := int32((i % 10) * 1000)
				offsetY := int32((i / 10) * 1000)
				x := offsetX
				y := offsetY

				player, _ := model.NewPlayer(uint32(0x10000000+i), int64(i+1), 1, "Player"+itoa(i), 10, 0, 1)
				player.SetLocation(model.Location{X: x, Y: y, Z: 0, Heading: 0})

				// Add to world grid
				worldObj := model.NewWorldObject(player.ObjectID(), player.Name(), player.Location())
				if err := worldInstance.AddObject(worldObj); err != nil {
					continue
				}

				// Register with ClientManager
				clientManager.Register("account"+itoa(i), mockClient)
				clientManager.RegisterPlayer(player, mockClient)
				mockClient.SetActivePlayer(player)

				// Register with VisibilityManager
				visibilityMgr.RegisterPlayer(player)

				if i == 0 {
					testPlayer = player
				}
			}

			// Trigger batch update to build visibility cache
			visibilityMgr.UpdateAll()

			// Create GameClient for sendVisibleObjectsInfo
			conn := testutil.NewMockConn()
			key := make([]byte, 16)
			for i := range key {
				key[i] = byte(i + 1)
			}
			client, err := NewGameClient(conn, key, nil, 0, 0)
			if err != nil {
				b.Fatal(err)
			}
			client.SetState(ClientStateInGame)

			// Skip firstPacket encryption (authentication)
			dummyBuf := make([]byte, 1024)
			_, _ = client.Encryption().EncryptPacket(dummyBuf, 2, 8)

			b.ResetTimer()
			b.ReportAllocs()

			for range b.N {
				err := handler.sendVisibleObjectsInfo(client, testPlayer)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
