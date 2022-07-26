package server

import "sync"

var (
	ConnMap map[string]Conner
	sr      sync.RWMutex
)

type Conner interface {
	Name() string
	Close() error
	IoLoop() error
}

func IsHas(c Conner) bool {
	_, ok := ConnMap[c.Name()]
	return ok
}

func AddConn(c Conner) error {
	sr.RLock()
	if IsHas(c) {
		sr.RUnlock()
		return c.Close()
	}
	sr.RUnlock()
	sr.Lock()
	defer sr.Unlock()
	if IsHas(c) {
		return c.Close()
	}
	ConnMap[c.Name()] = c
	return nil
}

func AllCones() map[string]Conner {
	sr.RLock()
	defer sr.RUnlock()
	return ConnMap
}

func RemoveConn(c Conner) {
	sr.Lock()
	defer sr.Unlock()
	delete(ConnMap, c.Name())
}
