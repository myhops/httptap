package bufpool

import (
	"bytes"
	"log/slog"
	"sync"
)

const bufSize = 4 * 1024

type BufferPool struct {
	pool sync.Pool
	size int
}

var Default = New()

func New() *BufferPool {
	b := &BufferPool{
		size: bufSize,
	}
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
	if c:= b.Cap(); c > p.size {
		slog.Default().Debug("discard large Buffer", slog.String("packager", "BufferPool"),  slog.Int("cap",c))
		return
	}
	p.pool.Put(b)
}

func Get() *bytes.Buffer {
	return Default.Get()
}

func Put(b *bytes.Buffer) {
	Default.Put(b)
}
