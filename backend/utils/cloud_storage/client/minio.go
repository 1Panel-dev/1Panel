package client

import (
	"context"
	"crypto/tls"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/1Panel-dev/1Panel/backend/constant"
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

func (m minIoClient) Exist(path string) (bool, error) {
	if _, err := m.client.GetObject(context.Background(), m.bucket, path, minio.GetObjectOptions{}); err != nil {
		return false, err
	}
	return true, nil
}

func (m minIoClient) Size(path string) (int64, error) {
	obj, err := m.client.GetObject(context.Background(), m.bucket, path, minio.GetObjectOptions{})
	if err != nil {
		return 0, err
	}
	file, err := obj.Stat()
	if err != nil {
		return 0, err
	}
	return file.Size, nil
}

func (m minIoClient) Delete(path string) (bool, error) {
	object, err := m.client.GetObject(context.Background(), m.bucket, path, minio.GetObjectOptions{})
	if err != nil {
		return false, err
	}
	info, err := object.Stat()
	if err != nil {
		return false, err
	}
	if err = m.client.RemoveObject(context.Background(), m.bucket, path, minio.RemoveObjectOptions{
		GovernanceBypass: true,
		VersionID:        info.VersionID,
	}); err != nil {
		return false, err
	}
	return true, nil
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

func (m minIoClient) Download(src, target string) (bool, error) {
	object, err := m.client.GetObject(context.Background(), m.bucket, src, minio.GetObjectOptions{})
	if err != nil {
		return false, err
	}
	defer object.Close()
	localFile, err := os.Create(target)
	if err != nil {
		return false, err
	}
	defer localFile.Close()
	if _, err = io.Copy(localFile, object); err != nil {
		return false, err
	}
	return true, nil
}

func (m minIoClient) ListObjects(prefix string) ([]string, error) {
	opts := minio.ListObjectsOptions{
		Recursive: true,
		Prefix:    prefix,
	}

	var result []string
	for object := range m.client.ListObjects(context.Background(), m.bucket, opts) {
		if object.Err != nil {
			continue
		}
		result = append(result, object.Key)
	}
	return result, nil
}
