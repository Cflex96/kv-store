package main

import "sync"

type MemoryStore struct {
	store map[string]string
	mut   *sync.RWMutex
}

func New() *MemoryStore {
	return &MemoryStore{make(map[string]string), &sync.RWMutex{}}
}

func (s *MemoryStore) Set(key, value string) {
	s.mut.Lock()
	defer s.mut.Unlock()
	s.store[key] = value
}

func (s *MemoryStore) Get(key string) string {
	s.mut.RLock()
	defer s.mut.RUnlock()
	return s.store[key]
}

func (s *MemoryStore) Delete(key string) bool {
	s.mut.Lock()
	defer s.mut.Unlock()
	if s.store[key] != "" {
		delete(s.store, key)
		return true
	}
	return false
}
