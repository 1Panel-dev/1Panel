package terminal

import (
	"bytes"
	"sync"
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

// Take atomically copies and clears the buffer so concurrent writes are not lost.
func (w *safeBuffer) Take() []byte {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.buffer.Len() == 0 {
		return nil
	}
	out := bytes.Clone(w.buffer.Bytes())
	w.buffer.Reset()
	return out
}

const (
	WsMsgCmd       = "cmd"
	WsMsgResize    = "resize"
	WsMsgHeartbeat = "heartbeat"
	WsMsgAINotice  = "ai_notice"
	WsMsgSession   = "session"
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
	ID        string `json:"id,omitempty"`        // WsMsgSession
	Pinned    *bool  `json:"pinned,omitempty"`    // WsMsgSession
}

func setQuit(ch chan bool) {
	ch <- true
}
