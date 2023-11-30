package client

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type sftpClient struct {
	Bucket string
	Vars   map[string]interface{}
	client *sftp.Client
}

func NewSftpClient(vars map[string]interface{}) (*sftpClient, error) {
	if _, ok := vars["address"]; !ok {
		return nil, constant.ErrInvalidParams
	}
	if _, ok := vars["port"].(float64); !ok {
		return nil, constant.ErrInvalidParams
	}
	if _, ok := vars["password"]; !ok {
		return nil, constant.ErrInvalidParams
	}
	if _, ok := vars["username"]; !ok {
		return nil, constant.ErrInvalidParams
	}
	var bucket string
	if _, ok := vars["bucket"]; ok {
		bucket = vars["bucket"].(string)
	} else {
		return nil, constant.ErrInvalidParams
	}

	port, err := strconv.Atoi(strconv.FormatFloat(vars["port"].(float64), 'G', -1, 64))
	if err != nil {
		return nil, err
	}
	sftpC, err := connect(vars["username"].(string), vars["password"].(string), vars["address"].(string), port)
	if err != nil {
		return nil, err
	}
	return &sftpClient{Bucket: bucket, client: sftpC, Vars: vars}, nil
}

func (s sftpClient) Upload(src, target string) (bool, error) {
	defer s.client.Close()

	srcFile, err := os.Open(src)
	if err != nil {
		return false, err
	}
	defer srcFile.Close()

	targetFilePath := s.Bucket + "/" + target
	targetDir, _ := path.Split(targetFilePath)
	if _, err = s.client.Stat(targetDir); err != nil {
		if os.IsNotExist(err) {
			if err = s.client.MkdirAll(targetDir); err != nil {
				return false, err
			}
		} else {
			return false, err
		}
	}
	dstFile, err := s.client.Create(targetFilePath)
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
	defer s.client.Close()
	srcFile, err := s.client.Open(s.Bucket + "/" + src)
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
	defer s.client.Close()
	srcFile, err := s.client.Open(s.Bucket + "/" + path)
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

func (s sftpClient) Delete(filePath string) (bool, error) {
	defer s.client.Close()
	targetFilePath := s.Bucket + "/" + filePath
	if err := s.client.Remove(targetFilePath); err != nil {
		return false, err
	}
	return true, nil
}

func connect(user, password, host string, port int) (*sftp.Client, error) {
	var (
		auth         []ssh.AuthMethod
		addr         string
		clientConfig *ssh.ClientConfig
		sshClient    *ssh.Client
		sftpClient   *sftp.Client
		err          error
	)
	auth = make([]ssh.AuthMethod, 0)
	auth = append(auth, ssh.Password(password))
	clientConfig = &ssh.ClientConfig{
		User:    user,
		Auth:    auth,
		Timeout: 30 * time.Second,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}
	addr = fmt.Sprintf("%s:%d", host, port)

	if sshClient, err = ssh.Dial("tcp", addr, clientConfig); err != nil {
		return nil, err
	}
	if sftpClient, err = sftp.NewClient(sshClient); err != nil {
		return nil, err
	}
	return sftpClient, nil
}

func (s sftpClient) ListObjects(prefix string) ([]string, error) {
	defer s.client.Close()
	files, err := s.client.ReadDir(s.Bucket + "/" + prefix)
	if err != nil {
		return nil, err
	}
	var result []string
	for _, file := range files {
		result = append(result, file.Name())
	}
	return result, nil
}
