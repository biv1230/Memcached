package server

import (
	"bufio"
	"bytes"
	"context"
	"net"
	"sync"
	"time"

	"Memcached/internal"
	"Memcached/warehouse"
)

type Conner interface {
	Name() string
	Close() error
	IoLoop() error
	Send(m *warehouse.Message) error
}

var Connects *connects

type connects struct {
	ctx    context.Context
	cancel context.CancelFunc

	tcpAddr        string
	remoteAddrList []string

	syncCheck time.Duration

	ConnMap map[string]Conner
	sync.RWMutex
}

func (c *connects) start() error {
	internal.Lg.Infof("tcp listening: [%s]", c.tcpAddr)
	listener, err := net.Listen("tcp", c.tcpAddr)
	if err != nil {
		internal.Lg.Errorf("listen (%s) failed - %s", c.tcpAddr, err)
		return err
	}
	for {
		select {
		case <-c.ctx.Done():
			internal.Lg.Infof("TCP: closing %s", listener.Addr())
			return listener.Close()
		default:
			con, err := listener.Accept()
			if err != nil {
				internal.Lg.Errorf("client connect error - %s", err)
			}
			go func() {
				c.handler(con)
			}()
		}
	}
}

func (c *connects) handler(conn net.Conn) {
	internal.Lg.Infof("TCP: new client(%s)", conn.RemoteAddr().String())

	bf, wf := bufio.NewReader(conn), bufio.NewWriter(conn)

	com, err := ReadCommand(bf)
	if err != nil {
		internal.Lg.Errorf("failed to read protocol version - %s", err)
		//internal.FaiConner.WriteTo(conn)
		conn.Close()
		return
	}
	if !bytes.Equal(com.Name, IdentifyBytes) {
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
		if _, err := SucConner.WriteTo(wf); err != nil {
			internal.Lg.Errorf("failed to read protocol version - %s", err)
			conn.Close()
			return
		}
		p = NewClientV1(c.ctx, conn, bf, wf, string(com.Params[1]), c.tcpAddr)
		c.addConn(p)

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
		c.removeConn(p)
	case string(internal.ClientV2):

	}
	p.Close()
}

func (c *connects) Close() error {
	c.cancel()
	return nil
}

func CMStart(ctx context.Context, tcpAddr string, remoteAddrList []string, syncCheck time.Duration) {
	Connects = &connects{
		tcpAddr:        tcpAddr,
		remoteAddrList: remoteAddrList,
		syncCheck:      syncCheck,
		ConnMap:        make(map[string]Conner),
	}
	Connects.ctx, Connects.cancel = context.WithCancel(ctx)
	go Connects.start()
	go Connects.connRemotes()
}

func (c *connects) syncCheckConnes() {
	ticker := time.NewTicker(c.syncCheck)
	for {
		select {
		case <-c.ctx.Done():
			ticker.Stop()
			goto exit
		case <-ticker.C:
			c.connRemotes()
		}
	}
exit:
}

func (c *connects) connRemotes() {
	for _, addr := range c.remoteAddrList {
		if addr != c.tcpAddr && !c.isHas(addr) {
			con, err := c.connRemoter(c.ctx, addr)
			if err != nil {
				internal.Lg.Errorf("[%s] remoter err:[%s]", addr, err)
			} else {
				c.addConn(con)
			}
		}
	}
}

func (c *connects) connRemoter(ctx context.Context, remoteAddr string) (*clientV1, error) {
	conn, err := net.DialTimeout("tcp", remoteAddr, time.Second)
	if err != nil {
		return nil, err
	}
	r, w := bufio.NewReader(conn), bufio.NewWriter(conn)
	com := Identify(internal.ClientV1, []byte(c.tcpAddr))
	if _, err := com.WriteTo(w); err != nil {
		internal.Lg.Errorf("identify error:[%s]", err)
		conn.Close()
		return nil, err
	}
	rCom, err := ReadCommand(r)
	if err != nil {
		internal.Lg.Errorf("wait return error:[%s]", err)
		conn.Close()
		return nil, err
	}
	if bytes.Equal(rCom.Name, SucConnBytes) {
		internal.Lg.Infof("[%s] connect suc !!!", conn.RemoteAddr())
		c := NewClientV1(ctx, conn, r, w, remoteAddr, c.tcpAddr)
		go c.IoLoop()
		return c, nil
	} else {
		internal.Lg.Errorf("wait return error:[%s]", err)
		conn.Close()
		return nil, err
	}
}

func (c *connects) isHas(name string) bool {
	_, ok := c.ConnMap[name]
	return ok
}

func (c *connects) addConn(con Conner) {
	c.RLock()
	if c.isHas(con.Name()) {
		c.RUnlock()
		internal.Lg.Infof("[%s] conner has exist", con.Name())
		c.Close()
		return
	}
	c.RUnlock()
	c.Lock()
	defer c.Unlock()
	if c.isHas(con.Name()) {
		internal.Lg.Infof("[%s] conner has exist", con.Name())
		c.Close()
	} else {
		c.ConnMap[con.Name()] = con
	}
}

func (c *connects) allCones() map[string]Conner {
	c.RLock()
	defer c.RUnlock()
	return c.ConnMap
}

func (c *connects) removeConn(con Conner) {
	c.Lock()
	defer c.Unlock()
	delete(c.ConnMap, con.Name())
}

func (c *connects) Notice(m *warehouse.Message) {
	c.RLock()
	defer c.RUnlock()
	for _, co := range c.ConnMap {
		con := co
		go func() {
			if err := con.Send(m); err != nil {
				internal.Lg.Errorf("message send err:[%s]", err)
			}
		}()
	}
}
