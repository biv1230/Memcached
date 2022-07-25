package server

import (
	"Memcached/internal"
	"context"
	"io"
	"net"
)

type TcpID []byte

type protocol interface {
	IoLoop() error
}

type TcpServer struct {
	ID     TcpID
	Ctx    context.Context
	Cancel context.CancelFunc

	lg internal.Logger

	TCPAddress string
}

func NewTcpServer(ctx context.Context, addr string, logger internal.Logger) *TcpServer {
	ts := TcpServer{
		ID:         []byte(addr),
		lg:         logger,
		TCPAddress: addr,
	}
	ts.Ctx, ts.Cancel = context.WithCancel(ctx)
	return &ts
}

func (ts *TcpServer) handler(conn net.Conn) {
	ts.lg.Infof("TCP: new client(%s)", conn.RemoteAddr().String())
	ts.lg.Infof("TCP: new client(%s)", conn.LocalAddr())

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
	case string(internal.ClientV2):
		p = NewClientV2()
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

	case string(internal.ClientV2):

	}
	conn.Close()
}

func (ts *TcpServer) Start() error {
	ts.lg.Infof("tcp listening: [%s]", ts.TCPAddress)
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
				ts.handler(c)
			}()
		}
	}
}
