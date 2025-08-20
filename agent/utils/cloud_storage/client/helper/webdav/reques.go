package webdav

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

func (c *Client) req(method, path string, body io.Reader, intercept func(*http.Request)) (rs *http.Response, err error) {
	var redo bool
	var r *http.Request
	var uri = PathEscape(Join(c.root, path))
	auth, body := c.auth.NewAuthenticator(body)
	defer auth.Close()

	for {
		if r, err = http.NewRequest(method, uri, body); err != nil {
			err = fmt.Errorf("handle request with uri: %s, method: %s failed, err: %v", uri, method, err)
			return
		}

		for k, vals := range c.headers {
			for _, v := range vals {
				r.Header.Add(k, v)
			}
		}

		if err = auth.Authorize(c.c, r, path); err != nil {
			return
		}

		if intercept != nil {
			intercept(r)
		}

		if rs, err = c.c.Do(r); err != nil {
			err = fmt.Errorf("do request for resp with uri: %s, method: %s failed, err: %v", uri, method, err)
			return
		}

		if redo, err = auth.Verify(c.c, rs, path); err != nil {
			rs.Body.Close()
			return nil, err
		}
		if redo {
			rs.Body.Close()
			if body, err = r.GetBody(); err != nil {
				return nil, err
			}
			continue
		}
		break
	}

	return rs, err
}

func (c *Client) propfind(path string, self bool, body string, resp interface{}, parse func(resp interface{}) error) error {
	rs, err := c.req("PROPFIND", path, strings.NewReader(body), func(rq *http.Request) {
		if self {
			rq.Header.Add("Depth", "0")
		} else {
			rq.Header.Add("Depth", "1")
		}
		rq.Header.Add("Content-Type", "application/xml;charset=UTF-8")
		rq.Header.Add("Accept", "application/xml,text/xml")
		rq.Header.Add("Accept-Charset", "utf-8")
		// TODO add support for 'gzip,deflate;q=0.8,q=0.7'
		rq.Header.Add("Accept-Encoding", "")
	})
	if err != nil {
		return err
	}
	defer rs.Body.Close()

	if rs.StatusCode != 207 {
		return NewPathError("PROPFIND", path, rs.StatusCode)
	}

	return parseXML(rs.Body, resp, parse)
}

func (c *Client) put(path string, stream io.Reader, contentLength int64) (status int, err error) {
	rs, err := c.req("PUT", path, stream, func(r *http.Request) {
		r.ContentLength = contentLength
	})
	if err != nil {
		return
	}
	defer rs.Body.Close()

	status = rs.StatusCode
	return
}
