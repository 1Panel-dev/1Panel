package websocket

import (
	"sync/atomic"

	"github.com/gorilla/websocket"
)

const MaxMessageQuenue = 32

type Client struct {
	ID     string
	Socket *websocket.Conn
	Msg    chan []byte
	closed atomic.Bool
}

func NewWsClient(ID string, socket *websocket.Conn) *Client {
	return &Client{
		ID:     ID,
		Socket: socket,
		Msg:    make(chan []byte, 32),
		closed: atomic.Bool{},
	}
}

func (c *Client) Read() {
	defer func() {
		c.Close()
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

func (c *Client) SendPayload(res []byte) {
	if c.closed.Load() {
		return
	}
	select {
	case c.Msg <- res:
		return
	default:
		select {
		case <-c.Msg:
		default:
		}
		c.Msg <- res
	}
}
