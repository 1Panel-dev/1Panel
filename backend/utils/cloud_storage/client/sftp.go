package client

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"path"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type sftpClient struct {
	bucket   string
	connInfo string
	config   *ssh.ClientConfig
}

func NewSftpClient(vars map[string]interface{}) (*sftpClient, error) {
	address := loadParamFromVars("address", true, vars)
	port := loadParamFromVars("port", false, vars)
	password := loadParamFromVars("password", true, vars)
	username := loadParamFromVars("username", true, vars)
	bucket := loadParamFromVars("bucket", true, vars)

	auth := []ssh.AuthMethod{ssh.Password(password)}
	clientConfig := &ssh.ClientConfig{
		User:    username,
		Auth:    auth,
		Timeout: 30 * time.Second,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}

	return &sftpClient{bucket: bucket, connInfo: fmt.Sprintf("%s:%s", address, port), config: clientConfig}, nil
}

func (s sftpClient) Upload(src, target string) (bool, error) {
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

	srcFile, err := os.Open(src)
	if err != nil {
		return false, err
	}
	defer srcFile.Close()

	targetFilePath := s.bucket + "/" + target
	targetDir, _ := path.Split(targetFilePath)
	if _, err = client.Stat(targetDir); err != nil {
		if os.IsNotExist(err) {
			if err = client.MkdirAll(targetDir); err != nil {
				return false, err
			}
		} else {
			return false, err
		}
	}
	dstFile, err := client.Create(targetFilePath)
	if err != nil {
		return false, err
	}
	defer dstFile.Close()

	reader := bufio.NewReaderSize(srcFile, 128*1024*1024)
	for {
		chunk, err := reader.Peek(8 * 1024 * 1024)
		if len(chunk) != 0 {
			_, _ = dstFile.Write(chunk)
			_, _ = reader.Discard(len(chunk))
		}
		if err != nil {
			break
		}
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
	srcFile, err := client.Open(s.bucket + "/" + src)
	if err != nil {
		return false, err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(target)
	if err != nil {
		return false, err
	}
	defer dstFile.Close()

	if _, err = srcFile.WriteTo(dstFile); err != nil {
		return false, err
	}
	return true, err
}

func (s sftpClient) Exist(path string) (bool, error) {
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

	srcFile, err := client.Open(s.bucket + "/" + path)
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

func (s sftpClient) Size(path string) (int64, error) {
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

	files, err := client.Stat(s.bucket + "/" + path)
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

	targetFilePath := s.bucket + "/" + filePath
	if err := client.Remove(targetFilePath); err != nil {
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

	files, err := client.ReadDir(s.bucket + "/" + prefix)
	if err != nil {
		return nil, err
	}
	var result []string
	for _, file := range files {
		result = append(result, file.Name())
	}
	return result, nil
}
