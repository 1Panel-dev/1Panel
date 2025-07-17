package email

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/smtp"
	"strings"
)

// SMTPConfig holds SMTP connection info
type SMTPConfig struct {
	Host       string
	Port       int // 465, 587, 25
	Username   string
	Password   string
	From       string
	Encryption string // "ssl", "starttls", "none"
	Recipient  string
}

// EmailMessage represents an email
type EmailMessage struct {
	Subject string
	Body    string // HTML or plain text
	IsHTML  bool
}

// SendMail sends the email using the given config
func SendMail(config SMTPConfig, message EmailMessage, transport *http.Transport) error {
	// 验证配置
	if err := validateConfig(config); err != nil {
		return err
	}

	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)
	toList := parseRecipients(config.Recipient)

	// 构建邮件内容
	msg, err := buildMessage(config, message, toList)
	if err != nil {
		return err
	}

	// 根据加密类型选择发送方式
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

// 验证配置有效性
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

// 检查加密类型是否有效
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

// 构建邮件内容
func buildMessage(config SMTPConfig, message EmailMessage, toList []string) (string, error) {
	headers := make(map[string]string)
	headers["From"] = config.From
	headers["To"] = strings.Join(toList, ",")
	headers["Subject"] = message.Subject

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

// 验证邮件头安全性
func isValidHeader(key, value string) bool {
	return !strings.ContainsAny(key, "\r\n") && !strings.ContainsAny(value, "\r\n")
}

// SSL/TLS方式发送邮件
func sendWithSSL(config SMTPConfig, addr string, toList []string, msg string, transport *http.Transport) error {
	var err error
	var conn net.Conn
	if transport != nil {
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

	// 创建SMTP客户端
	client, err := smtp.NewClient(tlsConn, config.Host)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Quit()

	return sendEmailWithClient(client, config, toList, msg)
}

// STARTTLS方式发送邮件
func sendWithStartTLS(config SMTPConfig, addr string, toList []string, msg string, transport *http.Transport) error {
	var err error
	var conn net.Conn
	if transport != nil {
		conn, err = transport.DialContext(context.Background(), "tcp", addr)
	} else {
		conn, err = net.Dial("tcp", addr)
	}
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}
	defer conn.Close()

	// 创建SMTP客户端
	client, err := smtp.NewClient(conn, config.Host)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Quit()

	// 启用TLS
	if err = client.StartTLS(&tls.Config{ServerName: config.Host}); err != nil {
		return fmt.Errorf("failed to start TLS: %w", err)
	}

	return sendEmailWithClient(client, config, toList, msg)
}

// 明文方式发送邮件
func sendPlaintext(config SMTPConfig, addr string, toList []string, msg string, transport *http.Transport) error {
	var err error
	var conn net.Conn
	if transport != nil {
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

// 使用SMTP客户端发送邮件
func sendEmailWithClient(client *smtp.Client, config SMTPConfig, toList []string, msg string) error {
	auth := smtp.PlainAuth("", config.Username, config.Password, config.Host)
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}
	if err := client.Mail(config.Username); err != nil {
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
