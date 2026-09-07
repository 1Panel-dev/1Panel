package terminal

import (
	"errors"
	"sort"
	"sync"
)

// sessions is the process wide registry of live sessions, keyed by id.
// Open stores, Close deletes.
var sessions sync.Map

var errSessionNotFound = errors.New("terminal session not found")

// Lookup resolves id for owner. Foreign and unknown sessions look the same.
func Lookup(id, owner string) (*Session, bool) {
	v, ok := sessions.Load(id)
	if !ok {
		return nil, false
	}
	s := v.(*Session)
	if s.Owner != "" && owner != "" && s.Owner != owner {
		return nil, false
	}
	return s, true
}

// List returns owner's sessions, oldest first.
func List(owner string) []Info {
	var out []Info
	sessions.Range(func(_, v any) bool {
		s := v.(*Session)
		if s.Owner == "" || owner == "" || s.Owner == owner {
			out = append(out, s.Info())
		}
		return true
	})
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.Before(out[j].CreatedAt) })
	return out
}

// CloseAll ends every live session; core calls it when the panel user logs out.
func CloseAll() {
	sessions.Range(func(_, v any) bool {
		v.(*Session).Close()
		return true
	})
}

// CloseSession closes id on behalf of owner.
func CloseSession(id, owner string) error {
	s, ok := Lookup(id, owner)
	if !ok {
		return errSessionNotFound
	}
	s.Close()
	return nil
}
