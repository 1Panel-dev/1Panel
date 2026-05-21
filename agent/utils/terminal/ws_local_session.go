package terminal

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"sync"

	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/i18n"
	terminalai "github.com/1Panel-dev/1Panel/agent/utils/terminal/ai"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
)

type LocalWsSession struct {
	slave  *LocalCommand
	wsConn *websocket.Conn

	allowCtrlC    bool
	writeMutex    sync.Mutex
	lang          string
	aiInterceptor *aiInputInterceptor
	aiVersion     uint64
}

func NewLocalWsSession(cols, rows int, wsConn *websocket.Conn, slave *LocalCommand, allowCtrlC bool) (*LocalWsSession, error) {
	if err := slave.ResizeTerminal(cols, rows); err != nil {
		global.LOG.Errorf("ssh pty change windows size failed, err: %v", err)
	}
	lang := i18n.GetLanguageFromDB()

	return &LocalWsSession{
		slave:  slave,
		wsConn: wsConn,

		allowCtrlC:    allowCtrlC,
		lang:          lang,
		aiInterceptor: newAIInputInterceptor("", lang),
		aiVersion:     terminalai.CurrentTerminalRuntimeVersion(),
	}, nil
}

func (sws *LocalWsSession) Start(quitChan chan bool) {
	go sws.handleSlaveEvent(quitChan)
	go sws.receiveWsMsg(quitChan)
}

func (sws *LocalWsSession) handleSlaveEvent(exitCh chan bool) {
	defer setQuit(exitCh)
	defer global.LOG.Debug("thread of handle slave event has exited now")

	buffer := make([]byte, 1024)
	for {
		select {
		case <-exitCh:
			return
		default:
			n, _ := sws.slave.Read(buffer)
			_ = sws.masterWrite(buffer[:n])
		}
	}
}

func (sws *LocalWsSession) masterWrite(data []byte) error {
	defer func() {
		if r := recover(); r != nil {
			global.LOG.Errorf("A panic occurred during write ws message to master, error message: %v", r)
		}
	}()
	sws.writeMutex.Lock()
	defer sws.writeMutex.Unlock()
	wsData, err := json.Marshal(WsMsg{
		Type: WsMsgCmd,
		Data: base64.StdEncoding.EncodeToString(data),
	})
	if err != nil {
		return errors.Wrapf(err, "failed to encoding to json")
	}
	err = sws.wsConn.WriteMessage(websocket.TextMessage, wsData)
	if err != nil {
		return errors.Wrapf(err, "failed to write to master")
	}
	return nil
}

func (sws *LocalWsSession) receiveWsMsg(exitCh chan bool) {
	defer func() {
		if r := recover(); r != nil {
			setQuit(exitCh)
			global.LOG.Errorf("A panic occurred during receive ws message, error message: %v", r)
		}
	}()
	wsConn := sws.wsConn
	defer setQuit(exitCh)
	defer global.LOG.Debug("thread of receive ws msg has exited now")
	for {
		select {
		case <-exitCh:
			return
		default:
			_, wsData, err := wsConn.ReadMessage()
			if err != nil {
				global.LOG.Errorf("reading webSocket message failed, err: %v", err)
				return
			}
			msgObj := WsMsg{}
			_ = json.Unmarshal(wsData, &msgObj)
			switch msgObj.Type {
			case WsMsgResize:
				if msgObj.Cols > 0 && msgObj.Rows > 0 {
					if err := sws.slave.ResizeTerminal(msgObj.Cols, msgObj.Rows); err != nil {
						global.LOG.Errorf("ssh pty change windows size failed, err: %v", err)
					}
				}
			case WsMsgCmd:
				decodeBytes, err := base64.StdEncoding.DecodeString(msgObj.Data)
				if err != nil {
					global.LOG.Errorf("websock cmd string base64 decoding failed, err: %v", err)
				}
				if isEnterInput(decodeBytes) {
					sws.ensureAIInterceptor()
					if sws.aiInterceptor != nil {
						sws.aiInterceptor.SetCurrentLine(msgObj.Line)
					}
					if generated, handled := sws.aiInterceptor.HandleEnter(sws.notifyAIThinking, sws.notifyAIDone, sws.notifyAIError); handled {
						sws.sendWebsocketInputCommandToSshSessionStdinPipe(buildAIPastePayload(generated))
						continue
					}
				}
				sws.sendWebsocketInputCommandToSshSessionStdinPipe(decodeBytes)
			case WsMsgHeartbeat:
				sws.writeMutex.Lock()
				err = wsConn.WriteMessage(websocket.TextMessage, wsData)
				sws.writeMutex.Unlock()
				if err != nil {
					global.LOG.Errorf("ssh sending heartbeat to webSocket failed, err: %v", err)
				}
			}
		}
	}
}

func (sws *LocalWsSession) ensureAIInterceptor() {
	if sws == nil || sws.aiInterceptor != nil {
		return
	}
	currentVersion := terminalai.CurrentTerminalRuntimeVersion()
	if sws.aiVersion == currentVersion {
		return
	}
	sws.aiVersion = currentVersion
	sws.aiInterceptor = newAIInputInterceptor("", sws.lang)
}

func (sws *LocalWsSession) notifyAIThinking() {
	if sws == nil {
		return
	}
	if err := sws.writeAINotice("info", i18n.GetMsgByKeyAndLang(sws.lang, "TerminalAIThinking")); err != nil {
		global.LOG.Errorf("write terminal ai thinking message failed, err: %v", err)
	}
}

func (sws *LocalWsSession) notifyAIDone(message string) {
	if sws == nil || strings.TrimSpace(message) == "" {
		return
	}
	if err := sws.writeAINotice("success", message); err != nil {
		global.LOG.Errorf("write terminal ai done message failed, err: %v", err)
	}
}

func (sws *LocalWsSession) notifyAIError(message string) {
	if sws == nil || strings.TrimSpace(message) == "" {
		return
	}
	if err := sws.writeAINotice("error", message); err != nil {
		global.LOG.Errorf("write terminal ai error message failed, err: %v", err)
	}
}

func (sws *LocalWsSession) writeAINotice(level, message string) error {
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
	sws.writeMutex.Lock()
	defer sws.writeMutex.Unlock()
	return sws.wsConn.WriteMessage(websocket.TextMessage, wsData)
}

func (sws *LocalWsSession) sendWebsocketInputCommandToSshSessionStdinPipe(cmdBytes []byte) {
	if _, err := sws.slave.Write(cmdBytes); err != nil {
		global.LOG.Errorf("ws cmd bytes write to ssh.stdin pipe failed, err: %v", err)
	}
}

func buildAIPastePayload(generated string) []byte {
	payload := []byte{lineClearControl}
	generated = strings.TrimSpace(generated)
	if generated == "" {
		return payload
	}
	payload = append(payload, []byte("\x1b[200~")...)
	payload = append(payload, []byte(generated)...)
	payload = append(payload, []byte("\x1b[201~")...)
	return payload
}
