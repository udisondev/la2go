package protocol

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/udisondev/la2go/internal/constants"
	"github.com/udisondev/la2go/internal/crypto"
)

// WritePacket encrypts payload in-place and writes the packet to w.
// Precondition: payload lives at buf[constants.PacketHeaderSize : constants.PacketHeaderSize+payloadLen].
// buf must have enough room for header + payload + encryption padding.
func WritePacket(w io.Writer, enc *crypto.LoginEncryption, buf []byte, payloadLen int) error {
	needed := constants.PacketHeaderSize + payloadLen + constants.PacketBufferPadding
	if len(buf) < needed {
		return fmt.Errorf("write packet: buffer too small (need %d, have %d)", needed, len(buf))
	}

	clear(buf[constants.PacketHeaderSize+payloadLen : needed])

	encSize, err := enc.EncryptPacket(buf, constants.PacketHeaderSize, payloadLen)
	if err != nil {
		return fmt.Errorf("encrypting packet: %w", err)
	}

	totalLen := constants.PacketHeaderSize + encSize
	binary.LittleEndian.PutUint16(buf[:constants.PacketHeaderSize], uint16(totalLen))

	if _, err := w.Write(buf[:totalLen]); err != nil {
		return fmt.Errorf("writing packet: %w", err)
	}
	return nil
}

// ReadPacket reads one packet from r into buf.
// Returns a subslice of buf with the decrypted payload (without the length header).
func ReadPacket(r io.Reader, enc *crypto.LoginEncryption, buf []byte) ([]byte, error) {
	var header [constants.PacketHeaderSize]byte
	if _, err := io.ReadFull(r, header[:]); err != nil {
		return nil, fmt.Errorf("reading packet header: %w", err)
	}

	totalLen := int(binary.LittleEndian.Uint16(header[:]))
	if totalLen < constants.PacketHeaderSize {
		return nil, fmt.Errorf("invalid packet length: %d", totalLen)
	}

	payloadLen := totalLen - constants.PacketHeaderSize
	if payloadLen == 0 {
		return nil, fmt.Errorf("empty packet")
	}

	if payloadLen > len(buf) {
		return nil, fmt.Errorf("packet payload %d exceeds buffer size %d", payloadLen, len(buf))
	}

	payload := buf[:payloadLen]
	if _, err := io.ReadFull(r, payload); err != nil {
		return nil, fmt.Errorf("reading packet payload: %w", err)
	}

	ok, err := enc.DecryptPacket(payload, 0, payloadLen)
	if err != nil {
		return nil, fmt.Errorf("decrypting packet: %w", err)
	}
	if !ok {
		return nil, fmt.Errorf("packet checksum verification failed")
	}

	return payload, nil
}
