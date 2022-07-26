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
	fromID string
	ctx    context.Context
	cancel context.CancelFunc
	net.Conn

	lg internal.Logger

	r *bufio.Reader
	w *bufio.Writer
}

func NewClientV1(ctx context.Context, lg internal.Logger, conn net.Conn, fromID string) *clientV1 {
	c := clientV1{
		Conn:   conn,
		lg:     lg,
		r:      bufio.NewReader(conn),
		w:      bufio.NewWriter(conn),
		fromID: fromID,
	}
	c.ctx, c.cancel = context.WithCancel(ctx)
	return &c
}

func (c *clientV1) Name() string {
	return c.fromID
}

func (c *clientV1) Close() error {
	return c.Conn.Close()
}

func (c *clientV1) IoLoop() error {
	for {
		select {
		case <-c.ctx.Done():
			c.lg.Infof("%s close connect", c.RemoteAddr())
			goto exit
		default:
			com, err := internal.ReadCommand(c.r)
			if err != nil {
				c.lg.Errorf("[%s] %s", c.RemoteAddr(), err)
				continue
			}
			c.ExecCommand(com)
		}
	}
exit:
	c.Close()
	return fmt.Errorf("%s close connect", c.RemoteAddr())
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
