package email

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/smtp"
	"strings"
	"time"
)

type SMTPConfig struct {
	Host       string
	Port       int
	Sender     string
	Username   string
	Password   string
	From       string
	Encryption string
	Recipient  string
}

type EmailMessage struct {
	Subject string
	Body    string
	IsHTML  bool
}

type loginAuth struct {
	username, password string
}

func LoginAuth(username, password string) smtp.Auth {
	return &loginAuth{username, password}
}

func (a *loginAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	return "LOGIN", []byte{}, nil
}

func (a *loginAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		switch string(fromServer) {
		case "Username:":
			return []byte(a.username), nil
		case "Password:":
			return []byte(a.password), nil
		default:
			return nil, fmt.Errorf("unknown server challenge: %s", fromServer)
		}
	}
	return nil, nil
}

func SendMail(config SMTPConfig, message EmailMessage, transport *http.Transport) error {
	if err := validateConfig(config); err != nil {
		return err
	}

	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)
	toList := parseRecipients(config.Recipient)

	msg, err := buildMessage(config, message, toList)
	if err != nil {
		return err
	}

	switch strings.ToLower(config.Encryption) {
	case "ssl":
		return sendWithSSL(config, addr, toList, msg, transport)
	case "starttls", "tls":
		return sendWithStartTLS(config, addr, toList, msg, transport)
	case "none":
		return sendPlaintext(config, addr, toList, msg, transport)
	default:
		return fmt.Errorf("unsupported encryption type: %s", config.Encryption)
	}
}

func validateConfig(config SMTPConfig) error {
	if config.Host == "" {
		return fmt.Errorf("SMTP host is required")
	}
	if config.Port <= 0 {
		return fmt.Errorf("invalid SMTP port: %d", config.Port)
	}
	if config.Username == "" {
		return fmt.Errorf("SMTP username is required")
	}
	if config.Password == "" {
		return fmt.Errorf("SMTP password is required")
	}
	if config.From == "" {
		return fmt.Errorf("SMTP from address is required")
	}
	if config.Recipient == "" {
		return fmt.Errorf("SMTP recipient is required")
	}
	if !isValidEncryption(config.Encryption) {
		return fmt.Errorf("invalid encryption type: %s. Allowed: ssl, starttls, none", config.Encryption)
	}
	return nil
}

func isValidEncryption(enc string) bool {
	enc = strings.ToLower(enc)
	return enc == "ssl" || enc == "starttls" || enc == "none" || enc == "tls"
}

func parseRecipients(recipient string) []string {
	toList := strings.Split(recipient, ",")
	for i := range toList {
		toList[i] = strings.TrimSpace(toList[i])
	}
	return toList
}

func buildMessage(config SMTPConfig, message EmailMessage, toList []string) (string, error) {
	headers := make(map[string]string)
	headers["From"] = config.From
	headers["To"] = strings.Join(toList, ",")
	headers["Subject"] = message.Subject
	headers["Date"] = time.Now().UTC().Format(time.RFC1123Z)

	if message.IsHTML {
		headers["MIME-version"] = "1.0"
		headers["Content-Type"] = "text/html; charset=\"UTF-8\""
	} else {
		headers["Content-Type"] = "text/plain; charset=\"UTF-8\""
	}

	var msg strings.Builder
	for k, v := range headers {
		if !isValidHeader(k, v) {
			return "", fmt.Errorf("invalid header: %s: %s", k, v)
		}
		msg.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	msg.WriteString("\r\n" + message.Body)

	return msg.String(), nil
}

func isValidHeader(key, value string) bool {
	return !strings.ContainsAny(key, "\r\n") && !strings.ContainsAny(value, "\r\n")
}

func sendWithSSL(config SMTPConfig, addr string, toList []string, msg string, transport *http.Transport) error {
	var err error
	var conn net.Conn
	if transport != nil && transport.DialContext != nil {
		conn, err = transport.DialContext(context.Background(), "tcp", addr)
	} else {
		conn, err = net.Dial("tcp", addr)
	}
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}
	defer conn.Close()
	tlsConfig := &tls.Config{
		ServerName: config.Host,
	}
	tlsConn := tls.Client(conn, tlsConfig)
	if err := tlsConn.Handshake(); err != nil {
		return fmt.Errorf("TLS handshake failed: %w", err)
	}

	client, err := smtp.NewClient(tlsConn, config.Host)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Quit()
	if err := tryAuth(client, config.Username, config.Password, config.Host); err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}
	return sendEmailWithClient(client, config, toList, msg)
}

func sendWithStartTLS(config SMTPConfig, addr string, toList []string, msg string, transport *http.Transport) error {
	var err error
	var conn net.Conn
	if transport != nil && transport.DialContext != nil {
		conn, err = transport.DialContext(context.Background(), "tcp", addr)
	} else {
		conn, err = net.Dial("tcp", addr)
	}
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, config.Host)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Quit()

	if err = client.StartTLS(&tls.Config{ServerName: config.Host}); err != nil {
		return fmt.Errorf("failed to start TLS: %w", err)
	}
	if err := tryAuth(client, config.Username, config.Password, config.Host); err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}
	return sendEmailWithClient(client, config, toList, msg)
}

func sendPlaintext(config SMTPConfig, addr string, toList []string, msg string, transport *http.Transport) error {
	var err error
	var conn net.Conn
	if transport != nil && transport.DialContext != nil {
		conn, err = transport.DialContext(context.Background(), "tcp", addr)
	} else {
		conn, err = net.Dial("tcp", addr)
	}
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, config.Host)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}

	return sendEmailWithClient(client, config, toList, msg)
}

func sendEmailWithClient(client *smtp.Client, config SMTPConfig, toList []string, msg string) error {
	if err := client.Mail(config.Sender); err != nil {
		return fmt.Errorf("setting sender failed: %w", err)
	}
	for _, addr := range toList {
		if err := client.Rcpt(addr); err != nil {
			return fmt.Errorf("adding recipient %s failed: %w", addr, err)
		}
	}
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("preparing data failed: %w", err)
	}
	defer w.Close()

	if _, err := w.Write([]byte(msg)); err != nil {
		return fmt.Errorf("writing message failed: %w", err)
	}

	return nil
}

func tryAuth(client *smtp.Client, username, password, host string) error {
	ok, authCap := client.Extension("AUTH")
	if !ok {
		return fmt.Errorf("server does not support AUTH")
	}
	authCap = strings.ToUpper(authCap)
	if strings.Contains(authCap, "PLAIN") {
		auth := smtp.PlainAuth("", username, password, host)
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("plain auth failed: %w", err)
		}
		return nil
	}
	if strings.Contains(authCap, "LOGIN") {
		if err := client.Auth(LoginAuth(username, password)); err != nil {
			return fmt.Errorf("login auth failed: %w", err)
		}
		return nil
	}

	return fmt.Errorf("no supported auth mechanism, server supports: %s", authCap)
}
