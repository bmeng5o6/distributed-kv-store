package store

import (
	"sync"
)

type Store struct {
	db  map[string]string
	mu  sync.RWMutex
	wal *WAL
}

func NewStore(path string) (*Store, error) {
	wal, err := NewWAL(path)
	if err != nil {
		return nil, err
	}

	store := &Store{
		db: make(map[string]string), wal: wal,
	}

	ops, err := Replay(path)
	if err != nil {
		return nil, err
	}

	for _, op := range ops {
		key := op[1]
		value := op[2]

		switch op[0] {
		case "PUT":
			store.db[key] = value
		case "DELETE":
			delete(store.db, key)
		}
	}

	return store, nil
}

func (s *Store) Get(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	val, ok := s.db[key]

	return val, ok
}

func (s *Store) Put(key string, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	err := s.wal.Append("PUT", key, value)
	if err != nil {
		return err
	}

	s.db[key] = value
	return nil
}

func (s *Store) Delete(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, ok := s.db[key]
	if !ok {
		return false
	}

	s.wal.Append("DELETE", key, "")
	delete(s.db, key)
	return true
}
