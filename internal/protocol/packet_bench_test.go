package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"testing"

	"github.com/udisondev/la2go/internal/constants"
	"github.com/udisondev/la2go/internal/crypto"
)

// BenchmarkReadPacket_Full measures full packet read with Blowfish decrypt for different packet sizes.
func BenchmarkReadPacket_Full(b *testing.B) {
	sizes := []int{64, 128, 256, 512, 1024}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("size=%d", size), func(b *testing.B) {
			b.ReportAllocs()

			// Setup encryption
			key := make([]byte, 16)
			for i := range key {
				key[i] = byte(i + 1)
			}

			// Create payload (opcode + data)
			payload := make([]byte, size)
			payload[0] = 0x0E // Opcode (ProtocolVersion)
			for i := 1; i < size; i++ {
				payload[i] = byte(i % 256)
			}

			// Encrypt payload (in-place, with padding)
			encForPrep, err := crypto.NewLoginEncryption(key)
			if err != nil {
				b.Fatal(err)
			}

			// Skip first packet (Init with XOR) to use checksum encryption (GameServer mode)
			// First packet uses XOR encryption which requires extra space (8+8+8=24 bytes minimum)
			dummy := make([]byte, 32)
			_, err = encForPrep.EncryptPacket(dummy, 0, 8)
			if err != nil {
				b.Fatal(err)
			}

			buf := make([]byte, size+constants.PacketBufferPadding)
			copy(buf, payload)
			encSize, err := encForPrep.EncryptPacket(buf, 0, size)
			if err != nil {
				b.Fatal(err)
			}

			// Create packet with header
			packetData := make([]byte, constants.PacketHeaderSize+encSize)
			binary.LittleEndian.PutUint16(packetData[:constants.PacketHeaderSize], uint16(constants.PacketHeaderSize+encSize))
			copy(packetData[constants.PacketHeaderSize:], buf[:encSize])

			// Buffer for ReadPacket
			readBuf := make([]byte, 8192)

			b.SetBytes(int64(size))
			b.ResetTimer()

			for range b.N {
				// Create fresh encryption for each iteration (avoid state pollution)
				enc, err := crypto.NewLoginEncryption(key)
				if err != nil {
					b.Fatal(err)
				}

				// Skip first packet (Init with XOR) to use checksum decryption (GameServer mode)
				// First packet uses XOR encryption which requires extra space (8+8+8=24 bytes minimum)
				dummy := make([]byte, 32)
				_, err = enc.EncryptPacket(dummy, 0, 8)
				if err != nil {
					b.Fatal(err)
				}

				// Create reader mock
				reader := &mockReader{data: packetData}

				_, err = ReadPacket(reader, enc, readBuf)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkWritePacket_Full measures full packet write with Blowfish encrypt for different packet sizes.
func BenchmarkWritePacket_Full(b *testing.B) {
	sizes := []int{64, 128, 256, 512, 1024}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("size=%d", size), func(b *testing.B) {
			b.ReportAllocs()

			// Setup encryption key
			key := make([]byte, 16)
			for i := range key {
				key[i] = byte(i + 1)
			}

			// Create payload
			payload := make([]byte, size)
			payload[0] = 0x0E // Opcode
			for i := 1; i < size; i++ {
				payload[i] = byte(i % 256)
			}

			// Writer mock
			writer := &mockWriter{}

			b.SetBytes(int64(size))
			b.ResetTimer()

			for range b.N {
				// Create fresh encryption for each iteration (avoid state pollution)
				enc, err := crypto.NewLoginEncryption(key)
				if err != nil {
					b.Fatal(err)
				}

				// Skip first packet (Init with XOR) to use checksum encryption (GameServer mode)
				// First packet uses XOR encryption which requires extra space (8+8+8=24 bytes minimum)
				dummy := make([]byte, 32)
				_, err = enc.EncryptPacket(dummy, 0, 8)
				if err != nil {
					b.Fatal(err)
				}

				// Buffer for WritePacket (header + payload + padding)
				buf := make([]byte, constants.PacketHeaderSize+size+constants.PacketBufferPadding)
				copy(buf[constants.PacketHeaderSize:], payload)

				if err := WritePacket(writer, enc, buf, size); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkRoundTripPacket measures full write→read cycle (e2e overhead).
func BenchmarkRoundTripPacket(b *testing.B) {
	sizes := []int{128, 256, 512}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("size=%d", size), func(b *testing.B) {
			b.ReportAllocs()

			// Create key
			key := make([]byte, 16)
			for i := range key {
				key[i] = byte(i + 1)
			}

			// Create payload
			payload := make([]byte, size)
			payload[0] = 0x0E // Opcode
			for i := 1; i < size; i++ {
				payload[i] = byte(i % 256)
			}

			b.SetBytes(int64(size))
			b.ResetTimer()

			for range b.N {
				// Create fresh encryption for each iteration (avoid state pollution)
				encWrite, err := crypto.NewLoginEncryption(key)
				if err != nil {
					b.Fatal(err)
				}

				encRead, err := crypto.NewLoginEncryption(key)
				if err != nil {
					b.Fatal(err)
				}

				// Skip first packet (Init with XOR) for both write and read (GameServer mode)
				// First packet uses XOR encryption which requires extra space (8+8+8=24 bytes minimum)
				dummy := make([]byte, 32)
				_, err = encWrite.EncryptPacket(dummy, 0, 8)
				if err != nil {
					b.Fatal(err)
				}

				dummyRead := make([]byte, 32)
				_, err = encRead.EncryptPacket(dummyRead, 0, 8)
				if err != nil {
					b.Fatal(err)
				}

				// Write
				writeBuf := make([]byte, constants.PacketHeaderSize+size+constants.PacketBufferPadding)
				copy(writeBuf[constants.PacketHeaderSize:], payload)
				writer := &bytes.Buffer{}

				if err := WritePacket(writer, encWrite, writeBuf, size); err != nil {
					b.Fatal(err)
				}

				// Read
				reader := bytes.NewReader(writer.Bytes())
				readBuf := make([]byte, 8192)

				_, err = ReadPacket(reader, encRead, readBuf)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// mockReader is a minimal io.Reader mock for benchmarks.
type mockReader struct {
	data []byte
	pos  int
}

func (m *mockReader) Read(p []byte) (int, error) {
	if m.pos >= len(m.data) {
		return 0, io.EOF
	}
	n := copy(p, m.data[m.pos:])
	m.pos += n
	return n, nil
}

// mockWriter is a minimal io.Writer mock that discards data.
type mockWriter struct{}

func (m *mockWriter) Write(p []byte) (int, error) {
	return len(p), nil
}

// BenchmarkEncryptInPlace measures encryption overhead for single packet.
func BenchmarkEncryptInPlace(b *testing.B) {
	dynamicKey := []byte{
		0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
		0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10,
	}
	enc, err := crypto.NewLoginEncryption(dynamicKey)
	if err != nil {
		b.Fatalf("NewLoginEncryption failed: %v", err)
	}

	// Skip firstPacket
	dummyBuf := make([]byte, 1024)
	_, _ = enc.EncryptPacket(dummyBuf, constants.PacketHeaderSize, 8)

	// Test payload
	payload := make([]byte, 256) // Typical NpcInfo size
	for i := range payload {
		payload[i] = byte(i)
	}

	buf := make([]byte, 1024)

	b.ResetTimer()
	b.ReportAllocs()

	for range b.N {
		copy(buf[constants.PacketHeaderSize:], payload)
		_, err := EncryptInPlace(enc, buf, len(payload))
		if err != nil {
			b.Fatalf("EncryptInPlace failed: %v", err)
		}
	}
}

// BenchmarkWriteEncrypted measures TCP write overhead for pre-encrypted packet.
func BenchmarkWriteEncrypted(b *testing.B) {
	// Pre-encrypted packet (simulated)
	encryptedData := make([]byte, 256)
	for i := range encryptedData {
		encryptedData[i] = byte(i)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for range b.N {
		var output bytes.Buffer
		err := WriteEncrypted(&output, encryptedData, len(encryptedData))
		if err != nil {
			b.Fatalf("WriteEncrypted failed: %v", err)
		}
	}
}

// BenchmarkWriteBatch_10 measures batched write for 10 packets.
func BenchmarkWriteBatch_10(b *testing.B) {
	packets := make([][]byte, 10)
	for i := range 10 {
		pkt := make([]byte, 256)
		for j := range pkt {
			pkt[j] = byte(i + j)
		}
		packets[i] = pkt
	}

	b.ResetTimer()
	b.ReportAllocs()

	for range b.N {
		var output bytes.Buffer
		err := WriteBatch(&output, packets)
		if err != nil {
			b.Fatalf("WriteBatch failed: %v", err)
		}
	}
}

// BenchmarkWriteBatch_450 measures batched write for 450 packets (sendVisibleObjectsInfo scenario).
func BenchmarkWriteBatch_450(b *testing.B) {
	packets := make([][]byte, 450)
	for i := range 450 {
		pkt := make([]byte, 256)
		for j := range pkt {
			pkt[j] = byte(i + j)
		}
		packets[i] = pkt
	}

	b.ResetTimer()
	b.ReportAllocs()

	for range b.N {
		var output bytes.Buffer
		err := WriteBatch(&output, packets)
		if err != nil {
			b.Fatalf("WriteBatch failed: %v", err)
		}
	}
}

// BenchmarkWritePacket_Sequential measures sequential WritePacket calls (baseline).
func BenchmarkWritePacket_Sequential(b *testing.B) {
	dynamicKey := []byte{
		0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
		0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10,
	}
	enc, err := crypto.NewLoginEncryption(dynamicKey)
	if err != nil {
		b.Fatalf("NewLoginEncryption failed: %v", err)
	}

	// Skip firstPacket
	dummyBuf := make([]byte, 1024)
	_, _ = enc.EncryptPacket(dummyBuf, constants.PacketHeaderSize, 8)

	payload := make([]byte, 256)
	for i := range payload {
		payload[i] = byte(i)
	}

	buf := make([]byte, 1024)

	b.ResetTimer()
	b.ReportAllocs()

	for range b.N {
		var output bytes.Buffer
		copy(buf[constants.PacketHeaderSize:], payload)
		err := WritePacket(&output, enc, buf, len(payload))
		if err != nil {
			b.Fatalf("WritePacket failed: %v", err)
		}
	}
}
