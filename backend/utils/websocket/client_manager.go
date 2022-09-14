package websocket

import "sync"

type Manager struct {
	Group                map[string]*Client
	Lock                 sync.Mutex
	Register, UnRegister chan *Client
	ClientCount          uint
}

func (m *Manager) Start() {
	for {
		select {
		case client := <-m.Register:
			m.Lock.Lock()
			m.Group[client.ID] = client
			m.ClientCount++
			m.Lock.Unlock()
		case client := <-m.UnRegister:
			m.Lock.Lock()
			if _, ok := m.Group[client.ID]; ok {
				close(client.Msg)
				delete(m.Group, client.ID)
				m.ClientCount--
			}
			m.Lock.Unlock()
		}
	}
}

func (m *Manager) RegisterClient(client *Client) {
	m.Register <- client
}

func (m *Manager) UnRegisterClient(client *Client) {
	m.UnRegister <- client
}
