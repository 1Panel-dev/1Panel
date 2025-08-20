package webdav

import (
	"fmt"
	"net/http"
)

type BasicAuth struct {
	user string
	pw   string
}

func (b *BasicAuth) Authorize(c *http.Client, rq *http.Request, path string) error {
	rq.SetBasicAuth(b.user, b.pw)
	return nil
}

func (b *BasicAuth) Verify(c *http.Client, rs *http.Response, path string) (redo bool, err error) {
	if rs.StatusCode == 401 {
		err = NewPathError("Authorize", path, rs.StatusCode)
	}
	return
}

func (b *BasicAuth) Close() error {
	return nil
}

func (b *BasicAuth) Clone() Authenticator {
	// no copy due to read only access
	return b
}

func (b *BasicAuth) String() string {
	return fmt.Sprintf("BasicAuth login: %s", b.user)
}
