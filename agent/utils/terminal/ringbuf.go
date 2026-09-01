package terminal

import (
	"bytes"
	"sync"
)

// defaultRingSize is the default capacity of a session output ring buffer.
const defaultRingSize = 256 * 1024

// ringBuffer is a fixed-capacity byte ring; writes drop the oldest bytes when full.
type ringBuffer struct {
	mu      sync.Mutex
	buf     []byte
	start   int  // index of the oldest byte
	size    int  // number of bytes currently stored
	wrapped bool // true once at least one byte has been dropped
}

func newRingBuffer(capacity int) *ringBuffer {
	if capacity <= 0 {
		capacity = defaultRingSize
	}
	return &ringBuffer{buf: make([]byte, capacity)}
}

// Write appends p, dropping oldest bytes when full. Always reports len(p).
func (r *ringBuffer) Write(p []byte) (int, error) {
	n := len(p)
	if n == 0 {
		return 0, nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	capacity := len(r.buf)
	if n >= capacity {
		if n > capacity || r.size > 0 {
			r.wrapped = true
		}
		copy(r.buf, p[n-capacity:])
		r.start = 0
		r.size = capacity
		return n, nil
	}

	end := (r.start + r.size) % capacity
	written := copy(r.buf[end:], p)
	if written < n {
		copy(r.buf, p[written:])
	}
	if r.size+n > capacity {
		r.wrapped = true
		r.start = (r.start + (r.size + n - capacity)) % capacity
		r.size = capacity
	} else {
		r.size += n
	}
	return n, nil
}

// Snapshot returns buffered bytes, oldest first.
// After overflow, starts after the first '\n' so replay is not mid-escape or mid-rune.
func (r *ringBuffer) Snapshot() []byte {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.size == 0 {
		return nil
	}
	out := make([]byte, r.size)
	n := copy(out, r.buf[r.start:min(r.start+r.size, len(r.buf))])
	if n < r.size {
		copy(out[n:], r.buf[:r.size-n])
	}
	if !r.wrapped {
		return out
	}
	if idx := bytes.IndexByte(out, '\n'); idx >= 0 {
		return out[idx+1:]
	}
	return out
}

// Len returns the number of buffered bytes.
func (r *ringBuffer) Len() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.size
}

// Reset drops every buffered byte.
func (r *ringBuffer) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.start = 0
	r.size = 0
	r.wrapped = false
}
