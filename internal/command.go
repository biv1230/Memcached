package internal

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

var (
	CacheBeforeBytes []byte = []byte("BEFORE")
	CachingBytes     []byte = []byte("CACHING")
	CacheAfterBytes  []byte = []byte("AFTER")
)

var (
	ByteSpace   = []byte(" ")
	ByteNewLine = []byte("\n")
)

type Command struct {
	Name   []byte
	Params [][]byte
	Body   []byte
}

func NewCommand(name []byte, params [][]byte, body []byte) *Command {
	return &Command{
		name,
		params,
		body,
	}
}

func (c *Command) WriteTo(w io.Writer) (int, error) {
	var total int
	var buf [4]byte

	n, err := w.Write(c.Name)
	total += n
	if err != nil {
		return total, err
	}

	for _, param := range c.Params {
		n, err = w.Write(ByteSpace)
		total += n
		if err != nil {
			return total, err
		}
		n, err = w.Write(param)
		total += n
		if err != nil {
			return total, err
		}
	}

	n, err = w.Write(ByteNewLine)
	total += n
	if err != nil {
		return total, err
	}
	if c.Body != nil {
		bufs := buf[:]
		binary.BigEndian.PutUint32(bufs, uint32(len(c.Body)))
		n, err = w.Write(bufs)
		total += n
		if err != nil {
			return total, err
		}
		n, err = w.Write(c.Body)
		total += n
		if err != nil {
			return total, err
		}
	}
	return total, nil
}

func DecodeCommand(r io.Reader, p []byte) (*Command, error) {
	params := bytes.Split(p, ByteSpace)
	var buf [4]byte
	var body []byte
	bufs := buf[:]
	_, err := io.ReadFull(r, bufs)
	if err != nil {
		return nil, err
	}
	l := binary.BigEndian.Uint32(bufs)
	if l > 0 {
		body = make([]byte, l)
		if n, err := io.ReadFull(r, body); err != nil || n != int(l) {
			return nil, errors.New(fmt.Sprintf("err:[%s],body len[%d],get[%d]", err, l, n))
		}
	} else {
		body = nil
	}
	return &Command{
		Name:   params[0],
		Params: params[1:],
		Body:   body,
	}, nil
}

func CacheBefore(key []byte) *Command {
	params := [][]byte{key}
	return &Command{CacheBeforeBytes, params, nil}
}
