package warehouse

import (
	"Memcached/internal"
	"context"
	"encoding/binary"
	"sync"
	"time"
)

const (
	InitStatus uint8 = iota
	NormalStatus
	CloseStatus
)

var Cache *caches

func Start(ctxParent context.Context, cap uint8) {
	if cap <= 0 {
		cap = 10
	}
	Cache = newCaches(ctxParent, cap)
}

type caches struct {
	ctx    context.Context
	cancel context.CancelFunc
	stores []*store
	status uint8

	cachingKeys map[string]*cachingProcess
}

func newCaches(ctxParent context.Context, cap uint8) *caches {
	ctx, cancel := context.WithCancel(ctxParent)
	stores := make([]*store, 0, cap)
	var i uint8
	for i = 0; i < cap; i++ {
		s := newStore(ctx, i)
		stores = append(stores, s)
	}
	return &caches{
		ctx:         ctx,
		cancel:      cancel,
		stores:      stores,
		cachingKeys: make(map[string]*cachingProcess),
	}
}

func (m *caches) determineStore(key []byte) uint8 {
	l := key[len(key)-1]
	return uint8(l) % uint8(len(m.stores))
}

func (m *caches) BeforeAdd(key []byte) {
	newCachingProcess(string(key))
}

func (m *caches) Add(msg *Message) {
	index := m.determineStore(msg.Key)
	m.stores[index].save(msg)
	closeCachingProcess(string(msg.Key))
}

func (m *caches) AddBody(body []byte) error {
	msg, err := DecodeMessage(body)
	if err != nil {
		internal.Lg.Errorf("body decode err:[%s]", err)
		return err
	}
	m.Add(msg)
	return nil
}

func (m *caches) Get(key []byte) *Message {
	index := m.determineStore(key)
	keyStr := string(key)
	msg := m.stores[index].get(keyStr)
	if msg == nil {
		if cp, ok := m.cachingKeys[keyStr]; ok {
			t := time.NewTimer(500 * time.Millisecond)
			select {
			case <-cp.process:
			case <-t.C:
			}
			return m.stores[index].get(keyStr)
		}
		return nil
	}
	return msg
}

func (m *caches) Delete(key []byte) {
	index := m.determineStore(key)
	m.stores[index].delete(string(key))
}

func (m *caches) Len() int {
	l := 0
	for _, v := range m.stores {
		l += v.len()
	}
	return l
}

func (m *caches) Cap() ([]byte, error) {
	b := internal.BufferPoolGet()
	defer internal.BufferPoolSet(b)
	err := binary.Write(b, binary.BigEndian, uint8(len(m.stores)))
	if err != nil {
		internal.Lg.Error(err)
		return nil, err
	}
	return b.Bytes(), nil
}

func (m *caches) Range(fn func(msg *Message) error) error {
	for _, v := range m.stores {
		items := v.getAllKQ()
		for _, item := range items {
			msg := v.get(item.key)
			if err := fn(msg); err != nil {
				internal.Lg.Error(err)
			}
		}
	}
	return nil
}

func (m *caches) SetStatus(status uint8) {
	m.status = status
}

type store struct {
	ctx    context.Context
	cancel context.CancelFunc
	index  uint8
	sync.RWMutex
	maps       map[string]*Message
	kq         *keyQueue
	syncDelete time.Duration
}

func newStore(ctx context.Context, index uint8) *store {
	ctx, cancel := context.WithCancel(ctx)
	return &store{
		ctx:    ctx,
		cancel: cancel,
		index:  index,
		maps:   make(map[string]*Message),
		kq:     newKewQueue(),
	}
}

func (s *store) save(m *Message) {
	key := string(m.Key)
	om := s.get(key)
	isNewKeyItem := false

	expiredTime := time.Now().Add(time.Duration(m.HoldTime) * time.Second).UnixNano()

	if om != nil {
		m.kt, om.kt = om.kt, nil
		m.kt.pastTime = expiredTime
	} else {
		isNewKeyItem = true
		m.kt = newKeyItem(key, expiredTime)
	}
	s.Lock()
	defer s.Unlock()
	s.maps[key] = m
	if isNewKeyItem {
		s.kq.add(m.kt)
	} else {
		s.kq.fix(m.kt.index)
	}
}

func (s *store) len() int {
	s.RLocker()
	defer s.RUnlock()
	return len(s.maps)
}

func (s *store) get(key string) *Message {
	s.RLock()
	defer s.RUnlock()
	if k, ok := s.maps[key]; ok {
		return k
	} else {
		return nil
	}
}

func (s *store) delete(key string) {
	m := s.get(key)
	if m != nil {
		s.Lock()
		defer s.Unlock()
		delete(s.maps, key)
		s.kq.remove(m.kt.index)
	}
}

func (s *store) first() *keyItem {
	s.RLock()
	s.RUnlock()
	if s.kq.len() > 0 {
		return s.kq.first()
	}
	return nil
}

func (s *store) deleteExpirationMessage(now int64) {
	for f := s.first(); f != nil && f.pastTime <= now; {
		s.delete(f.key)
	}
}

func (s *store) syncDeleteExpiration() {
	t := time.NewTicker(s.syncDelete)
	for {
		select {
		case <-s.ctx.Done():
			goto exit
		case now := <-t.C:
			s.deleteExpirationMessage(now.UnixNano())
		}
	}
exit:
	internal.Lg.Info("sync delete expiration message close")
}

func (s *store) getAllKQ() []*keyItem {
	s.RLock()
	defer s.RUnlock()
	dst := make([]*keyItem, len(s.kq.items))
	copy(dst, s.kq.items)
	return dst
}

type cachingProcess struct {
	process chan uint
}

func newCachingProcess(key string) {
	cp := &cachingProcess{
		process: make(chan uint),
	}
	go func() {
		Cache.cachingKeys[key] = cp

		t := time.NewTimer(500 * time.Millisecond)
		select {
		case <-t.C:
			internal.Lg.Errorf("超时")
			close(cp.process)
		case <-cp.process:
			internal.Lg.Errorf("save suc")
		}
		delete(Cache.cachingKeys, key)
	}()
}

func closeCachingProcess(key string) {
	if cp, ok := Cache.cachingKeys[key]; ok {
		close(cp.process)
	}
}
