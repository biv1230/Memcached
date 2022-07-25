package server

import (
	"Memcached/internal"
	"bufio"
	"context"
	"errors"
	"fmt"
	"net"
)

type clientV1 struct {
	name   string
	ctx    context.Context
	cancel context.CancelFunc
	net.Conn

	lg internal.Logger

	r *bufio.Reader
	w *bufio.Writer
}

func NewClientV1(ctx context.Context, lg internal.Logger, conn net.Conn) *clientV1 {
	c := clientV1{
		Conn: conn,
		lg:   lg,
		r:    bufio.NewReader(conn),
		w:    bufio.NewWriter(conn),
	}
	c.ctx, c.cancel = context.WithCancel(ctx)
	return &c
}

func (c *clientV1) Name() string {
	return c.name
}

func (c *clientV1) IoLoop() error {
	var err error
	var line []byte
	for {
		select {
		case <-c.ctx.Done():
			err = errors.New(fmt.Sprintf("%s close connect", c.RemoteAddr()))
			goto exit
		default:
			line, err = c.r.ReadBytes('\n')
			if err != nil {
				c.lg.Infof("%s read err: %s", c.RemoteAddr(), err)
				goto exit
			}
			line = line[:len(line)-1]
			if len(line) > 0 && line[len(line)-1] == '\r' {
				line = line[:len(line)-1]
			}
			com, err := internal.DecodeCommand(c.r, line)
			if err != nil {
				c.lg.Errorf("[%s]command decode error %s", c.RemoteAddr(), err)
				continue
			}
			c.ExecCommand(com)
		}
	}
exit:
	c.Close()
	return err
}

func (c *clientV1) writerLoop(commChan <-chan *internal.Command) error {
	var err error
	for {
		select {
		case <-c.ctx.Done():
			c.lg.Infof("%s close connect", c.RemoteAddr())
			err = errors.New(fmt.Sprintf("%s close connect", c.RemoteAddr()))
			goto exit
		case comm := <-commChan:
			c.ExecCommand(comm)
		}
	}
exit:
	c.Close()
	return err
}

func (c *clientV1) ExecCommand(comm *internal.Command) error {
	return nil
}
