package idempotency

import "sync"

type Store struct {
	mu   sync.Mutex
	seen map[string]struct{}
}

func NewStore() *Store {
	return &Store{seen: make(map[string]struct{})}
}

func (s *Store) Seen(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.seen[id]; ok {
		return true
	}
	s.seen[id] = struct{}{}
	return false
}

func (s *Store) Forget(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.seen, id)
}
