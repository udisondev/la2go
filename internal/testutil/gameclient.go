package testutil

import (
	"fmt"
	"io"
	"net"
	"testing"
	"time"

	"github.com/udisondev/la2go/internal/constants"
	"github.com/udisondev/la2go/internal/crypto"
	"github.com/udisondev/la2go/internal/gameserver/clientpackets"
	"github.com/udisondev/la2go/internal/gameserver/packet"
	"github.com/udisondev/la2go/internal/login"
	"github.com/udisondev/la2go/internal/protocol"
)

// GameClient is a test helper for connecting to GameServer.
// Handles encryption and packet parsing.
type GameClient struct {
	t          testing.TB
	conn       net.Conn
	encryption *crypto.LoginEncryption
	sessionID  int32
}

// NewGameClient connects to the GameServer and reads the KeyPacket.
func NewGameClient(t testing.TB, addr string) (*GameClient, error) {
	t.Helper()

	// Set connection deadline
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("connecting to %s: %w", addr, err)
	}

	conn = NewConnWithDeadline(conn, 5*time.Second)

	client := &GameClient{
		t:    t,
		conn: conn,
	}

	// Read KeyPacket (sent in plaintext, NOT encrypted)
	if err := client.readKeyPacket(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("reading KeyPacket: %w", err)
	}

	return client, nil
}

// readKeyPacket reads and parses the KeyPacket (opcode 0x00).
// KeyPacket structure:
// - byte: opcode (0x00)
// - byte: protocol version (0x01)
// - byte[16]: Blowfish key
func (c *GameClient) readKeyPacket() error {
	// Read KeyPacket (18 bytes: 1 opcode + 1 protocol + 16 key)
	keyData := make([]byte, 18)
	if _, err := io.ReadFull(c.conn, keyData); err != nil {
		return fmt.Errorf("reading KeyPacket data: %w", err)
	}

	// Verify opcode
	if keyData[0] != 0x00 {
		return fmt.Errorf("invalid KeyPacket opcode: expected 0x00, got 0x%02X", keyData[0])
	}

	// Verify protocol version
	if keyData[1] != 0x01 {
		return fmt.Errorf("invalid protocol version: expected 0x01, got 0x%02X", keyData[1])
	}

	// Extract Blowfish key (16 bytes)
	blowfishKey := keyData[2:18]

	// Initialize encryption
	enc, err := crypto.NewLoginEncryption(blowfishKey)
	if err != nil {
		return fmt.Errorf("creating encryption: %w", err)
	}

	c.encryption = enc

	return nil
}

// SendProtocolVersion sends a ProtocolVersion packet (opcode 0x0E).
func (c *GameClient) SendProtocolVersion(revision int32) error {
	w := packet.NewWriter(32)

	if err := w.WriteByte(clientpackets.OpcodeProtocolVersion); err != nil {
		return fmt.Errorf("writing opcode: %w", err)
	}

	w.WriteInt(revision)

	data := w.Bytes()

	// Prepare buffer with header space and padding (Blowfish requires 8-byte blocks)
	buf := make([]byte, constants.DefaultSendBufSize)
	copy(buf[constants.PacketHeaderSize:], data)

	if err := protocol.WritePacket(c.conn, c.encryption, buf, len(data)); err != nil {
		return fmt.Errorf("writing ProtocolVersion packet: %w", err)
	}

	return nil
}

// SendAuthLogin sends an AuthLogin packet (opcode 0x08).
func (c *GameClient) SendAuthLogin(accountName string, sessionKey login.SessionKey) error {
	w := packet.NewWriter(256)

	if err := w.WriteByte(clientpackets.OpcodeAuthLogin); err != nil {
		return fmt.Errorf("writing opcode: %w", err)
	}

	// Account name (UTF-16LE null-terminated)
	w.WriteString(accountName)

	// SessionKey (4×int32)
	w.WriteInt(sessionKey.PlayOkID1)
	w.WriteInt(sessionKey.PlayOkID2)
	w.WriteInt(sessionKey.LoginOkID1)
	w.WriteInt(sessionKey.LoginOkID2)

	// Unknown fields (4×int32, all zeros)
	for range 4 {
		w.WriteInt(0)
	}

	data := w.Bytes()

	// Prepare buffer with header space and padding
	buf := make([]byte, constants.DefaultSendBufSize)
	copy(buf[constants.PacketHeaderSize:], data)

	if err := protocol.WritePacket(c.conn, c.encryption, buf, len(data)); err != nil {
		return fmt.Errorf("writing AuthLogin packet: %w", err)
	}

	return nil
}

// ReadPacket reads and decrypts a packet from the server.
// Returns the decrypted payload (without header).
func (c *GameClient) ReadPacket() ([]byte, error) {
	buf := make([]byte, constants.DefaultReadBufSize)
	payload, err := protocol.ReadPacket(c.conn, c.encryption, buf)
	if err != nil {
		return nil, fmt.Errorf("reading packet: %w", err)
	}
	return payload, nil
}

// ReadPacketWithOpcode reads a packet and verifies the opcode.
func (c *GameClient) ReadPacketWithOpcode(expectedOpcode byte) ([]byte, error) {
	payload, err := c.ReadPacket()
	if err != nil {
		return nil, err
	}

	if len(payload) < 1 {
		return nil, fmt.Errorf("packet too short: %d bytes", len(payload))
	}

	if payload[0] != expectedOpcode {
		return nil, fmt.Errorf("unexpected opcode: expected 0x%02X, got 0x%02X", expectedOpcode, payload[0])
	}

	return payload[1:], nil // return body without opcode
}

// Close closes the connection.
func (c *GameClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// SessionID returns the session ID (for testing).
func (c *GameClient) SessionID() int32 {
	return c.sessionID
}

// Encryption returns the encryption context (for testing).
func (c *GameClient) Encryption() *crypto.LoginEncryption {
	return c.encryption
}
