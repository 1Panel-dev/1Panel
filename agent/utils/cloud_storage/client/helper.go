package client

import (
	"context"
	"fmt"
	"net/http"

	"github.com/1Panel-dev/1Panel/agent/global"
)

type deadlineTransport struct {
	ctx  context.Context
	base http.RoundTripper
}

func (t deadlineTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if err := t.ctx.Err(); err != nil {
		return nil, err
	}
	return t.base.RoundTrip(req.Clone(t.ctx))
}

func loadParamFromVars(key string, vars map[string]interface{}) string {
	if _, ok := vars[key]; !ok {
		if key != "bucket" && key != "port" && key != "authMode" && key != "passPhrase" {
			global.LOG.Errorf("load param %s from vars failed, err: not exist!", key)
		}
		return ""
	}

	return fmt.Sprintf("%v", vars[key])
}
