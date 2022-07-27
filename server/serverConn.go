package server

import (
	"Memcached/internal"
	"bufio"
	"bytes"
	"net"
	"time"
)

type clientV3 struct {
	name       string
	remoteAddr string
	net.Conn

	r *bufio.Reader
	w *bufio.Writer
}

func NewClientV3(remoteAddr, CmID string) *clientV3 {
	return &clientV3{
		remoteAddr: remoteAddr,
	}
}

func ConnOtherServer(remoteAddr, CmID string) (*clientV3, error) {
	c := NewClientV3(remoteAddr, CmID)
	conn, err := net.DialTimeout("tcp", c.remoteAddr, 2*time.Second)
	if err != nil {
		return nil, err
	}
	c.r, c.w = bufio.NewReader(conn), bufio.NewWriter(conn)
	com := internal.Identify(internal.ClientV1, []byte(CmID))
	if _, err := com.WriteTo(c.w); err != nil {
		internal.Lg.Errorf("identify error:[%s]", err)
		conn.Close()
		return nil, err
	}
	rCom, err := internal.ReadCommand(c.r)
	if err != nil {
		internal.Lg.Errorf("wait return error:[%s]", err)
		conn.Close()
		return nil, err
	}
	if bytes.Equal(rCom.Name, internal.SucConnBytes) {
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

func (c *clientV3) IoLoop() error {
	return nil
}
