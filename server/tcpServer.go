package server

import (
	"Memcached/internal"
	"bufio"
	"bytes"
	"context"
	"net"
)

type Server struct {
	ctx    context.Context
	cancel context.CancelFunc

	mc         *TcpConnects
	TCPAddress string
}

func NewTcpServer(ctx context.Context, addr string) *Server {
	ts := Server{
		TCPAddress: addr,
	}
	ts.ctx, ts.cancel = context.WithCancel(ctx)
	return &ts
}

func (ts *Server) handler(conn net.Conn) {
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
		p = NewClientV1(ts.ctx, conn, bf, string(com.Params[1]))
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

func (ts *Server) Close() error {
	ts.cancel()
	return nil
}

func (ts *Server) Start() error {
	internal.Lg.Infof("tcp listening: [%s]", ts.TCPAddress)
	listener, err := net.Listen("tcp", ts.TCPAddress)
	if err != nil {
		internal.Lg.Errorf("listen (%s) failed - %s", ts.TCPAddress, err)
		return err
	}
	for {
		select {
		case <-ts.ctx.Done():
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
