package gameserver

import (
	"fmt"
	"sync"

	"github.com/udisondev/la2go/internal/constants"
	"github.com/udisondev/la2go/internal/crypto"
	"github.com/udisondev/la2go/internal/protocol"
)

// BytePool is a pool of reusable []byte buffers.
// Reduces GC pressure by reusing allocations.
type BytePool struct {
	pool sync.Pool
}

// NewBytePool creates a buffer pool with the specified default capacity for new slices.
func NewBytePool(defaultCap int) *BytePool {
	p := &BytePool{}
	p.pool.New = func() any {
		return make([]byte, 0, defaultCap)
	}
	return p
}

// Get returns a slice of length size, preferably from the pool.
func (p *BytePool) Get(size int) []byte {
	b := p.pool.Get().([]byte)
	if cap(b) < size {
		p.pool.Put(b)
		return make([]byte, size)
	}
	b = b[:size]
	clear(b)
	return b
}

// Put returns the slice to the pool for reuse.
func (p *BytePool) Put(b []byte) {
	if b == nil {
		return
	}
	p.pool.Put(b[:0])
}

// EncryptToPooled encrypts payload into a fresh buffer from pool.
// Returns ready-to-send encrypted packet (pool-backed).
// OWNERSHIP: caller owns returned slice; MUST return to pool via pool.Put() after use.
// Thread-safe: does not modify input payload.
func (p *BytePool) EncryptToPooled(enc *crypto.LoginEncryption, payload []byte, payloadLen int) ([]byte, error) {
	// First-packet overhead: +8 (static) + up to 7 (pad to 8) + 8 (final) = +23
	// Subsequent-packet overhead: +4 (checksum) + up to 7 (pad to 8) = +11
	// Use worst case (24) for safety. PacketBufferPadding=16 is only enough for subsequent packets.
	const encryptionOverhead = 24
	needed := constants.PacketHeaderSize + payloadLen + encryptionOverhead
	buf := p.Get(needed)
	copy(buf[constants.PacketHeaderSize:], payload[:payloadLen])
	encSize, err := protocol.EncryptInPlace(enc, buf, payloadLen)
	if err != nil {
		p.Put(buf)
		return nil, fmt.Errorf("encrypting to pooled buffer: %w", err)
	}
	return buf[:encSize], nil
}
