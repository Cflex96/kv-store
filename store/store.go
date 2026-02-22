package store

import "sync"

type MemoryStore struct {
	store map[string]Type
	mut   *sync.RWMutex
}

func New() *MemoryStore {
	return &MemoryStore{make(map[string]Type), &sync.RWMutex{}}
}

func (s *MemoryStore) Set(key string, value Type) {
	s.mut.Lock()
	defer s.mut.Unlock()
	s.store[key] = value
}

func (s *MemoryStore) Get(key string) Type {
	s.mut.RLock()
	defer s.mut.RUnlock()
	return s.store[key]
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
