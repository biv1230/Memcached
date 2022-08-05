package internal

import (
	"bufio"
	"bytes"
	"io"
	"sync"
)

var (
	bufferPool  sync.Pool
	bufioReader sync.Pool
	bufioWriter sync.Pool
)

func init() {
	bufferPool.New = func() any {
		return &bytes.Buffer{}
	}
	bufioReader.New = func() any {
		return bufio.NewReader(nil)
	}
	bufioWriter.New = func() any {
		return bufio.NewWriter(nil)
	}

}

func BufferPoolGet() *bytes.Buffer {
	return bufferPool.Get().(*bytes.Buffer)
}

func BufferPoolSet(b *bytes.Buffer) {
	b.Reset()
	bufferPool.Put(b)
}

func BufioReaderGet(r io.Reader) *bufio.Reader {
	b := bufioReader.Get().(*bufio.Reader)
	b.Reset(r)
	return b
}

func BufioReaderPut(r *bufio.Reader) {
	r.Reset(nil)
	bufioReader.Put(r)
}

func BufioWriterGet(w io.Writer) *bufio.Writer {
	ret := bufioWriter.Get().(*bufio.Writer)
	ret.Reset(w)
	return ret
}

func BufioWriterPut(w *bufio.Writer) {
	w.Reset(nil)
	bufioWriter.Put(w)
}
