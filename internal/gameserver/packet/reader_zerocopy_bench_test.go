package packet

import (
	"testing"
)

// BenchmarkReadBytes_ZeroCopy — zero-copy ReadBytes (current implementation)
func BenchmarkReadBytes_ZeroCopy(b *testing.B) {
	sizes := []int{16, 64, 256}

	for _, size := range sizes {
		b.Run("size="+string(rune('0'+size/10)), func(b *testing.B) {
			b.ReportAllocs()

			data := make([]byte, size)
			for i := range data {
				data[i] = byte(i % 256)
			}

			b.ResetTimer()
			for range b.N {
				r := NewReader(data)
				_, err := r.ReadBytes(size)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkReadBytesCopy — mutable copy version
func BenchmarkReadBytesCopy(b *testing.B) {
	sizes := []int{16, 64, 256}

	for _, size := range sizes {
		b.Run("size="+string(rune('0'+size/10)), func(b *testing.B) {
			b.ReportAllocs()

			data := make([]byte, size)
			for i := range data {
				data[i] = byte(i % 256)
			}

			b.ResetTimer()
			for range b.N {
				r := NewReader(data)
				_, err := r.ReadBytesCopy(size)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkReadBytes_ZeroCopy_vs_Copy — direct comparison
func BenchmarkReadBytes_ZeroCopy_vs_Copy(b *testing.B) {
	size := 64

	data := make([]byte, size)
	for i := range data {
		data[i] = byte(i % 256)
	}

	b.Run("ZeroCopy", func(b *testing.B) {
		b.ReportAllocs()

		b.ResetTimer()
		for range b.N {
			r := NewReader(data)
			_, err := r.ReadBytes(size)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("Copy", func(b *testing.B) {
		b.ReportAllocs()

		b.ResetTimer()
		for range b.N {
			r := NewReader(data)
			_, err := r.ReadBytesCopy(size)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkReader_ReadBytes_Multiple — realistic workload (multiple ReadBytes in packet)
func BenchmarkReader_ReadBytes_Multiple(b *testing.B) {
	// Simulate packet with multiple byte arrays
	data := make([]byte, 256)
	for i := range data {
		data[i] = byte(i % 256)
	}

	b.Run("ZeroCopy_3_reads", func(b *testing.B) {
		b.ReportAllocs()

		b.ResetTimer()
		for range b.N {
			r := NewReader(data)
			_, _ = r.ReadBytes(16) // Header
			_, _ = r.ReadBytes(64) // Payload 1
			_, _ = r.ReadBytes(32) // Payload 2
		}
	})

	b.Run("Copy_3_reads", func(b *testing.B) {
		b.ReportAllocs()

		b.ResetTimer()
		for range b.N {
			r := NewReader(data)
			_, _ = r.ReadBytesCopy(16) // Header
			_, _ = r.ReadBytesCopy(64) // Payload 1
			_, _ = r.ReadBytesCopy(32) // Payload 2
		}
	})
}
