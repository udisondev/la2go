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
