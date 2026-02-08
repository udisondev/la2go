package login

import (
	"fmt"
	"testing"
)

// BenchmarkSessionManager_Store — запись сессии (write lock)
func BenchmarkSessionManager_Store(b *testing.B) {
	b.ReportAllocs()

	sm := NewSessionManager()
	sk := SessionKey{LoginOkID1: 1, LoginOkID2: 2, PlayOkID1: 3, PlayOkID2: 4}

	b.ResetTimer()
	for i := range b.N {
		account := fmt.Sprintf("user_%d", i)
		sm.Store(account, sk, nil)
	}
}

// BenchmarkSessionManager_Validate — RLock на каждый PlayerAuthRequest (P1 hotpath)
func BenchmarkSessionManager_Validate(b *testing.B) {
	b.ReportAllocs()

	sm := NewSessionManager()
	sk := SessionKey{LoginOkID1: 1, LoginOkID2: 2, PlayOkID1: 3, PlayOkID2: 4}
	sm.Store("test_account", sk, nil)

	b.ResetTimer()
	for range b.N {
		if !sm.Validate("test_account", sk, false) {
			b.Fatal("validation failed")
		}
	}
}

// BenchmarkSessionManager_Validate_WithLicence — проверка всех 4 ключей
func BenchmarkSessionManager_Validate_WithLicence(b *testing.B) {
	b.ReportAllocs()

	sm := NewSessionManager()
	sk := SessionKey{LoginOkID1: 1, LoginOkID2: 2, PlayOkID1: 3, PlayOkID2: 4}
	sm.Store("test_account", sk, nil)

	b.ResetTimer()
	for range b.N {
		if !sm.Validate("test_account", sk, true) {
			b.Fatal("validation failed")
		}
	}
}

// BenchmarkSessionManager_Validate_NotFound — негативный случай (аккаунт не найден)
func BenchmarkSessionManager_Validate_NotFound(b *testing.B) {
	b.ReportAllocs()

	sm := NewSessionManager()
	sk := SessionKey{LoginOkID1: 1, LoginOkID2: 2, PlayOkID1: 3, PlayOkID2: 4}

	b.ResetTimer()
	for range b.N {
		if sm.Validate("non_existent", sk, false) {
			b.Fatal("unexpected success")
		}
	}
}

// BenchmarkSessionManager_Validate_WithManyAccounts — валидация в большой таблице
func BenchmarkSessionManager_Validate_WithManyAccounts(b *testing.B) {
	accountCounts := []int{100, 1000, 10000, 50000}

	for _, count := range accountCounts {
		b.Run(fmt.Sprintf("accounts=%d", count), func(b *testing.B) {
			b.ReportAllocs()

			sm := NewSessionManager()

			// Заполняем SessionManager
			for i := range count {
				sk := SessionKey{
					LoginOkID1: int32(i),
					LoginOkID2: int32(i + 1),
					PlayOkID1:  int32(i + 2),
					PlayOkID2:  int32(i + 3),
				}
				sm.Store(fmt.Sprintf("user_%d", i), sk, nil)
			}

			// Проверяем mid-range аккаунт
			targetAccount := fmt.Sprintf("user_%d", count/2)
			targetSK := SessionKey{
				LoginOkID1: int32(count / 2),
				LoginOkID2: int32(count/2 + 1),
				PlayOkID1:  int32(count/2 + 2),
				PlayOkID2:  int32(count/2 + 3),
			}

			b.ResetTimer()
			for range b.N {
				if !sm.Validate(targetAccount, targetSK, false) {
					b.Fatal("validation failed")
				}
			}
		})
	}
}

// BenchmarkSessionManager_Validate_Concurrent — реальная параллельная нагрузка (P1)
func BenchmarkSessionManager_Validate_Concurrent(b *testing.B) {
	b.ReportAllocs()

	sm := NewSessionManager()

	// Заполняем 1000 сессий
	for i := range 1000 {
		sk := SessionKey{
			LoginOkID1: int32(i),
			LoginOkID2: int32(i + 1),
			PlayOkID1:  int32(i + 2),
			PlayOkID2:  int32(i + 3),
		}
		sm.Store(fmt.Sprintf("user_%d", i), sk, nil)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		// Каждая горутина проверяет свой аккаунт
		sk := SessionKey{
			LoginOkID1: 500,
			LoginOkID2: 501,
			PlayOkID1:  502,
			PlayOkID2:  503,
		}
		for pb.Next() {
			if !sm.Validate("user_500", sk, false) {
				b.Fatal("validation failed")
			}
		}
	})
}

// BenchmarkSessionManager_Concurrent_ReadWrite — смешанная нагрузка (90% read, 10% write)
func BenchmarkSessionManager_Concurrent_ReadWrite(b *testing.B) {
	b.ReportAllocs()

	sm := NewSessionManager()

	// Заполняем 1000 сессий
	for i := range 1000 {
		sk := SessionKey{
			LoginOkID1: int32(i),
			LoginOkID2: int32(i + 1),
			PlayOkID1:  int32(i + 2),
			PlayOkID2:  int32(i + 3),
		}
		sm.Store(fmt.Sprintf("user_%d", i), sk, nil)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		sk := SessionKey{
			LoginOkID1: 500,
			LoginOkID2: 501,
			PlayOkID1:  502,
			PlayOkID2:  503,
		}

		iteration := 0
		for pb.Next() {
			if iteration%10 == 0 {
				// 10% — write (Store)
				sm.Store("user_500", sk, nil)
			} else {
				// 90% — read (Validate)
				sm.Validate("user_500", sk, false)
			}
			iteration++
		}
	})
}

// BenchmarkSessionManager_Remove — удаление сессии (write lock)
func BenchmarkSessionManager_Remove(b *testing.B) {
	b.ReportAllocs()

	// Setup: создаем N аккаунтов для удаления
	sm := NewSessionManager()
	sk := SessionKey{LoginOkID1: 1, LoginOkID2: 2, PlayOkID1: 3, PlayOkID2: 4}

	for i := range b.N {
		account := fmt.Sprintf("user_%d", i)
		sm.Store(account, sk, nil)
	}

	b.ResetTimer()
	for i := range b.N {
		account := fmt.Sprintf("user_%d", i)
		sm.Remove(account)
	}
}

// BenchmarkSessionManager_Count — подсчет активных сессий (read lock)
func BenchmarkSessionManager_Count(b *testing.B) {
	b.ReportAllocs()

	sm := NewSessionManager()
	sk := SessionKey{LoginOkID1: 1, LoginOkID2: 2, PlayOkID1: 3, PlayOkID2: 4}

	// Заполняем 1000 сессий
	for i := range 1000 {
		sm.Store(fmt.Sprintf("user_%d", i), sk, nil)
	}

	b.ResetTimer()
	for range b.N {
		count := sm.Count()
		if count != 1000 {
			b.Fatalf("expected 1000, got %d", count)
		}
	}
}

// BenchmarkSessionManager_CleanExpired — очистка устаревших сессий
func BenchmarkSessionManager_CleanExpired(b *testing.B) {
	b.ReportAllocs()

	sm := NewSessionManager()
	sk := SessionKey{LoginOkID1: 1, LoginOkID2: 2, PlayOkID1: 3, PlayOkID2: 4}

	// Заполняем 10000 сессий
	for i := range 10000 {
		sm.Store(fmt.Sprintf("user_%d", i), sk, nil)
	}

	b.ResetTimer()
	for range b.N {
		sm.CleanExpired(0) // TTL=0 — удалит все сессии
		// Re-populate для следующей итерации
		for i := range 10000 {
			sm.Store(fmt.Sprintf("user_%d", i), sk, nil)
		}
	}
}
