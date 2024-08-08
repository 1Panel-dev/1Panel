package client

import (
	"context"
	"crypto/tls"
	"net/http"
	"os"
	"strings"

	"github.com/1Panel-dev/1Panel/core/constant"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type minIoClient struct {
	bucket string
	client *minio.Client
}

func NewMinIoClient(vars map[string]interface{}) (*minIoClient, error) {
	endpoint := loadParamFromVars("endpoint", vars)
	accessKeyID := loadParamFromVars("accessKey", vars)
	secretAccessKey := loadParamFromVars("secretKey", vars)
	bucket := loadParamFromVars("bucket", vars)
	ssl := strings.Split(endpoint, ":")[0]
	if len(ssl) == 0 || (ssl != "https" && ssl != "http") {
		return nil, constant.ErrInvalidParams
	}

	secure := false
	tlsConfig := &tls.Config{}
	if ssl == "https" {
		secure = true
		tlsConfig.InsecureSkipVerify = true
	}
	var transport http.RoundTripper = &http.Transport{
		TLSClientConfig: tlsConfig,
	}
	client, err := minio.New(strings.ReplaceAll(endpoint, ssl+"://", ""), &minio.Options{
		Creds:     credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure:    secure,
		Transport: transport,
	})
	if err != nil {
		return nil, err
	}
	return &minIoClient{bucket: bucket, client: client}, nil
}

func (m minIoClient) ListBuckets() ([]interface{}, error) {
	buckets, err := m.client.ListBuckets(context.Background())
	if err != nil {
		return nil, err
	}
	var result []interface{}
	for _, bucket := range buckets {
		result = append(result, bucket.Name)
	}
	return result, err
}

func (m minIoClient) Upload(src, target string) (bool, error) {
	file, err := os.Open(src)
	if err != nil {
		return false, err
	}
	defer file.Close()

	fileStat, err := file.Stat()
	if err != nil {
		return false, err
	}
	_, err = m.client.PutObject(context.Background(), m.bucket, target, file, fileStat.Size(), minio.PutObjectOptions{ContentType: "application/octet-stream"})
	if err != nil {
		return false, err
	}
	return true, nil
}
