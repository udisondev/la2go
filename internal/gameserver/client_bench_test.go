package gameserver

import (
	"testing"

	"github.com/udisondev/la2go/internal/login"
	"github.com/udisondev/la2go/internal/testutil"
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
