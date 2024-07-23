package websocket

import (
	"github.com/gorilla/websocket"
)

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
		ProcessData(c, message)
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
