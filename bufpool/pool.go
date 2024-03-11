package bufpool

import (
	"bytes"
	"sync"
)

type BufferPool struct {
	pool sync.Pool
}

var Default = New()

func New() *BufferPool {
	b := &BufferPool{}
	b.pool.New = func() any {
		return &bytes.Buffer{}
	}
	return b
}

func (p *BufferPool) Get() *bytes.Buffer {
	b, ok := p.pool.Get().(*bytes.Buffer)
	if !ok {
		panic("BufferPool contains element of bad type")
	}
	b.Reset()
	return b
}

func (p *BufferPool) Put(b *bytes.Buffer) {
	if b == nil {
		return
	}
	p.pool.Put(b)
}

func Get()  *bytes.Buffer {
	return Default.Get()
}

func Put(b *bytes.Buffer) {
	Default.Put(b)
}
