package maps

import (
	"sync"
)

type SyncMap[Key comparable, Value any] struct {
	internalMap map[Key]Value
	mtx         sync.Mutex
}

func NewSyncMap[K comparable, V any]() *SyncMap[K, V] {
	return &SyncMap[K, V]{
		internalMap: make(map[K]V),
	}
}

func (s *SyncMap[Key, Value]) Put(key Key, value Value) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	s.internalMap[key] = value
}

func (s *SyncMap[Key, Value]) Has(key Key) bool {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	_, exists := s.internalMap[key]
	return exists
}

func (s *SyncMap[Key, Value]) Get(key Key) (Value, bool) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	value, ok := s.internalMap[key]
	return value, ok
}

func (s *SyncMap[Key, Value]) Len() int {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	return len(s.internalMap)
}

func (s *SyncMap[Key, Value]) Delete(key Key) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	delete(s.internalMap, key)
}

func (s *SyncMap[Key, Value]) Foreach(f func(key Key, value Value)) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	for key, value := range s.internalMap {
		f(key, value)
	}
}
