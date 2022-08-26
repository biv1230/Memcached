package command

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

var (
	NewLine     byte = '\n'
	ByteSpace        = []byte(" ")
	ByteNewLine      = []byte{NewLine}

	FileInfoCommand     = "FILEINFOSTAT"
	FileInfoCommandByte = []byte(FileInfoCommand)
)

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

func (c *Command) String() string {
	params := ""
	for _, b := range params {
		params += string(b)
	}
	return fmt.Sprintf("Name=%s,Params=%s,Body=%s", c.Name, params, c.Body)
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
	body = make([]byte, l)
	if n, err := io.ReadFull(r, body); err != nil || n != int(l) {
		if err != io.EOF {
			return nil, errors.New(fmt.Sprintf("err:[%s],body len[%d],get[%d]", err, l, n))
		}
	}
	return &Command{
		Name:   params[0],
		Params: params[1:],
		Body:   body,
	}, nil
}
