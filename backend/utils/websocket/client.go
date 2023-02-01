package websocket

import (
	"encoding/json"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/files"
	"github.com/gorilla/websocket"
)

type WsMsg struct {
	Type string
	Keys []string
}

type Client struct {
	ID     string
	Socket *websocket.Conn
	Msg    chan []byte
}

func NewWsClient(ID string, socket *websocket.Conn) *Client {
	return &Client{
		ID:     ID,
		Socket: socket,
		Msg:    make(chan []byte, 100),
	}
}

func (c *Client) Read() {
	defer func() {
		close(c.Msg)
	}()

	for {
		_, message, err := c.Socket.ReadMessage()
		if err != nil {
			return
		}
		msg := &WsMsg{}
		_ = json.Unmarshal(message, msg)
		ProcessData(c, msg)
	}
}

func (c *Client) Write() {
	defer func() {
		c.Socket.Close()
	}()

	for {
		message, ok := <-c.Msg
		if !ok {
			return
		}
		_ = c.Socket.WriteMessage(websocket.TextMessage, message)
	}
}

func ProcessData(c *Client, msg *WsMsg) {

	if msg.Type == "wget" {
		var res []files.Process
		for _, k := range msg.Keys {
			value, err := global.CACHE.Get(k)
			if err != nil {
				global.LOG.Errorf("get cache error,err %s", err.Error())
				return
			}

			process := &files.Process{}
			_ = json.Unmarshal(value, process)
			res = append(res, *process)
		}
		reByte, _ := json.Marshal(res)
		c.Msg <- reByte
	}
}
