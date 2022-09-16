package client

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/1Panel-dev/1Panel/constant"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type sftpClient struct {
	Vars map[string]interface{}
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
	return &sftpClient{
		Vars: vars,
	}, nil
}

func (s sftpClient) Upload(src, target string) (bool, error) {
	bucket, err := s.getBucket()
	if err != nil {
		return false, err
	}
	port, err := strconv.Atoi(strconv.FormatFloat(s.Vars["port"].(float64), 'G', -1, 64))
	if err != nil {
		return false, err
	}
	sftpC, err := connect(s.Vars["username"].(string), s.Vars["password"].(string), s.Vars["address"].(string), port)
	if err != nil {
		return false, err
	}
	defer sftpC.Close()
	srcFile, err := os.Open(src)
	if err != nil {
		return false, err
	}
	defer srcFile.Close()

	targetFilePath := bucket + "/" + target
	remotePath, _ := path.Split(targetFilePath)
	_, err = sftpC.Stat(remotePath)
	if err != nil {
		if os.IsNotExist(err) {
			err = sftpC.MkdirAll(remotePath)
			if err != nil {
				return false, err
			}
		} else {
			return false, err
		}
	}

	dstFile, err := sftpC.Create(targetFilePath)
	if err != nil {
		return false, err
	}
	defer dstFile.Close()
	ff, err := ioutil.ReadAll(srcFile)
	if err != nil {
		return false, err
	}
	_, _ = dstFile.Write(ff)
	return true, nil
}

func (s sftpClient) ListBuckets() ([]interface{}, error) {
	var result []interface{}
	return result, nil
}

func (s sftpClient) Download(src, target string) (bool, error) {
	bucket, err := s.getBucket()
	if err != nil {
		return false, err
	}
	port, err := strconv.Atoi(strconv.FormatFloat(s.Vars["port"].(float64), 'G', -1, 64))
	if err != nil {
		return false, err
	}
	sftpC, err := connect(s.Vars["username"].(string), s.Vars["password"].(string), s.Vars["address"].(string), port)
	if err != nil {
		return false, err
	}
	defer sftpC.Close()
	srcFile, err := sftpC.Open(bucket + "/" + src)
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
	bucket, err := s.getBucket()
	if err != nil {
		return false, err
	}
	port, err := strconv.Atoi(strconv.FormatFloat(s.Vars["port"].(float64), 'G', -1, 64))
	if err != nil {
		return false, err
	}
	sftpC, err := connect(s.Vars["username"].(string), s.Vars["password"].(string), s.Vars["address"].(string), port)
	if err != nil {
		return false, err
	}
	defer sftpC.Close()
	srcFile, err := sftpC.Open(bucket + "/" + path)
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
	bucket, err := s.getBucket()
	if err != nil {
		return false, err
	}
	port, err := strconv.Atoi(strconv.FormatFloat(s.Vars["port"].(float64), 'G', -1, 64))
	if err != nil {
		return false, err
	}
	sftpC, err := connect(s.Vars["username"].(string), s.Vars["password"].(string), s.Vars["address"].(string), port)
	if err != nil {
		return false, err
	}
	defer sftpC.Close()
	targetFilePath := bucket + "/" + filePath
	err = sftpC.Remove(targetFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return true, nil
		} else {
			return false, err
		}
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

func (s sftpClient) getBucket() (string, error) {
	if _, ok := s.Vars["bucket"]; ok {
		return s.Vars["bucket"].(string), nil
	} else {
		return "", constant.ErrInvalidParams
	}
}
