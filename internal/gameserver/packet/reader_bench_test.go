package packet

import (
	"encoding/binary"
	"testing"
	"unicode/utf16"
)

// BenchmarkReader_ReadByte — чтение одного байта (P0 hotpath)
func BenchmarkReader_ReadByte(b *testing.B) {
	b.ReportAllocs()

	data := make([]byte, 1024)
	for i := range data {
		data[i] = byte(i % 256)
	}

	b.ResetTimer()
	for range b.N {
		r := NewReader(data)
		for range 100 {
			if _, err := r.ReadByte(); err != nil {
				b.Fatal(err)
			}
		}
	}
}

// BenchmarkReader_ReadInt — чтение int32 (P0 hotpath, часто используется)
func BenchmarkReader_ReadInt(b *testing.B) {
	b.ReportAllocs()

	data := make([]byte, 1024)
	for i := 0; i < len(data)/4; i++ {
		binary.LittleEndian.PutUint32(data[i*4:], uint32(i))
	}

	b.ResetTimer()
	for range b.N {
		r := NewReader(data)
		for range 50 {
			if _, err := r.ReadInt(); err != nil {
				b.Fatal(err)
			}
		}
	}
}

// BenchmarkReader_ReadString_Short — чтение короткой строки (UTF-16LE, ~10 символов)
func BenchmarkReader_ReadString_Short(b *testing.B) {
	b.ReportAllocs()

	// Подготовка: "TestUser" в UTF-16LE + null terminator
	str := "TestUser"
	runes := []rune(str)
	utf16Encoded := utf16.Encode(runes)

	data := make([]byte, 0, len(utf16Encoded)*2+2)
	buf := make([]byte, 2)
	for _, r := range utf16Encoded {
		binary.LittleEndian.PutUint16(buf, r)
		data = append(data, buf...)
	}
	// Null terminator
	data = append(data, 0, 0)

	b.ResetTimer()
	for range b.N {
		r := NewReader(data)
		if _, err := r.ReadString(); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkReader_ReadString_Long — чтение длинной строки (UTF-16LE, ~100 символов)
func BenchmarkReader_ReadString_Long(b *testing.B) {
	b.ReportAllocs()

	// Подготовка: длинная строка (100 символов)
	str := "ThisIsAVeryLongAccountNameThatMightBeUsedInSomeEdgeCasesForTestingPurposesAndPerformanceAnalysisOf"
	runes := []rune(str)
	utf16Encoded := utf16.Encode(runes)

	data := make([]byte, 0, len(utf16Encoded)*2+2)
	buf := make([]byte, 2)
	for _, r := range utf16Encoded {
		binary.LittleEndian.PutUint16(buf, r)
		data = append(data, buf...)
	}
	// Null terminator
	data = append(data, 0, 0)

	b.ResetTimer()
	for range b.N {
		r := NewReader(data)
		if _, err := r.ReadString(); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkReader_ReadBytes — чтение массива байт
func BenchmarkReader_ReadBytes(b *testing.B) {
	sizes := []int{16, 64, 256, 1024}

	for _, size := range sizes {
		b.Run("size="+string(rune(size)), func(b *testing.B) {
			b.ReportAllocs()

			data := make([]byte, size*2)
			for i := range data {
				data[i] = byte(i % 256)
			}

			b.ResetTimer()
			for range b.N {
				r := NewReader(data)
				if _, err := r.ReadBytes(size); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkReader_MixedPacket — реалистичный packet (AuthLogin)
// Packet структура: string (account) + 4×int32 (SessionKey) + 4×int32 (unknown)
func BenchmarkReader_MixedPacket(b *testing.B) {
	b.ReportAllocs()

	// Подготовка: account name "TestUser123"
	str := "TestUser123"
	runes := []rune(str)
	utf16Encoded := utf16.Encode(runes)

	data := make([]byte, 0, 256)
	buf := make([]byte, 2)

	// Write string
	for _, r := range utf16Encoded {
		binary.LittleEndian.PutUint16(buf, r)
		data = append(data, buf...)
	}
	// Null terminator
	data = append(data, 0, 0)

	// Write 8×int32 (SessionKey + unknown fields)
	intBuf := make([]byte, 4)
	for i := 0; i < 8; i++ {
		binary.LittleEndian.PutUint32(intBuf, uint32(i+1))
		data = append(data, intBuf...)
	}

	b.ResetTimer()
	for range b.N {
		r := NewReader(data)

		// Read string
		if _, err := r.ReadString(); err != nil {
			b.Fatal(err)
		}

		// Read 8×int32
		for range 8 {
			if _, err := r.ReadInt(); err != nil {
				b.Fatal(err)
			}
		}
	}
}
