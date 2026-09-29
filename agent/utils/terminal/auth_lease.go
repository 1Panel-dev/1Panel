package terminal

import "time"

func (s *Session) renewAuthLease(until time.Time) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	select {
	case <-s.done:
		return false
	default:
	}
	now := time.Now()
	if s.authLeaseExpired || (!s.authLeaseUntil.IsZero() && !now.Before(s.authLeaseUntil)) {
		return false
	}
	if until.IsZero() {
		return s.authLeaseUntil.IsZero()
	}
	if !now.Before(until) {
		return false
	}
	s.authLeaseUntil = until
	s.authLeaseVersion++
	version := s.authLeaseVersion
	if s.authLeaseTimer != nil {
		s.authLeaseTimer.Stop()
	}
	s.authLeaseTimer = time.AfterFunc(time.Until(until), func() {
		s.mu.Lock()
		expired := s.authLeaseVersion == version && !time.Now().Before(s.authLeaseUntil)
		if expired {
			s.authLeaseExpired = true
		}
		s.mu.Unlock()
		if expired {
			s.Close()
		}
	})
	return true
}
