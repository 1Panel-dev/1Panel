package terminal

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/i18n"
	"github.com/gorilla/websocket"
)

// Half-open detection uses protocol level ping/pong, which browsers answer
// without JavaScript, so background tab timer throttling cannot trip it.
const (
	pingInterval = 30 * time.Second
	pongWait     = 75 * time.Second
	writeWait    = 5 * time.Second
)

var errAttachmentClosed = errors.New("terminal attachment is closed")

// attachment is one websocket connection bound to a Session.
type attachment struct {
	sess *Session
	ws   *websocket.Conn

	writeMu sync.Mutex
	cursor  uint64 // ring offset of the next byte to send; guarded by writeMu

	done      chan struct{}
	closeOnce sync.Once
}

// Run reads client messages until the websocket fails or this attachment is closed.
// On return the session is detached: cleanly if the client sent close code 1000.
func (a *attachment) Run() {
	clean := false
	defer func() {
		if r := recover(); r != nil {
			global.LOG.Errorf("[A panic occurred during receive ws message, error message: %v", r)
		}
		a.close(websocket.CloseNormalClosure, "")
		a.sess.detach(a, clean)
	}()

	_ = a.ws.SetReadDeadline(time.Now().Add(pongWait))
	a.ws.SetPongHandler(func(string) error {
		return a.ws.SetReadDeadline(time.Now().Add(pongWait))
	})
	go a.pingLoop()

	// close() shuts the websocket, which is what ends this loop.
	for {
		_, wsData, err := a.ws.ReadMessage()
		if err != nil {
			clean = websocket.IsCloseError(err, websocket.CloseNormalClosure)
			return
		}
		_ = a.ws.SetReadDeadline(time.Now().Add(pongWait))
		msgObj := WsMsg{}
		_ = json.Unmarshal(wsData, &msgObj)
		switch msgObj.Type {
		case WsMsgResize:
			if msgObj.Cols > 0 && msgObj.Rows > 0 {
				a.sess.resize(msgObj.Cols, msgObj.Rows)
			}
		case WsMsgCmd:
			decodeBytes, err := base64.StdEncoding.DecodeString(msgObj.Data)
			if err != nil {
				global.LOG.Errorf("websock cmd string base64 decoding failed, err: %v", err)
			}
			if isEnterInput(decodeBytes) {
				interceptor := a.sess.ensureAIInterceptor()
				if interceptor != nil {
					interceptor.SetCurrentLine(msgObj.Line)
				}
				if generated, handled := interceptor.HandleEnter(a.notifyAIThinking, a.notifyAIDone, a.notifyAIError); handled {
					if payload, err := buildAIPastePayload(generated); err != nil {
						global.LOG.Errorf("ai generated command rejected before ssh.stdin pipe write, err: %v", err)
					} else {
						a.sess.writeInput(payload)
					}
					continue
				}
			}
			a.sess.writeInput(decodeBytes)
		case WsMsgHeartbeat:
			if err := a.write(wsData); err != nil {
				global.LOG.Errorf("ssh sending heartbeat to webSocket failed, err: %v", err)
			}
		}
	}
}

// pingLoop keeps the read deadline honest; a ping that cannot be sent ends the attachment.
func (a *attachment) pingLoop() {
	tick := time.NewTicker(pingInterval)
	defer tick.Stop()
	for {
		select {
		case <-a.done:
			return
		case <-tick.C:
			if err := a.ws.WriteControl(websocket.PingMessage, nil, time.Now().Add(writeWait)); err != nil {
				a.close(websocket.CloseInternalServerErr, "ping failed")
				return
			}
		}
	}
}

// write sends one text message, serialized against every other writer.
func (a *attachment) write(data []byte) error {
	a.writeMu.Lock()
	defer a.writeMu.Unlock()
	return a.writeLocked(data)
}

// writeLocked sends one text message; the caller owns writeMu.
func (a *attachment) writeLocked(data []byte) error {
	select {
	case <-a.done:
		return errAttachmentClosed
	default:
	}
	_ = a.ws.SetWriteDeadline(time.Now().Add(writeWait))
	return a.ws.WriteMessage(websocket.TextMessage, data)
}

// close sends a close frame and tears the websocket down. Idempotent.
func (a *attachment) close(code int, reason string) {
	a.closeOnce.Do(func() {
		defer func() {
			if r := recover(); r != nil {
				global.LOG.Errorf("a panic occurred during close ws attachment, error message: %v", r)
			}
		}()
		close(a.done)
		sendClose(a.ws, code, reason)
		_ = a.ws.Close()
	})
}

func (a *attachment) notifyAIThinking() {
	if err := a.writeAINotice("info", i18n.GetMsgByKeyAndLang(a.sess.lang, "TerminalAIThinking")); err != nil {
		global.LOG.Errorf("write terminal ai thinking message failed, err: %v", err)
	}
}

func (a *attachment) notifyAIDone(message string) {
	if err := a.writeAINotice("success", message); err != nil {
		global.LOG.Errorf("write terminal ai done message failed, err: %v", err)
	}
}

func (a *attachment) notifyAIError(message string) {
	if err := a.writeAINotice("error", message); err != nil {
		global.LOG.Errorf("write terminal ai error message failed, err: %v", err)
	}
}

func (a *attachment) writeAINotice(level, message string) error {
	if strings.TrimSpace(message) == "" {
		return nil
	}
	wsData, err := json.Marshal(WsMsg{
		Type:    WsMsgAINotice,
		Level:   strings.TrimSpace(level),
		Message: strings.TrimSpace(message),
	})
	if err != nil {
		return err
	}
	return a.write(wsData)
}
