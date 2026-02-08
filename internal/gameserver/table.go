package gameserver

import (
	"bytes"
	"context"
	"fmt"
	"sync"

	"github.com/udisondev/la2go/internal/db"
)

// GameServerTable — registry всех зарегистрированных GameServer'ов.
// Thread-safe через sync.RWMutex.
// Использует bitmap для O(1) поиска свободных ID (вместо O(N) линейного сканирования).
type GameServerTable struct {
	mu         sync.RWMutex
	servers    map[int]*GameServerInfo
	db         *db.DB
	freeBitmap [2]uint64 // 128 бит для ID 1..127 (бит 0 не используется)
}

// NewGameServerTable создаёт новую таблицу GameServer'ов.
// Инициализирует freeBitmap с всеми битами=1 (все ID свободны).
func NewGameServerTable(database *db.DB) *GameServerTable {
	return &GameServerTable{
		servers: make(map[int]*GameServerInfo),
		db:      database,
		// Инициализируем все биты=1 (все ID свободны)
		freeBitmap: [2]uint64{^uint64(0), ^uint64(0)},
	}
}

// Register регистрирует GameServer с указанным ID.
// Возвращает false, если ID уже занят.
func (gst *GameServerTable) Register(id int, info *GameServerInfo) bool {
	gst.mu.Lock()
	defer gst.mu.Unlock()

	if _, exists := gst.servers[id]; exists {
		return false
	}

	gst.servers[id] = info
	gst.markIDUsed(id)
	return true
}

// GetByID возвращает GameServerInfo по ID.
func (gst *GameServerTable) GetByID(id int) (*GameServerInfo, bool) {
	gst.mu.RLock()
	defer gst.mu.RUnlock()

	info, ok := gst.servers[id]
	return info, ok
}

// RegisterWithFirstAvailableID регистрирует GameServer с первым свободным ID (1..maxID).
// Возвращает присвоенный ID и true при успехе.
// Использует bitmap для быстрого O(1) поиска свободного ID.
func (gst *GameServerTable) RegisterWithFirstAvailableID(info *GameServerInfo, maxID int) (int, bool) {
	gst.mu.Lock()
	defer gst.mu.Unlock()

	// Ищем первый свободный ID через bitmap
	id := gst.firstAvailableID(maxID)
	if id == 0 {
		return 0, false
	}

	info.SetID(id)
	gst.servers[id] = info
	gst.markIDUsed(id)
	return id, true
}

// firstAvailableID находит первый свободный ID в диапазоне 1..maxID через bitmap.
// Возвращает 0 если свободных ID нет.
func (gst *GameServerTable) firstAvailableID(maxID int) int {
	// Проверяем первые 64 бита (ID 0..63)
	if maxID >= 1 && maxID <= 63 {
		for id := 1; id <= maxID; id++ {
			if gst.freeBitmap[0]&(1<<id) != 0 {
				return id
			}
		}
		return 0
	}

	// Проверяем первые 64 бита
	for id := 1; id <= 63; id++ {
		if gst.freeBitmap[0]&(1<<id) != 0 {
			return id
		}
	}

	// Проверяем вторые 64 бита (ID 64..127)
	for id := 64; id <= maxID && id <= 127; id++ {
		bitPos := id - 64
		if gst.freeBitmap[1]&(1<<bitPos) != 0 {
			return id
		}
	}

	return 0
}

// markIDUsed помечает ID как занятый (сбрасывает бит в 0).
func (gst *GameServerTable) markIDUsed(id int) {
	if id < 64 {
		gst.freeBitmap[0] &^= 1 << id
	} else {
		bitPos := id - 64
		gst.freeBitmap[1] &^= 1 << bitPos
	}
}

// markIDFree помечает ID как свободный (устанавливает бит в 1).
func (gst *GameServerTable) markIDFree(id int) {
	if id < 64 {
		gst.freeBitmap[0] |= 1 << id
	} else {
		bitPos := id - 64
		gst.freeBitmap[1] |= 1 << bitPos
	}
}

// ValidateHexID проверяет, что HexID совпадает с зарегистрированным для данного ID.
func (gst *GameServerTable) ValidateHexID(id int, hexID []byte) bool {
	gst.mu.RLock()
	defer gst.mu.RUnlock()

	info, ok := gst.servers[id]
	if !ok {
		return false
	}

	return bytes.Equal(info.HexID(), hexID)
}

// List возвращает список всех зарегистрированных GameServer'ов (копию).
func (gst *GameServerTable) List() []*GameServerInfo {
	gst.mu.RLock()
	defer gst.mu.RUnlock()

	servers := make([]*GameServerInfo, 0, len(gst.servers))
	for _, info := range gst.servers {
		servers = append(servers, info)
	}

	return servers
}

// Remove удаляет GameServer по ID.
func (gst *GameServerTable) Remove(id int) {
	gst.mu.Lock()
	defer gst.mu.Unlock()

	delete(gst.servers, id)
	gst.markIDFree(id)
}

// LoadFromDB загружает зарегистрированные GameServer'ы из БД.
func (gst *GameServerTable) LoadFromDB(ctx context.Context) error {
	if gst.db == nil {
		return fmt.Errorf("database is nil")
	}

	rows, err := gst.db.Pool().Query(ctx, "SELECT server_id, hexid, host FROM gameservers")
	if err != nil {
		return fmt.Errorf("querying gameservers: %w", err)
	}
	defer rows.Close()

	gst.mu.Lock()
	defer gst.mu.Unlock()

	for rows.Next() {
		var id int
		var hexIDStr string
		var host *string

		if err := rows.Scan(&id, &hexIDStr, &host); err != nil {
			return fmt.Errorf("scanning gameserver row: %w", err)
		}

		// Конвертируем hex string в []byte
		hexID, err := hexStringToBytes(hexIDStr)
		if err != nil {
			return fmt.Errorf("parsing hexid for server %d: %w", id, err)
		}

		info := NewGameServerInfo(id, hexID)
		gst.servers[id] = info
		gst.markIDUsed(id)
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterating gameservers: %w", err)
	}

	return nil
}

// hexStringToBytes конвертирует hex string (например "00000001") в []byte.
func hexStringToBytes(hexStr string) ([]byte, error) {
	if len(hexStr)%2 != 0 {
		return nil, fmt.Errorf("hex string must have even length")
	}

	result := make([]byte, len(hexStr)/2)
	for i := 0; i < len(hexStr); i += 2 {
		var b byte
		_, err := fmt.Sscanf(hexStr[i:i+2], "%02x", &b)
		if err != nil {
			return nil, fmt.Errorf("parsing hex byte at position %d: %w", i, err)
		}
		result[i/2] = b
	}

	return result, nil
}
