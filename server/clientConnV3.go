package server

import (
	"Memcached/internal"
	"bufio"
	"bytes"
	"context"
	"net"
	"time"
)

type clientV3 struct {
	ctx    context.Context
	cancel context.CancelFunc

	name       string
	remoteAddr string
	net.Conn

	r *bufio.Reader
	w *bufio.Writer

	cache *warehouse.Caches
}

func NewClientV3(ctx context.Context, remoteAddr string) *clientV3 {
	c := &clientV3{
		remoteAddr: remoteAddr,
	}
	c.ctx, c.cancel = context.WithCancel(ctx)
	return c
}

func ConnOtherServer(ctx context.Context, remoteAddr, CmID string) (*clientV3, error) {
	c := NewClientV3(ctx, remoteAddr)
	conn, err := net.DialTimeout("tcp", c.remoteAddr, time.Second)
	if err != nil {
		return nil, err
	}
	c.r, c.w = bufio.NewReader(conn), bufio.NewWriter(conn)
	com := Identify(internal.ClientV1, []byte(CmID))
	if _, err := com.WriteTo(c.w); err != nil {
		internal.Lg.Errorf("identify error:[%s]", err)
		conn.Close()
		return nil, err
	}
	rCom, err := ReadCommand(c.r)
	if err != nil {
		internal.Lg.Errorf("wait return error:[%s]", err)
		conn.Close()
		return nil, err
	}
	if bytes.Equal(rCom.Name, SucConnBytes) {
		internal.Lg.Infof("[%s] connect suc !!!", conn.RemoteAddr())
		c.name = c.remoteAddr
		go c.IoLoop()
		return c, nil
	} else {
		internal.Lg.Errorf("wait return error:[%s]", err)
		conn.Close()
		return nil, err
	}
}

func (c *clientV3) Name() string {
	return c.name
}

func (c *clientV3) Close() error {
	return c.Close()
}

func (c *clientV3) ExecCommand(com *Command) error {

	return nil
}

func (c *clientV3) IoLoop() error {
	for {
		select {
		case <-c.ctx.Done():
			goto exit
		default:
			rCom, err := ReadCommand(c.r)
			if err != nil {
				internal.Lg.Errorf("read from [%s] err:[%s]", c.RemoteAddr(), err)
				goto exit
			}
			c.ExecCommand(rCom)
		}
	}
exit:
	internal.Lg.Infof("[%s]--->[%s] conn close", c.LocalAddr(), c.RemoteAddr())
	return c.Close()
}
