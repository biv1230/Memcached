package internal

import (
	"bytes"
	"sync"
)

var (
	bufferPool sync.Pool
)

func init() {
	bufferPool.New = func() any {
		return &bytes.Buffer{}
	}
}

func BufferPoolGet() *bytes.Buffer {
	return bufferPool.Get().(*bytes.Buffer)
}

func BufferPoolSet(b *bytes.Buffer) {
	b.Reset()
	bufferPool.Put(b)
}
