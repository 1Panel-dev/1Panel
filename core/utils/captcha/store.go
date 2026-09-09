package captcha

import (
	"container/list"
	"errors"
	"sync"
	"time"
)

type challenge struct {
	id        string
	answer    string
	expiresAt time.Time
}

type memoryStore struct {
	mu         sync.Mutex
	entries    map[string]*list.Element
	order      *list.List
	limit      int
	expiration time.Duration
	now        func() time.Time
}

func newMemoryStore(limit int, expiration time.Duration) *memoryStore {
	return &memoryStore{
		entries: make(map[string]*list.Element), order: list.New(),
		limit: limit, expiration: expiration, now: time.Now,
	}
}

func (s *memoryStore) put(id, answer string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.now()
	s.removeExpired(now)
	if len(s.entries) >= s.limit {
		return errors.New("captcha store is full")
	}
	if _, exists := s.entries[id]; exists {
		return errors.New("captcha ID already exists")
	}
	s.entries[id] = s.order.PushBack(challenge{id: id, answer: answer, expiresAt: now.Add(s.expiration)})
	return nil
}

func (s *memoryStore) consume(id string) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.now()
	s.removeExpired(now)
	e, ok := s.entries[id]
	if !ok {
		return ""
	}
	value := e.Value.(challenge)
	s.remove(e)
	if !now.Before(value.expiresAt) {
		return ""
	}
	return value.answer
}

func (s *memoryStore) removeExpired(now time.Time) {
	for e := s.order.Front(); e != nil; e = s.order.Front() {
		if now.Before(e.Value.(challenge).expiresAt) {
			return
		}
		s.remove(e)
	}
}

func (s *memoryStore) remove(e *list.Element) {
	delete(s.entries, e.Value.(challenge).id)
	s.order.Remove(e)
}
