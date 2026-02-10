package gameserver

import (
	"context"
	"testing"

	"github.com/udisondev/la2go/internal/gameserver/clientpackets"
	"github.com/udisondev/la2go/internal/gameserver/packet"
	"github.com/udisondev/la2go/internal/login"
	"github.com/udisondev/la2go/internal/model"
	"github.com/udisondev/la2go/internal/testutil"
)

// mockCharacterRepository is a mock implementation of CharacterRepository for benchmarks.
type mockCharacterRepository struct{}

func (m *mockCharacterRepository) LoadByAccountName(ctx context.Context, accountName string) ([]*model.Player, error) {
	// Return empty slice for benchmarks (character loading not tested here)
	return []*model.Player{}, nil
}

// BenchmarkHandler_HandlePacket_ProtocolVersion measures full packet flow for ProtocolVersion (simplest packet).
func BenchmarkHandler_HandlePacket_ProtocolVersion(b *testing.B) {
	sessionManager := login.NewSessionManager()
	clientManager := NewClientManager()
	charRepo := &mockCharacterRepository{}
	handler := NewHandler(sessionManager, clientManager, charRepo)

	conn := testutil.NewMockConn()
	key := make([]byte, 16)
	for i := range key {
		key[i] = byte(i + 1)
	}
	client, _ := NewGameClient(conn, key)
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
	handler := NewHandler(sessionManager, clientManager, charRepo)

	conn := testutil.NewMockConn()
	key := make([]byte, 16)
	for i := range key {
		key[i] = byte(i + 1)
	}
	client, _ := NewGameClient(conn, key)
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
	handler := NewHandler(sessionManager, clientManager, charRepo)

	conn := testutil.NewMockConn()
	key := make([]byte, 16)
	for i := range key {
		key[i] = byte(i + 1)
	}

	for _, tc := range testCases {
		name := tc.state.String() + "_" + opcodeString(tc.opcode)
		b.Run(name, func(b *testing.B) {
			client, _ := NewGameClient(conn, key)
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
	handler := NewHandler(sessionManager, clientManager, charRepo)

	conn := testutil.NewMockConn()
	key := make([]byte, 16)
	for i := range key {
		key[i] = byte(i + 1)
	}
	client, _ := NewGameClient(conn, key)
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
