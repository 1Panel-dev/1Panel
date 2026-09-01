package terminal

import (
	"errors"
	"slices"
	"sync"
	"time"

	gossh "golang.org/x/crypto/ssh"
)

var (
	// ErrSessionNotFound covers unknown ids and foreign sessions.
	ErrSessionNotFound = errors.New("terminal session not found")
	ErrPinDisabled     = errors.New("terminal session pinning is disabled")
	ErrPinLimit        = errors.New("pinned terminal session limit reached")
)

const (
	DefaultKeepAlive = 30 * time.Minute
	DefaultMaxPinned = 10
)

const (
	managerInterval  = 30 * time.Second
	keepaliveTimeout = 10 * time.Second
	// grace for sessions that were registered but never attached
	unpinnedGrace = time.Minute
)

// Config is the session lifetime settings.
type Config struct {
	// KeepAlive is how long a pinned session survives without a websocket; 0 disables pinning.
	KeepAlive time.Duration
	MaxPinned int
	RingSize  int
}

// Manager owns every live terminal session of this agent.
type Manager struct {
	mu       sync.Mutex
	sessions map[string]*Session
	config   func() Config
	now      func() time.Time

	// serializes Pin's quota check-and-set
	pinMu sync.Mutex

	startOnce sync.Once
}

// DefaultManager is the process wide session registry.
var DefaultManager = NewManager()

// NewManager returns an empty registry; the background loop starts on first Register.
func NewManager() *Manager {
	return &Manager{
		sessions: make(map[string]*Session),
		now:      time.Now,
	}
}

// SetConfigProvider installs the settings callback. It is called lazily.
func (m *Manager) SetConfigProvider(fn func() Config) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.config = fn
}

// Config returns current settings, substituting defaults for unusable values.
func (m *Manager) Config() Config {
	m.mu.Lock()
	fn := m.config
	m.mu.Unlock()

	cfg := Config{KeepAlive: DefaultKeepAlive, MaxPinned: DefaultMaxPinned, RingSize: defaultRingSize}
	if fn != nil {
		cfg = fn()
	}
	cfg.KeepAlive = max(cfg.KeepAlive, 0)
	cfg.MaxPinned = max(cfg.MaxPinned, 0)
	if cfg.RingSize <= 0 {
		cfg.RingSize = defaultRingSize
	}
	return cfg
}

// Register adds s to the registry and makes it remove itself once it closes.
func (m *Manager) Register(s *Session) {
	if s == nil {
		return
	}
	s.SetOnClosed(m.remove)

	m.mu.Lock()
	m.sessions[s.ID] = s
	m.mu.Unlock()

	m.start()

	// close hook may have missed a session that already died
	select {
	case <-s.Done():
		m.remove(s)
	default:
	}
}

// OpenSession creates a session on client and registers it.
func (m *Manager) OpenSession(client *gossh.Client, opts SessionOptions) (*Session, error) {
	if opts.RingSize <= 0 {
		opts.RingSize = m.Config().RingSize
	}
	sess, err := NewSession(client, opts)
	if err != nil {
		return nil, err
	}
	m.Register(sess)
	return sess, nil
}

// Get returns the session with that id, if it is still alive.
func (m *Manager) Get(id string) (*Session, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	sess, ok := m.sessions[id]
	return sess, ok
}

// List returns sessions visible to owner, oldest first. Empty owner sees all.
func (m *Manager) List(owner string) []SessionInfo {
	cfg := m.Config()
	sessions := m.snapshot()
	infos := make([]SessionInfo, 0, len(sessions))
	for _, sess := range sessions {
		info := sess.Info()
		if !ownerMatches(info.Owner, owner) {
			continue
		}
		info.ExpiresAt = expiresAt(info, cfg)
		infos = append(infos, info)
	}
	slices.SortFunc(infos, func(a, b SessionInfo) int { return a.CreatedAt.Compare(b.CreatedAt) })
	return infos
}

// Pin pins or unpins a session. Unpinning a detached session closes it.
func (m *Manager) Pin(id, owner string, pinned bool) error {
	sess, err := m.Lookup(id, owner)
	if err != nil {
		return err
	}
	if !pinned {
		sess.SetPinned(false)
		return nil
	}

	cfg := m.Config()
	if cfg.KeepAlive <= 0 {
		return ErrPinDisabled
	}
	m.pinMu.Lock()
	defer m.pinMu.Unlock()
	if m.pinnedCount(sess) >= cfg.MaxPinned {
		return ErrPinLimit
	}
	sess.SetPinned(true)
	return nil
}

// Close terminates a session owned by owner.
func (m *Manager) Close(id, owner string) error {
	sess, err := m.Lookup(id, owner)
	if err != nil {
		return err
	}
	sess.Close()
	return nil
}

// ownerMatches is true if either side is empty (unknown) or they are equal.
func ownerMatches(sessionOwner, caller string) bool {
	return sessionOwner == "" || caller == "" || sessionOwner == caller
}

// Lookup resolves an id for one caller, hiding foreign sessions.
func (m *Manager) Lookup(id, owner string) (*Session, error) {
	sess, ok := m.Get(id)
	if !ok || !ownerMatches(sess.Owner, owner) {
		return nil, ErrSessionNotFound
	}
	return sess, nil
}

// remove is the session close hook; must not take a session lock or block.
func (m *Manager) remove(s *Session) {
	if s == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if current, ok := m.sessions[s.ID]; ok && current == s {
		delete(m.sessions, s.ID)
	}
}

// snapshot copies sessions so Close (which calls remove) cannot deadlock on m.mu.
func (m *Manager) snapshot() []*Session {
	m.mu.Lock()
	defer m.mu.Unlock()
	sessions := make([]*Session, 0, len(m.sessions))
	for _, sess := range m.sessions {
		sessions = append(sessions, sess)
	}
	return sessions
}

// pinnedCount counts the pinned sessions, ignoring exclude.
func (m *Manager) pinnedCount(exclude *Session) int {
	count := 0
	for _, sess := range m.snapshot() {
		if sess == exclude {
			continue
		}
		if sess.Pinned() {
			count++
		}
	}
	return count
}

// setNow overrides the clock the reaper reads, for tests.
func (m *Manager) setNow(fn func() time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.now = fn
}

func (m *Manager) timeNow() time.Time {
	m.mu.Lock()
	fn := m.now
	m.mu.Unlock()
	if fn == nil {
		return time.Now()
	}
	return fn()
}

// start launches the background loop exactly once.
func (m *Manager) start() {
	m.startOnce.Do(func() { go m.loop() })
}

// loop reaps expired sessions and probes the live ones forever.
func (m *Manager) loop() {
	for range time.Tick(managerInterval) {
		m.reap()
		m.keepaliveAll()
	}
}

// reap closes every detached session whose grace period elapsed.
func (m *Manager) reap() {
	sessions := m.snapshot()
	if len(sessions) == 0 {
		return
	}
	cfg := m.Config()
	now := m.timeNow()
	for _, sess := range sessions {
		if exp := expiresAt(sess.Info(), cfg); !exp.IsZero() && !exp.After(now) {
			sess.Close()
		}
	}
}

// expiresAt is when a detached session will be reaped, or zero if attached.
// Never-attached sessions measure grace from CreatedAt.
func expiresAt(info SessionInfo, cfg Config) time.Time {
	if info.Attached {
		return time.Time{}
	}
	since := info.DetachedAt
	if since.IsZero() {
		since = info.CreatedAt
	}
	grace := unpinnedGrace
	if info.Pinned {
		grace = cfg.KeepAlive
	}
	return since.Add(grace)
}

// keepaliveAll probes each session in its own goroutine; a failed probe closes that session.
func (m *Manager) keepaliveAll() {
	for _, sess := range m.snapshot() {
		go func() {
			result := make(chan error, 1)
			go func() { result <- sess.keepalive() }()
			select {
			case err := <-result:
				if err != nil {
					sess.Close()
				}
			case <-time.After(keepaliveTimeout):
				sess.Close()
			case <-sess.Done():
			}
		}()
	}
}
