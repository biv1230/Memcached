package server

import (
	"Memcached/internal"
	"bufio"
	"bytes"
	"context"
	"net"
)

type TcpServer struct {
	Ctx    context.Context
	Cancel context.CancelFunc

	mc         *ConnManager
	TCPAddress string
}

func NewTcpServer(ctx context.Context, addr string) *TcpServer {
	ts := TcpServer{
		TCPAddress: addr,
	}
	ts.Ctx, ts.Cancel = context.WithCancel(ctx)
	return &ts
}

func (ts *TcpServer) handler(conn net.Conn) {
	internal.Lg.Infof("TCP: new client(%s)", conn.RemoteAddr().String())

	bf, wf := bufio.NewReader(conn), bufio.NewWriter(conn)

	com, err := internal.ReadCommand(bf)
	if err != nil {
		internal.Lg.Errorf("failed to read protocol version - %s", err)
		//internal.FaiConner.WriteTo(conn)
		conn.Close()
		return
	}
	if !bytes.Equal(com.Name, internal.IdentifyBytes) {
		internal.Lg.Errorf("failed to read protocol version - %s", err)
		//internal.FaiConner.WriteTo(conn)
		conn.Close()
		return
	}

	protocolMagic := string(com.Params[0])
	internal.Lg.Infof("client(%s): protocol magic [%s]", conn.RemoteAddr(), protocolMagic)
	var p Conner

	switch protocolMagic {
	case internal.ClientV1Str:
		if _, err := internal.SucConner.WriteTo(wf); err != nil {
			internal.Lg.Errorf("failed to read protocol version - %s", err)
			conn.Close()
			return
		}
		p = NewClientV1(ts.Ctx, conn, bf, string(com.Params[1]))
		ts.mc.AddConn(p)

	case internal.ClientV2Str:
		p = NewClientV2()
	default:
		conn.Close()
		internal.Lg.Errorf("client(%s) bad protocol magic %s", conn.RemoteAddr(), protocolMagic)
		return
	}

	err = p.IoLoop()
	if err != nil {
		internal.Lg.Errorf("client(%s) - %s", conn.RemoteAddr(), err)
	}
	switch protocolMagic {
	case string(internal.ClientV1):
		ts.mc.RemoveConn(p)
	case string(internal.ClientV2):

	}
	p.Close()
}

func (ts *TcpServer) Start() error {
	internal.Lg.Infof("tcp listening: [%s]", ts.TCPAddress)
	listener, err := net.Listen("tcp", ts.TCPAddress)
	if err != nil {
		internal.Lg.Errorf("listen (%s) failed - %s", ts.TCPAddress, err)
		return err
	}
	for {
		select {
		case <-ts.Ctx.Done():
			internal.Lg.Infof("TCP: closing %s", listener.Addr())
			return listener.Close()
		default:
			c, err := listener.Accept()
			if err != nil {
				internal.Lg.Errorf("client connect error - %s", err)
			}
			go func() {
				ts.handler(c)
			}()
		}
	}
}
