package server

import (
	"Memcached/internal"
	"Memcached/warehouse"
	"bufio"
	"context"
	"fmt"
	"net"
)

type clientV1 struct {
	fromID string
	toID   string

	ctx    context.Context
	cancel context.CancelFunc
	net.Conn

	r *bufio.Reader
	w *bufio.Writer
}

func NewClientV1(ctx context.Context, conn net.Conn, r *bufio.Reader, w *bufio.Writer, fromID, toID string) *clientV1 {
	c := clientV1{
		Conn:   conn,
		r:      r,
		w:      w,
		fromID: fromID,
		toID:   toID,
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
			com, err := ReadCommand(c.r)
			if err != nil {
				internal.Lg.Errorf("[%s] %s", c.RemoteAddr(), err)
				goto exit
			}
			if err := ReceiveCommandExec(com); err != nil {
				internal.Lg.Errorf("[%s] error message: [%s]", c.RemoteAddr(), err)
			}
		}
	}
exit:
	c.Close()
	return fmt.Errorf("%s close connect", c.RemoteAddr())
}

func (c *clientV1) Send(m *warehouse.Message) error {
	body, err := m.ToByte()
	if err != nil {
		internal.Lg.Errorf("message to byte array err:[%s]", err)
		return err
	}
	if _, err := c.w.Write(body); err != nil {
		internal.Lg.Errorf("message to byte array err:[%s]", err)
		return err
	}
	c.w.Flush()
	return nil
}
