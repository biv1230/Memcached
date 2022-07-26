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

	SyncTimeout time.Duration
}

type ConnManager struct {
	Ctx    context.Context
	Cancel context.CancelFunc

	*Config
	ts      *TcpServer
	ConnMap map[string]Conner
	sync.RWMutex
}

func CMStart(ctx context.Context, cf *Config) (*ConnManager, error) {
	mc := &ConnManager{
		Config: cf,
	}
	mc.Ctx, mc.Cancel = context.WithCancel(ctx)
	mc.ts = NewTcpServer(mc.Ctx, cf.TcpServerAddr)

	mc.connRemotes()

	return mc, nil
}

func (sr *ConnManager) connRemotes() {
	for _, addr := range sr.RemoteAddrArr {
		if addr != sr.ts.TCPAddress && !sr.IsHas(addr) {
			c, err := ConnOtherServer(addr)
			if err != nil {
				internal.Lg.Errorf("[%s] remoter err:", addr, err)
			} else {
				sr.AddConn(c)
			}
		}
	}
}

func (sr *ConnManager) IsHas(name string) bool {
	_, ok := sr.ConnMap[name]
	return ok
}

func (sr *ConnManager) AddConn(c Conner) {
	sr.RLock()
	if sr.IsHas(c.Name()) {
		sr.RUnlock()
		internal.Lg.Infof("[%s] conner has exist", c.Name())
		c.Close()
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

func (sr *ConnManager) AllCones() map[string]Conner {
	sr.RLock()
	defer sr.RUnlock()
	return sr.ConnMap
}

func (sr *ConnManager) RemoveConn(c Conner) {
	sr.Lock()
	defer sr.Unlock()
	delete(sr.ConnMap, c.Name())
}
