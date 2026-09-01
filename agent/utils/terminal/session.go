package terminal

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/i18n"
	terminalai "github.com/1Panel-dev/1Panel/agent/utils/terminal/ai"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	gossh "golang.org/x/crypto/ssh"
)

// Session kinds handled by this package.
const (
	SessionKindLocal = "local"
	SessionKindSSH   = "ssh"
)

// closeCodeAttachedElsewhere is sent when another websocket attaches to this session.
const closeCodeAttachedElsewhere = 4409

// comboFlushInterval is the output coalescing interval.
const comboFlushInterval = 60 * time.Millisecond

// logoutOutput is the shell output that marks a finished login shell.
var logoutOutput = []byte{13, 10, 108, 111, 103, 111, 117, 116, 13, 10}

var errSessionClosed = errors.New("terminal session is closed")

// shellBackend is the interactive shell a Session drives.
type shellBackend interface {
	Write(p []byte) (int, error)
	Resize(cols, rows int) error
	Keepalive() error
	Wait() error
	Close() error
}

// SessionOptions describes a session that is about to be created.
type SessionOptions struct {
	Kind     string
	HostID   uint
	Title    string
	Owner    string
	Cols     int
	Rows     int
	InitCmd  string
	RingSize int
}

// SessionInfo is an immutable view of a session state.
type SessionInfo struct {
	ID           string
	Kind         string
	HostID       uint
	Title        string
	Owner        string
	Pinned       bool
	Attached     bool
	CreatedAt    time.Time
	LastActiveAt time.Time
	DetachedAt   time.Time
	// set by Manager.List for detached sessions; Session.Info leaves it zero
	ExpiresAt time.Time
}

// Session owns one shell; a websocket is only a detachable attachment.
type Session struct {
	ID     string
	Kind   string
	HostID uint
	Title  string
	Owner  string

	mu           sync.Mutex
	pinned       bool
	createdAt    time.Time
	lastActiveAt time.Time
	detachedAt   time.Time
	cols         int
	rows         int
	attached     *attachment

	backend shellBackend
	combo   *safeBuffer
	ring    *ringBuffer
	// serializes flushCombo
	flushMu sync.Mutex

	lang          string
	aiInterceptor *aiInputInterceptor
	aiVersion     uint64

	done      chan struct{}
	closeOnce sync.Once
	onClosed  func(*Session)
}

// NewSession opens a shell on client and starts buffering its output.
func NewSession(client *gossh.Client, opts SessionOptions) (*Session, error) {
	combo := new(safeBuffer)
	backend, err := newSSHBackend(client, opts.Cols, opts.Rows, opts.InitCmd, combo)
	if err != nil {
		return nil, err
	}
	return newSessionWithBackend(backend, combo, opts), nil
}

// newSessionWithBackend starts the output pump and shell watcher.
func newSessionWithBackend(backend shellBackend, out *safeBuffer, opts SessionOptions) *Session {
	now := time.Now()
	lang := i18n.GetLanguageFromDB()
	sess := &Session{
		ID:     uuid.NewString(),
		Kind:   opts.Kind,
		HostID: opts.HostID,
		Title:  opts.Title,
		Owner:  opts.Owner,

		createdAt:    now,
		lastActiveAt: now,
		cols:         opts.Cols,
		rows:         opts.Rows,

		backend: backend,
		combo:   out,
		ring:    newRingBuffer(opts.RingSize),

		lang:          lang,
		aiInterceptor: newAIInputInterceptor("", lang),
		aiVersion:     terminalai.CurrentTerminalRuntimeVersion(),

		done: make(chan struct{}),
	}
	go sess.pump()
	go sess.waitBackend()
	return sess
}

// Done is closed once the session is terminated.
func (s *Session) Done() <-chan struct{} {
	return s.done
}

// SetOnClosed registers a hook invoked once when the session terminates.
func (s *Session) SetOnClosed(fn func(*Session)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.onClosed = fn
}

// SetPinned marks whether the session survives losing its websocket.
// Unpinning a detached session closes it immediately.
func (s *Session) SetPinned(pinned bool) {
	s.mu.Lock()
	s.pinned = pinned
	detached := s.attached == nil
	s.mu.Unlock()
	if !pinned && detached {
		s.Close()
	}
}

// Pinned reports whether the session survives the loss of its attachment.
func (s *Session) Pinned() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.pinned
}

// IsAttached reports whether a websocket is currently bound to the session.
func (s *Session) IsAttached() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.attached != nil
}

// Info returns a snapshot of the session state.
func (s *Session) Info() SessionInfo {
	s.mu.Lock()
	defer s.mu.Unlock()
	return SessionInfo{
		ID:           s.ID,
		Kind:         s.Kind,
		HostID:       s.HostID,
		Title:        s.Title,
		Owner:        s.Owner,
		Pinned:       s.pinned,
		Attached:     s.attached != nil,
		CreatedAt:    s.createdAt,
		LastActiveAt: s.lastActiveAt,
		DetachedAt:   s.detachedAt,
	}
}

// Attach binds ws to the session, kicking any previous attachment.
func (s *Session) Attach(ws *websocket.Conn, cols, rows int) (*attachment, error) {
	if ws == nil {
		return nil, errors.New("nil websocket connection")
	}
	att := &attachment{sess: s, ws: ws, done: make(chan struct{})}

	s.mu.Lock()
	select {
	case <-s.done:
		s.mu.Unlock()
		return nil, errSessionClosed
	default:
	}
	previous := s.attached
	s.attached = att
	if cols > 0 {
		s.cols = cols
	}
	if rows > 0 {
		s.rows = rows
	}
	s.detachedAt = time.Time{}
	s.lastActiveAt = time.Now()
	pinned := s.pinned
	newCols, newRows := s.cols, s.rows
	// Snapshot under lock so a concurrent flush is either replayed or delivered live, never both.
	replay := s.ring.Snapshot()
	// Hold writeMu across unlock so hello+replay go out before live output.
	att.writeMu.Lock()
	s.mu.Unlock()

	if previous != nil {
		previous.close(closeCodeAttachedElsewhere, "attached elsewhere")
	}

	err := func() error {
		defer att.writeMu.Unlock()
		hello, err := json.Marshal(WsMsg{Type: WsMsgSession, ID: s.ID, Pinned: &pinned})
		if err != nil {
			return err
		}
		if err := att.writeLocked(websocket.TextMessage, hello); err != nil {
			return err
		}
		if len(replay) == 0 {
			return nil
		}
		wsData, err := json.Marshal(WsMsg{Type: WsMsgCmd, Data: base64.StdEncoding.EncodeToString(replay)})
		if err != nil {
			return err
		}
		return att.writeLocked(websocket.TextMessage, wsData)
	}()
	if err != nil {
		att.close(websocket.CloseInternalServerErr, "attach failed")
		s.mu.Lock()
		if s.attached == att {
			s.attached = nil
		}
		s.mu.Unlock()
		return nil, err
	}

	if err := s.backend.Resize(newCols, newRows); err != nil {
		global.LOG.Errorf("ssh pty change windows size failed, err: %v", err)
	}
	return att, nil
}

// onAttachmentClosed is called once the attachment loop returned.
func (s *Session) onAttachmentClosed(a *attachment) {
	s.mu.Lock()
	if s.attached != a {
		s.mu.Unlock()
		return
	}
	s.attached = nil
	pinned := s.pinned
	s.detachedAt = time.Now()
	s.mu.Unlock()

	if !pinned {
		s.Close()
	}
}

// Close terminates the shell and any attachment. Idempotent.
func (s *Session) Close() {
	s.closeOnce.Do(func() {
		close(s.done)
		if s.backend != nil {
			_ = s.backend.Close()
		}
		s.mu.Lock()
		att := s.attached
		s.attached = nil
		onClosed := s.onClosed
		s.mu.Unlock()

		if att != nil {
			att.close(websocket.CloseNormalClosure, "")
		}
		if onClosed != nil {
			onClosed(s)
		}
	})
}

// keepalive probes the shell backend connection.
func (s *Session) keepalive() error {
	if s.backend == nil {
		return nil
	}
	return s.backend.Keepalive()
}

// resize forwards a window size change to the shell.
func (s *Session) resize(cols, rows int) {
	s.mu.Lock()
	s.cols = cols
	s.rows = rows
	s.lastActiveAt = time.Now()
	s.mu.Unlock()
	if err := s.backend.Resize(cols, rows); err != nil {
		global.LOG.Errorf("ssh pty change windows size failed, err: %v", err)
	}
}

// writeInput forwards client input to the shell stdin.
func (s *Session) writeInput(data []byte) {
	s.mu.Lock()
	s.lastActiveAt = time.Now()
	s.mu.Unlock()
	if _, err := s.backend.Write(data); err != nil {
		global.LOG.Errorf("ws cmd bytes write to ssh.stdin pipe failed, err: %v", err)
	}
}

// ensureAIInterceptor rebuilds the interceptor when AI runtime settings change.
func (s *Session) ensureAIInterceptor() *aiInputInterceptor {
	s.mu.Lock()
	defer s.mu.Unlock()
	currentVersion := terminalai.CurrentTerminalRuntimeVersion()
	if s.aiInterceptor == nil || s.aiVersion != currentVersion {
		s.aiVersion = currentVersion
		s.aiInterceptor = newAIInputInterceptor("", s.lang)
	}
	return s.aiInterceptor
}

// pump periodically flushes coalesced shell output.
func (s *Session) pump() {
	defer func() {
		if r := recover(); r != nil {
			global.LOG.Errorf("a panic occurred during send combo output, error message: %v", r)
		}
	}()
	tick := time.NewTicker(comboFlushInterval)
	defer tick.Stop()
	for {
		select {
		case <-s.done:
			return
		case <-tick.C:
			if bs := s.flushCombo(); bytes.Equal(bs, logoutOutput) {
				s.Close()
				return
			}
		}
	}
}

// flushCombo writes combo output into the ring and the current attachment.
func (s *Session) flushCombo() []byte {
	s.flushMu.Lock()
	defer s.flushMu.Unlock()
	bs := s.combo.Take()
	if len(bs) == 0 {
		return nil
	}
	s.mu.Lock()
	if _, err := s.ring.Write(bs); err != nil {
		global.LOG.Errorf("combo output to ring buffer failed, err: %v", err)
	}
	att := s.attached
	s.lastActiveAt = time.Now()
	s.mu.Unlock()
	if att == nil {
		return bs
	}
	wsData, err := json.Marshal(WsMsg{Type: WsMsgCmd, Data: base64.StdEncoding.EncodeToString(bs)})
	if err != nil {
		global.LOG.Errorf("encoding combo output to json failed, err: %v", err)
		return bs
	}
	if err := att.write(websocket.TextMessage, wsData); err != nil {
		global.LOG.Errorf("ssh sending combo output to webSocket failed, err: %v", err)
		att.close(websocket.CloseInternalServerErr, "write failed")
	}
	return bs
}

// waitBackend drains leftover combo after the shell exits, then closes the session.
func (s *Session) waitBackend() {
	_ = s.backend.Wait()
	s.flushCombo()
	s.Close()
}
