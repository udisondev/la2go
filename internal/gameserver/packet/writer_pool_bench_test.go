package packet

import (
	"testing"
)

// BenchmarkWriterPool_Get — получение Writer из pool
func BenchmarkWriterPool_Get(b *testing.B) {
	b.ReportAllocs()

	b.ResetTimer()
	for range b.N {
		w := Get()
		w.Put()
	}
}

// BenchmarkWriterPool_WriteString_Reuse — переиспользование Writer через pool
func BenchmarkWriterPool_WriteString_Reuse(b *testing.B) {
	b.ReportAllocs()

	str := "TestUser"

	b.ResetTimer()
	for range b.N {
		w := Get()
		w.WriteString(str)
		_ = w.Bytes()
		w.Put()
	}
}

// BenchmarkWriter_ManualEncoding_vs_NewWriter — сравнение pool vs NewWriter для realistic workload
func BenchmarkWriter_ManualEncoding_vs_NewWriter(b *testing.B) {
	b.Run("Pool_Get_Put", func(b *testing.B) {
		b.ReportAllocs()

		b.ResetTimer()
		for range b.N {
			w := Get()
			w.WriteInt(0x12345678)
			w.WriteString("TestUserAccount")
			w.WriteShort(100)
			_ = w.Bytes()
			w.Put()
		}
	})

	b.Run("NewWriter_each_time", func(b *testing.B) {
		b.ReportAllocs()

		b.ResetTimer()
		for range b.N {
			w := NewWriter(256)
			w.WriteInt(0x12345678)
			w.WriteString("TestUserAccount")
			w.WriteShort(100)
			_ = w.Bytes()
		}
	})
}

// BenchmarkWriter_WriteInt_Manual — оценка manual encoding для WriteInt
func BenchmarkWriter_WriteInt_Manual(b *testing.B) {
	b.ReportAllocs()

	b.ResetTimer()
	for range b.N {
		w := Get()
		for range 50 {
			w.WriteInt(0x12345678)
		}
		w.Put()
	}
}

// BenchmarkWriter_WriteString_Manual_Short — оценка manual encoding для короткой строки
func BenchmarkWriter_WriteString_Manual_Short(b *testing.B) {
	b.ReportAllocs()

	str := "TestUser"

	b.ResetTimer()
	for range b.N {
		w := Get()
		w.WriteString(str)
		w.Put()
	}
}

// BenchmarkWriter_WriteString_Manual_Long — оценка manual encoding для длинной строки
func BenchmarkWriter_WriteString_Manual_Long(b *testing.B) {
	b.ReportAllocs()

	str := "ThisIsAVeryLongAccountNameThatMightBeUsedInSomeEdgeCasesForTestingPurposesAndPerformanceAnalysisOf"

	b.ResetTimer()
	for range b.N {
		w := Get()
		w.WriteString(str)
		w.Put()
	}
}

// BenchmarkWriter_WriteString_Unicode — тест на Unicode с surrogates (emoji)
func BenchmarkWriter_WriteString_Unicode(b *testing.B) {
	b.ReportAllocs()

	str := "Hello🌍World🚀Test"

	b.ResetTimer()
	for range b.N {
		w := Get()
		w.WriteString(str)
		w.Put()
	}
}
