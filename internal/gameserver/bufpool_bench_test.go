package gameserver

import (
	"testing"
)

// BenchmarkBytePool_Get — получение буфера из пула (P0 hotpath)
func BenchmarkBytePool_Get(b *testing.B) {
	b.ReportAllocs()

	pool := NewBytePool(1024)

	b.ResetTimer()
	for range b.N {
		buf := pool.Get(1024)
		pool.Put(buf)
	}
}

// BenchmarkBytePool_Get_SmallBuffer — маленький буфер (256 байт)
func BenchmarkBytePool_Get_SmallBuffer(b *testing.B) {
	b.ReportAllocs()

	pool := NewBytePool(1024)

	b.ResetTimer()
	for range b.N {
		buf := pool.Get(256)
		pool.Put(buf)
	}
}

// BenchmarkBytePool_Get_LargeBuffer — большой буфер (8192 байт)
func BenchmarkBytePool_Get_LargeBuffer(b *testing.B) {
	b.ReportAllocs()

	pool := NewBytePool(1024)

	b.ResetTimer()
	for range b.N {
		buf := pool.Get(8192)
		pool.Put(buf)
	}
}

// BenchmarkBytePool_Get_ExactCapacity — точное совпадение capacity (оптимальный случай)
func BenchmarkBytePool_Get_ExactCapacity(b *testing.B) {
	b.ReportAllocs()

	pool := NewBytePool(4096)

	b.ResetTimer()
	for range b.N {
		buf := pool.Get(4096)
		pool.Put(buf)
	}
}

// BenchmarkBytePool_vs_MakeSlice — сравнение pool vs make() каждый раз
func BenchmarkBytePool_vs_MakeSlice(b *testing.B) {
	b.Run("BytePool", func(b *testing.B) {
		b.ReportAllocs()

		pool := NewBytePool(1024)

		b.ResetTimer()
		for range b.N {
			buf := pool.Get(1024)
			// Simulate usage
			for i := range buf {
				buf[i] = byte(i % 256)
			}
			pool.Put(buf)
		}
	})

	b.Run("make_each_time", func(b *testing.B) {
		b.ReportAllocs()

		b.ResetTimer()
		for range b.N {
			buf := make([]byte, 1024)
			// Simulate usage
			for i := range buf {
				buf[i] = byte(i % 256)
			}
		}
	})
}

// BenchmarkBytePool_Concurrent — параллельный доступ к пулу (реалистичный сценарий)
func BenchmarkBytePool_Concurrent(b *testing.B) {
	b.ReportAllocs()

	pool := NewBytePool(1024)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			buf := pool.Get(1024)
			// Simulate packet processing
			for i := range buf {
				buf[i] = byte(i % 256)
			}
			pool.Put(buf)
		}
	})
}

// BenchmarkBytePool_Concurrent_MixedSizes — параллельный доступ с разными размерами
func BenchmarkBytePool_Concurrent_MixedSizes(b *testing.B) {
	b.ReportAllocs()

	pool := NewBytePool(2048)

	sizes := []int{256, 512, 1024, 2048, 4096}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			size := sizes[i%len(sizes)]
			buf := pool.Get(size)
			// Simulate packet processing
			for j := range buf {
				buf[j] = byte(j % 256)
			}
			pool.Put(buf)
			i++
		}
	})
}
