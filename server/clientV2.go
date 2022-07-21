package server

import (
	"bufio"
	"context"
	"net"

	"Memcached/internal"
)

type clientV2 struct {
	ctx    context.Context
	cancel context.CancelFunc
	net.Conn

	lg internal.Logger

	r *bufio.Reader
	w *bufio.Writer
}

func NewClientV2() *clientV2 {
	return nil
}

func (c *clientV2) IoLoop() error {
	return nil
}

func (c *clientV2) WriterLoop() error {
	return nil
}
