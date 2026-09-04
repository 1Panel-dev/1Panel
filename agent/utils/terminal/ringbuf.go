package terminal

import (
	"bytes"
	"sync"
)

// ringSize is the output retained per session. Reattach replays at most this much.
// ponytail: fixed; make it a setting if someone asks for a bigger tail.
const ringSize = 128 * 1024

// ringBuffer is a fixed capacity byte ring addressed by absolute write offset.
// Write never blocks and overwrites the oldest bytes; readers that fall behind
// skip ahead and are told they lost data.
type ringBuffer struct {
	mu      sync.Mutex
	buf     []byte
	written uint64 // total bytes ever written; offset of the next byte
}

func newRingBuffer() *ringBuffer {
	return &ringBuffer{buf: make([]byte, ringSize)}
}

// Write appends p, dropping the oldest bytes when full. Always reports len(p).
func (r *ringBuffer) Write(p []byte) (int, error) {
	n := len(p)
	if n == 0 {
		return 0, nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	capacity := len(r.buf)
	if n > capacity {
		// only the tail can survive; account for the skipped bytes so offsets stay absolute
		r.written += uint64(n - capacity)
		p = p[n-capacity:]
	}
	pos := int(r.written % uint64(capacity))
	k := copy(r.buf[pos:], p)
	copy(r.buf, p[k:])
	r.written += uint64(len(p))
	return n, nil
}

// Oldest is the offset of the oldest byte still retained.
func (r *ringBuffer) Oldest() uint64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.oldestLocked()
}

func (r *ringBuffer) oldestLocked() uint64 {
	if capacity := uint64(len(r.buf)); r.written > capacity {
		return r.written - capacity
	}
	return 0
}

// ReadFrom returns every byte from offset onward and the offset to continue from.
// If offset was already overwritten, reading starts at the oldest retained byte
// and lost is true. Whenever the start is not the true beginning of output the
// result is aligned to the next '\n' so replay never begins mid escape sequence.
func (r *ringBuffer) ReadFrom(offset uint64) (data []byte, next uint64, lost bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	oldest := r.oldestLocked()
	if offset > r.written {
		offset = r.written
	}
	if offset < oldest {
		offset = oldest
		lost = true
	}
	n := int(r.written - offset)
	if n == 0 {
		return nil, r.written, lost
	}
	out := make([]byte, n)
	start := int(offset % uint64(len(r.buf)))
	k := copy(out, r.buf[start:min(start+n, len(r.buf))])
	if k < n {
		copy(out[k:], r.buf[:n-k])
	}
	if offset == oldest && oldest > 0 {
		if idx := bytes.IndexByte(out, '\n'); idx >= 0 && idx+1 < len(out) {
			out = out[idx+1:]
		}
	}
	return out, r.written, lost
}
