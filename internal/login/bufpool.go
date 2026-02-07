package login

import "sync"

// BytePool — пул переиспользуемых []byte буферов.
// Снижает давление на GC за счёт повторного использования аллокаций.
type BytePool struct {
	pool sync.Pool
}

// NewBytePool создаёт пул с указанной начальной ёмкостью для новых слайсов.
func NewBytePool(defaultCap int) *BytePool {
	p := &BytePool{}
	p.pool.New = func() any {
		return make([]byte, 0, defaultCap)
	}
	return p
}

// Get возвращает слайс длиной size, по возможности из пула.
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

// Put возвращает слайс в пул для повторного использования.
func (p *BytePool) Put(b []byte) {
	if b == nil {
		return
	}
	p.pool.Put(b[:0])
}
