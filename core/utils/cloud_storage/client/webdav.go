package client

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/studio-b12/gowebdav"
)

type webDAVClient struct {
	Bucket string
	client *gowebdav.Client
}

func NewWebDAVClient(vars map[string]interface{}) (*webDAVClient, error) {
	address := loadParamFromVars("address", vars)
	port := loadParamFromVars("port", vars)
	password := loadParamFromVars("password", vars)
	username := loadParamFromVars("username", vars)
	bucket := loadParamFromVars("bucket", vars)

	url := fmt.Sprintf("%s:%s", address, port)
	if len(port) == 0 {
		url = address
	}
	client := gowebdav.NewClient(url, username, password)
	tlsConfig := &tls.Config{}
	if strings.HasPrefix(address, "https") {
		tlsConfig.InsecureSkipVerify = true
	}
	var transport http.RoundTripper = &http.Transport{
		TLSClientConfig: tlsConfig,
	}
	client.SetTransport(transport)
	if err := client.Connect(); err != nil {
		return nil, err
	}
	return &webDAVClient{Bucket: bucket, client: client}, nil
}

func (s webDAVClient) Upload(src, target string) (bool, error) {
	targetFilePath := path.Join(s.Bucket, target)
	srcFile, err := os.Open(src)
	if err != nil {
		return false, err
	}
	defer srcFile.Close()

	if err := s.client.WriteStream(targetFilePath, srcFile, 0644); err != nil {
		return false, err
	}
	return true, nil
}

func (s webDAVClient) ListBuckets() ([]interface{}, error) {
	var result []interface{}
	return result, nil
}

func (s webDAVClient) Download(src, target string) (bool, error) {
	srcPath := path.Join(s.Bucket, src)
	info, err := s.client.Stat(srcPath)
	if err != nil {
		return false, err
	}
	targetStat, err := os.Stat(target)
	if err == nil {
		if info.Size() == targetStat.Size() {
			return true, nil
		}
	}
	file, err := os.Create(target)
	if err != nil {
		return false, err
	}
	defer file.Close()
	reader, _ := s.client.ReadStream(srcPath)
	if _, err := io.Copy(file, reader); err != nil {
		return false, err
	}
	return true, err
}

func (s webDAVClient) Exist(pathItem string) (bool, error) {
	if _, err := s.client.Stat(path.Join(s.Bucket, pathItem)); err != nil {
		return false, err
	}
	return true, nil
}

func (s webDAVClient) Size(pathItem string) (int64, error) {
	file, err := s.client.Stat(path.Join(s.Bucket, pathItem))
	if err != nil {
		return 0, err
	}
	return file.Size(), nil
}

func (s webDAVClient) Delete(pathItem string) (bool, error) {
	if err := s.client.Remove(path.Join(s.Bucket, pathItem)); err != nil {
		return false, err
	}
	return true, nil
}

func (s webDAVClient) ListObjects(prefix string) ([]string, error) {
	files, err := s.client.ReadDir(path.Join(s.Bucket, prefix))
	if err != nil {
		return nil, err
	}
	var result []string
	for _, file := range files {
		result = append(result, file.Name())
	}
	return result, nil
}
