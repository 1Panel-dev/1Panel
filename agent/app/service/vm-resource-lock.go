package service

import (
	"context"
	"sync"
	"time"

	"github.com/1Panel-dev/1Panel/agent/global"
)

type vmResourceKey struct{ kind, name string }
type vmResourceGate struct {
	token chan struct{}
	refs  int
}

type vmResourceLockSet struct {
	mu       sync.Mutex
	gates    map[vmResourceKey]*vmResourceGate
	active   map[string]int
	revision map[string]uint64
}

var vmResourceLocks = newVMResourceLockSet()

func newVMResourceLockSet() *vmResourceLockSet {
	return &vmResourceLockSet{gates: make(map[vmResourceKey]*vmResourceGate), active: make(map[string]int), revision: make(map[string]uint64)}
}
func vmResourceOperationContext(force bool) (context.Context, context.CancelFunc) {
	timeout := 2 * time.Minute
	if force {
		timeout = 10 * time.Second
	}
	return context.WithTimeout(context.Background(), timeout)
}
func (s *vmResourceLockSet) lock(ctx context.Context, kind, name string) (func(), error) {
	key := vmResourceKey{kind, name}
	s.mu.Lock()
	gate := s.gates[key]
	if gate == nil {
		gate = &vmResourceGate{token: make(chan struct{}, 1)}
		gate.token <- struct{}{}
		s.gates[key] = gate
	}
	gate.refs++
	s.mu.Unlock()
	select {
	case <-ctx.Done():
		s.mu.Lock()
		s.dropRef(key, gate)
		s.mu.Unlock()
		return nil, ctx.Err()
	case <-gate.token:
	}
	s.mu.Lock()
	if err := ctx.Err(); err != nil {
		gate.token <- struct{}{}
		s.dropRef(key, gate)
		s.mu.Unlock()
		return nil, err
	}
	s.active[kind]++
	s.revision[kind]++
	s.mu.Unlock()
	return func() {
		s.mu.Lock()
		s.active[kind]--
		s.revision[kind]++
		gate.token <- struct{}{}
		s.dropRef(key, gate)
		s.mu.Unlock()
	}, nil
}
func (s *vmResourceLockSet) dropRef(key vmResourceKey, gate *vmResourceGate) {
	gate.refs--
	if gate.refs == 0 {
		delete(s.gates, key)
	}
}

func (s *vmResourceLockSet) lockForDelete(ctx context.Context, kind, name string, force bool) (func(), bool, error) {
	unlock, err := s.lock(ctx, kind, name)
	if err == nil {
		return unlock, true, nil
	}
	if !force {
		return nil, false, err
	}
	if global.LOG != nil {
		global.LOG.Warnf("[xpack] force delete VM %s %q metadata after waiting for resource lock failed: %v", kind, name, err)
	}
	s.mu.Lock()
	s.revision[kind]++
	return func() { s.revision[kind]++; s.mu.Unlock() }, false, nil
}

func (s *vmResourceLockSet) snapshot(kind string) (uint64, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.revision[kind], s.active[kind] == 0
}

func (s *vmResourceLockSet) applySnapshot(kind string, revision uint64, apply func() error) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.active[kind] != 0 || s.revision[kind] != revision {
		return nil
	}
	defer func() { s.revision[kind]++ }()
	return apply()
}
