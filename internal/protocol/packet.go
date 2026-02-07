package protocol

import (
	"encoding/binary"
	"fmt"
	"io"
	"log/slog"

	"github.com/udisondev/la2go/internal/crypto"
)

// WritePacket encrypts payload in-place and writes the packet to w.
// Precondition: payload lives at buf[2 : 2+payloadLen].
// buf must have enough room for 2-byte length header + payload + encryption padding (~16 bytes).
func WritePacket(w io.Writer, enc *crypto.LoginEncryption, buf []byte, payloadLen int) error {
	needed := 2 + payloadLen + 16
	if len(buf) < needed {
		return fmt.Errorf("write packet: buffer too small (need %d, have %d)", needed, len(buf))
	}

	clear(buf[2+payloadLen : needed])

	encSize, err := enc.EncryptPacket(buf, 2, payloadLen)
	if err != nil {
		return fmt.Errorf("encrypting packet: %w", err)
	}

	totalLen := 2 + encSize
	binary.LittleEndian.PutUint16(buf[:2], uint16(totalLen))

	if _, err := w.Write(buf[:totalLen]); err != nil {
		return fmt.Errorf("writing packet: %w", err)
	}
	return nil
}

// ReadPacket reads one packet from r into buf.
// Returns a subslice of buf with the decrypted payload (without the 2-byte length header).
func ReadPacket(r io.Reader, enc *crypto.LoginEncryption, buf []byte) ([]byte, error) {
	var header [2]byte
	if _, err := io.ReadFull(r, header[:]); err != nil {
		return nil, fmt.Errorf("reading packet header: %w", err)
	}

	totalLen := int(binary.LittleEndian.Uint16(header[:]))
	if totalLen < 2 {
		return nil, fmt.Errorf("invalid packet length: %d", totalLen)
	}

	payloadLen := totalLen - 2
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
		slog.Warn("packet checksum verification failed")
	}

	return payload, nil
}
