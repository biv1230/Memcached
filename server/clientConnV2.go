package server

import (
	"Memcached/warehouse"
	"bufio"
	"context"
	"net"
)

type clientV2 struct {
	name   string
	ctx    context.Context
	cancel context.CancelFunc
	net.Conn

	r *bufio.Reader
	w *bufio.Writer
}

func NewClientV2() *clientV2 {
	return nil
}

func (c *clientV2) Name() string {
	return ""
}

func (c *clientV2) Close() error {
	c.cancel()
	return c.Conn.Close()
}

func (c *clientV2) IoLoop() error {
	return nil
}

func (c *clientV2) Send(w *warehouse.Message) error {
	return nil
}
