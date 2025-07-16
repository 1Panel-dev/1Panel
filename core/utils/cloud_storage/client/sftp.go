package client

import (
	"fmt"
	"io"
	"net"
	"os"
	"path"
	"time"

	"github.com/1Panel-dev/1Panel/core/global"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type sftpClient struct {
	bucket   string
	connInfo string
	config   *ssh.ClientConfig
}

func NewSftpClient(vars map[string]interface{}) (*sftpClient, error) {
	address := loadParamFromVars("address", vars)
	port := loadParamFromVars("port", vars)
	if len(port) == 0 {
		global.LOG.Errorf("load param port from vars failed, err: not exist!")
	}
	authMode := loadParamFromVars("authMode", vars)
	passPhrase := loadParamFromVars("passPhrase", vars)
	password := loadParamFromVars("password", vars)
	bucket := loadParamFromVars("bucket", vars)

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
	username := loadParamFromVars("username", vars)

	clientConfig := &ssh.ClientConfig{
		User:    username,
		Auth:    auth,
		Timeout: 30 * time.Second,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}
	if _, err := ssh.Dial("tcp", fmt.Sprintf("%s:%s", address, port), clientConfig); err != nil {
		return nil, err
	}

	return &sftpClient{connInfo: fmt.Sprintf("%s:%s", address, port), config: clientConfig, bucket: bucket}, nil
}

func (s sftpClient) Upload(src, target string) (bool, error) {
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

	targetFilePath := path.Join(s.bucket, target)
	targetDir, _ := path.Split(targetFilePath)
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
	dstFile, err := client.Create(path.Join(s.bucket, target))
	if err != nil {
		return false, err
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return false, err
	}
	return true, nil
}

func (s sftpClient) ListBuckets() ([]interface{}, error) {
	var result []interface{}
	return result, nil
}

func (s sftpClient) Delete(filePath string) (bool, error) {
	sshClient, err := ssh.Dial("tcp", s.connInfo, s.config)
	if err != nil {
		return false, err
	}
	client, err := sftp.NewClient(sshClient)
	if err != nil {
		return false, err
	}
	defer client.Close()
	defer sshClient.Close()

	if err := client.Remove(filePath); err != nil {
		return false, err
	}
	return true, nil
}
