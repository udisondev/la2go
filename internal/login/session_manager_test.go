package login

import (
	"sync"
	"testing"
	"time"
)

func TestSessionManager_StoreAndValidate(t *testing.T) {
	sm := NewSessionManager()
	key := SessionKey{
		LoginOkID1: 123,
		LoginOkID2: 456,
		PlayOkID1:  789,
		PlayOkID2:  101112,
	}

	// Store session
	sm.Store("testuser", key, nil)

	// Validate with SHOW_LICENCE=true (все 4 ключа)
	if !sm.Validate("testuser", key, true) {
		t.Error("Expected validation to pass with all 4 keys")
	}

	// Validate with SHOW_LICENCE=false (только PlayOk)
	if !sm.Validate("testuser", key, false) {
		t.Error("Expected validation to pass with PlayOk keys only")
	}
}

func TestSessionManager_Validate_ShowLicenceFalse(t *testing.T) {
	sm := NewSessionManager()
	key := SessionKey{
		LoginOkID1: 123,
		LoginOkID2: 456,
		PlayOkID1:  789,
		PlayOkID2:  101112,
	}

	sm.Store("testuser", key, nil)

	// Неправильные LoginOk ключи, но правильные PlayOk
	wrongKey := SessionKey{
		LoginOkID1: 999, // неправильный
		LoginOkID2: 888, // неправильный
		PlayOkID1:  789, // правильный
		PlayOkID2:  101112, // правильный
	}

	// С showLicence=false должно пройти (проверяются только PlayOk)
	if !sm.Validate("testuser", wrongKey, false) {
		t.Error("Expected validation to pass with correct PlayOk keys (showLicence=false)")
	}

	// С showLicence=true должно упасть (проверяются все 4)
	if sm.Validate("testuser", wrongKey, true) {
		t.Error("Expected validation to fail with wrong LoginOk keys (showLicence=true)")
	}
}

func TestSessionManager_ValidateNonExistent(t *testing.T) {
	sm := NewSessionManager()
	key := SessionKey{
		LoginOkID1: 123,
		LoginOkID2: 456,
		PlayOkID1:  789,
		PlayOkID2:  101112,
	}

	// Несуществующий аккаунт → false
	if sm.Validate("nonexistent", key, true) {
		t.Error("Expected validation to fail for non-existent account")
	}
}

func TestSessionManager_ValidateWrongKeys(t *testing.T) {
	sm := NewSessionManager()
	key := SessionKey{
		LoginOkID1: 123,
		LoginOkID2: 456,
		PlayOkID1:  789,
		PlayOkID2:  101112,
	}

	sm.Store("testuser", key, nil)

	// Неправильные ключи → false
	wrongKey := SessionKey{
		LoginOkID1: 999,
		LoginOkID2: 888,
		PlayOkID1:  777,
		PlayOkID2:  666,
	}

	if sm.Validate("testuser", wrongKey, true) {
		t.Error("Expected validation to fail with wrong keys")
	}
}

func TestSessionManager_Remove(t *testing.T) {
	sm := NewSessionManager()
	key := SessionKey{
		LoginOkID1: 123,
		LoginOkID2: 456,
		PlayOkID1:  789,
		PlayOkID2:  101112,
	}

	sm.Store("testuser", key, nil)

	// Validate перед удалением
	if !sm.Validate("testuser", key, true) {
		t.Error("Expected validation to pass before removal")
	}

	// Remove
	sm.Remove("testuser")

	// Validate после удаления → false
	if sm.Validate("testuser", key, true) {
		t.Error("Expected validation to fail after removal")
	}
}

func TestSessionManager_ConcurrentAccess(t *testing.T) {
	sm := NewSessionManager()
	key := SessionKey{
		LoginOkID1: 123,
		LoginOkID2: 456,
		PlayOkID1:  789,
		PlayOkID2:  101112,
	}

	var wg sync.WaitGroup
	accounts := 100

	// Concurrent Store
	for i := range accounts {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			account := "user" + string(rune('0'+idx%10))
			sm.Store(account, key, nil)
		}(i)
	}

	// Concurrent Validate
	for i := range accounts {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			account := "user" + string(rune('0'+idx%10))
			sm.Validate(account, key, true)
		}(i)
	}

	// Concurrent Remove
	for i := range accounts {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			account := "user" + string(rune('0'+idx%10))
			sm.Remove(account)
		}(i)
	}

	wg.Wait()

	// Если дошли сюда без panic, значит thread-safe
}

func TestSessionManager_ExpiredSessions(t *testing.T) {
	sm := NewSessionManager()
	key := SessionKey{
		LoginOkID1: 123,
		LoginOkID2: 456,
		PlayOkID1:  789,
		PlayOkID2:  101112,
	}

	// Создаём сессию с прошедшим временем
	info := &SessionInfo{
		SessionKey: key,
		Client:     nil,
		CreatedAt:  time.Now().Add(-2 * time.Hour),
	}
	sm.sessions.Store("testuser", info)

	// CleanExpired с TTL 1 час
	sm.CleanExpired(1 * time.Hour)

	// Сессия должна быть удалена
	if sm.Validate("testuser", key, true) {
		t.Error("Expected expired session to be removed")
	}
}

func TestSessionManager_Count(t *testing.T) {
	sm := NewSessionManager()
	key := SessionKey{
		LoginOkID1: 123,
		LoginOkID2: 456,
		PlayOkID1:  789,
		PlayOkID2:  101112,
	}

	// Изначально 0
	if sm.Count() != 0 {
		t.Errorf("Expected count=0, got %d", sm.Count())
	}

	// Добавляем сессии
	sm.Store("user1", key, nil)
	sm.Store("user2", key, nil)
	sm.Store("user3", key, nil)

	if sm.Count() != 3 {
		t.Errorf("Expected count=3, got %d", sm.Count())
	}

	// Удаляем одну
	sm.Remove("user2")

	if sm.Count() != 2 {
		t.Errorf("Expected count=2, got %d", sm.Count())
	}
}
