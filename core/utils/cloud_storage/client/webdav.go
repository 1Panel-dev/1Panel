package client

import (
	"crypto/tls"
	"fmt"
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
