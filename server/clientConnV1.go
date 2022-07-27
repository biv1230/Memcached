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

	r *bufio.Reader
	w *bufio.Writer
}

func NewClientV1(ctx context.Context, conn net.Conn, r *bufio.Reader, fromID string) *clientV1 {
	c := clientV1{
		Conn:   conn,
		r:      r,
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
	c.cancel()
	return c.Conn.Close()
}

func (c *clientV1) IoLoop() error {
	for {
		select {
		case <-c.ctx.Done():
			internal.Lg.Infof("%s close connect", c.RemoteAddr())
			goto exit
		default:
			com, err := internal.ReadCommand(c.r)
			if err != nil {
				internal.Lg.Errorf("[%s] %s", c.RemoteAddr(), err)
				goto exit
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
			internal.Lg.Infof("%s close connect", c.RemoteAddr())
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

func (c *clientV1) ExecCommand(com *internal.Command) error {
	switch string(com.Params[1]) {
	case internal.IdentifyString:

	}
	return nil
}
