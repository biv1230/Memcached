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

var Connects *connects

type connects struct {
	ctx    context.Context
	cancel context.CancelFunc

	tcpAddr        string
	remoteAddrList []string

	syncCheck time.Duration

	ConnMap map[string]*clientV1
	sync.RWMutex
}

func (c *connects) start() error {
	internal.Lg.Infof("TCP start on: [%s]", c.tcpAddr)
	listener, err := net.Listen("tcp", c.tcpAddr)
	if err != nil {
		internal.Lg.Errorf("listen (%s) failed - [%s]", c.tcpAddr, err)
		return err
	}
	for {
		select {
		case <-c.ctx.Done():
			internal.Lg.Infof("TCP: closing %s", listener.Addr())
			return listener.Close()
		default:
			if con, err := listener.Accept(); err != nil {
				internal.Lg.Errorf("client connect error - %s", err)
			} else {
				go func() {
					c.handler(con)
				}()
			}
		}
	}
}

func (c *connects) handler(conn net.Conn) {
	internal.Lg.Infof("TCP: new client-(%s)", conn.RemoteAddr())

	bf, wf := bufio.NewReader(conn), bufio.NewWriter(conn)

	com, err := ReadCommand(bf)
	if err != nil {
		internal.Lg.Errorf("failed to read protocol version - %s", err)
		conn.Close()
		return
	}
	if !bytes.Equal(com.Name, IdentifyBytes) {
		internal.Lg.Errorf("client(%s) bad identify", conn.RemoteAddr())
		conn.Close()
		return
	}
	if _, err := SucCommand.WriteTo(wf); err != nil {
		internal.Lg.Errorf("failed to write protocol version - %s", err)
		conn.Close()
		return
	}

	p := NewClientV1(c.ctx, conn, bf, wf, string(com.Params[0]), c.tcpAddr)
	c.addConn(p)

	err = p.IoLoop()
	if err != nil {
		internal.Lg.Errorf("client(%s) - %s", conn.RemoteAddr(), err)
	}
	c.removeConn(p)
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
		ConnMap:        make(map[string]*clientV1),
	}
	Connects.ctx, Connects.cancel = context.WithCancel(ctx)
	go func() {
		if err := Connects.start(); err != nil {
			Connects.Close()
		}
	}()
	go Connects.syncCheckConnes()
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
	internal.Lg.Info("sync check connes stop")
}

func (c *connects) connRemotes() {
	conns := c.allCones()
	for _, addr := range c.remoteAddrList {
		if addr != c.tcpAddr {
			if con, ok := conns[addr]; ok {
				if _, err := PingCommand.WriteTo(con.w); err == nil {
					continue
				} else {
					internal.Lg.Errorf("[%s] remoter write err:[%s]", addr, err)
					c.removeConn(con)
				}
			}
			nCon, err := c.connRemoter(c.ctx, addr)
			if err != nil {
				internal.Lg.Errorf("[%s] remoter err:[%s]", addr, err)
			} else {
				c.addConn(nCon)
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
	com := Identify([]byte(c.tcpAddr))
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
		v1 := NewClientV1(ctx, conn, r, w, remoteAddr, c.tcpAddr)
		go v1.IoLoop()
		return v1, nil
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

func (c *connects) addConn(p *clientV1) {
	c.RLock()
	if c.isHas(p.Name()) {
		c.RUnlock()
		internal.Lg.Infof("[%s] conner has exist", p.Name())
		p.Close()
		return
	}
	c.RUnlock()
	c.Lock()
	defer c.Unlock()
	if c.isHas(p.Name()) {
		internal.Lg.Infof("[%s] conner has exist", p.Name())
		p.Close()
	} else {
		c.ConnMap[p.Name()] = p
	}
}

func (c *connects) allCones() map[string]*clientV1 {
	c.RLock()
	defer c.RUnlock()
	return c.ConnMap
}

func (c *connects) removeConn(p *clientV1) {
	c.Lock()
	defer c.Unlock()
	delete(c.ConnMap, p.Name())
}

// send message to all service
func (c *connects) Notice(m *warehouse.Message) {
	for _, co := range c.allCones() {
		con := co
		go func() {
			if err := con.Send(m); err != nil {
				internal.Lg.Errorf("message send err:[%s]", err)
			}
		}()
	}
}
