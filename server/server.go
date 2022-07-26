package server

import (
	"Memcached/internal"
	"bufio"
	"context"
	"net"
)

type TcpServer struct {
	Ctx    context.Context
	Cancel context.CancelFunc

	lg internal.Logger

	TCPAddress string
}

func NewTcpServer(ctx context.Context, addr string, logger internal.Logger) *TcpServer {
	ts := TcpServer{
		lg:         logger,
		TCPAddress: addr,
	}
	ts.Ctx, ts.Cancel = context.WithCancel(ctx)
	internal.Lg = logger
	return &ts
}

func (ts *TcpServer) handler(conn net.Conn) {
	ts.lg.Infof("TCP: new client(%s)", conn.RemoteAddr().String())

	bf := bufio.NewReader(conn)

	com, err := internal.ReadCommand(bf)
	if err != nil {
		ts.lg.Errorf("failed to read protocol version - %s", err)
		conn.Close()
		return
	}
	protocolMagic := string(com.Params[0])
	ts.lg.Infof("client(%s): protocol magic [%s]", conn.RemoteAddr(), protocolMagic)
	var p Conner

	switch protocolMagic {
	case string(internal.ClientV1):
		p = NewClientV1(ts.Ctx, ts.lg, conn, string(com.Params[1]))
		AddConn(p)

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
