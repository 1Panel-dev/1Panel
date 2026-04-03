package terminal

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/i18n"
	"github.com/gorilla/websocket"
	"golang.org/x/crypto/ssh"
)

type safeBuffer struct {
	buffer bytes.Buffer
	mu     sync.Mutex
}

func (w *safeBuffer) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.buffer.Write(p)
}
func (w *safeBuffer) Bytes() []byte {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.buffer.Bytes()
}
func (w *safeBuffer) Reset() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.buffer.Reset()
}

const (
	WsMsgCmd       = "cmd"
	WsMsgResize    = "resize"
	WsMsgHeartbeat = "heartbeat"
	WsMsgAINotice  = "ai_notice"
)

type WsMsg struct {
	Type      string `json:"type"`
	Data      string `json:"data,omitempty"`      // WsMsgCmd
	Line      string `json:"line,omitempty"`      // WsMsgCmd
	Level     string `json:"level,omitempty"`     // WsMsgAINotice
	Message   string `json:"message,omitempty"`   // WsMsgAINotice
	Cols      int    `json:"cols,omitempty"`      // WsMsgResize
	Rows      int    `json:"rows,omitempty"`      // WsMsgResize
	Timestamp int    `json:"timestamp,omitempty"` // WsMsgHeartbeat
}

type LogicSshWsSession struct {
	stdinPipe     io.WriteCloser
	comboOutput   *safeBuffer
	logBuff       *safeBuffer
	session       *ssh.Session
	wsConn        *websocket.Conn
	writeMutex    sync.Mutex
	lang          string
	isAdmin       bool
	IsFlagged     bool
	aiInterceptor *aiInputInterceptor
}

func NewLogicSshWsSession(cols, rows int, sshClient *ssh.Client, wsConn *websocket.Conn, initCmd string) (*LogicSshWsSession, error) {
	sshSession, err := sshClient.NewSession()
	if err != nil {
		return nil, err
	}

	stdinP, err := sshSession.StdinPipe()
	if err != nil {
		return nil, err
	}

	comboWriter := new(safeBuffer)
	logBuf := new(safeBuffer)
	sshSession.Stdout = comboWriter
	sshSession.Stderr = comboWriter

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}
	if err := sshSession.RequestPty("xterm", rows, cols, modes); err != nil {
		return nil, err
	}
	if err := sshSession.Shell(); err != nil {
		return nil, err
	}
	if len(initCmd) != 0 {
		time.Sleep(100 * time.Millisecond)
		_, _ = stdinP.Write([]byte(initCmd + "\n"))
	}
	lang := i18n.GetLanguageFromDB()
	return &LogicSshWsSession{
		stdinPipe:     stdinP,
		comboOutput:   comboWriter,
		logBuff:       logBuf,
		session:       sshSession,
		wsConn:        wsConn,
		lang:          lang,
		isAdmin:       true,
		IsFlagged:     false,
		aiInterceptor: newAIInputInterceptor("", lang),
	}, nil
}

func (sws *LogicSshWsSession) Close() {
	if sws.session != nil {
		sws.session.Close()
	}
	if sws.logBuff != nil {
		sws.logBuff = nil
	}
	if sws.comboOutput != nil {
		sws.comboOutput = nil
	}
}

func (sws *LogicSshWsSession) Start(quitChan chan bool) {
	go sws.receiveWsMsg(quitChan)
	go sws.sendComboOutput(quitChan)
}

func (sws *LogicSshWsSession) receiveWsMsg(exitCh chan bool) {
	defer func() {
		if r := recover(); r != nil {
			global.LOG.Errorf("[A panic occurred during receive ws message, error message: %v", r)
		}
	}()
	wsConn := sws.wsConn
	defer setQuit(exitCh)
	for {
		select {
		case <-exitCh:
			return
		default:
			_, wsData, err := wsConn.ReadMessage()
			if err != nil {
				return
			}
			msgObj := WsMsg{}
			_ = json.Unmarshal(wsData, &msgObj)
			switch msgObj.Type {
			case WsMsgResize:
				if msgObj.Cols > 0 && msgObj.Rows > 0 {
					if err := sws.session.WindowChange(msgObj.Rows, msgObj.Cols); err != nil {
						global.LOG.Errorf("ssh pty change windows size failed, err: %v", err)
					}
				}
			case WsMsgCmd:
				decodeBytes, err := base64.StdEncoding.DecodeString(msgObj.Data)
				if err != nil {
					global.LOG.Errorf("websock cmd string base64 decoding failed, err: %v", err)
				}
				if isEnterInput(decodeBytes) {
					if sws.aiInterceptor != nil {
						sws.aiInterceptor.SetCurrentLine(msgObj.Line)
					}
					if generated, handled := sws.aiInterceptor.HandleEnter(sws.notifyAIThinking, sws.notifyAIDone, sws.notifyAIError); handled {
						payload := []byte{lineClearControl}
						if strings.TrimSpace(generated) != "" {
							payload = append(payload, []byte(generated)...)
						}
						sws.sendWebsocketInputCommandToSshSessionStdinPipe(payload)
						continue
					}
				}
				sws.sendWebsocketInputCommandToSshSessionStdinPipe(decodeBytes)
			case WsMsgHeartbeat:
				err = sws.writeWSMessage(websocket.TextMessage, wsData)
				if err != nil {
					global.LOG.Errorf("ssh sending heartbeat to webSocket failed, err: %v", err)
				}
			}
		}
	}
}

func (sws *LogicSshWsSession) notifyAIThinking() {
	if sws == nil {
		return
	}
	if err := sws.writeAINotice("info", i18n.GetMsgByKeyAndLang(sws.lang, "TerminalAIThinking")); err != nil {
		global.LOG.Errorf("write terminal ai thinking message failed, err: %v", err)
	}
}

func (sws *LogicSshWsSession) notifyAIDone(message string) {
	if sws == nil || strings.TrimSpace(message) == "" {
		return
	}
	if err := sws.writeAINotice("success", message); err != nil {
		global.LOG.Errorf("write terminal ai done message failed, err: %v", err)
	}
}

func (sws *LogicSshWsSession) notifyAIError(message string) {
	if sws == nil || strings.TrimSpace(message) == "" {
		return
	}
	if err := sws.writeAINotice("error", message); err != nil {
		global.LOG.Errorf("write terminal ai error message failed, err: %v", err)
	}
}

func (sws *LogicSshWsSession) sendWebsocketInputCommandToSshSessionStdinPipe(cmdBytes []byte) {
	if _, err := sws.stdinPipe.Write(cmdBytes); err != nil {
		global.LOG.Errorf("ws cmd bytes write to ssh.stdin pipe failed, err: %v", err)
	}
}

func (sws *LogicSshWsSession) writeAINotice(level, message string) error {
	if sws == nil || strings.TrimSpace(message) == "" {
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
	return sws.writeWSMessage(websocket.TextMessage, wsData)
}

func (sws *LogicSshWsSession) writeWSMessage(messageType int, data []byte) error {
	sws.writeMutex.Lock()
	defer sws.writeMutex.Unlock()
	return sws.wsConn.WriteMessage(messageType, data)
}

func (sws *LogicSshWsSession) sendComboOutput(exitCh chan bool) {
	defer setQuit(exitCh)

	tick := time.NewTicker(time.Millisecond * time.Duration(60))
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			if sws.comboOutput == nil {
				return
			}
			bs := sws.comboOutput.Bytes()
			if len(bs) > 0 {
				wsData, err := json.Marshal(WsMsg{
					Type: WsMsgCmd,
					Data: base64.StdEncoding.EncodeToString(bs),
				})
				if err != nil {
					global.LOG.Errorf("encoding combo output to json failed, err: %v", err)
					continue
				}
				err = sws.writeWSMessage(websocket.TextMessage, wsData)
				if err != nil {
					global.LOG.Errorf("ssh sending combo output to webSocket failed, err: %v", err)
				}
				_, err = sws.logBuff.Write(bs)
				if err != nil {
					global.LOG.Errorf("combo output to log buffer failed, err: %v", err)
				}
				sws.comboOutput.buffer.Reset()
			}
			if string(bs) == string([]byte{13, 10, 108, 111, 103, 111, 117, 116, 13, 10}) {
				sws.Close()
				return
			}

		case <-exitCh:
			return
		}
	}
}

func (sws *LogicSshWsSession) Wait(quitChan chan bool) {
	if err := sws.session.Wait(); err != nil {
		setQuit(quitChan)
	}
}

func setQuit(ch chan bool) {
	ch <- true
}
