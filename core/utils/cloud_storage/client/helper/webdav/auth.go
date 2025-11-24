package webdav

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync"
)

type AuthFactory func(c *http.Client, rs *http.Response, path string) (auth Authenticator, err error)

type Authorizer interface {
	NewAuthenticator(body io.Reader) (Authenticator, io.Reader)
	AddAuthenticator(key string, fn AuthFactory)
}

type Authenticator interface {
	Authorize(c *http.Client, rq *http.Request, path string) error
	Verify(c *http.Client, rs *http.Response, path string) (redo bool, err error)
	Clone() Authenticator
	io.Closer
}

type authfactory struct {
	key    string
	create AuthFactory
}

type authorizer struct {
	factories  []authfactory
	defAuthMux sync.Mutex
	defAuth    Authenticator
}

type preemptiveAuthorizer struct {
	auth Authenticator
}

type authShim struct {
	factory AuthFactory
	body    io.Reader
	auth    Authenticator
}

type negoAuth struct {
	auths                   []Authenticator
	setDefaultAuthenticator func(auth Authenticator)
}

type nullAuth struct{}

type noAuth struct{}

func NewAutoAuth(login string, secret string) Authorizer {
	fmap := make([]authfactory, 0)
	az := &authorizer{factories: fmap, defAuthMux: sync.Mutex{}, defAuth: &nullAuth{}}

	az.AddAuthenticator("basic", func(c *http.Client, rs *http.Response, path string) (auth Authenticator, err error) {
		return &BasicAuth{user: login, pw: secret}, nil
	})

	az.AddAuthenticator("digest", func(c *http.Client, rs *http.Response, path string) (auth Authenticator, err error) {
		return NewDigestAuth(login, secret, rs)
	})

	az.AddAuthenticator("passport1.4", func(c *http.Client, rs *http.Response, path string) (auth Authenticator, err error) {
		return NewPassportAuth(c, login, secret, rs.Request.URL.String(), &rs.Header)
	})

	return az
}

func NewEmptyAuth() Authorizer {
	fmap := make([]authfactory, 0)
	az := &authorizer{factories: fmap, defAuthMux: sync.Mutex{}, defAuth: &nullAuth{}}
	return az
}

func NewPreemptiveAuth(auth Authenticator) Authorizer {
	return &preemptiveAuthorizer{auth: auth}
}

func (a *authorizer) NewAuthenticator(body io.Reader) (Authenticator, io.Reader) {
	var retryBuf io.Reader = body
	if body != nil {
		if _, ok := retryBuf.(io.Seeker); ok {
			body = io.NopCloser(body)
		} else {
			buff := &bytes.Buffer{}
			retryBuf = buff
			body = io.TeeReader(body, buff)
		}
	}
	a.defAuthMux.Lock()
	defAuth := a.defAuth.Clone()
	a.defAuthMux.Unlock()

	return &authShim{factory: a.factory, body: retryBuf, auth: defAuth}, body
}

func (a *authorizer) AddAuthenticator(key string, fn AuthFactory) {
	key = strings.ToLower(key)
	for _, f := range a.factories {
		if f.key == key {
			panic("Authenticator exists: " + key)
		}
	}
	a.factories = append(a.factories, authfactory{key, fn})
}

func (a *authorizer) factory(c *http.Client, rs *http.Response, path string) (auth Authenticator, err error) {
	headers := rs.Header.Values("Www-Authenticate")
	if len(headers) > 0 {
		auths := make([]Authenticator, 0)
		for _, f := range a.factories {
			for _, header := range headers {
				headerLower := strings.ToLower(header)
				if strings.Contains(headerLower, f.key) {
					rs.Header.Set("Www-Authenticate", header)
					if auth, err = f.create(c, rs, path); err == nil {
						auths = append(auths, auth)
						break
					}
				}
			}
		}

		switch len(auths) {
		case 0:
			return nil, NewPathError("NoAuthenticator", path, rs.StatusCode)
		case 1:
			auth = auths[0]
		default:
			auth = &negoAuth{auths: auths, setDefaultAuthenticator: a.setDefaultAuthenticator}
		}
	} else {
		auth = &noAuth{}
	}

	a.setDefaultAuthenticator(auth)

	return auth, nil
}

func (a *authorizer) setDefaultAuthenticator(auth Authenticator) {
	a.defAuthMux.Lock()
	a.defAuth.Close()
	a.defAuth = auth
	a.defAuthMux.Unlock()
}

func (s *authShim) Authorize(c *http.Client, rq *http.Request, path string) error {
	if err := s.auth.Authorize(c, rq, path); err != nil {
		return err
	}
	body := s.body
	rq.GetBody = func() (io.ReadCloser, error) {
		if body != nil {
			if sk, ok := body.(io.Seeker); ok {
				if _, err := sk.Seek(0, io.SeekStart); err != nil {
					return nil, err
				}
			}
			return io.NopCloser(body), nil
		}
		return nil, nil
	}
	return nil
}

func (s *authShim) Verify(c *http.Client, rs *http.Response, path string) (redo bool, err error) {
	redo, err = s.auth.Verify(c, rs, path)
	if err != nil && errors.Is(err, ErrAuthChanged) {
		if auth, aerr := s.factory(c, rs, path); aerr == nil {
			s.auth.Close()
			s.auth = auth
			return true, nil
		} else {
			return false, aerr
		}
	}
	return
}

func (s *authShim) Close() error {
	s.auth.Close()
	s.auth, s.factory = nil, nil
	if s.body != nil {
		if closer, ok := s.body.(io.Closer); ok {
			return closer.Close()
		}
	}
	return nil
}

func (s *authShim) Clone() Authenticator {
	return &noAuth{}
}

func (s *authShim) String() string {
	return "AuthShim"
}

func (n *negoAuth) Authorize(c *http.Client, rq *http.Request, path string) error {
	if len(n.auths) == 0 {
		return NewPathError("NoAuthenticator", path, 400)
	}
	return n.auths[0].Authorize(c, rq, path)
}

func (n *negoAuth) Verify(c *http.Client, rs *http.Response, path string) (redo bool, err error) {
	if len(n.auths) == 0 {
		return false, NewPathError("NoAuthenticator", path, 400)
	}
	redo, err = n.auths[0].Verify(c, rs, path)
	if err != nil {
		if len(n.auths) > 1 {
			n.auths[0].Close()
			n.auths = n.auths[1:]
			return true, nil
		}
	} else if redo {
		return
	} else {
		auth := n.auths[0]
		n.auths = n.auths[1:]
		n.setDefaultAuthenticator(auth)
		return
	}

	return false, NewPathError("NoAuthenticator", path, rs.StatusCode)
}

func (n *negoAuth) Close() error {
	for _, a := range n.auths {
		a.Close()
	}
	n.setDefaultAuthenticator = nil
	return nil
}

func (n *negoAuth) Clone() Authenticator {
	auths := make([]Authenticator, len(n.auths))
	for i, e := range n.auths {
		auths[i] = e.Clone()
	}
	return &negoAuth{auths: auths, setDefaultAuthenticator: n.setDefaultAuthenticator}
}

func (n *negoAuth) String() string {
	return "NegoAuth"
}

func (n *noAuth) Authorize(c *http.Client, rq *http.Request, path string) error {
	return nil
}

func (n *noAuth) Verify(c *http.Client, rs *http.Response, path string) (redo bool, err error) {
	if "" != rs.Header.Get("Www-Authenticate") {
		err = ErrAuthChanged
	}
	return
}

func (n *noAuth) Close() error {
	return nil
}

func (n *noAuth) Clone() Authenticator {
	return n
}

func (n *noAuth) String() string {
	return "NoAuth"
}

func (n *nullAuth) Authorize(c *http.Client, rq *http.Request, path string) error {
	rq.Header.Set(XInhibitRedirect, "1")
	return nil
}

func (n *nullAuth) Verify(c *http.Client, rs *http.Response, path string) (redo bool, err error) {
	return true, ErrAuthChanged
}

func (n *nullAuth) Close() error {
	return nil
}

func (n *nullAuth) Clone() Authenticator {
	return n
}

func (n *nullAuth) String() string {
	return "NullAuth"
}

func (b *preemptiveAuthorizer) NewAuthenticator(body io.Reader) (Authenticator, io.Reader) {
	return b.auth.Clone(), body
}

func (b *preemptiveAuthorizer) AddAuthenticator(key string, fn AuthFactory) {
	panic("You're funny! A preemptive authorizer may only have a single authentication method")
}
