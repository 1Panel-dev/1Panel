// HTTPS Listener For HTTP Redirect
//
// Adapted from net/http
//
// BSD-3-clause license
package hlfhr

import (
	"bytes"
	"fmt"
	"net"
)

type Listener struct {
	net.Listener

	// HttpOnHttpsPortErrorHandler handles HTTP requests sent to an HTTPS port.
	hlfhr_httpOnHttpsPortErrorHandler *func(b []byte, conn net.Conn)

	// Default 4096 Bytes
	hflhr_readFirstRequestBytesLen int
}

func NewListener(inner net.Listener, srv *Server) net.Listener {
	var l *Listener
	if innerThisListener, ok := inner.(*Listener); ok {
		l = innerThisListener
	} else {
		l = &Listener{
			Listener: inner,
		}
	}
	l.hlfhr_httpOnHttpsPortErrorHandler = &srv.Hflhr_HttpOnHttpsPortErrorHandler
	l.hflhr_readFirstRequestBytesLen = srv.Hflhr_ReadFirstRequestBytesLen
	if l.hflhr_readFirstRequestBytesLen == 0 {
		l.hflhr_readFirstRequestBytesLen = 4096
	}
	return l
}

func (l *Listener) Accept() (net.Conn, error) {
	c, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}
	// Hijacking net.Conn
	return newConn(c, l), nil
}

type conn struct {
	net.Conn

	// HttpOnHttpsPortErrorHandler handles HTTP requests sent to an HTTPS port.
	hlfhr_HttpOnHttpsPortErrorHandler *func(b []byte, conn net.Conn)

	// Default 4096
	hflhr_readFirstRequestBytesLen int

	isWritten                 bool
	isNotFirstRead            bool
	firstReadBytesForRedirect []byte
}

func newConn(inner net.Conn, l *Listener) net.Conn {
	c := &conn{
		Conn:                              inner,
		hlfhr_HttpOnHttpsPortErrorHandler: l.hlfhr_httpOnHttpsPortErrorHandler,
		hflhr_readFirstRequestBytesLen:    l.hflhr_readFirstRequestBytesLen,
		isWritten:                         false,
		isNotFirstRead:                    false,
		firstReadBytesForRedirect:         nil,
	}
	if c.hflhr_readFirstRequestBytesLen == 0 {
		c.hflhr_readFirstRequestBytesLen = 4096
	}
	return c
}

func (c *conn) Read(b []byte) (n int, err error) {
	if c.isNotFirstRead {
		return c.Conn.Read(b)
	}
	c.isNotFirstRead = true

	// Default 576 Bytes
	if len(b) <= 5 {
		return c.Conn.Read(b)
	}
	if len(b) >= c.hflhr_readFirstRequestBytesLen {
		n, err = c.Conn.Read(b)
		if err == nil && compiledRegexp_tlsRecordHeaderLooksLikeHTTP.Match(b) {
			// Cache for redirect
			c.firstReadBytesForRedirect = b[:n]
		}
		return
	}

	// Read 5 Bytes Header
	rb5 := make([]byte, 5)
	rb5n, err := c.Conn.Read(rb5)
	if err != nil {
		return 0, err
	}
	rb5 = rb5[:rb5n]

	// More Bytes Length
	var rb4kLen int
	looksLikeHttp := compiledRegexp_tlsRecordHeaderLooksLikeHTTP.Match(rb5)
	if looksLikeHttp {
		// HTTP Read 4096 Bytes Cache for redirect
		rb4kLen = c.hflhr_readFirstRequestBytesLen - rb5n
	} else {
		// HTTPS Default 576 Bytes
		rb4kLen = len(b) - rb5n
	}

	// Read More Bytes
	rb4k := make([]byte, rb4kLen)
	rb4kn, err := c.Conn.Read(rb4k)
	if err != nil {
		return 0, err
	}
	rb4k = rb4k[:rb4kn]

	rbAll := append(rb5, rb4k...)
	if looksLikeHttp {
		// Cache for redirect
		c.firstReadBytesForRedirect = rbAll
	}

	return bytes.NewBuffer(rbAll).Read(b)
}

// Hijacking the Write function to achieve redirection
func (c *conn) Write(b []byte) (n int, err error) {
	if !c.isWritten {
		c.isWritten = true
		// redirect
		if c.firstReadBytesForRedirect != nil {
			defer func() {
				c.firstReadBytesForRedirect = nil
			}()
			if bytes.Equal(b, []byte("HTTP/1.0 400 Bad Request\r\n\r\nClient sent an HTTP request to an HTTPS server.\n")) {
				n = len(b)
				frb := c.firstReadBytesForRedirect
				// handler
				if c.hlfhr_HttpOnHttpsPortErrorHandler != nil {
					if handler := *c.hlfhr_HttpOnHttpsPortErrorHandler; handler != nil {
						handler(frb, c.Conn)
						return
					}
				}
				// 302 Found
				host, path, ok := ReadReqHostPath(frb)
				if ok {
					c.Conn.Write([]byte(fmt.Sprint("HTTP/1.1 302 Found\r\nLocation: https://", host, path, "\r\nConnection: close\r\n\r\nRedirect to HTTPS\n")))
					return
				}
				// script
				c.Conn.Write([]byte("HTTP/1.1 400 Bad Request\r\nContent-Type: text/html\r\nConnection: close\r\n\r\n<script> location.protocol = 'https:' </script>\n"))
				return
			}
		}
	}
	return c.Conn.Write(b)
}
