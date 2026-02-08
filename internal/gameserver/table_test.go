package gameserver

import (
	"sync"
	"testing"
)

func TestGameServerTable_Register(t *testing.T) {
	table := NewGameServerTable(nil)
	info := NewGameServerInfo(1, []byte{0x01, 0x02})

	// Успешная регистрация
	if !table.Register(1, info) {
		t.Error("Expected Register to succeed for new ID")
	}

	// Дубликат ID → false
	info2 := NewGameServerInfo(1, []byte{0x03, 0x04})
	if table.Register(1, info2) {
		t.Error("Expected Register to fail for duplicate ID")
	}
}

func TestGameServerTable_GetByID(t *testing.T) {
	table := NewGameServerTable(nil)
	info := NewGameServerInfo(1, []byte{0x01, 0x02})

	// Регистрируем
	table.Register(1, info)

	// Получаем существующий
	retrieved, ok := table.GetByID(1)
	if !ok {
		t.Error("Expected GetByID to return true for existing ID")
	}
	if retrieved.ID() != 1 {
		t.Errorf("Expected ID=1, got %d", retrieved.ID())
	}

	// Несуществующий ID
	_, ok = table.GetByID(999)
	if ok {
		t.Error("Expected GetByID to return false for non-existent ID")
	}
}

func TestGameServerTable_RegisterWithFirstAvailableID(t *testing.T) {
	table := NewGameServerTable(nil)

	// Регистрируем серверы с ID 1, 2, 4
	table.Register(1, NewGameServerInfo(1, []byte{0x01}))
	table.Register(2, NewGameServerInfo(2, []byte{0x02}))
	table.Register(4, NewGameServerInfo(4, []byte{0x04}))

	// Новый сервер без ID → должен получить ID=3 (первый свободный)
	info := NewGameServerInfo(0, []byte{0x99})
	availableID, ok := table.RegisterWithFirstAvailableID(info, 10)
	if !ok {
		t.Error("Expected RegisterWithFirstAvailableID to succeed")
	}
	if availableID != 3 {
		t.Errorf("Expected first available ID=3, got %d", availableID)
	}
	if info.ID() != 3 {
		t.Errorf("Expected info.ID() to be updated to 3, got %d", info.ID())
	}

	// Проверяем, что зарегистрирован под ID=3
	retrieved, ok := table.GetByID(3)
	if !ok || retrieved.ID() != 3 {
		t.Error("Expected server to be registered with ID=3")
	}
}

func TestGameServerTable_RegisterWithFirstAvailableID_NoFreeSlots(t *testing.T) {
	table := NewGameServerTable(nil)

	// Заполняем все слоты 1-5
	for i := 1; i <= 5; i++ {
		table.Register(i, NewGameServerInfo(i, []byte{byte(i)}))
	}

	// Попытка зарегистрировать с maxID=5 → должно вернуть false
	info := NewGameServerInfo(0, []byte{0x99})
	_, ok := table.RegisterWithFirstAvailableID(info, 5)
	if ok {
		t.Error("Expected RegisterWithFirstAvailableID to fail when no free slots")
	}
}

func TestGameServerTable_ValidateHexID(t *testing.T) {
	table := NewGameServerTable(nil)
	hexID := []byte{0x01, 0x02, 0x03, 0x04}
	info := NewGameServerInfo(1, hexID)

	table.Register(1, info)

	// Правильный HexID → true
	if !table.ValidateHexID(1, hexID) {
		t.Error("Expected ValidateHexID to return true for correct hexID")
	}

	// Неправильный HexID → false
	wrongHexID := []byte{0x99, 0x99, 0x99, 0x99}
	if table.ValidateHexID(1, wrongHexID) {
		t.Error("Expected ValidateHexID to return false for wrong hexID")
	}

	// Несуществующий ID → false
	if table.ValidateHexID(999, hexID) {
		t.Error("Expected ValidateHexID to return false for non-existent ID")
	}
}

func TestGameServerTable_ConcurrentRegister(t *testing.T) {
	table := NewGameServerTable(nil)
	var wg sync.WaitGroup

	// Concurrent Register разных ID
	for i := 1; i <= 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			info := NewGameServerInfo(id, []byte{byte(id)})
			table.Register(id, info)
		}(i)
	}

	wg.Wait()

	// Проверяем, что все зарегистрированы
	for i := 1; i <= 100; i++ {
		if _, ok := table.GetByID(i); !ok {
			t.Errorf("Expected ID=%d to be registered", i)
		}
	}
}

func TestGameServerTable_ConcurrentGetByID(t *testing.T) {
	table := NewGameServerTable(nil)
	info := NewGameServerInfo(1, []byte{0x01})
	table.Register(1, info)

	var wg sync.WaitGroup

	// Concurrent GetByID
	for range 100 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			retrieved, ok := table.GetByID(1)
			if !ok {
				t.Error("Expected GetByID to return true")
			}
			if retrieved.ID() != 1 {
				t.Error("Expected ID=1")
			}
		}()
	}

	wg.Wait()
}

func TestGameServerTable_List(t *testing.T) {
	table := NewGameServerTable(nil)

	// Регистрируем 3 сервера
	table.Register(1, NewGameServerInfo(1, []byte{0x01}))
	table.Register(2, NewGameServerInfo(2, []byte{0x02}))
	table.Register(3, NewGameServerInfo(3, []byte{0x03}))

	// Получаем список
	servers := table.List()
	if len(servers) != 3 {
		t.Errorf("Expected 3 servers, got %d", len(servers))
	}

	// Проверяем, что все ID присутствуют
	ids := make(map[int]bool)
	for _, info := range servers {
		ids[info.ID()] = true
	}
	for i := 1; i <= 3; i++ {
		if !ids[i] {
			t.Errorf("Expected ID=%d in list", i)
		}
	}
}

func TestGameServerTable_Remove(t *testing.T) {
	table := NewGameServerTable(nil)
	info := NewGameServerInfo(1, []byte{0x01})

	table.Register(1, info)

	// Проверяем, что существует
	if _, ok := table.GetByID(1); !ok {
		t.Error("Expected server to exist before removal")
	}

	// Удаляем
	table.Remove(1)

	// Проверяем, что удалён
	if _, ok := table.GetByID(1); ok {
		t.Error("Expected server to be removed")
	}
}

// Mock для тестирования LoadFromDB (без реального DB)
func TestGameServerTable_LoadFromDB_Mock(t *testing.T) {
	// Этот тест будет реализован когда появится DB интерфейс
	// Пока что пропускаем
	t.Skip("LoadFromDB requires database interface")
}
