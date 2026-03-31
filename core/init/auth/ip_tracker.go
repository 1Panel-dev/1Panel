package auth

import (
	"sync"
	"time"
)

const (
	MaxIPCount        = 100
	ExpireDuration    = 30 * time.Minute
	MaxFailedAttempts = 10
	LockDuration      = 5 * time.Minute
)

type IPRecord struct {
	NeedCaptcha bool
	FailedCount int
	LockedUntil time.Time
	LastUpdate  time.Time
}

type IPTracker struct {
	records map[string]*IPRecord
	ipOrder []string
	mu      sync.RWMutex
}

func NewIPTracker() *IPTracker {
	return &IPTracker{
		records: make(map[string]*IPRecord),
		ipOrder: make([]string, 0),
	}
}

func (t *IPTracker) NeedCaptcha(ip string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	record, exists := t.records[ip]
	if !exists {
		return false
	}

	if !t.prepareRecordUnsafe(ip, record) {
		t.removeIPUnsafe(ip)
		return false
	}

	return record.NeedCaptcha
}

func (t *IPTracker) IsLocked(ip string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	record, exists := t.records[ip]
	if !exists {
		return false
	}

	if !t.prepareRecordUnsafe(ip, record) {
		t.removeIPUnsafe(ip)
		return false
	}

	return time.Now().Before(record.LockedUntil)
}

func (t *IPTracker) SetNeedCaptcha(ip string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if record, exists := t.records[ip]; exists {
		if !t.prepareRecordUnsafe(ip, record) {
			t.removeIPUnsafe(ip)
		} else {
			record.NeedCaptcha = true
			record.LastUpdate = time.Now()
			return
		}
	}

	if len(t.records) >= MaxIPCount {
		t.removeOldestUnsafe()
	}

	t.records[ip] = &IPRecord{
		NeedCaptcha: true,
		LastUpdate:  time.Now(),
	}
	t.ipOrder = append(t.ipOrder, ip)
}

func (t *IPTracker) RecordFailure(ip string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if record, exists := t.records[ip]; exists {
		if !t.prepareRecordUnsafe(ip, record) {
			t.removeIPUnsafe(ip)
		} else {
			record.FailedCount++
			record.LastUpdate = time.Now()
			if record.FailedCount >= MaxFailedAttempts {
				record.LockedUntil = time.Now().Add(LockDuration)
			}
			return
		}
	}

	if len(t.records) >= MaxIPCount {
		t.removeOldestUnsafe()
	}

	t.records[ip] = &IPRecord{
		FailedCount: 1,
		LastUpdate:  time.Now(),
	}
	t.ipOrder = append(t.ipOrder, ip)
}

func (t *IPTracker) Clear(ip string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.removeIPUnsafe(ip)
}

func (t *IPTracker) removeIPUnsafe(ip string) {
	delete(t.records, ip)

	for i, storedIP := range t.ipOrder {
		if storedIP == ip {
			t.ipOrder = append(t.ipOrder[:i], t.ipOrder[i+1:]...)
			break
		}
	}
}

func (t *IPTracker) removeOldestUnsafe() {
	if len(t.ipOrder) == 0 {
		return
	}

	oldestIP := t.ipOrder[0]
	delete(t.records, oldestIP)
	t.ipOrder = t.ipOrder[1:]
}

func (t *IPTracker) prepareRecordUnsafe(ip string, record *IPRecord) bool {
	if time.Since(record.LastUpdate) > ExpireDuration {
		return false
	}
	if !record.LockedUntil.IsZero() && !time.Now().Before(record.LockedUntil) {
		record.FailedCount = 0
		record.LockedUntil = time.Time{}
		record.LastUpdate = time.Now()
	}
	return true
}
