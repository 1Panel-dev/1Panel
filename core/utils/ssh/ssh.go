package ssh

import (
	"fmt"
	"strings"
	"time"

	gossh "golang.org/x/crypto/ssh"
)

type ConnInfo struct {
	User        string        `json:"user"`
	Addr        string        `json:"addr"`
	Port        int           `json:"port"`
	AuthMode    string        `json:"authMode"`
	Password    string        `json:"password"`
	PrivateKey  []byte        `json:"privateKey"`
	PassPhrase  []byte        `json:"passPhrase"`
	DialTimeOut time.Duration `json:"dialTimeOut"`
}

type SSHClient struct {
	Client *gossh.Client `json:"client"`
}

func NewClient(c ConnInfo) (*SSHClient, error) {
	if strings.Contains(c.Addr, ":") {
		c.Addr = fmt.Sprintf("[%s]", c.Addr)
	}
	config := &gossh.ClientConfig{}
	config.SetDefaults()
	addr := fmt.Sprintf("%s:%d", c.Addr, c.Port)
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
	client, err := gossh.Dial(proto, addr, config)
	if nil != err {
		return nil, err
	}
	return &SSHClient{Client: client}, nil
}

func (c *SSHClient) Run(shell string) (string, error) {
	session, err := c.Client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()
	buf, err := session.CombinedOutput(shell)

	return string(buf), err
}

func (c *SSHClient) SudoHandleCmd() string {
	if _, err := c.Run("sudo -n ls"); err == nil {
		return "sudo "
	}
	return ""
}

func (c *SSHClient) Runf(shell string, args ...interface{}) (string, error) {
	session, err := c.Client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()
	buf, err := session.CombinedOutput(fmt.Sprintf(shell, args...))

	return string(buf), err
}

func (c *SSHClient) Close() {
	_ = c.Client.Close()
}

func makePrivateKeySigner(privateKey []byte, passPhrase []byte) (gossh.Signer, error) {
	if len(passPhrase) != 0 {
		return gossh.ParsePrivateKeyWithPassphrase(privateKey, passPhrase)
	}
	return gossh.ParsePrivateKey(privateKey)
}
