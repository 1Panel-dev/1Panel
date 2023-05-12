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
	Vars   map[string]interface{}
	client *minio.Client
}

func NewMinIoClient(vars map[string]interface{}) (*minIoClient, error) {
	var endpoint string
	var accessKeyID string
	var secretAccessKey string
	if _, ok := vars["endpoint"]; ok {
		endpoint = vars["endpoint"].(string)
	} else {
		return nil, constant.ErrInvalidParams
	}
	if _, ok := vars["accessKey"]; ok {
		accessKeyID = vars["accessKey"].(string)
	} else {
		return nil, constant.ErrInvalidParams
	}
	if _, ok := vars["secretKey"]; ok {
		secretAccessKey = vars["secretKey"].(string)
	} else {
		return nil, constant.ErrInvalidParams
	}
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
	return &minIoClient{
		client: client,
		Vars:   vars,
	}, nil
}

func (minIo minIoClient) ListBuckets() ([]interface{}, error) {
	buckets, err := minIo.client.ListBuckets(context.Background())
	if err != nil {
		return nil, err
	}
	var result []interface{}
	for _, bucket := range buckets {
		result = append(result, bucket.Name)
	}
	return result, err
}

func (minIo minIoClient) Exist(path string) (bool, error) {
	if _, ok := minIo.Vars["bucket"]; ok {
		_, err := minIo.client.GetObject(context.Background(), minIo.Vars["bucket"].(string), path, minio.GetObjectOptions{})
		if err != nil {
			return false, err
		}
		return true, nil
	} else {
		return false, constant.ErrInvalidParams
	}
}

func (minIo minIoClient) Delete(path string) (bool, error) {
	if _, ok := minIo.Vars["bucket"]; ok {
		object, err := minIo.client.GetObject(context.Background(), minIo.Vars["bucket"].(string), path, minio.GetObjectOptions{})
		if err != nil {
			return false, err
		}
		info, err := object.Stat()
		if err != nil {
			return false, err
		}
		err = minIo.client.RemoveObject(context.Background(), minIo.Vars["bucket"].(string), path, minio.RemoveObjectOptions{
			GovernanceBypass: true,
			VersionID:        info.VersionID,
		})
		if err != nil {
			return false, err
		}
		return true, nil
	} else {
		return false, constant.ErrInvalidParams
	}
}

func (minIo minIoClient) Upload(src, target string) (bool, error) {
	var bucket string
	if _, ok := minIo.Vars["bucket"]; ok {
		bucket = minIo.Vars["bucket"].(string)
	} else {
		return false, constant.ErrInvalidParams
	}

	file, err := os.Open(src)
	if err != nil {
		return false, err
	}
	defer file.Close()

	fileStat, err := file.Stat()
	if err != nil {
		return false, err
	}
	_, err = minIo.client.PutObject(context.Background(), bucket, target, file, fileStat.Size(), minio.PutObjectOptions{ContentType: "application/octet-stream"})
	if err != nil {
		return false, err
	}
	return true, nil
}

func (minIo minIoClient) Download(src, target string) (bool, error) {
	if _, ok := minIo.Vars["bucket"]; ok {
		object, err := minIo.client.GetObject(context.Background(), minIo.Vars["bucket"].(string), src, minio.GetObjectOptions{})
		if err != nil {
			return false, err
		}
		localFile, err := os.Create(target)
		if err != nil {
			return false, err
		}
		if _, err = io.Copy(localFile, object); err != nil {
			return false, err
		}
		return true, nil
	} else {
		return false, constant.ErrInvalidParams
	}
}

func (minIo *minIoClient) GetBucket() (string, error) {
	if _, ok := minIo.Vars["bucket"]; ok {
		return minIo.Vars["bucket"].(string), nil
	} else {
		return "", constant.ErrInvalidParams
	}
}

func (minIo minIoClient) ListObjects(prefix string) ([]interface{}, error) {
	bucket, err := minIo.GetBucket()
	if err != nil {
		return nil, constant.ErrInvalidParams
	}
	opts := minio.ListObjectsOptions{
		Recursive: true,
		Prefix:    prefix,
	}

	var result []interface{}
	for object := range minIo.client.ListObjects(context.Background(), bucket, opts) {
		if object.Err != nil {
			continue
		}
		result = append(result, object.Key)
	}
	return result, nil
}
