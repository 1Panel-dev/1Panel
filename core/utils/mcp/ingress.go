package mcp

import (
	"sync"
	"time"
)

var ingress struct {
	sync.Mutex
	start    int64
	requests int
}

func AdmitIngress(now time.Time) error {
	ingress.Lock()
	defer ingress.Unlock()
	if now.Unix() >= ingress.start+60 {
		ingress.start = now.Unix()
		ingress.requests = 0
	}
	if ingress.requests >= 1200 {
		err := Fail("MCP_RATE_LIMITED", 429, "MCP entrance request limit exceeded")
		err.RetryAfterSeconds = int(ingress.start + 60 - now.Unix())
		if err.RetryAfterSeconds < 1 {
			err.RetryAfterSeconds = 1
		}
		return err
	}
	ingress.requests++
	return nil
}
