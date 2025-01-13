package client

import (
	"fmt"
	"io"
	"net"
	"os"
	"path"
	"time"

	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type sftpClient struct {
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
	if _, err := ssh.Dial("tcp", fmt.Sprintf("%s:%s", address, port), clientConfig); err != nil {
		return nil, err
	}

	return &sftpClient{connInfo: fmt.Sprintf("%s:%s", address, port), config: clientConfig}, nil
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

	targetDir, _ := path.Split(target)
	if _, err = client.Stat(targetDir); err != nil {
		if os.IsNotExist(err) {
			if err = client.MkdirAll(targetDir); err != nil {
				return false, err
			}
		} else {
			return false, err
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

func (s sftpClient) ListBuckets() ([]interface{}, error) {
	var result []interface{}
	return result, nil
}

func (s sftpClient) Download(src, target string) (bool, error) {
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

	srcFile, err := client.Open(src)
	if err != nil {
		return false, err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(target)
	if err != nil {
		return false, err
	}
	defer dstFile.Close()

	if _, err = io.Copy(dstFile, srcFile); err != nil {
		return false, err
	}
	return true, err
}

func (s sftpClient) Exist(filePath string) (bool, error) {
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

	srcFile, err := client.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		} else {
			return false, err
		}
	}
	defer srcFile.Close()
	return true, err
}

func (s sftpClient) Size(filePath string) (int64, error) {
	sshClient, err := ssh.Dial("tcp", s.connInfo, s.config)
	if err != nil {
		return 0, err
	}
	client, err := sftp.NewClient(sshClient)
	if err != nil {
		return 0, err
	}
	defer client.Close()
	defer sshClient.Close()

	files, err := client.Stat(filePath)
	if err != nil {
		return 0, err
	}
	return files.Size(), nil
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

func (s sftpClient) ListObjects(prefix string) ([]string, error) {
	sshClient, err := ssh.Dial("tcp", s.connInfo, s.config)
	if err != nil {
		return nil, err
	}
	client, err := sftp.NewClient(sshClient)
	if err != nil {
		return nil, err
	}
	defer client.Close()
	defer sshClient.Close()

	files, err := client.ReadDir(prefix)
	if err != nil {
		return nil, err
	}
	var result []string
	for _, file := range files {
		result = append(result, file.Name())
	}
	return result, nil
}
