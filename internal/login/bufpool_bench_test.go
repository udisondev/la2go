package login

import (
	"testing"
)

// BenchmarkBytePool_GetPut — базовый тест эффективности sync.Pool
func BenchmarkBytePool_GetPut(b *testing.B) {
	b.ReportAllocs()

	pool := NewBytePool(512)

	b.ResetTimer()
	for range b.N {
		buf := pool.Get(256)
		pool.Put(buf)
	}
}

// BenchmarkBytePool_GetPut_Sizes — разные размеры буферов
func BenchmarkBytePool_GetPut_Sizes(b *testing.B) {
	sizes := []int{64, 128, 256, 512, 1024, 2048, 4096}

	for _, size := range sizes {
		b.Run(formatSize(size), func(b *testing.B) {
			b.ReportAllocs()

			pool := NewBytePool(size)

			b.ResetTimer()
			for range b.N {
				buf := pool.Get(size)
				pool.Put(buf)
			}
		})
	}
}

// BenchmarkBytePool_vs_MakeSlice — сравнение pool vs прямая аллокация
func BenchmarkBytePool_vs_MakeSlice(b *testing.B) {
	sizes := []int{64, 128, 256, 512, 1024, 2048}

	for _, size := range sizes {
		b.Run("pool/"+formatSize(size), func(b *testing.B) {
			b.ReportAllocs()

			pool := NewBytePool(size)

			b.ResetTimer()
			for range b.N {
				buf := pool.Get(size)
				pool.Put(buf)
			}
		})

		b.Run("make/"+formatSize(size), func(b *testing.B) {
			b.ReportAllocs()

			b.ResetTimer()
			for range b.N {
				_ = make([]byte, size)
			}
		})
	}
}

// BenchmarkBytePool_Concurrent — производительность под параллельной нагрузкой
func BenchmarkBytePool_Concurrent(b *testing.B) {
	b.ReportAllocs()

	pool := NewBytePool(512)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			buf := pool.Get(256)
			pool.Put(buf)
		}
	})
}

// BenchmarkBytePool_Concurrent_Sizes — параллельная нагрузка с разными размерами
func BenchmarkBytePool_Concurrent_Sizes(b *testing.B) {
	sizes := []int{64, 128, 256, 512, 1024, 2048}

	for _, size := range sizes {
		b.Run(formatSize(size), func(b *testing.B) {
			b.ReportAllocs()

			pool := NewBytePool(size)

			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					buf := pool.Get(size)
					pool.Put(buf)
				}
			})
		})
	}
}

// BenchmarkBytePool_Concurrent_vs_MakeSlice — параллельное сравнение pool vs make
func BenchmarkBytePool_Concurrent_vs_MakeSlice(b *testing.B) {
	size := 512

	b.Run("pool", func(b *testing.B) {
		b.ReportAllocs()

		pool := NewBytePool(size)

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				buf := pool.Get(size)
				pool.Put(buf)
			}
		})
	})

	b.Run("make", func(b *testing.B) {
		b.ReportAllocs()

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = make([]byte, size)
			}
		})
	})
}

// BenchmarkBytePool_RealWorkload — имитация реального использования (Get → fill → Put)
func BenchmarkBytePool_RealWorkload(b *testing.B) {
	b.ReportAllocs()

	pool := NewBytePool(512)

	b.ResetTimer()
	for range b.N {
		buf := pool.Get(256)

		// Имитируем работу с буфером (заполнение данными)
		for i := range buf {
			buf[i] = byte(i)
		}

		pool.Put(buf)
	}
}

// BenchmarkBytePool_RealWorkload_Concurrent — реальная рабочая нагрузка параллельно
func BenchmarkBytePool_RealWorkload_Concurrent(b *testing.B) {
	b.ReportAllocs()

	pool := NewBytePool(512)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			buf := pool.Get(256)

			for i := range buf {
				buf[i] = byte(i)
			}

			pool.Put(buf)
		}
	})
}

// BenchmarkBytePool_OversizedRequest — тест случая, когда запрашиваемый размер > pool capacity
func BenchmarkBytePool_OversizedRequest(b *testing.B) {
	b.ReportAllocs()

	pool := NewBytePool(256) // Маленький pool

	b.ResetTimer()
	for range b.N {
		buf := pool.Get(1024) // Запрашиваем больше, чем capacity
		pool.Put(buf)
	}
}

// BenchmarkBytePool_Clear — overhead от clear(buf) в Get
func BenchmarkBytePool_Clear(b *testing.B) {
	b.ReportAllocs()

	sizes := []int{64, 128, 256, 512, 1024, 2048}

	for _, size := range sizes {
		b.Run(formatSize(size), func(b *testing.B) {
			b.ReportAllocs()

			buf := make([]byte, size)

			b.ResetTimer()
			for range b.N {
				clear(buf)
			}
		})
	}
}

// formatSize форматирует размер в байтах для имени бенчмарка
func formatSize(size int) string {
	switch {
	case size >= 1024:
		return string(rune('0'+size/1024)) + "KB"
	case size >= 64:
		return string(rune('0'+size/64)) + "x64B"
	default:
		return "small"
	}
}
