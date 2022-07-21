package server

import (
	"context"
	"io"
	"net"
	"sync"

	"Memcached/internal"
)

type protocol interface {
	IoLoop() error
}

type TcpServer struct {
	Ctx    context.Context
	Cancel context.CancelFunc

	lg internal.Logger

	TCPAddress string

	clientV1Connects sync.Map
	clientV2Connects sync.Map
}

func (ts *TcpServer) Handler(conn net.Conn) {
	ts.lg.Infof("TCP: new client(%s)", conn.RemoteAddr())

	buf := make([]byte, 4)
	_, err := io.ReadFull(conn, buf)
	if err != nil {
		ts.lg.Errorf("failed to read protocol version - %s", err)
		conn.Close()
		return
	}
	protocolMagic := string(buf)
	ts.lg.Infof("client(%s): protocol magic [%s]", conn.RemoteAddr(), protocolMagic)
	var p protocol

	switch protocolMagic {
	case string(internal.ClientV1):
		p = NewClientV1(ts.Ctx, ts.lg, conn)
		ts.clientV1Connects.Store(conn.RemoteAddr(), p)
	case string(internal.ClientV2):
		p = NewClientV2()
		ts.clientV2Connects.Store(conn.RemoteAddr(), p)
	default:
		conn.Close()
		ts.lg.Errorf("client(%s) bad protocol magic %s", conn.RemoteAddr(), protocolMagic)
		return
	}

	err = p.IoLoop()
	if err != nil {
		ts.lg.Errorf("client(%s) - %s", conn.RemoteAddr(), err)
	}
	switch protocolMagic {
	case string(internal.ClientV1):
		ts.clientV1Connects.Delete(conn.RemoteAddr())
	case string(internal.ClientV2):
		ts.clientV2Connects.Delete(conn.RemoteAddr())
	}
	conn.Close()
}

func (ts *TcpServer) Start() error {
	listener, err := net.Listen("tcp", ts.TCPAddress)
	if err != nil {
		ts.lg.Errorf("listen (%s) failed - %s", ts.TCPAddress, err)
		return err
	}
	for {
		select {
		case <-ts.Ctx.Done():
			ts.lg.Infof("TCP: closing %s", listener.Addr())
			return listener.Close()
		default:
			c, err := listener.Accept()
			if err != nil {
				ts.lg.Errorf("client connect error - %s", err)
			}
			go func() {
				ts.Handler(c)
			}()
		}
	}
}
