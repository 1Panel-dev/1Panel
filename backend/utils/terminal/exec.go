package terminal

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"sync"

	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
)

type ExecWsSession struct {
	conn   net.Conn
	wsConn *websocket.Conn

	writeMutex sync.Mutex
}

func NewExecConn(cols, rows int, wsConn *websocket.Conn, hijacked net.Conn, commands ...string) (*ExecWsSession, error) {
	_, _ = hijacked.Write([]byte(fmt.Sprintf("stty cols %d rows %d && clear \r", cols, rows)))
	for _, command := range commands {
		_, _ = hijacked.Write([]byte(fmt.Sprintf("%s \r", command)))
	}

	return &ExecWsSession{
		conn:   hijacked,
		wsConn: wsConn,
	}, nil
}

func (sws *ExecWsSession) Start(ctx context.Context, quitChan chan bool) {
	go sws.handleSlaveEvent(ctx, quitChan)
	go sws.receiveWsMsg(ctx, quitChan)
}

func (sws *ExecWsSession) handleSlaveEvent(ctx context.Context, exitCh chan bool) {
	defer setQuit(exitCh)

	buffer := make([]byte, 1024)
	for {
		n, err := sws.conn.Read(buffer)
		if err != nil && errors.Is(err, net.ErrClosed) {
			return
		}

		if err := sws.masterWrite(buffer[:n]); err != nil {
			if errors.Is(err, websocket.ErrCloseSent) {
				return
			}
		}
	}
}

func (sws *ExecWsSession) masterWrite(data []byte) error {
	sws.writeMutex.Lock()
	defer sws.writeMutex.Unlock()
	err := sws.wsConn.WriteMessage(websocket.TextMessage, data)
	if err != nil {
		return errors.Wrapf(err, "failed to write to master")
	}

	return nil
}

func (sws *ExecWsSession) receiveWsMsg(ctx context.Context, exitCh chan bool) {
	wsConn := sws.wsConn
	defer setQuit(exitCh)
	for {
		_, wsData, err := wsConn.ReadMessage()
		if err != nil {
			return
		}
		msgObj := wsMsg{}
		_ = json.Unmarshal(wsData, &msgObj)
		switch msgObj.Type {
		case wsMsgResize:
			if msgObj.Cols > 0 && msgObj.Rows > 0 {
				sws.ResizeTerminal(msgObj.Rows, msgObj.Cols)
			}
		case wsMsgCmd:
			decodeBytes, err := base64.StdEncoding.DecodeString(msgObj.Cmd)
			if err != nil {
				global.LOG.Errorf("websock cmd string base64 decoding failed, err: %v", err)
				return
			}
			sws.sendWebsocketInputCommandToSshSessionStdinPipe(decodeBytes)
		case wsMsgClose:
			_, _ = sws.conn.Write([]byte("exit\r"))
			return
		}
	}
}

func (sws *ExecWsSession) sendWebsocketInputCommandToSshSessionStdinPipe(cmdBytes []byte) {
	_, _ = sws.conn.Write(cmdBytes)
}

func (sws *ExecWsSession) ResizeTerminal(rows int, cols int) {
	_, _ = sws.conn.Write([]byte(fmt.Sprintf("stty cols %d rows %d && clear \r", cols, rows)))
}
