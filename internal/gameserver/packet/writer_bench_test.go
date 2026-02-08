package packet

import (
	"testing"
)

// BenchmarkWriter_WriteByte — запись одного байта (P0 hotpath)
func BenchmarkWriter_WriteByte(b *testing.B) {
	b.ReportAllocs()

	b.ResetTimer()
	for range b.N {
		w := NewWriter(1024)
		for range 100 {
			if err := w.WriteByte(0x42); err != nil {
				b.Fatal(err)
			}
		}
	}
}

// BenchmarkWriter_WriteInt — запись int32 (P0 hotpath, часто используется)
func BenchmarkWriter_WriteInt(b *testing.B) {
	b.ReportAllocs()

	b.ResetTimer()
	for range b.N {
		w := NewWriter(1024)
		for range 50 {
			w.WriteInt(0x12345678)
		}
	}
}

// BenchmarkWriter_WriteString_Short — запись короткой строки (UTF-16LE, ~10 символов)
func BenchmarkWriter_WriteString_Short(b *testing.B) {
	b.ReportAllocs()

	str := "TestUser"

	b.ResetTimer()
	for range b.N {
		w := NewWriter(256)
		w.WriteString(str)
	}
}

// BenchmarkWriter_WriteString_Long — запись длинной строки (UTF-16LE, ~100 символов)
func BenchmarkWriter_WriteString_Long(b *testing.B) {
	b.ReportAllocs()

	str := "ThisIsAVeryLongAccountNameThatMightBeUsedInSomeEdgeCasesForTestingPurposesAndPerformanceAnalysisOf"

	b.ResetTimer()
	for range b.N {
		w := NewWriter(512)
		w.WriteString(str)
	}
}

// BenchmarkWriter_WriteBytes — запись массива байт
func BenchmarkWriter_WriteBytes(b *testing.B) {
	sizes := []int{16, 64, 256, 1024}

	for _, size := range sizes {
		b.Run("size="+string(rune(size)), func(b *testing.B) {
			b.ReportAllocs()

			data := make([]byte, size)
			for i := range data {
				data[i] = byte(i % 256)
			}

			b.ResetTimer()
			for range b.N {
				w := NewWriter(size * 2)
				w.WriteBytes(data)
			}
		})
	}
}

// BenchmarkWriter_MixedPacket — реалистичный packet (KeyPacket response example)
// Packet структура: byte (opcode) + byte (unknown) + 16×byte (blowfish key)
func BenchmarkWriter_MixedPacket(b *testing.B) {
	b.ReportAllocs()

	key := make([]byte, 16)
	for i := range key {
		key[i] = byte(i + 1)
	}

	b.ResetTimer()
	for range b.N {
		w := NewWriter(256)

		// Write opcode
		if err := w.WriteByte(0x2E); err != nil {
			b.Fatal(err)
		}

		// Write unknown byte
		if err := w.WriteByte(0x01); err != nil {
			b.Fatal(err)
		}

		// Write blowfish key
		w.WriteBytes(key)
	}
}

// BenchmarkWriter_Reset — переиспользование Writer через Reset
func BenchmarkWriter_Reset(b *testing.B) {
	b.ReportAllocs()

	w := NewWriter(256)

	b.ResetTimer()
	for range b.N {
		w.WriteInt(0x12345678)
		w.WriteString("TestUser")
		_ = w.Bytes()
		w.Reset()
	}
}

// BenchmarkWriter_vs_NewWriter — сравнение Reset vs создание нового Writer
func BenchmarkWriter_vs_NewWriter(b *testing.B) {
	b.Run("NewWriter_each_time", func(b *testing.B) {
		b.ReportAllocs()

		b.ResetTimer()
		for range b.N {
			w := NewWriter(256)
			w.WriteInt(0x12345678)
			w.WriteString("TestUser")
			_ = w.Bytes()
		}
	})

	b.Run("Reset_reuse", func(b *testing.B) {
		b.ReportAllocs()

		w := NewWriter(256)

		b.ResetTimer()
		for range b.N {
			w.WriteInt(0x12345678)
			w.WriteString("TestUser")
			_ = w.Bytes()
			w.Reset()
		}
	})
}
