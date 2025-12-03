package cloud_storage

import (
	"io"
	"net"
	"os"
	"path"
	"time"

	"github.com/1Panel-dev/1Panel/core/global"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type SftpClient struct {
	connInfo string
	config   *ssh.ClientConfig
}

func NewSftpClient(vars map[string]interface{}) (*SftpClient, error) {
	address := loadParamFromVars("address", vars)
	port := loadParamFromVars("port", vars)
	if len(port) == 0 {
		global.LOG.Errorf("load param port from vars failed, err: not exist!")
	}
	authMode := loadParamFromVars("authMode", vars)
	passPhrase := loadParamFromVars("passPhrase", vars)
	username := loadParamFromVars("username", vars)
	password := loadParamFromVars("password", vars)

	var auth []ssh.AuthMethod
	if authMode == "key" {
		var signer ssh.Signer
		var err error
		if len(passPhrase) != 0 {
			signer, err = ssh.ParsePrivateKeyWithPassphrase([]byte(password), []byte(passPhrase))
		} else {
			signer, err = ssh.ParsePrivateKey([]byte(password))
		}
		if err != nil {
			return nil, err
		}
		auth = []ssh.AuthMethod{ssh.PublicKeys(signer)}
	} else {
		auth = []ssh.AuthMethod{ssh.Password(password)}
	}
	clientConfig := &ssh.ClientConfig{
		User:    username,
		Auth:    auth,
		Timeout: 30 * time.Second,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}
	addr := net.JoinHostPort(address, port)
	if _, err := ssh.Dial("tcp", addr, clientConfig); err != nil {
		return nil, err
	}

	return &SftpClient{connInfo: addr, config: clientConfig}, nil
}

func (s SftpClient) Upload(src, target string) (bool, error) {
	sshClient, err := ssh.Dial("tcp", s.connInfo, s.config)
	if err != nil {
		return false, err
	}
	defer sshClient.Close()
	client, err := sftp.NewClient(sshClient)
	if err != nil {
		return false, err
	}
	defer client.Close()

	srcFile, err := os.Open(src)
	if err != nil {
		return false, err
	}
	defer srcFile.Close()

	targetDir, _ := path.Split(target)
	if len(targetDir) != 0 {
		if _, err = client.Stat(targetDir); err != nil {
			if os.IsNotExist(err) {
				if err = client.MkdirAll(targetDir); err != nil {
					return false, err
				}
			} else {
				return false, err
			}
		}
	}
	dstFile, err := client.Create(target)
	if err != nil {
		return false, err
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return false, err
	}
	return true, nil
}
