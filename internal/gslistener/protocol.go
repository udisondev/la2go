package gslistener

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/udisondev/la2go/internal/constants"
	"github.com/udisondev/la2go/internal/crypto"
)

// WritePacket encrypts payload in-place and writes the packet to w.
// Precondition: payload lives at buf[constants.PacketHeaderSize : constants.PacketHeaderSize+payloadLen].
// buf must have enough room for header + payload + checksum + padding.
//
// GS↔LS protocol format:
// - constants.PacketHeaderSize-byte length header (LE)
// - encrypted payload (payload + checksum + padding to multiple of constants.PacketPaddingAlign)
func WritePacket(w io.Writer, cipher *crypto.BlowfishCipher, buf []byte, payloadLen int) error {
	minBufSize := constants.PacketHeaderSize + constants.PacketBufferPadding
	if payloadLen < 0 || payloadLen > len(buf)-minBufSize {
		return fmt.Errorf("invalid payload length: %d", payloadLen)
	}

	// Padding до кратности constants.PacketPaddingAlign
	dataSize := payloadLen + constants.PacketChecksumSize
	padding := (constants.PacketPaddingAlign - (dataSize % constants.PacketPaddingAlign)) % constants.PacketPaddingAlign
	encryptedSize := dataSize + padding

	// Append checksum
	crypto.AppendChecksum(buf, constants.PacketHeaderSize, encryptedSize)

	// Encrypt in-place
	cipher.Encrypt(buf, constants.PacketHeaderSize, encryptedSize)

	// Write length header (LE)
	totalSize := constants.PacketHeaderSize + encryptedSize
	binary.LittleEndian.PutUint16(buf[0:constants.PacketHeaderSize], uint16(totalSize))

	// Write packet
	_, err := w.Write(buf[0:totalSize])
	if err != nil {
		return fmt.Errorf("writing packet: %w", err)
	}

	return nil
}

// ReadPacket reads one packet from r into buf.
// Returns a subslice of buf with the decrypted payload (без checksum и padding).
//
// GS↔LS protocol format:
// - constants.PacketHeaderSize-byte length header (LE)
// - encrypted payload
func ReadPacket(r io.Reader, cipher *crypto.BlowfishCipher, buf []byte) ([]byte, error) {
	// Read length header
	var header [constants.PacketHeaderSize]byte
	_, err := io.ReadFull(r, header[:])
	if err != nil {
		return nil, fmt.Errorf("reading packet header: %w", err)
	}

	totalLen := binary.LittleEndian.Uint16(header[:])
	if totalLen < constants.PacketHeaderSize {
		return nil, fmt.Errorf("invalid packet length: %d", totalLen)
	}

	encryptedSize := int(totalLen) - constants.PacketHeaderSize
	if encryptedSize > len(buf) {
		return nil, fmt.Errorf("packet too large: %d bytes (buffer: %d)", encryptedSize, len(buf))
	}

	// Read encrypted payload
	payload := buf[0:encryptedSize]
	_, err = io.ReadFull(r, payload)
	if err != nil {
		return nil, fmt.Errorf("reading encrypted payload: %w", err)
	}

	// Decrypt in-place
	cipher.Decrypt(buf, 0, encryptedSize)

	// Verify checksum
	valid := crypto.VerifyChecksum(buf, 0, encryptedSize)
	if !valid {
		return nil, fmt.Errorf("checksum verification failed")
	}

	// Payload = decrypted data - checksum
	payloadLen := encryptedSize - constants.PacketChecksumSize
	return buf[0:payloadLen], nil
}
