package gameserver

import (
	"testing"

	"github.com/udisondev/la2go/internal/model"
	"github.com/udisondev/la2go/internal/testutil"
	"github.com/udisondev/la2go/internal/world"
)

// TestGameClient_GetCharacters_CacheHit verifies cache returns same data without calling loader
func TestGameClient_GetCharacters_CacheHit(t *testing.T) {
	conn := testutil.NewMockConn()
	key := make([]byte, 16)
	for i := range key {
		key[i] = byte(i + 1)
	}

	client, err := NewGameClient(conn, key)
	if err != nil {
		t.Fatal(err)
	}

	mockPlayers := createTestPlayers(3)
	accountName := "TestUser"
	loaderCallCount := 0

	// First call — cache miss, should call loader
	result1, err := client.GetCharacters(accountName, func(name string) ([]*model.Player, error) {
		loaderCallCount++
		if name != accountName {
			t.Errorf("loader called with wrong account name: got %q, want %q", name, accountName)
		}
		return mockPlayers, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if loaderCallCount != 1 {
		t.Errorf("loader should be called once, got %d calls", loaderCallCount)
	}
	if len(result1) != 3 {
		t.Errorf("expected 3 players, got %d", len(result1))
	}

	// Second call — cache hit, should NOT call loader
	result2, err := client.GetCharacters(accountName, func(name string) ([]*model.Player, error) {
		loaderCallCount++
		t.Error("loader should not be called on cache hit")
		return nil, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if loaderCallCount != 1 {
		t.Errorf("loader should be called only once (cache hit), got %d calls", loaderCallCount)
	}
	if len(result2) != 3 {
		t.Errorf("expected 3 players from cache, got %d", len(result2))
	}

	// Verify same slice reference (ownership transfer)
	if &result1[0] != &result2[0] {
		t.Error("cache should return same slice reference")
	}
}

// TestGameClient_GetCharacters_DifferentAccount verifies cache is per-account
func TestGameClient_GetCharacters_DifferentAccount(t *testing.T) {
	conn := testutil.NewMockConn()
	key := make([]byte, 16)
	for i := range key {
		key[i] = byte(i + 1)
	}

	client, err := NewGameClient(conn, key)
	if err != nil {
		t.Fatal(err)
	}

	account1Players := createTestPlayers(2)
	account2Players := createTestPlayers(5)

	// Load for account1
	result1, err := client.GetCharacters("account1", func(name string) ([]*model.Player, error) {
		if name != "account1" {
			t.Errorf("expected account1, got %q", name)
		}
		return account1Players, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result1) != 2 {
		t.Errorf("expected 2 players for account1, got %d", len(result1))
	}

	// Load for account2 — should call loader (different account)
	result2, err := client.GetCharacters("account2", func(name string) ([]*model.Player, error) {
		if name != "account2" {
			t.Errorf("expected account2, got %q", name)
		}
		return account2Players, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result2) != 5 {
		t.Errorf("expected 5 players for account2, got %d", len(result2))
	}

	// Cache should now have account2 data
	result3, err := client.GetCharacters("account2", func(name string) ([]*model.Player, error) {
		t.Error("loader should not be called (cache hit for account2)")
		return nil, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result3) != 5 {
		t.Errorf("expected 5 players from cache, got %d", len(result3))
	}
}

// TestGameClient_ClearCharacterCache verifies cache clearing
func TestGameClient_ClearCharacterCache(t *testing.T) {
	conn := testutil.NewMockConn()
	key := make([]byte, 16)
	for i := range key {
		key[i] = byte(i + 1)
	}

	client, err := NewGameClient(conn, key)
	if err != nil {
		t.Fatal(err)
	}

	mockPlayers := createTestPlayers(3)
	accountName := "TestUser"
	loaderCallCount := 0

	// First call — populate cache
	_, err = client.GetCharacters(accountName, func(name string) ([]*model.Player, error) {
		loaderCallCount++
		return mockPlayers, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if loaderCallCount != 1 {
		t.Errorf("loader should be called once, got %d calls", loaderCallCount)
	}

	// Clear cache
	client.ClearCharacterCache()

	// Next call — should call loader again (cache cleared)
	_, err = client.GetCharacters(accountName, func(name string) ([]*model.Player, error) {
		loaderCallCount++
		return mockPlayers, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if loaderCallCount != 2 {
		t.Errorf("loader should be called twice (after cache clear), got %d calls", loaderCallCount)
	}
}

// TestGameClient_GetCharacters_LoaderError verifies error propagation
func TestGameClient_GetCharacters_LoaderError(t *testing.T) {
	conn := testutil.NewMockConn()
	key := make([]byte, 16)
	for i := range key {
		key[i] = byte(i + 1)
	}

	client, err := NewGameClient(conn, key)
	if err != nil {
		t.Fatal(err)
	}

	// Loader returns error
	_, err = client.GetCharacters("TestUser", func(name string) ([]*model.Player, error) {
		return nil, testutil.ErrSimulated
	})
	if err != testutil.ErrSimulated {
		t.Errorf("expected ErrSimulated, got %v", err)
	}

	// Cache should NOT be populated after error
	loaderCallCount := 0
	_, err = client.GetCharacters("TestUser", func(name string) ([]*model.Player, error) {
		loaderCallCount++
		return nil, testutil.ErrSimulated
	})
	if err != testutil.ErrSimulated {
		t.Errorf("expected ErrSimulated, got %v", err)
	}
	if loaderCallCount != 1 {
		t.Errorf("loader should be called again after previous error, got %d calls", loaderCallCount)
	}
}

// createTestPlayers creates N test players for cache testing
func createTestPlayers(count int) []*model.Player {
	players := make([]*model.Player, count)
	for i := range count {
		objectID := world.IDGenerator().NextPlayerID()
		player, err := model.NewPlayer(
			objectID,
			int64(i+1),  // characterID
			int64(1),    // accountID
			"TestChar",  // name
			1,           // level
			0,           // raceID
			0,           // classID
		)
		if err != nil {
			panic(err)
		}
		players[i] = player
	}
	return players
}
