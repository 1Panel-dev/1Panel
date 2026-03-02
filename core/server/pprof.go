//go:build pprof

package server

import (
	"net/http"
	_ "net/http/pprof"
	"os"

	"github.com/1Panel-dev/1Panel/core/global"
)

func startPprof() {
	addr := os.Getenv("1PANEL_PPROF_ADDR")
	if addr == "" {
		addr = ":6060"
	}
	go func() {
		global.LOG.Infof("pprof server listening on %s", addr)
		if err := http.ListenAndServe(addr, nil); err != nil {
			global.LOG.Errorf("pprof server error: %v", err)
		}
	}()
}
