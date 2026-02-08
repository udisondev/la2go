package gameserver

import (
	"fmt"
	"testing"
)

// BenchmarkGameServerTable_Register — регистрация с указанным ID (write lock)
func BenchmarkGameServerTable_Register(b *testing.B) {
	b.ReportAllocs()

	gsTable := NewGameServerTable(nil)

	b.ResetTimer()
	for i := range b.N {
		hexID := []byte(fmt.Sprintf("hex_%d", i))
		info := NewGameServerInfo(i, hexID)
		if !gsTable.Register(i, info) {
			b.Fatalf("failed to register server %d", i)
		}
	}
}

// BenchmarkGameServerTable_RegisterWithFirstAvailableID — P1 проблема: O(maxID) поиск
func BenchmarkGameServerTable_RegisterWithFirstAvailableID(b *testing.B) {
	b.ReportAllocs()

	// Тест worst case: 60 серверов уже зарегистрировано, ищем свободный ID
	gsTable := NewGameServerTable(nil)
	for i := 1; i <= 60; i++ {
		hexID := []byte(fmt.Sprintf("hex_%d", i))
		info := NewGameServerInfo(i, hexID)
		gsTable.Register(i, info)
	}

	b.ResetTimer()
	for range b.N {
		hexID := []byte("new_server")
		info := NewGameServerInfo(0, hexID)
		id, ok := gsTable.RegisterWithFirstAvailableID(info, 127)
		if !ok || id != 61 {
			b.Fatalf("expected id=61, got %d, ok=%v", id, ok)
		}
		// Cleanup после каждой итерации
		gsTable.Remove(61)
	}
}

// BenchmarkGameServerTable_RegisterWithFirstAvailableID_Scenarios — разные степени заполненности
func BenchmarkGameServerTable_RegisterWithFirstAvailableID_Scenarios(b *testing.B) {
	scenarios := []struct {
		name      string
		fillCount int
		maxID     int
	}{
		{"empty", 0, 127},
		{"10%", 12, 127},
		{"50%", 63, 127},
		{"90%", 114, 127},
		{"almost_full", 126, 127},
	}

	for _, sc := range scenarios {
		b.Run(sc.name, func(b *testing.B) {
			b.ReportAllocs()

			gsTable := NewGameServerTable(nil)

			// Заполняем таблицу
			for i := 1; i <= sc.fillCount; i++ {
				hexID := []byte(fmt.Sprintf("hex_%d", i))
				info := NewGameServerInfo(i, hexID)
				gsTable.Register(i, info)
			}

			expectedID := sc.fillCount + 1

			b.ResetTimer()
			for range b.N {
				hexID := []byte("new_server")
				info := NewGameServerInfo(0, hexID)
				id, ok := gsTable.RegisterWithFirstAvailableID(info, sc.maxID)
				if !ok {
					b.Fatal("failed to register")
				}
				if id != expectedID {
					b.Fatalf("expected id=%d, got %d", expectedID, id)
				}
				// Cleanup
				gsTable.Remove(id)
			}
		})
	}
}

// BenchmarkGameServerTable_GetByID — читающая операция (RLock)
func BenchmarkGameServerTable_GetByID(b *testing.B) {
	b.ReportAllocs()

	gsTable := NewGameServerTable(nil)
	hexID := []byte("test_hex")
	info := NewGameServerInfo(1, hexID)
	gsTable.Register(1, info)

	b.ResetTimer()
	for range b.N {
		_, ok := gsTable.GetByID(1)
		if !ok {
			b.Fatal("server not found")
		}
	}
}

// BenchmarkGameServerTable_GetByID_WithManyServers — чтение в большой таблице
func BenchmarkGameServerTable_GetByID_WithManyServers(b *testing.B) {
	serverCounts := []int{10, 50, 100, 127}

	for _, count := range serverCounts {
		b.Run(fmt.Sprintf("servers=%d", count), func(b *testing.B) {
			b.ReportAllocs()

			gsTable := NewGameServerTable(nil)

			// Заполняем таблицу
			for i := 1; i <= count; i++ {
				hexID := []byte(fmt.Sprintf("hex_%d", i))
				info := NewGameServerInfo(i, hexID)
				gsTable.Register(i, info)
			}

			// Проверяем mid-range сервер
			targetID := count / 2

			b.ResetTimer()
			for range b.N {
				_, ok := gsTable.GetByID(targetID)
				if !ok {
					b.Fatal("server not found")
				}
			}
		})
	}
}

// BenchmarkGameServerTable_GetByID_Concurrent — параллельная читающая нагрузка
func BenchmarkGameServerTable_GetByID_Concurrent(b *testing.B) {
	b.ReportAllocs()

	gsTable := NewGameServerTable(nil)

	// Заполняем 100 серверов
	for i := 1; i <= 100; i++ {
		hexID := []byte(fmt.Sprintf("hex_%d", i))
		info := NewGameServerInfo(i, hexID)
		gsTable.Register(i, info)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, ok := gsTable.GetByID(50)
			if !ok {
				b.Fatal("server not found")
			}
		}
	})
}

// BenchmarkGameServerTable_ValidateHexID — проверка HexID (RLock)
func BenchmarkGameServerTable_ValidateHexID(b *testing.B) {
	b.ReportAllocs()

	gsTable := NewGameServerTable(nil)
	hexID := []byte("test_hex_id_123")
	info := NewGameServerInfo(1, hexID)
	gsTable.Register(1, info)

	b.ResetTimer()
	for range b.N {
		if !gsTable.ValidateHexID(1, hexID) {
			b.Fatal("validation failed")
		}
	}
}

// BenchmarkGameServerTable_ValidateHexID_Concurrent — параллельная валидация
func BenchmarkGameServerTable_ValidateHexID_Concurrent(b *testing.B) {
	b.ReportAllocs()

	gsTable := NewGameServerTable(nil)

	// Заполняем 100 серверов
	for i := 1; i <= 100; i++ {
		hexID := []byte(fmt.Sprintf("hex_%d", i))
		info := NewGameServerInfo(i, hexID)
		gsTable.Register(i, info)
	}

	testHexID := []byte("hex_50")

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if !gsTable.ValidateHexID(50, testHexID) {
				b.Fatal("validation failed")
			}
		}
	})
}

// BenchmarkGameServerTable_List — получение списка всех серверов (копия)
func BenchmarkGameServerTable_List(b *testing.B) {
	serverCounts := []int{10, 50, 100, 127}

	for _, count := range serverCounts {
		b.Run(fmt.Sprintf("servers=%d", count), func(b *testing.B) {
			b.ReportAllocs()

			gsTable := NewGameServerTable(nil)

			// Заполняем таблицу
			for i := 1; i <= count; i++ {
				hexID := []byte(fmt.Sprintf("hex_%d", i))
				info := NewGameServerInfo(i, hexID)
				gsTable.Register(i, info)
			}

			b.ResetTimer()
			for range b.N {
				list := gsTable.List()
				if len(list) != count {
					b.Fatalf("expected %d servers, got %d", count, len(list))
				}
			}
		})
	}
}

// BenchmarkGameServerTable_Remove — удаление сервера (write lock)
func BenchmarkGameServerTable_Remove(b *testing.B) {
	b.ReportAllocs()

	// Setup: создаем N серверов для удаления
	gsTable := NewGameServerTable(nil)

	for i := range b.N {
		hexID := []byte(fmt.Sprintf("hex_%d", i))
		info := NewGameServerInfo(i, hexID)
		gsTable.Register(i, info)
	}

	b.ResetTimer()
	for i := range b.N {
		gsTable.Remove(i)
	}
}

// BenchmarkGameServerTable_Concurrent_ReadWrite — смешанная нагрузка (90% read, 10% write)
func BenchmarkGameServerTable_Concurrent_ReadWrite(b *testing.B) {
	b.ReportAllocs()

	gsTable := NewGameServerTable(nil)

	// Заполняем 100 серверов
	for i := 1; i <= 100; i++ {
		hexID := []byte(fmt.Sprintf("hex_%d", i))
		info := NewGameServerInfo(i, hexID)
		gsTable.Register(i, info)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		iteration := 0
		for pb.Next() {
			if iteration%10 == 0 {
				// 10% — write (Register новый сервер)
				hexID := []byte(fmt.Sprintf("new_hex_%d", iteration))
				info := NewGameServerInfo(101+iteration/10, hexID)
				gsTable.Register(101+iteration/10, info)
			} else {
				// 90% — read (GetByID)
				gsTable.GetByID(50)
			}
			iteration++
		}
	})
}
