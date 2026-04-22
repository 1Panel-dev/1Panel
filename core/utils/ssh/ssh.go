package ssh

import (
	"errors"
	"fmt"
	"net"
	"path"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/core/app/repo"
	"github.com/1Panel-dev/1Panel/core/global"
	"github.com/1Panel-dev/1Panel/core/utils/encrypt"
	gossh "golang.org/x/crypto/ssh"
	"golang.org/x/net/proxy"
)

type ConnInfo struct {
	User       string `json:"user"`
	Addr       string `json:"addr"`
	Port       int    `json:"port"`
	AuthMode   string `json:"authMode"`
	Password   string `json:"password"`
	PrivateKey []byte `json:"privateKey"`
	PassPhrase []byte `json:"passPhrase"`

	UseProxy    bool          `json:"useProxy"`
	DialTimeOut time.Duration `json:"dialTimeOut"`
}

type SSHClient struct {
	Client   *gossh.Client `json:"client"`
	SudoItem string        `json:"sudoItem"`
}

func NewClient(c ConnInfo) (*SSHClient, error) {
	config := &gossh.ClientConfig{}
	config.SetDefaults()
	addr := net.JoinHostPort(c.Addr, fmt.Sprintf("%d", c.Port))
	config.User = c.User
	if c.AuthMode == "password" {
		config.Auth = []gossh.AuthMethod{gossh.Password(c.Password)}
	} else {
		signer, err := makePrivateKeySigner(c.PrivateKey, c.PassPhrase)
		if err != nil {
			return nil, err
		}
		config.Auth = []gossh.AuthMethod{gossh.PublicKeys(signer)}
	}
	if c.DialTimeOut == 0 {
		c.DialTimeOut = 5 * time.Second
	}
	config.Timeout = c.DialTimeOut

	config.HostKeyCallback = gossh.InsecureIgnoreHostKey()
	proto := "tcp"
	if strings.Contains(c.Addr, ":") {
		proto = "tcp6"
	}
	client, err := DialWithTimeout(proto, addr, c.UseProxy, config)
	if nil != err {
		return nil, err
	}
	sshClient := &SSHClient{Client: client}
	if c.User == "root" {
		return sshClient, nil
	}
	if _, err := sshClient.Run("sudo -n ls"); err == nil {
		sshClient.SudoItem = "sudo"
	}
	return sshClient, nil
}

func (c *SSHClient) Run(shell string) (string, error) {
	shell = c.SudoItem + " " + shell
	shell = strings.ReplaceAll(shell, " && ", fmt.Sprintf(" && %s ", c.SudoItem))
	session, err := c.Client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()
	buf, err := session.CombinedOutput(shell)

	return string(buf), err
}

func (c *SSHClient) CpFileWithCheck(src, dst string) error {
	localMd5, err := c.Runf("md5sum %s | awk '{print $1}'", src)
	if err != nil {
		global.LOG.Debugf("load md5sum with src for %s failed, std: %s, err: %v", path.Base(src), localMd5, err)
		localMd5 = ""
	}
	for i := 0; i < 3; i++ {
		std, cpErr := c.Runf("cp %s %s", src, dst)
		if err != nil {
			err = fmt.Errorf("cp file %s failed, std: %s, err: %v", src, std, cpErr)
			continue
		}
		if len(strings.TrimSpace(localMd5)) == 0 {
			return nil
		}
		remoteMd5, errDst := c.Runf("md5sum %s | awk '{print $1}'", dst)
		if errDst != nil {
			global.LOG.Debugf("load md5sum with dst for %s failed, std: %s, err: %v", path.Base(src), remoteMd5, errDst)
			return nil
		}
		if strings.TrimSpace(localMd5) == strings.TrimSpace(remoteMd5) {
			return nil
		}
		err = errors.New("cp file failed, file is not match!")
	}

	return err
}

func (c *SSHClient) SudoHandleCmd() string {
	if _, err := c.Run("sudo -n ls"); err == nil {
		return "sudo "
	}
	return ""
}

func (c *SSHClient) IsRoot(user string) bool {
	if user == "root" {
		return true
	}
	_, err := c.Run("sudo -n true")
	return err == nil
}

func (c *SSHClient) Runf(shell string, args ...interface{}) (string, error) {
	shell = c.SudoItem + " " + shell
	shell = strings.ReplaceAll(shell, " && ", fmt.Sprintf(" && %s ", c.SudoItem))
	session, err := c.Client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()
	buf, err := session.CombinedOutput(fmt.Sprintf(shell, args...))

	return string(buf), err
}

func (c *SSHClient) Close() {
	if c.Client != nil {
		_ = c.Client.Close()
	}
}

func makePrivateKeySigner(privateKey []byte, passPhrase []byte) (gossh.Signer, error) {
	if len(passPhrase) != 0 {
		return gossh.ParsePrivateKeyWithPassphrase(privateKey, passPhrase)
	}
	return gossh.ParsePrivateKey(privateKey)
}

func (c *SSHClient) RunWithStreamOutput(command string, outputCallback func(string)) error {
	session, err := c.Client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create SSH session: %w", err)
	}
	defer session.Close()

	stdout, err := session.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to set up stdout pipe: %w", err)
	}

	stderr, err := session.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to set up stderr pipe: %w", err)
	}

	if err := session.Start(command); err != nil {
		return fmt.Errorf("failed to start command: %w", err)
	}

	stdoutCh := make(chan string, 100)
	stderrCh := make(chan string, 100)
	doneCh := make(chan struct{})

	go func() {
		buffer := make([]byte, 1024)
		for {
			n, err := stdout.Read(buffer)
			if err != nil {
				close(stdoutCh)
				return
			}
			if n > 0 {
				stdoutCh <- string(buffer[:n])
			}
		}
	}()

	go func() {
		buffer := make([]byte, 1024)
		for {
			n, err := stderr.Read(buffer)
			if err != nil {
				close(stderrCh)
				return
			}
			if n > 0 {
				stderrCh <- string(buffer[:n])
			}
		}
	}()

	go func() {
		for {
			select {
			case stdoutOutput, ok := <-stdoutCh:
				if !ok {
					stdoutCh = nil
					if stderrCh == nil {
						close(doneCh)
						return
					}
					continue
				}
				if outputCallback != nil {
					outputCallback(stdoutOutput)
				}

			case stderrOutput, ok := <-stderrCh:
				if !ok {
					stderrCh = nil
					if stdoutCh == nil {
						close(doneCh)
						return
					}
					continue
				}
				if outputCallback != nil {
					outputCallback(stderrOutput)
				}
			}
		}
	}()

	err = session.Wait()
	<-doneCh

	return err
}

func DialWithTimeout(network, addr string, useProxy bool, config *gossh.ClientConfig) (*gossh.Client, error) {
	var conn net.Conn
	var err error
	if useProxy {
		conn, err = loadSSHConnByProxy(network, addr, config.Timeout)
	} else {
		conn, err = net.DialTimeout(network, addr, config.Timeout)
	}
	if err != nil {
		return nil, err
	}
	_ = conn.SetDeadline(time.Now().Add(config.Timeout))
	c, chans, reqs, err := gossh.NewClientConn(conn, addr, config)
	if err != nil {
		return nil, err
	}
	if err := conn.SetDeadline(time.Time{}); err != nil {
		conn.Close()
		return nil, fmt.Errorf("clear deadline failed: %v", err)
	}
	return gossh.NewClient(c, chans, reqs), nil
}

func loadSSHConnByProxy(network, addr string, timeout time.Duration) (net.Conn, error) {
	settingRepo := repo.NewISettingRepo()
	proxyType, err := settingRepo.Get(repo.WithByKey("ProxyType"))
	if err != nil {
		return nil, fmt.Errorf("get proxy type from db failed, err: %v", err)
	}
	if len(proxyType.Value) == 0 {
		return nil, fmt.Errorf("get proxy type from db failed, err: %v", err)
	}
	proxyUrl, _ := settingRepo.Get(repo.WithByKey("ProxyUrl"))
	port, _ := settingRepo.Get(repo.WithByKey("ProxyPort"))
	user, _ := settingRepo.Get(repo.WithByKey("ProxyUser"))
	passwd, _ := settingRepo.Get(repo.WithByKey("ProxyPasswd"))

	pass, _ := encrypt.StringDecrypt(passwd.Value)
	proxyItem := fmt.Sprintf("%s:%s", proxyUrl.Value, port.Value)
	switch proxyType.Value {
	case "http", "https":
		item := HTTPProxyDialer{
			Type:     proxyType.Value,
			URL:      proxyItem,
			User:     user.Value,
			Password: pass,
		}
		return HTTPDial(item, network, addr)
	case "socks5":
		var auth *proxy.Auth
		if len(user.Value) == 0 {
			auth = nil
		} else {
			auth = &proxy.Auth{
				User:     user.Value,
				Password: pass,
			}
		}
		dialer, err := proxy.SOCKS5("tcp", proxyItem, auth, &net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		})
		if err != nil {
			return nil, fmt.Errorf("new socks5 proxy failed, err: %v", err)
		}
		return dialer.Dial(network, addr)
	default:
		conn, err := net.DialTimeout(network, addr, timeout)
		if err != nil {
			return nil, err
		}
		return conn, nil
	}
}
