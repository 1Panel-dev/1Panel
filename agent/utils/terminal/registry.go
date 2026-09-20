package terminal

import (
	"errors"
	"sort"
	"strings"
	"sync"
	"time"
)

var sessions sync.Map

var errSessionNotFound = errors.New("terminal session not found")

const (
	HeaderUserID         = "X-Panel-User-ID"
	HeaderAuthSessionID  = "X-Panel-Auth-Session-ID"
	HeaderAuthLeaseUntil = "X-Panel-Auth-Lease-Until"
)

type Identity struct {
	UserID         string
	AuthSessionID  string
	AuthLeaseUntil time.Time
}

func (i Identity) Valid() bool {
	if i.UserID == "" || i.AuthSessionID == "" {
		return false
	}
	if strings.HasPrefix(i.AuthSessionID, "api-key:") && i.AuthLeaseUntil.IsZero() {
		return false
	}
	return i.AuthLeaseUntil.IsZero() || time.Now().Before(i.AuthLeaseUntil)
}

func registerSession(s *Session) {
	sessions.Store(s.ID, s)
}

func unregisterSession(s *Session) {
	sessions.CompareAndDelete(s.ID, s)
}

func Lookup(id string, identity Identity) (*Session, bool) {
	if !identity.Valid() {
		return nil, false
	}
	v, ok := sessions.Load(id)
	if !ok {
		return nil, false
	}
	s := v.(*Session)
	if s.UserID != identity.UserID || s.AuthSessionID != identity.AuthSessionID {
		return nil, false
	}
	return s, true
}

func List(identity Identity) []Info {
	if !identity.Valid() {
		return nil
	}
	var out []Info
	sessions.Range(func(_, v any) bool {
		s := v.(*Session)
		if s.UserID == identity.UserID && s.AuthSessionID == identity.AuthSessionID {
			out = append(out, s.Info())
		}
		return true
	})
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.Before(out[j].CreatedAt) })
	return out
}

func CloseSession(id string, identity Identity) error {
	s, ok := Lookup(id, identity)
	if !ok {
		return errSessionNotFound
	}
	s.Close()
	return nil
}

func Revoke(scope, userID, authSessionID string) int {
	closed := 0
	sessions.Range(func(_, v any) bool {
		s := v.(*Session)
		match := false
		switch scope {
		case "auth_session":
			match = userID != "" && authSessionID != "" && s.UserID == userID && s.AuthSessionID == authSessionID
		case "user":
			match = userID != "" && s.UserID == userID
		case "all":
			match = true
		}
		if match {
			closed++
			s.Close()
		}
		return true
	})
	return closed
}
