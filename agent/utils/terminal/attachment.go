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

var errAttachmentClosed = errors.New("terminal attachment is closed")

// attachment is one websocket connection bound to a Session.
type attachment struct {
	sess      *Session
	ws        *websocket.Conn
	writeMu   sync.Mutex
	done      chan struct{}
	closeOnce sync.Once
}

// Run reads client messages until the websocket fails or this attachment is closed.
func (a *attachment) Run() {
	defer func() {
		if r := recover(); r != nil {
			global.LOG.Errorf("[A panic occurred during receive ws message, error message: %v", r)
		}
		a.close(websocket.CloseNormalClosure, "")
		a.sess.onAttachmentClosed(a)
	}()

	for {
		select {
		case <-a.done:
			return
		case <-a.sess.done:
			return
		default:
		}
		_, wsData, err := a.ws.ReadMessage()
		if err != nil {
			return
		}
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
			if err := a.write(websocket.TextMessage, wsData); err != nil {
				global.LOG.Errorf("ssh sending heartbeat to webSocket failed, err: %v", err)
			}
		}
	}
}

// write sends one websocket message, serialized against every other writer.
func (a *attachment) write(msgType int, data []byte) error {
	a.writeMu.Lock()
	defer a.writeMu.Unlock()
	return a.writeLocked(msgType, data)
}

// writeLocked sends one websocket message, the caller owns writeMu.
func (a *attachment) writeLocked(msgType int, data []byte) error {
	select {
	case <-a.done:
		return errAttachmentClosed
	default:
	}
	return a.ws.WriteMessage(msgType, data)
}

// close sends a close frame and tears the websocket down. It is idempotent.
func (a *attachment) close(code int, reason string) {
	a.closeOnce.Do(func() {
		defer func() {
			if r := recover(); r != nil {
				global.LOG.Errorf("a panic occurred during close ws attachment, error message: %v", r)
			}
		}()
		close(a.done)
		a.writeMu.Lock()
		_ = a.ws.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(code, reason), time.Now().Add(time.Second))
		a.writeMu.Unlock()
		_ = a.ws.Close()
	})
}

func (a *attachment) notifyAIThinking() {
	if a == nil {
		return
	}
	if err := a.writeAINotice("info", i18n.GetMsgByKeyAndLang(a.sess.lang, "TerminalAIThinking")); err != nil {
		global.LOG.Errorf("write terminal ai thinking message failed, err: %v", err)
	}
}

func (a *attachment) notifyAIDone(message string) {
	if a == nil || strings.TrimSpace(message) == "" {
		return
	}
	if err := a.writeAINotice("success", message); err != nil {
		global.LOG.Errorf("write terminal ai done message failed, err: %v", err)
	}
}

func (a *attachment) notifyAIError(message string) {
	if a == nil || strings.TrimSpace(message) == "" {
		return
	}
	if err := a.writeAINotice("error", message); err != nil {
		global.LOG.Errorf("write terminal ai error message failed, err: %v", err)
	}
}

func (a *attachment) writeAINotice(level, message string) error {
	if a == nil || strings.TrimSpace(message) == "" {
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
	return a.write(websocket.TextMessage, wsData)
}
