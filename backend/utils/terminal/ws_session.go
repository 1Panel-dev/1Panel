package terminal

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"sync"
	"time"

	"github.com/1Panel-dev/1Panel/backend/global"
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
	wsMsgCmd    = "cmd"
	wsMsgResize = "resize"
	wsMsgClose  = "close"
)

type wsMsg struct {
	Type string `json:"type"`
	Cmd  string `json:"cmd"`
	Cols int    `json:"cols"`
	Rows int    `json:"rows"`
}

type LogicSshWsSession struct {
	stdinPipe       io.WriteCloser
	comboOutput     *safeBuffer
	logBuff         *safeBuffer
	inputFilterBuff *safeBuffer
	session         *ssh.Session
	wsConn          *websocket.Conn
	isAdmin         bool
	IsFlagged       bool
}

func NewLogicSshWsSession(cols, rows int, isAdmin bool, sshClient *ssh.Client, wsConn *websocket.Conn) (*LogicSshWsSession, error) {
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
	inputBuf := new(safeBuffer)
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
	return &LogicSshWsSession{
		stdinPipe:       stdinP,
		comboOutput:     comboWriter,
		logBuff:         logBuf,
		inputFilterBuff: inputBuf,
		session:         sshSession,
		wsConn:          wsConn,
		isAdmin:         isAdmin,
		IsFlagged:       false,
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
			msgObj := wsMsg{}
			_ = json.Unmarshal(wsData, &msgObj)
			switch msgObj.Type {
			case wsMsgResize:
				if msgObj.Cols > 0 && msgObj.Rows > 0 {
					if err := sws.session.WindowChange(msgObj.Rows, msgObj.Cols); err != nil {
						global.LOG.Errorf("ssh pty change windows size failed, err: %v", err)
					}
				}
			case wsMsgCmd:
				decodeBytes, err := base64.StdEncoding.DecodeString(msgObj.Cmd)
				if err != nil {
					global.LOG.Errorf("websock cmd string base64 decoding failed, err: %v", err)
				}
				sws.sendWebsocketInputCommandToSshSessionStdinPipe(decodeBytes)
			}
		}
	}
}

func (sws *LogicSshWsSession) sendWebsocketInputCommandToSshSessionStdinPipe(cmdBytes []byte) {
	if _, err := sws.stdinPipe.Write(cmdBytes); err != nil {
		global.LOG.Errorf("ws cmd bytes write to ssh.stdin pipe failed, err: %v", err)
	}
}

func (sws *LogicSshWsSession) sendComboOutput(exitCh chan bool) {
	wsConn := sws.wsConn
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
				err := wsConn.WriteMessage(websocket.TextMessage, bs)
				if err != nil {
					global.LOG.Errorf("ssh sending combo output to webSocket failed, err: %v", err)
				}
				_, err = sws.logBuff.Write(bs)
				if err != nil {
					global.LOG.Errorf("combo output to log buffer failed, err: %v", err)
				}
				sws.comboOutput.buffer.Reset()
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
