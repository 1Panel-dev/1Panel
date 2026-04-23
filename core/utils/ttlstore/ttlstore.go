package ttlstore

import (
	"sync"
	"time"
)

type Item[T any] struct {
	Value     T
	ExpiresAt time.Time
}

type Store[T any] struct {
	mu         sync.Mutex
	items      map[string]Item[T]
	ttl        time.Duration
	maxEntries int
	newID      func() string
	now        func() time.Time
}

func New[T any](ttl time.Duration, maxEntries int, newID func() string) *Store[T] {
	return &Store[T]{
		items:      make(map[string]Item[T]),
		ttl:        ttl,
		maxEntries: maxEntries,
		newID:      newID,
		now:        time.Now,
	}
}

func (s *Store[T]) Set(value T) string {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.cleanupExpiredLocked()
	if len(s.items) >= s.maxEntries {
		s.removeOldestLocked()
	}

	itemID := s.newID()
	s.items[itemID] = Item[T]{
		Value:     value,
		ExpiresAt: s.now().Add(s.ttl),
	}
	return itemID
}

func (s *Store[T]) Get(itemID string) (T, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	item, ok := s.items[itemID]
	if !ok {
		var zero T
		return zero, false
	}
	if s.now().After(item.ExpiresAt) {
		delete(s.items, itemID)
		var zero T
		return zero, false
	}
	return item.Value, true
}

func (s *Store[T]) Update(itemID string, fn func(*T) bool) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	item, ok := s.items[itemID]
	if !ok {
		return false
	}
	if s.now().After(item.ExpiresAt) {
		delete(s.items, itemID)
		return false
	}

	keep := fn(&item.Value)
	if !keep {
		delete(s.items, itemID)
		return true
	}

	s.items[itemID] = item
	return true
}

func (s *Store[T]) Delete(itemID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.items, itemID)
}

func (s *Store[T]) cleanupExpiredLocked() {
	now := s.now()
	for id, item := range s.items {
		if now.After(item.ExpiresAt) {
			delete(s.items, id)
		}
	}
}

func (s *Store[T]) removeOldestLocked() {
	var oldestID string
	var oldestTime time.Time
	for id, item := range s.items {
		if oldestID == "" || item.ExpiresAt.Before(oldestTime) {
			oldestID = id
			oldestTime = item.ExpiresAt
		}
	}
	if oldestID != "" {
		delete(s.items, oldestID)
	}
}
