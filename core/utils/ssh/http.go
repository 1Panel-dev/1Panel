package ssh

import (
	"bufio"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/core/global"
)

type HTTPProxyDialer struct {
	Type     string
	URL      string
	User     string
	Password string
}

func HTTPDial(dialer HTTPProxyDialer, network, addr string) (net.Conn, error) {
	var conn net.Conn
	var err error

	global.LOG.Debugf("Dialing HTTP proxy %s for %s", dialer.URL, addr)
	dialer.URL = strings.TrimPrefix(dialer.URL, dialer.Type+"://")
	if dialer.Type == "https" {
		conn, err = tls.DialWithDialer(
			&net.Dialer{Timeout: 30 * time.Second},
			network,
			dialer.URL,
			&tls.Config{InsecureSkipVerify: true},
		)
	} else {
		conn, err = net.DialTimeout(network, dialer.URL, 30*time.Second)
	}
	if err != nil {
		return nil, err
	}
	conn.SetDeadline(time.Now().Add(30 * time.Second))
	connectReq := fmt.Sprintf("CONNECT %s HTTP/1.1\r\n", addr)
	connectReq += fmt.Sprintf("Host: %s\r\n", addr)
	connectReq += "User-Agent: Go-ssh-client/1.0\r\n"

	if dialer.User != "" {
		auth := base64.StdEncoding.EncodeToString(
			[]byte(dialer.User + ":" + dialer.Password),
		)
		connectReq += fmt.Sprintf("Proxy-Authorization: Basic %s\r\n", auth)
	}
	connectReq += "Connection: keep-alive\r\n\r\n"
	if _, err := conn.Write([]byte(connectReq)); err != nil {
		conn.Close()
		return nil, err
	}
	reader := bufio.NewReader(conn)
	response, err := reader.ReadString('\n')
	if err != nil {
		conn.Close()
		return nil, err
	}
	if !strings.HasPrefix(response, "HTTP/1.1 200") &&
		!strings.HasPrefix(response, "HTTP/1.0 200") {
		conn.Close()
		return nil, fmt.Errorf("proxy connection failed: %s", strings.TrimSpace(response))
	}
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			conn.Close()
			return nil, err
		}
		if line == "\r\n" || line == "\n" {
			break
		}
	}
	conn.SetDeadline(time.Time{})

	return conn, nil
}
