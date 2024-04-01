// HTTPS Listener For HTTP Redirect
//
// Adapted from net/http
//
// BSD-3-clause license
package hlfhr

import (
	"context"
	"net"
	"net/http"
)

type Server struct {
	*http.Server

	// HttpOnHttpsPortErrorHandler handles HTTP requests sent to an HTTPS port.
	//
	// WriteString use:
	//   io.WriteString(conn, "HTTP/1.1 400 Bad Request\r\nConnection: close\r\nConnection: close\r\n\r\nClient sent an HTTP request to an HTTPS server.\n")
	// Parse the request Host header and path use:
	//   host, path, ok := hlfhr.ReadReqHostPath(b)
	// Parse the request use:
	//   req, err := hlfhr.ReadReq(b)
	Hflhr_HttpOnHttpsPortErrorHandler func(b []byte, conn net.Conn)

	// Default 4096 Bytes
	Hflhr_ReadFirstRequestBytesLen int

	hlfhr_shuttingDown bool
}

// New hlfhr Server
func New(srv *http.Server) *Server {
	return NewServer(srv)
}

// New hlfhr Server
func NewServer(srv *http.Server) *Server {
	return &Server{
		Server:                            srv,
		Hflhr_ReadFirstRequestBytesLen:    4096,
		Hflhr_HttpOnHttpsPortErrorHandler: nil,
	}
}

// ListenAndServeTLS listens on the TCP network address srv.Addr and
// then calls ServeTLS to handle requests on incoming TLS connections.
// Accepted connections are configured to enable TCP keep-alives.
//
// Filenames containing a certificate and matching private key for the
// server must be provided if neither the Server's TLSConfig.Certificates
// nor TLSConfig.GetCertificate are populated. If the certificate is
// signed by a certificate authority, the certFile should be the
// concatenation of the server's certificate, any intermediates, and
// the CA's certificate.
//
// If srv.Addr is blank, ":https" is used.
//
// ListenAndServeTLS always returns a non-nil error. After Shutdown or
// Close, the returned error is ErrServerClosed.
func (srv *Server) ListenAndServeTLS(certFile string, keyFile string) error {
	if srv.hlfhr_shuttingDown {
		return http.ErrServerClosed
	}
	addr := srv.Addr
	if addr == "" {
		addr = ":https"
	}

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	ln = NewListener(ln, srv)

	defer ln.Close()

	return srv.Server.ServeTLS(ln, certFile, keyFile)
}

// ListenAndServeTLS acts identically to ListenAndServe, except that it
// expects HTTPS connections. Additionally, files containing a certificate and
// matching private key for the server must be provided. If the certificate
// is signed by a certificate authority, the certFile should be the concatenation
// of the server's certificate, any intermediates, and the CA's certificate.
func ListenAndServeTLS(addr, certFile, keyFile string, handler http.Handler) error {
	srv := New(&http.Server{
		Addr:    addr,
		Handler: handler,
	})
	return srv.ListenAndServeTLS(certFile, keyFile)
}

// ServeTLS accepts incoming connections on the Listener l, creating a
// new service goroutine for each. The service goroutines perform TLS
// setup and then read requests, calling srv.Handler to reply to them.
//
// Files containing a certificate and matching private key for the
// server must be provided if neither the Server's
// TLSConfig.Certificates nor TLSConfig.GetCertificate are populated.
// If the certificate is signed by a certificate authority, the
// certFile should be the concatenation of the server's certificate,
// any intermediates, and the CA's certificate.
//
// ServeTLS always returns a non-nil error. After Shutdown or Close, the
// returned error is ErrServerClosed.
func (srv *Server) ServeTLS(l net.Listener, certFile string, keyFile string) error {
	l = NewListener(l, srv)
	return srv.Server.ServeTLS(l, certFile, keyFile)
}

// ServeTLS accepts incoming HTTPS connections on the listener l,
// creating a new service goroutine for each. The service goroutines
// read requests and then call handler to reply to them.
//
// The handler is typically nil, in which case the DefaultServeMux is used.
//
// Additionally, files containing a certificate and matching private key
// for the server must be provided. If the certificate is signed by a
// certificate authority, the certFile should be the concatenation
// of the server's certificate, any intermediates, and the CA's certificate.
//
// ServeTLS always returns a non-nil error.
func ServeTLS(l net.Listener, handler http.Handler, certFile, keyFile string) error {
	srv := New(&http.Server{Handler: handler})
	return srv.ServeTLS(l, certFile, keyFile)
}

// Close immediately closes all active net.Listeners and any
// connections in state StateNew, StateActive, or StateIdle. For a
// graceful shutdown, use Shutdown.
//
// Close does not attempt to close (and does not even know about)
// any hijacked connections, such as WebSockets.
//
// Close returns any error returned from closing the Server's
// underlying Listener(s).
func (s *Server) Close() error {
	s.hlfhr_shuttingDown = true
	return s.Server.Close()
}

// Shutdown gracefully shuts down the server without interrupting any
// active connections. Shutdown works by first closing all open
// listeners, then closing all idle connections, and then waiting
// indefinitely for connections to return to idle and then shut down.
// If the provided context expires before the shutdown is complete,
// Shutdown returns the context's error, otherwise it returns any
// error returned from closing the Server's underlying Listener(s).
//
// When Shutdown is called, Serve, ListenAndServe, and
// ListenAndServeTLS immediately return ErrServerClosed. Make sure the
// program doesn't exit and waits instead for Shutdown to return.
//
// Shutdown does not attempt to close nor wait for hijacked
// connections such as WebSockets. The caller of Shutdown should
// separately notify such long-lived connections of shutdown and wait
// for them to close, if desired. See RegisterOnShutdown for a way to
// register shutdown notification functions.
//
// Once Shutdown has been called on a server, it may not be reused;
// future calls to methods such as Serve will return ErrServerClosed.
func (s *Server) Shutdown(ctx context.Context) error {
	s.hlfhr_shuttingDown = true
	return s.Server.Shutdown(ctx)
}

func (s *Server) Hlfhr_IsShuttingDown() bool {
	return s.hlfhr_shuttingDown
}
