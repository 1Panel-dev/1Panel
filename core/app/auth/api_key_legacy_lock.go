package auth

import "sync"

var legacyAPIKeyMutationMu sync.Mutex

func WithLegacyAPIKeyMutation(change func() error) error {
	legacyAPIKeyMutationMu.Lock()
	defer legacyAPIKeyMutationMu.Unlock()
	return change()
}
