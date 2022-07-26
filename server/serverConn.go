package server

import (
	"Memcached/internal"
	"bufio"
	"bytes"
	"net"
	"time"
)

type clientV3 struct {
	toID       string
	remoteAddr string
	net.Conn

	r *bufio.Reader
	w *bufio.Writer
}

func NewClientV3(remoteAddr string) *clientV3 {
	return &clientV3{
		toID:       remoteAddr,
		remoteAddr: remoteAddr,
	}
}

func ConnOtherServer(remoteAddr string) (*clientV3, error) {
	c := NewClientV3(remoteAddr)
	conn, err := net.DialTimeout("tcp", c.remoteAddr, time.Second)
	if err != nil {
		return nil, err
	}
	c.r = bufio.NewReader(conn)
	c.w = bufio.NewWriter(conn)
	com := internal.Identify(internal.ClientV1, []byte(c.remoteAddr))
	if _, err := com.WriteTo(c.w); err != nil {
		internal.Lg.Errorf("identify error:[%s]", err)
		conn.Close()
		return nil, err
	}
	rcom, err := internal.ReadCommand(c.r)
	if err != nil {
		internal.Lg.Errorf("wait return error:[%s]", err)
		conn.Close()
		return nil, err
	}
	if bytes.Equal(rcom.Name, internal.SucConnBytes) {
		internal.Lg.Infof("[%s] connect suc !!!", conn.RemoteAddr())
		go c.IoLoop()
		return c, nil
	} else {
		internal.Lg.Errorf("wait return error:[%s]", err)
		conn.Close()
		return nil, err
	}
}

func (c *clientV3) Name() string {
	return c.toID
}

func (c *clientV3) Close() error {
	return c.Close()
}

func (c *clientV3) IoLoop() error {
	return nil
}
