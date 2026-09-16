package runtime

import (
	"context"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/1Panel-dev/1Panel/agent/utils/firewall/filter"
)

const inventoryTTL = 2 * time.Second

var inventoryGeneration atomic.Uint64

func InvalidateInventory() { inventoryGeneration.Add(1) }

type inventoryEntry struct {
	snapshots  []filter.Snapshot
	expires    time.Time
	generation uint64
}

type inventoryCache struct {
	mu      sync.Mutex
	entries map[string]inventoryEntry
	reading map[string]chan struct{}
}

func (e *Engine) ObserveInventory(ctx context.Context, scope filter.Scope, refresh bool) (filter.Snapshot, error) {
	snapshots, err := e.ObserveInventoryScopes(ctx, []filter.Scope{scope}, refresh)
	if err != nil {
		return filter.Snapshot{}, err
	}
	return snapshots[0], nil
}

func (e *Engine) ObserveInventoryScopes(ctx context.Context, scopes []filter.Scope, refresh bool) ([]filter.Snapshot, error) {
	read := func() ([]filter.Snapshot, error) {
		if len(scopes) == 1 {
			snapshot, err := e.adapter.Observe(ctx, scopes[0])
			return []filter.Snapshot{snapshot}, err
		}
		observer, ok := e.adapter.(filter.MultiScopeObserver)
		if !ok {
			return nil, filter.ErrAdapterUnavailable
		}
		return observer.ObserveScopes(ctx, scopes)
	}
	var snapshots []filter.Snapshot
	var err error
	if e.Provider() == filter.ProviderFirewalld || e.Provider() == filter.ProviderUFW {
		keys := make([]string, len(scopes))
		for i, scope := range scopes {
			keys[i] = scope.Normalize().Key()
		}
		snapshots, err = e.inventory.load(ctx, strings.Join(keys, "\n"), refresh, read)
	} else {
		snapshots, err = read()
	}
	if err != nil {
		return nil, err
	}
	result := append([]filter.Snapshot(nil), snapshots...)
	for i := range result {
		result[i].Rules = append([]filter.ObservedRule(nil), snapshots[i].Rules...)
		result[i].Notices = append([]filter.ScopeNotice(nil), snapshots[i].Notices...)
		if e.policy != nil {
			result[i], err = e.policy(ctx, result[i])
			if err != nil {
				return nil, err
			}
		}
	}
	return result, nil
}

func (c *inventoryCache) load(ctx context.Context, key string, refresh bool, read func() ([]filter.Snapshot, error)) ([]filter.Snapshot, error) {
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		c.mu.Lock()
		generation := inventoryGeneration.Load()
		if entry, ok := c.entries[key]; !refresh && ok && entry.generation == generation && time.Now().Before(entry.expires) {
			c.mu.Unlock()
			return entry.snapshots, nil
		}
		if done := c.reading[key]; done != nil {
			c.mu.Unlock()
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-done:
				refresh = false
				continue
			}
		}
		if c.reading == nil {
			c.reading = make(map[string]chan struct{})
		}
		done := make(chan struct{})
		c.reading[key] = done
		delete(c.entries, key)
		c.mu.Unlock()
		snapshots, err := read()
		c.mu.Lock()
		if err == nil && generation == inventoryGeneration.Load() {
			if len(c.entries) >= 32 {
				clear(c.entries)
			}
			if c.entries == nil {
				c.entries = make(map[string]inventoryEntry)
			}
			c.entries[key] = inventoryEntry{snapshots, time.Now().Add(inventoryTTL), generation}
		}
		delete(c.reading, key)
		close(done)
		c.mu.Unlock()
		return snapshots, err
	}
}
