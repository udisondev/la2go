package login

import (
	"testing"
	"time"
)

// TestSessionManager_StoreInfo_CustomTime тестирует helper для манипуляции временем.
func TestSessionManager_StoreInfo_CustomTime(t *testing.T) {
	sm := NewSessionManager()
	key := SessionKey{
		LoginOkID1: 123,
		LoginOkID2: 456,
		PlayOkID1:  789,
		PlayOkID2:  101112,
	}

	// Создаём сессию с прошедшим временем через StoreInfo
	pastTime := time.Now().Add(-2 * time.Hour)
	info := &SessionInfo{
		SessionKey: key,
		Client:     nil,
		CreatedAt:  pastTime,
	}
	sm.StoreInfo("testuser", info)

	// Проверяем что сессия сохранена
	if !sm.Validate("testuser", key, true) {
		t.Error("Expected session to be valid after StoreInfo")
	}

	// CleanExpired с TTL 1 час должен удалить сессию
	sm.CleanExpired(1 * time.Hour)

	// Сессия должна быть удалена
	if sm.Validate("testuser", key, true) {
		t.Error("Expected expired session to be removed")
	}
}

// TestSessionManager_StoreInfo_FutureTime тестирует сессию с будущим временем.
func TestSessionManager_StoreInfo_FutureTime(t *testing.T) {
	sm := NewSessionManager()
	key := SessionKey{
		LoginOkID1: 123,
		LoginOkID2: 456,
		PlayOkID1:  789,
		PlayOkID2:  101112,
	}

	// Создаём сессию с будущим временем
	futureTime := time.Now().Add(2 * time.Hour)
	info := &SessionInfo{
		SessionKey: key,
		Client:     nil,
		CreatedAt:  futureTime,
	}
	sm.StoreInfo("testuser", info)

	// CleanExpired с TTL 1 час НЕ должен удалить сессию
	sm.CleanExpired(1 * time.Hour)

	// Сессия должна остаться
	if !sm.Validate("testuser", key, true) {
		t.Error("Expected future session to remain after CleanExpired")
	}
}

// TestSessionManager_StoreInfo_MultipleExpirations тестирует удаление нескольких expired сессий.
func TestSessionManager_StoreInfo_MultipleExpirations(t *testing.T) {
	sm := NewSessionManager()

	now := time.Now()

	// Создаём 3 сессии с разным временем
	sessions := []struct {
		account   string
		createdAt time.Time
		shouldExpire bool
	}{
		{"user1", now.Add(-3 * time.Hour), true},  // expired
		{"user2", now.Add(-30 * time.Minute), false}, // not expired
		{"user3", now.Add(-2 * time.Hour), true},  // expired
	}

	for _, s := range sessions {
		key := SessionKey{LoginOkID1: 1, LoginOkID2: 2, PlayOkID1: 3, PlayOkID2: 4}
		info := &SessionInfo{
			SessionKey: key,
			CreatedAt:  s.createdAt,
		}
		sm.StoreInfo(s.account, info)
	}

	// Проверяем начальное состояние
	if sm.Count() != 3 {
		t.Errorf("Expected 3 sessions initially, got %d", sm.Count())
	}

	// CleanExpired с TTL 1 час
	sm.CleanExpired(1 * time.Hour)

	// Проверяем что остался только user2
	if sm.Count() != 1 {
		t.Errorf("Expected 1 session after cleanup, got %d", sm.Count())
	}

	key := SessionKey{LoginOkID1: 1, LoginOkID2: 2, PlayOkID1: 3, PlayOkID2: 4}
	if sm.Validate("user1", key, false) {
		t.Error("Expected user1 to be expired")
	}
	if !sm.Validate("user2", key, false) {
		t.Error("Expected user2 to remain")
	}
	if sm.Validate("user3", key, false) {
		t.Error("Expected user3 to be expired")
	}
}
