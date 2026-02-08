package gameserver

import "sync"

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
