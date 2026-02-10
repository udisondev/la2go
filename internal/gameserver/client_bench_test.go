package gameserver

import (
	"testing"

	"github.com/udisondev/la2go/internal/login"
	"github.com/udisondev/la2go/internal/model"
	"github.com/udisondev/la2go/internal/testutil"
	"github.com/udisondev/la2go/internal/world"
)

// BenchmarkGameClient_State — чтение state (P0 hotpath, mutex lock на каждый packet)
func BenchmarkGameClient_State(b *testing.B) {
	b.ReportAllocs()

	conn := testutil.NewMockConn()
	key := make([]byte, 16)
	for i := range key {
		key[i] = byte(i + 1)
	}

	client, err := NewGameClient(conn, key)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for range b.N {
		_ = client.State()
	}
}

// BenchmarkGameClient_SetState — изменение state (FSM transition)
func BenchmarkGameClient_SetState(b *testing.B) {
	b.ReportAllocs()

	conn := testutil.NewMockConn()
	key := make([]byte, 16)
	for i := range key {
		key[i] = byte(i + 1)
	}

	client, err := NewGameClient(conn, key)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := range b.N {
		// Чередуем состояния (CONNECTED → AUTHENTICATED → CONNECTED)
		if i%2 == 0 {
			client.SetState(ClientStateAuthenticated)
		} else {
			client.SetState(ClientStateConnected)
		}
	}
}

// BenchmarkGameClient_AccountName — чтение account name
func BenchmarkGameClient_AccountName(b *testing.B) {
	b.ReportAllocs()

	conn := testutil.NewMockConn()
	key := make([]byte, 16)
	for i := range key {
		key[i] = byte(i + 1)
	}

	client, err := NewGameClient(conn, key)
	if err != nil {
		b.Fatal(err)
	}

	client.SetAccountName("TestUser123")

	b.ResetTimer()
	for range b.N {
		_ = client.AccountName()
	}
}

// BenchmarkGameClient_SetAccountName — запись account name
func BenchmarkGameClient_SetAccountName(b *testing.B) {
	b.ReportAllocs()

	conn := testutil.NewMockConn()
	key := make([]byte, 16)
	for i := range key {
		key[i] = byte(i + 1)
	}

	client, err := NewGameClient(conn, key)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for range b.N {
		client.SetAccountName("TestUser123")
	}
}

// BenchmarkGameClient_SessionKey — чтение SessionKey
func BenchmarkGameClient_SessionKey(b *testing.B) {
	b.ReportAllocs()

	conn := testutil.NewMockConn()
	key := make([]byte, 16)
	for i := range key {
		key[i] = byte(i + 1)
	}

	client, err := NewGameClient(conn, key)
	if err != nil {
		b.Fatal(err)
	}

	sk := &login.SessionKey{
		PlayOkID1:  1,
		PlayOkID2:  2,
		LoginOkID1: 3,
		LoginOkID2: 4,
	}
	client.SetSessionKey(sk)

	b.ResetTimer()
	for range b.N {
		_ = client.SessionKey()
	}
}

// BenchmarkGameClient_SetSessionKey — запись SessionKey
func BenchmarkGameClient_SetSessionKey(b *testing.B) {
	b.ReportAllocs()

	conn := testutil.NewMockConn()
	key := make([]byte, 16)
	for i := range key {
		key[i] = byte(i + 1)
	}

	client, err := NewGameClient(conn, key)
	if err != nil {
		b.Fatal(err)
	}

	sk := &login.SessionKey{
		PlayOkID1:  1,
		PlayOkID2:  2,
		LoginOkID1: 3,
		LoginOkID2: 4,
	}

	b.ResetTimer()
	for range b.N {
		client.SetSessionKey(sk)
	}
}

// BenchmarkGameClient_Concurrent_StateAccess — параллельный доступ к state (реалистичный сценарий)
func BenchmarkGameClient_Concurrent_StateAccess(b *testing.B) {
	b.ReportAllocs()

	conn := testutil.NewMockConn()
	key := make([]byte, 16)
	for i := range key {
		key[i] = byte(i + 1)
	}

	client, err := NewGameClient(conn, key)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// 90% reads, 10% writes (реалистичная нагрузка)
			if b.N%10 == 0 {
				client.SetState(ClientStateAuthenticated)
			} else {
				_ = client.State()
			}
		}
	})
}

// BenchmarkGameClient_GetCharacters_CacheHit — cache hit performance (Phase 4.18 Opt 3)
// Измеряет производительность GetCharacters() при cache hit (expected: ~50ns)
func BenchmarkGameClient_GetCharacters_CacheHit(b *testing.B) {
	b.ReportAllocs()

	conn := testutil.NewMockConn()
	key := make([]byte, 16)
	for i := range key {
		key[i] = byte(i + 1)
	}

	client, err := NewGameClient(conn, key)
	if err != nil {
		b.Fatal(err)
	}

	// Simulate 3 characters (typical account)
	mockPlayers := createMockPlayers(3)
	accountName := "TestUser123"

	// Warm up cache
	_, err = client.GetCharacters(accountName, func(name string) ([]*model.Player, error) {
		return mockPlayers, nil
	})
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for range b.N {
		_, _ = client.GetCharacters(accountName, func(name string) ([]*model.Player, error) {
			b.Fatal("loader called on cache hit")
			return nil, nil
		})
	}
}

// BenchmarkGameClient_GetCharacters_CacheMiss — cache miss performance (Phase 4.18 Opt 3)
// Измеряет производительность GetCharacters() при cache miss (expected: ~500µs DB query)
func BenchmarkGameClient_GetCharacters_CacheMiss(b *testing.B) {
	b.ReportAllocs()

	conn := testutil.NewMockConn()
	key := make([]byte, 16)
	for i := range key {
		key[i] = byte(i + 1)
	}

	client, err := NewGameClient(conn, key)
	if err != nil {
		b.Fatal(err)
	}

	mockPlayers := createMockPlayers(3)

	b.ResetTimer()
	for range b.N {
		// Clear cache each iteration to simulate miss
		client.ClearCharacterCache()

		_, err := client.GetCharacters("TestUser123", func(name string) ([]*model.Player, error) {
			return mockPlayers, nil
		})
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkGameClient_GetCharacters_Concurrent — concurrent cache access (Phase 4.18 Opt 3)
// Симулирует concurrent login flow: 90% cache hits (2nd/3rd call), 10% cache miss (1st call)
func BenchmarkGameClient_GetCharacters_Concurrent(b *testing.B) {
	b.ReportAllocs()

	conn := testutil.NewMockConn()
	key := make([]byte, 16)
	for i := range key {
		key[i] = byte(i + 1)
	}

	client, err := NewGameClient(conn, key)
	if err != nil {
		b.Fatal(err)
	}

	mockPlayers := createMockPlayers(3)
	accountName := "TestUser123"

	// Warm up cache
	_, err = client.GetCharacters(accountName, func(name string) ([]*model.Player, error) {
		return mockPlayers, nil
	})
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = client.GetCharacters(accountName, func(name string) ([]*model.Player, error) {
				return mockPlayers, nil
			})
		}
	})
}

// createMockPlayers creates N mock players for testing GetCharacters cache
func createMockPlayers(count int) []*model.Player {
	players := make([]*model.Player, count)
	for i := range count {
		objectID := world.IDGenerator().NextPlayerID()
		player, err := model.NewPlayer(
			objectID,
			int64(i+1),      // characterID
			int64(1),        // accountID
			"TestChar",      // name
			1,               // level
			0,               // raceID (Human)
			0,               // classID (Fighter)
		)
		if err != nil {
			panic(err) // Should never happen in test
		}
		players[i] = player
	}
	return players
}
