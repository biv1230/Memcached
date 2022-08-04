package server

import (
	"Memcached/internal"
	"Memcached/warehouse"
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

const (
	IdentifyString    = "IDENTIFY"
	SucConnString     = "CONNECTSUC"
	FaiConnString     = "CONNECTFAI"
	CacheBeforeString = "BEFORE"
	CachingString     = "CACHING"
	CacheAfterString  = "AFTER"
)

var (
	NewLine     byte = '\n'
	ByteSpace        = []byte(" ")
	ByteNewLine      = []byte{NewLine}

	IdentifyBytes    = []byte(IdentifyString)
	SucConnBytes     = []byte(SucConnString)
	FaiConnBytes     = []byte(FaiConnString)
	CacheBeforeBytes = []byte(CacheBeforeString)
	CachingBytes     = []byte(CachingString)
	CacheAfterBytes  = []byte(CacheAfterString)

	SucConner *Command
	FaiConner *Command
)

func init() {
	SucConner, FaiConner = sucConn(), failConn()
}

type Command struct {
	Name   []byte
	Params [][]byte
	Body   []byte
}

func NewCommand(name []byte, params [][]byte, body []byte) *Command {
	if body == nil {
		body = []byte{}
	}
	return &Command{
		name,
		params,
		body,
	}
}

func (c *Command) WriteTo(w *bufio.Writer) (int, error) {
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
	return total, w.Flush()
}

func decodeBody(r *bufio.Reader, p []byte) (*Command, error) {
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
	return NewCommand(CacheBeforeBytes, params, nil)
}

func CacheAdd(key []byte, msg *warehouse.Message) (*Command, error) {
	params := [][]byte{key}
	body, err := msg.ToByte()
	if err != nil {
		internal.Lg.Errorf("message error %s", err)
		return nil, err
	}
	return NewCommand(CachingBytes, params, body), nil
}

func Identify(ver []byte, id []byte) *Command {
	params := [][]byte{ver, id}
	return NewCommand(IdentifyBytes, params, nil)
}

func sucConn() *Command {
	return NewCommand(SucConnBytes, nil, nil)
}

func failConn() *Command {
	return NewCommand(FaiConnBytes, nil, nil)
}

func ReadCommand(r *bufio.Reader) (*Command, error) {
	line, err := r.ReadBytes(NewLine)
	if err != nil {
		return nil, err
	}
	line = line[:len(line)-1]
	if len(line) > 0 && line[len(line)-1] == '\r' {
		line = line[:len(line)-1]
	}
	return decodeBody(r, line)
}
