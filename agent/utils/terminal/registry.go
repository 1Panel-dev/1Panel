package terminal

import (
	"errors"
	"fmt"
	"sort"
	"sync"
)

// sessions is the process wide registry of live sessions, keyed by id.
// Open stores, Close deletes.
var sessions sync.Map

const maxSessionsPerIdentity = 10

var (
	sessionSlotsMu sync.Mutex
	sessionSlots   = make(map[Identity]int)
)

var errSessionNotFound = errors.New("terminal session not found")

const (
	HeaderUserID        = "X-Panel-User-ID"
	HeaderAuthSessionID = "X-Panel-Auth-Session-ID"
)

type Identity struct {
	UserID        string
	AuthSessionID string
}

func (i Identity) Valid() bool {
	return i.UserID != "" && i.AuthSessionID != ""
}

func reserveSessionSlot(identity Identity) error {
	if !identity.Valid() {
		return errors.New("missing terminal identity")
	}
	sessionSlotsMu.Lock()
	defer sessionSlotsMu.Unlock()
	if sessionSlots[identity] >= maxSessionsPerIdentity {
		return fmt.Errorf("terminal session limit reached (maximum %d)", maxSessionsPerIdentity)
	}
	sessionSlots[identity]++
	return nil
}

func releaseSessionSlot(identity Identity) {
	sessionSlotsMu.Lock()
	defer sessionSlotsMu.Unlock()
	releaseSessionSlotLocked(identity)
}

func releaseSessionSlotLocked(identity Identity) {
	remaining := sessionSlots[identity] - 1
	if remaining <= 0 {
		delete(sessionSlots, identity)
		return
	}
	sessionSlots[identity] = remaining
}

func registerReservedSession(s *Session) {
	sessions.Store(s.ID, s)
}

func unregisterSession(s *Session) {
	sessionSlotsMu.Lock()
	defer sessionSlotsMu.Unlock()
	current, ok := sessions.Load(s.ID)
	if !ok || current != s {
		return
	}
	sessions.Delete(s.ID)
	releaseSessionSlotLocked(Identity{UserID: s.UserID, AuthSessionID: s.AuthSessionID})
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
