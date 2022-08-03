package server

import (
	"Memcached/internal"
	"context"
	"sync"
	"time"
)

type Conner interface {
	Name() string
	Close() error
	IoLoop() error
}

type Config struct {
	TcpServerAddr string
	RemoteAddrArr []string

	SyncCheck time.Duration
}

type TcpConnects struct {
	ID     string
	ctx    context.Context
	cancel context.CancelFunc

	*Config
	ts      *Server
	ConnMap map[string]Conner
	sync.RWMutex
}

func CMStart(ctx context.Context, cf *Config) (*TcpConnects, error) {
	mc := &TcpConnects{
		ID:      cf.TcpServerAddr,
		Config:  cf,
		ConnMap: make(map[string]Conner),
	}
	mc.ctx, mc.cancel = context.WithCancel(ctx)
	ts := NewTcpServer(mc.ctx, cf.TcpServerAddr)
	ts.mc = mc
	mc.ts = ts
	go mc.ts.Start()
	go mc.connRemotes()

	return mc, nil
}

func (sr *TcpConnects) syncCheckConnes() {
	ticker := time.NewTicker(sr.SyncCheck)
	for {
		select {
		case <-sr.ctx.Done():
			ticker.Stop()
			goto exit
		case <-ticker.C:
			sr.connRemotes()
		}
	}
exit:
}

func (sr *TcpConnects) connRemotes() {
	for _, addr := range sr.RemoteAddrArr {
		if addr != sr.ts.TCPAddress && !sr.IsHas(addr) {
			c, err := ConnOtherServer(sr.ctx, addr, sr.TcpServerAddr, sr)
			if err != nil {
				internal.Lg.Errorf("[%s] remoter err:", addr, err)
			} else {
				sr.AddConn(c)
			}
		}
	}
}

func (sr *TcpConnects) IsHas(name string) bool {
	_, ok := sr.ConnMap[name]
	return ok
}

func (sr *TcpConnects) AddConn(c Conner) {
	sr.RLock()
	if sr.IsHas(c.Name()) {
		sr.RUnlock()
		internal.Lg.Infof("[%s] conner has exist", c.Name())
		c.Close()
		return
	}
	sr.RUnlock()
	sr.Lock()
	defer sr.Unlock()
	if sr.IsHas(c.Name()) {
		internal.Lg.Infof("[%s] conner has exist", c.Name())
		c.Close()
	} else {
		sr.ConnMap[c.Name()] = c
	}
}

func (sr *TcpConnects) AllCones() map[string]Conner {
	sr.RLock()
	defer sr.RUnlock()
	return sr.ConnMap
}

func (sr *TcpConnects) RemoveConn(c Conner) {
	sr.Lock()
	defer sr.Unlock()
	delete(sr.ConnMap, c.Name())
}
