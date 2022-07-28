package cache

import (
	"sync"
	"time"
)

type store struct {
	index int
	sync.RWMutex
	maps map[string]*Message
	kq   *keyQueue
}

func newStore(index int) *store {
	return &store{
		index: index,
		maps:  make(map[string]*Message),
		kq:    newKewQueue(),
	}
}

func (s *store) Save(m *Message) {
	key := string(m.Key)
	om := s.get(key)
	isNewKeyItem := false
	if om != nil {
		m.kt, om.kt = om.kt, nil
		m.kt.pastTime = m.ExpirationTime
	} else {
		isNewKeyItem = true
		m.kt = newKeyItem(key, m.ExpirationTime)
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

func (s *store) Get(key string) *Message {
	return s.get(key)
}

func (s *store) Remove(key string) {
	s.delete(key)
}

func (s *store) get(key string) *Message {
	s.RLocker()
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

func (s *store) deleteExpirationMessage() {
	t := time.Now().UnixNano()
	for f := s.first(); f != nil && f.pastTime > t; {
		s.delete(f.key)
	}
}
