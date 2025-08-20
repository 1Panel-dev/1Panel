package webdav

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type PassportAuth struct {
	user            string
	pw              string
	cookies         []http.Cookie
	inhibitRedirect bool
}

func NewPassportAuth(c *http.Client, user, pw, partnerURL string, header *http.Header) (Authenticator, error) {
	p := &PassportAuth{
		user:            user,
		pw:              pw,
		inhibitRedirect: true,
	}
	err := p.genCookies(c, partnerURL, header)
	return p, err
}

func (p *PassportAuth) Authorize(c *http.Client, rq *http.Request, path string) error {
	if p.inhibitRedirect {
		rq.Header.Set(XInhibitRedirect, "1")
	} else {
		p.inhibitRedirect = true
	}
	for _, cookie := range p.cookies {
		rq.AddCookie(&cookie)
	}
	return nil
}

func (p *PassportAuth) Verify(c *http.Client, rs *http.Response, path string) (redo bool, err error) {
	switch rs.StatusCode {
	case 301, 302, 307, 308:
		redo = true
		if rs.Header.Get("Www-Authenticate") != "" {
			err = p.genCookies(c, rs.Request.URL.String(), &rs.Header)
		} else {
			p.inhibitRedirect = false
		}
	case 401:
		err = NewPathError("Authorize", path, rs.StatusCode)
	}
	return
}

func (p *PassportAuth) Close() error {
	return nil
}

func (p *PassportAuth) Clone() Authenticator {
	clonedCookies := make([]http.Cookie, len(p.cookies))
	copy(clonedCookies, p.cookies)

	return &PassportAuth{
		user:            p.user,
		pw:              p.pw,
		cookies:         clonedCookies,
		inhibitRedirect: true,
	}
}

func (p *PassportAuth) String() string {
	return fmt.Sprintf("PassportAuth login: %s", p.user)
}

func (p *PassportAuth) genCookies(c *http.Client, partnerUrl string, header *http.Header) error {
	baseAuthenticationServer := header.Get("Location")
	baseAuthenticationServerURL, err := url.Parse(baseAuthenticationServer)
	if err != nil {
		return err
	}

	authenticationServerUrl := url.URL{
		Scheme: baseAuthenticationServerURL.Scheme,
		Host:   baseAuthenticationServerURL.Host,
		Path:   "/login2.srf",
	}

	partnerServerChallenge := strings.Split(header.Get("Www-Authenticate"), " ")[1]

	req := http.Request{
		Method: "GET",
		URL:    &authenticationServerUrl,
		Header: http.Header{
			"Authorization": []string{"Passport1.4 sign-in=" + url.QueryEscape(p.user) + ",pwd=" + url.QueryEscape(p.pw) + ",OrgVerb=GET,OrgUrl=" + partnerUrl + "," + partnerServerChallenge},
		},
	}

	rs, err := c.Do(&req)
	if err != nil {
		return err
	}
	io.Copy(io.Discard, rs.Body)
	rs.Body.Close()
	if rs.StatusCode != 200 {
		return NewPathError("Authorize", "/", rs.StatusCode)
	}

	tokenResponseHeader := rs.Header.Get("Authentication-Info")
	if tokenResponseHeader == "" {
		return NewPathError("Authorize", "/", 401)
	}
	tokenResponseHeaderList := strings.Split(tokenResponseHeader, ",")
	token := ""
	for _, tokenResponseHeader := range tokenResponseHeaderList {
		if strings.HasPrefix(tokenResponseHeader, "from-PP='") {
			token = tokenResponseHeader
			break
		}
	}
	if token == "" {
		return NewPathError("Authorize", "/", 401)
	}

	origUrl, err := url.Parse(partnerUrl)
	if err != nil {
		return err
	}
	req = http.Request{
		Method: "GET",
		URL:    origUrl,
		Header: http.Header{
			"Authorization": []string{"Passport1.4 " + token},
		},
	}

	rs, err = c.Do(&req)
	if err != nil {
		return err
	}
	io.Copy(io.Discard, rs.Body)
	rs.Body.Close()
	if rs.StatusCode != 200 && rs.StatusCode != 302 {
		return NewPathError("Authorize", "/", rs.StatusCode)
	}

	cookies := rs.Header.Values("Set-Cookie")
	p.cookies = make([]http.Cookie, len(cookies))
	for i, cookie := range cookies {
		cookieParts := strings.Split(cookie, ";")
		cookieName := strings.Split(cookieParts[0], "=")[0]
		cookieValue := strings.Split(cookieParts[0], "=")[1]

		p.cookies[i] = http.Cookie{
			Name:  cookieName,
			Value: cookieValue,
		}
	}

	return nil
}
