package store

import "sync"

type MemoryStore struct {
	store map[string]Value
	mut   *sync.RWMutex
}

func New() *MemoryStore {
	return &MemoryStore{make(map[string]Value), &sync.RWMutex{}}
}

func (s *MemoryStore) Set(key string, value Value) {
	s.mut.Lock()
	defer s.mut.Unlock()
	s.store[key] = value
}

func (s *MemoryStore) Get(key string) (Value, bool) {
	s.mut.RLock()
	defer s.mut.RUnlock()
	val, ok := s.store[key]
	return val, ok
}

func (s *MemoryStore) Delete(key string) bool {
	s.mut.Lock()
	defer s.mut.Unlock()
	if s.store[key] != nil {
		delete(s.store, key)
		return true
	}
	return false
}
