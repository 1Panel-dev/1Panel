package terminal_session

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/1Panel-dev/1Panel/core/global"
)

type RevokeFunc func(scope, userID, authSessionID string) error

type resumableRevocation struct {
	err   error
	retry RevokeFunc
}

func (e resumableRevocation) Error() string               { return e.err.Error() }
func (e resumableRevocation) Unwrap() error               { return e.err }
func (e resumableRevocation) RetryRevokeFunc() RevokeFunc { return e.retry }

func PendingRevocation(err error, retry RevokeFunc) error {
	if err == nil {
		return nil
	}
	return resumableRevocation{err: err, retry: retry}
}

func nextRevokeAttempt(err error, fallback RevokeFunc) RevokeFunc {
	var resumable interface{ RetryRevokeFunc() RevokeFunc }
	if errors.As(err, &resumable) && resumable.RetryRevokeFunc() != nil {
		return resumable.RetryRevokeFunc()
	}
	return fallback
}

type revokeTarget struct{ scope, userID, authSessionID string }
type revokeRetry struct {
	target   revokeTarget
	revoke   RevokeFunc
	next     time.Time
	deadline time.Time
	attempt  int
	version  uint64
}

type revocationQueue struct {
	mu      sync.Mutex
	pending map[revokeTarget]revokeRetry
	limit   int
	version uint64
}

var terminalRevocations = revocationQueue{pending: make(map[revokeTarget]revokeRetry), limit: 256}
var startRevocationWorker sync.Once

func RevokeWithRetry(scope, userID, authSessionID string, revoke RevokeFunc) error {
	if revoke == nil || (scope != "all" && userID == "") || (scope == "auth_session" && authSessionID == "") ||
		(scope != "all" && scope != "user" && scope != "auth_session") {
		return errors.New("invalid terminal revocation")
	}
	target := revokeTarget{scope, userID, authSessionID}
	terminalRevocations.mu.Lock()
	previous, hadPrevious := terminalRevocations.pending[target]
	terminalRevocations.mu.Unlock()
	err := revoke(scope, userID, authSessionID)
	if err == nil {
		if hadPrevious {
			terminalRevocations.complete(previous, nil, time.Now())
		}
		return nil
	}
	if !terminalRevocations.enqueue(target, nextRevokeAttempt(err, revoke), time.Now()) {
		err = errors.Join(err, errors.New("terminal revocation retry capacity exceeded"))
	} else {
		startRevocationWorker.Do(func() { go runRevocationRetries() })
	}
	logRevokeFailure(target, err)
	return fmt.Errorf("terminal closure is pending: %w", err)
}

func (q *revocationQueue) enqueue(target revokeTarget, revoke RevokeFunc, now time.Time) bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	if _, exists := q.pending[target]; !exists && len(q.pending) >= q.limit {
		return false
	}
	q.version++
	q.pending[target] = revokeRetry{target: target, revoke: revoke, next: now.Add(5 * time.Second), deadline: now.Add(2 * time.Minute), version: q.version}
	return true
}

func (q *revocationQueue) due(now time.Time) []revokeRetry {
	q.mu.Lock()
	defer q.mu.Unlock()
	var jobs []revokeRetry
	for _, item := range q.pending {
		if !now.Before(item.next) {
			jobs = append(jobs, item)
		}
	}
	return jobs
}

func (q *revocationQueue) complete(item revokeRetry, err error, now time.Time) {
	q.mu.Lock()
	defer q.mu.Unlock()
	current, exists := q.pending[item.target]
	if !exists || current.version != item.version {
		return
	}
	if err == nil || !now.Before(item.deadline) {
		delete(q.pending, item.target)
		return
	}
	item.attempt++
	item.revoke = nextRevokeAttempt(err, item.revoke)
	item.next = now.Add(time.Duration(min(item.attempt+1, 6)) * 5 * time.Second)
	q.pending[item.target] = item
}

func runRevocationRetries() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for now := range ticker.C {
		for _, item := range terminalRevocations.due(now) {
			if time.Now().After(item.deadline) {
				terminalRevocations.complete(item, errors.New("revocation retry deadline reached"), time.Now())
				logRevokeFailure(item.target, errors.New("revocation retries exhausted; Agent authorization lease must expire"))
				continue
			}
			err := item.revoke(item.target.scope, item.target.userID, item.target.authSessionID)
			terminalRevocations.complete(item, err, time.Now())
			if err != nil {
				logRevokeFailure(item.target, err)
			}
		}
	}
}

func logRevokeFailure(target revokeTarget, err error) {
	if global.LOG != nil {
		global.LOG.Warnf("terminal revocation scope=%s user=%s session=%s: %v", target.scope, target.userID, target.authSessionID, err)
	}
}
