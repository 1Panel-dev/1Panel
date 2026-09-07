package terminal

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"sync"
	"time"

	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/i18n"
	terminalai "github.com/1Panel-dev/1Panel/agent/utils/terminal/ai"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	gossh "golang.org/x/crypto/ssh"
)

// Lifetime rules. A session outlives its websocket:
//   - the client closes its websocket with 1000  -> the shell is closed at once
//   - the websocket drops any other way (browser tab closed, network) -> the
//     shell waits graceTimeout for a reattach, then is closed
//
// ponytail: all fixed; promote to settings only if someone asks.
const (
	graceTimeout       = 30 * time.Minute
	revalidateInterval = 60 * time.Second
	revalidateGrace    = 30 * time.Second
	keepaliveInterval  = 30 * time.Second
	keepaliveTimeout   = 10 * time.Second
	pumpInterval       = 60 * time.Millisecond
)

// Websocket close codes of the session protocol; the frontend switches on them.
const (
	// CloseCodeSessionNotFound: the session is gone or not the caller's; do not retry.
	CloseCodeSessionNotFound = 4404
	// CloseCodeAttachedElsewhere: a newer websocket took over the session.
	CloseCodeAttachedElsewhere = 4409
	CloseCodeRevalidate = 4410
)

var errSessionClosed = errors.New("terminal session is closed")

// SessionOptions describes a session that is about to be created.
type SessionOptions struct {
	Identity Identity
	Kind     string
	Target   string
	Title    string
	HostID   uint // 0 = local shell
	Cols     int
	Rows     int
	InitCmd  string
}

// Info is the client visible snapshot of a session.
type Info struct {
	ID         string    `json:"id"`
	Kind       string    `json:"kind"`
	Title      string    `json:"title"`
	HostID     uint      `json:"hostId"`
	Attached   bool      `json:"attached"`
	CreatedAt  time.Time `json:"createdAt"`
	DetachedAt time.Time `json:"detachedAt"` // zero while attached
}

// Session owns one shell; a websocket is only a detachable attachment.
type Session struct {
	ID            string
	UserID        string
	AuthSessionID string
	Kind          string
	Target        string
	Title         string
	HostID        uint
	CreatedAt     time.Time

	mu                sync.Mutex
	attached          *attachment
	detachedAt        time.Time
	grace             *time.Timer
	revalidateCursor  uint64
	revalidatePending bool
	cols              int
	rows              int

	backend sessionBackend
	ring    *ringBuffer

	lang          string
	aiInterceptor *aiInputInterceptor
	aiVersion     uint64

	done    chan struct{}
	closeFn func()
}

type sessionBackend interface {
	io.Writer
	Resize(cols, rows int) error
	Wait() error
	Keepalive() error
	Close() error
}

// Serve drives ws until it ends: it reattaches to sessionID when given, and
// otherwise opens a fresh shell on the client that connect returns. A returned
// error has not been reported to the client yet.
func Serve(ws *websocket.Conn, sessionID string, opts SessionOptions, connect func() (*gossh.Client, error)) error {
	return serve(ws, sessionID, opts, func() (*Session, error) {
		client, err := connect()
		if err != nil {
			return nil, err
		}
		sess, err := Open(client, opts)
		if err != nil {
			_ = client.Close()
		}
		return sess, err
	})
}

func ServeCommand(ws *websocket.Conn, sessionID string, opts SessionOptions, connect func() (*LocalCommand, error)) error {
	return serve(ws, sessionID, opts, func() (*Session, error) {
		command, err := connect()
		if err != nil {
			return nil, err
		}
		sess, err := OpenCommand(command, opts)
		if err != nil {
			_ = command.Close()
		}
		return sess, err
	})
}

func serve(ws *websocket.Conn, sessionID string, opts SessionOptions, open func() (*Session, error)) error {
	if sessionID != "" {
		sess, ok := Lookup(sessionID, opts.Identity)
		if ok && sess.Kind == opts.Kind && sess.Target == opts.Target && sess.HostID == opts.HostID {
			att, err := sess.Attach(ws, opts.Cols, opts.Rows)
			if err == nil {
				att.Run()
				return nil
			}
			global.LOG.Errorf("attach terminal session %s failed, err: %v", sessionID, err)
		}
		sendClose(ws, CloseCodeSessionNotFound, "session not found")
		return nil
	}

	sess, err := open()
	if err != nil {
		return err
	}
	// no sess.Close() on return: a dirty disconnect leaves the shell alive for a reattach
	att, err := sess.Attach(ws, opts.Cols, opts.Rows)
	if err != nil {
		sess.Close()
		return err
	}
	att.Run()
	return nil
}

// Open starts a shell on client and registers the session.
func Open(client *gossh.Client, opts SessionOptions) (*Session, error) {
	if err := validateSessionOptions(opts); err != nil {
		return nil, err
	}
	if err := reserveSessionSlot(opts.Identity); err != nil {
		return nil, err
	}
	ring := newRingBuffer()
	backend, err := newSSHBackend(client, opts.Cols, opts.Rows, opts.InitCmd, ring)
	if err != nil {
		releaseSessionSlot(opts.Identity)
		return nil, err
	}
	return openBackend(backend, ring, opts), nil
}

func OpenCommand(command *LocalCommand, opts SessionOptions) (*Session, error) {
	if err := validateSessionOptions(opts); err != nil {
		return nil, err
	}
	if command == nil {
		return nil, errors.New("nil terminal command")
	}
	if err := reserveSessionSlot(opts.Identity); err != nil {
		return nil, err
	}
	ring := newRingBuffer()
	return openBackend(newCommandBackend(command, ring), ring, opts), nil
}

func openBackend(backend sessionBackend, ring *ringBuffer, opts SessionOptions) *Session {
	lang := i18n.GetLanguageFromDB()
	s := &Session{
		ID:            uuid.NewString(),
		UserID:        opts.Identity.UserID,
		AuthSessionID: opts.Identity.AuthSessionID,
		Kind:          opts.Kind,
		Target:        opts.Target,
		Title:         opts.Title,
		HostID:        opts.HostID,
		CreatedAt:     time.Now(),
		cols:          opts.Cols,
		rows:          opts.Rows,
		backend:       backend,
		ring:          ring,
		lang:          lang,
		aiInterceptor: newAIInputInterceptor("", lang),
		aiVersion:     terminalai.CurrentTerminalRuntimeVersion(),
		done:          make(chan struct{}),
	}
	s.closeFn = sync.OnceFunc(s.doClose)
	registerReservedSession(s)
	go s.pump()
	go s.keepaliveLoop()
	go s.waitBackend()
	return s
}

func validateSessionOptions(opts SessionOptions) error {
	if !opts.Identity.Valid() {
		return errors.New("missing terminal identity")
	}
	if opts.Kind != "local" && opts.Kind != "ssh" && opts.Kind != "container" {
		return errors.New("invalid terminal kind")
	}
	if opts.Kind == "container" && opts.Target == "" {
		return errors.New("missing container terminal target")
	}
	return nil
}

// Attach binds ws to the session, kicking any previous attachment, and replays
// the retained output tail before any live output.
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
	s.detachedAt = time.Time{}
	if s.grace != nil {
		s.grace.Stop()
		s.grace = nil
	}
	if cols > 0 {
		s.cols = cols
	}
	if rows > 0 {
		s.rows = rows
	}
	cols, rows = s.cols, s.rows
	att.cursor = s.ring.Oldest()
	if s.revalidatePending {
		att.cursor = s.revalidateCursor
	}
	s.revalidatePending = false
	// Hold writeMu across unlock so hello+replay go out before the pump can write.
	att.writeMu.Lock()
	s.mu.Unlock()

	if previous != nil {
		previous.close(CloseCodeAttachedElsewhere, "attached elsewhere")
	}

	err := func() error {
		defer att.writeMu.Unlock()
		hello, err := json.Marshal(WsMsg{Type: WsMsgSession, ID: s.ID})
		if err != nil {
			return err
		}
		if err := att.writeLocked(hello); err != nil {
			return err
		}
		data, next, _ := s.ring.ReadFrom(att.cursor)
		att.cursor = next
		if len(data) == 0 {
			return nil
		}
		return att.writeLocked(cmdMessage(data))
	}()
	if err != nil {
		att.close(websocket.CloseInternalServerErr, "attach failed")
		s.detach(att, false, false, 0)
		return nil, err
	}

	if err := s.backend.Resize(cols, rows); err != nil {
		global.LOG.Errorf("ssh pty change windows size failed, err: %v", err)
	}
	return att, nil
}

// detach unbinds a. A clean detach closes the shell; a dirty one arms the grace timer.
func (s *Session) detach(a *attachment, clean, revalidate bool, cursor uint64) {
	s.mu.Lock()
	if s.attached != a {
		s.mu.Unlock()
		return
	}
	s.attached = nil
	s.detachedAt = time.Now()
	s.revalidatePending = revalidate
	if revalidate {
		s.revalidateCursor = cursor
	}
	if !clean {
		timeout := graceTimeout
		if revalidate {
			timeout = revalidateGrace
		}
		s.grace = time.AfterFunc(timeout, s.Close)
	}
	s.mu.Unlock()
	global.LOG.Debugf("terminal session %s detached, clean=%v", s.ID, clean)
	if clean {
		s.Close()
	}
}

func (s *Session) markRevalidation(a *attachment, cursor uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.attached != a {
		return
	}
	s.revalidatePending = true
	s.revalidateCursor = cursor
}

// Close terminates the shell and any attachment. Idempotent.
func (s *Session) Close() { s.closeFn() }

func (s *Session) doClose() {
	close(s.done)
	s.mu.Lock()
	att := s.attached
	s.attached = nil
	if s.grace != nil {
		s.grace.Stop()
	}
	s.mu.Unlock()
	if att != nil {
		att.close(websocket.CloseNormalClosure, "")
	}
	unregisterSession(s)
	if err := s.backend.Close(); err != nil {
		global.LOG.Debugf("close terminal backend: %v", err)
	}
}

// Info snapshots the session for listing.
func (s *Session) Info() Info {
	s.mu.Lock()
	defer s.mu.Unlock()
	return Info{
		ID:         s.ID,
		Kind:       s.Kind,
		Title:      s.Title,
		HostID:     s.HostID,
		Attached:   s.attached != nil,
		CreatedAt:  s.CreatedAt,
		DetachedAt: s.detachedAt,
	}
}

// resize forwards a window size change to the shell.
func (s *Session) resize(cols, rows int) {
	s.mu.Lock()
	s.cols, s.rows = cols, rows
	s.mu.Unlock()
	if err := s.backend.Resize(cols, rows); err != nil {
		global.LOG.Errorf("ssh pty change windows size failed, err: %v", err)
	}
}

// writeInput forwards client input to the shell stdin.
func (s *Session) writeInput(data []byte) {
	if _, err := s.backend.Write(data); err != nil {
		global.LOG.Errorf("ws cmd bytes write to ssh.stdin pipe failed, err: %v", err)
	}
}

// ensureAIInterceptor rebuilds the interceptor when AI runtime settings change.
func (s *Session) ensureAIInterceptor() *aiInputInterceptor {
	s.mu.Lock()
	defer s.mu.Unlock()
	if v := terminalai.CurrentTerminalRuntimeVersion(); s.aiInterceptor == nil || s.aiVersion != v {
		s.aiVersion = v
		s.aiInterceptor = newAIInputInterceptor("", s.lang)
	}
	return s.aiInterceptor
}

// pump forwards new ring output to the current attachment.
func (s *Session) pump() {
	defer func() {
		if r := recover(); r != nil {
			global.LOG.Errorf("a panic occurred during send combo output, error message: %v", r)
		}
	}()
	tick := time.NewTicker(pumpInterval)
	defer tick.Stop()
	for {
		select {
		case <-s.done:
			return
		case <-tick.C:
			s.flush()
		}
	}
}

// flush sends everything the attachment has not seen yet. A client that fell
// behind the ring skips ahead and is told so; output is never queued unbounded.
func (s *Session) flush() {
	s.mu.Lock()
	att := s.attached
	s.mu.Unlock()
	if att == nil {
		return
	}
	att.writeMu.Lock()
	defer att.writeMu.Unlock()
	data, next, lost := s.ring.ReadFrom(att.cursor)
	if lost {
		notice := "\r\n\x1b[33m" + i18n.GetMsgByKeyAndLang(s.lang, "TerminalOutputTruncated") + "\x1b[m\r\n"
		if err := att.writeLocked(cmdMessage([]byte(notice))); err != nil {
			att.close(websocket.CloseInternalServerErr, "write failed")
			return
		}
	}
	if len(data) == 0 {
		return
	}
	if err := att.writeLocked(cmdMessage(data)); err != nil {
		global.LOG.Errorf("ssh sending combo output to webSocket failed, err: %v", err)
		att.close(websocket.CloseInternalServerErr, "write failed")
		return
	}
	att.cursor = next
}

// keepaliveLoop probes the shell connection; a failed or stuck probe closes the session.
func (s *Session) keepaliveLoop() {
	tick := time.NewTicker(keepaliveInterval)
	defer tick.Stop()
	for {
		select {
		case <-s.done:
			return
		case <-tick.C:
			result := make(chan error, 1)
			go func() { result <- s.backend.Keepalive() }()
			select {
			case err := <-result:
				if err != nil {
					global.LOG.Infof("terminal session %s keepalive failed: %v", s.ID, err)
					s.Close()
					return
				}
			case <-time.After(keepaliveTimeout):
				global.LOG.Infof("terminal session %s keepalive timed out", s.ID)
				s.Close()
				return
			case <-s.done:
				return
			}
		}
	}
}

// waitBackend closes the session once the shell exits, after a last flush.
func (s *Session) waitBackend() {
	_ = s.backend.Wait()
	s.flush()
	s.Close()
}

func cmdMessage(data []byte) []byte {
	msg, _ := json.Marshal(WsMsg{Type: WsMsgCmd, Data: base64.StdEncoding.EncodeToString(data)})
	return msg
}

// sendClose writes a close frame with code and reason, best effort.
func sendClose(ws *websocket.Conn, code int, reason string) {
	_ = ws.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(code, reason), time.Now().Add(time.Second))
}
