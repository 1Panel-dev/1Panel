package websocket

import (
	"sync"
	"sync/atomic"

	"github.com/gorilla/websocket"
)

const MaxMessageQuenue = 32

type Client struct {
	ID        string
	Socket    *websocket.Conn
	Msg       chan []byte
	closed    atomic.Bool
	closeOnce sync.Once
}

func NewWsClient(ID string, socket *websocket.Conn) *Client {
	return &Client{
		ID:     ID,
		Socket: socket,
		Msg:    make(chan []byte, MaxMessageQuenue),
	}
}

func (c *Client) Close() {
	c.closeOnce.Do(func() {
		c.closed.Store(true)
		close(c.Msg)
		c.Socket.Close()
	})
}

func (c *Client) Read() {
	defer c.Close()
	for {
		_, message, err := c.Socket.ReadMessage()
		if err != nil {
			return
		}
		ProcessData(c, message)
	}
}

func (c *Client) Write() {
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
	default:
	}
}
