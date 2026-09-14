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

const (
	graceTimeout       = 30 * time.Minute
	revalidateInterval = 60 * time.Second
	revalidateGrace    = 30 * time.Second
	keepaliveInterval  = 30 * time.Second
	keepaliveTimeout   = 10 * time.Second
	pumpInterval       = 60 * time.Millisecond
)

const (
	CloseCodeSessionNotFound   = 4404
	CloseCodeAttachedElsewhere = 4409
	CloseCodeRevalidate        = 4410
)

var errSessionClosed = errors.New("terminal session is closed")

type SessionOptions struct {
	Identity   Identity
	Kind       string
	Target     string
	Title      string
	Persistent bool
	HostID     uint // 0 = local shell
	Cols       int
	Rows       int
	InitCmd    string
}

type Info struct {
	ID         string    `json:"id"`
	Kind       string    `json:"kind"`
	Title      string    `json:"title"`
	Persistent bool      `json:"persistent"`
	HostID     uint      `json:"hostId"`
	Attached   bool      `json:"attached"`
	CreatedAt  time.Time `json:"createdAt"`
	DetachedAt time.Time `json:"detachedAt"` // zero while attached
}

type Session struct {
	ID            string
	UserID        string
	AuthSessionID string
	Kind          string
	Target        string
	Title         string
	Persistent    bool
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
		if ok && sess.Kind == opts.Kind && sess.Target == opts.Target && sess.Persistent == opts.Persistent && sess.HostID == opts.HostID {
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
	att, err := sess.Attach(ws, opts.Cols, opts.Rows)
	if err != nil {
		sess.Close()
		return err
	}
	att.Run()
	return nil
}

func Open(client *gossh.Client, opts SessionOptions) (*Session, error) {
	if err := validateSessionOptions(opts); err != nil {
		return nil, err
	}
	ring := newRingBuffer()
	backend, err := newSSHBackend(client, opts.Cols, opts.Rows, opts.InitCmd, ring)
	if err != nil {
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
		Persistent:    opts.Persistent,
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
	registerSession(s)
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
	shouldClose := clean || (!s.Persistent && !revalidate)
	if !shouldClose {
		timeout := graceTimeout
		if revalidate {
			timeout = revalidateGrace
		}
		s.grace = time.AfterFunc(timeout, s.Close)
	}
	s.mu.Unlock()
	if shouldClose {
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

func (s *Session) Info() Info {
	s.mu.Lock()
	defer s.mu.Unlock()
	return Info{
		ID:         s.ID,
		Kind:       s.Kind,
		Title:      s.Title,
		Persistent: s.Persistent,
		HostID:     s.HostID,
		Attached:   s.attached != nil,
		CreatedAt:  s.CreatedAt,
		DetachedAt: s.detachedAt,
	}
}

func (s *Session) resize(cols, rows int) {
	s.mu.Lock()
	s.cols, s.rows = cols, rows
	s.mu.Unlock()
	if err := s.backend.Resize(cols, rows); err != nil {
		global.LOG.Errorf("ssh pty change windows size failed, err: %v", err)
	}
}

func (s *Session) writeInput(data []byte) {
	if _, err := s.backend.Write(data); err != nil {
		global.LOG.Errorf("ws cmd bytes write to ssh.stdin pipe failed, err: %v", err)
	}
}

func (s *Session) ensureAIInterceptor() *aiInputInterceptor {
	s.mu.Lock()
	defer s.mu.Unlock()
	if v := terminalai.CurrentTerminalRuntimeVersion(); s.aiInterceptor == nil || s.aiVersion != v {
		s.aiVersion = v
		s.aiInterceptor = newAIInputInterceptor("", s.lang)
	}
	return s.aiInterceptor
}

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

func (s *Session) waitBackend() {
	_ = s.backend.Wait()
	s.flush()
	s.Close()
}

func cmdMessage(data []byte) []byte {
	msg, _ := json.Marshal(WsMsg{Type: WsMsgCmd, Data: base64.StdEncoding.EncodeToString(data)})
	return msg
}

func sendClose(ws *websocket.Conn, code int, reason string) {
	_ = ws.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(code, reason), time.Now().Add(time.Second))
}
