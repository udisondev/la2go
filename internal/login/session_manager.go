package login

import (
	"sync"
	"time"
)

// SessionManager управляет сессиями клиентов для relay между LS и GS.
// Thread-safe через sync.Map для оптимальной read performance.
type SessionManager struct {
	sessions sync.Map // map[string]*SessionInfo
}

// SessionInfo хранит информацию о сессии клиента.
// Экспортируется для тестирования (можно манипулировать CreatedAt).
type SessionInfo struct {
	SessionKey SessionKey
	Client     *Client
	CreatedAt  time.Time
}

// NewSessionManager создаёт новый SessionManager.
func NewSessionManager() *SessionManager {
	return &SessionManager{}
}

// Store сохраняет SessionKey для аккаунта.
func (sm *SessionManager) Store(account string, key SessionKey, client *Client) {
	info := &SessionInfo{
		SessionKey: key,
		Client:     client,
		CreatedAt:  time.Now(),
	}
	sm.sessions.Store(account, info)
}

// Validate проверяет SessionKey для аккаунта.
// showLicence=true: проверяются все 4 ключа (LoginOkID1/2, PlayOkID1/2)
// showLicence=false: проверяются только PlayOkID1/2
func (sm *SessionManager) Validate(account string, key SessionKey, showLicence bool) bool {
	val, ok := sm.sessions.Load(account)
	if !ok {
		return false
	}

	info := val.(*SessionInfo)

	if showLicence {
		// Проверяем все 4 ключа
		return info.SessionKey.LoginOkID1 == key.LoginOkID1 &&
			info.SessionKey.LoginOkID2 == key.LoginOkID2 &&
			info.SessionKey.PlayOkID1 == key.PlayOkID1 &&
			info.SessionKey.PlayOkID2 == key.PlayOkID2
	}

	// Проверяем только PlayOk ключи
	return info.SessionKey.PlayOkID1 == key.PlayOkID1 &&
		info.SessionKey.PlayOkID2 == key.PlayOkID2
}

// Remove удаляет сессию для аккаунта.
func (sm *SessionManager) Remove(account string) {
	sm.sessions.Delete(account)
}

// CleanExpired удаляет сессии старше ttl.
func (sm *SessionManager) CleanExpired(ttl time.Duration) {
	now := time.Now()
	sm.sessions.Range(func(key, value any) bool {
		account := key.(string)
		info := value.(*SessionInfo)
		if now.Sub(info.CreatedAt) > ttl {
			sm.sessions.Delete(account)
		}
		return true
	})
}

// Count возвращает количество активных сессий.
func (sm *SessionManager) Count() int {
	count := 0
	sm.sessions.Range(func(_, _ any) bool {
		count++
		return true
	})
	return count
}

// StoreInfo сохраняет готовый SessionInfo (для тестов с манипуляцией времени).
func (sm *SessionManager) StoreInfo(account string, info *SessionInfo) {
	sm.sessions.Store(account, info)
}
