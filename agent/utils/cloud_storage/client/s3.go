package client

import (
	"context"
	"crypto/tls"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

const (
	s3DefaultTimeout  = 30 * time.Second
	s3TransferTimeout = 24 * time.Hour
)

type s3Client struct {
	scType string
	bucket string
	client *minio.Client
}

func (s *s3Client) ctx(timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), timeout)
}

func NewS3Client(vars map[string]interface{}) (*s3Client, error) {
	accessKey := loadParamFromVars("accessKey", vars)
	secretKey := loadParamFromVars("secretKey", vars)
	endpoint := loadParamFromVars("endpoint", vars)
	region := loadParamFromVars("region", vars)
	bucket := loadParamFromVars("bucket", vars)
	scType := loadParamFromVars("scType", vars)
	if len(scType) == 0 {
		scType = "Standard"
	}
	mode := loadParamFromVars("mode", vars)
	if len(mode) == 0 {
		mode = "virtual hosted"
	}

	lookupStyle := minio.BucketLookupDNS
	if mode == "path" {
		lookupStyle = minio.BucketLookupPath
	}

	ssl := strings.Split(endpoint, ":")[0]
	secure := false
	tlsConfig := &tls.Config{}
	if ssl == "https" {
		secure = true
		tlsConfig.InsecureSkipVerify = true
	}
	var transport http.RoundTripper = &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	endpoint = strings.TrimPrefix(endpoint, ssl+"://")

	client, err := minio.New(endpoint, &minio.Options{
		Creds:        credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure:       secure,
		Region:       region,
		BucketLookup: lookupStyle,
		Transport:    transport,
	})
	if err != nil {
		return nil, err
	}
	return &s3Client{scType: scType, bucket: bucket, client: client}, nil
}

func (s s3Client) ListBuckets() ([]interface{}, error) {
	ctx, cancel := s.ctx(s3DefaultTimeout)
	defer cancel()
	buckets, err := s.client.ListBuckets(ctx)
	if err != nil {
		return nil, err
	}
	var result []interface{}
	for _, b := range buckets {
		result = append(result, b.Name)
	}
	return result, nil
}

func (s s3Client) Exist(path string) (bool, error) {
	ctx, cancel := s.ctx(s3DefaultTimeout)
	defer cancel()
	_, err := s.client.StatObject(ctx, s.bucket, path, minio.StatObjectOptions{})
	if err != nil {
		resp := minio.ToErrorResponse(err)
		if resp.StatusCode == 404 {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (s *s3Client) Size(path string) (int64, error) {
	ctx, cancel := s.ctx(s3DefaultTimeout)
	defer cancel()
	info, err := s.client.StatObject(ctx, s.bucket, path, minio.StatObjectOptions{})
	if err != nil {
		return 0, err
	}
	return info.Size, nil
}

func (s s3Client) Delete(path string) (bool, error) {
	ctx, cancel := s.ctx(s3DefaultTimeout)
	defer cancel()
	if err := s.client.RemoveObject(ctx, s.bucket, path, minio.RemoveObjectOptions{}); err != nil {
		return false, err
	}
	return true, nil
}

func (s s3Client) Upload(src, target string) (bool, error) {
	fileInfo, err := os.Stat(src)
	if err != nil {
		return false, err
	}
	file, err := os.Open(src)
	if err != nil {
		return false, err
	}
	defer file.Close()

	opts := minio.PutObjectOptions{
		StorageClass: s.scType,
	}

	const maxParts = 10000
	const defaultPartSize = 64 * 1024 * 1024 // 64 MiB
	partSize := uint64(defaultPartSize)
	if fileInfo.Size() > int64(maxParts)*int64(defaultPartSize) {
		partSize = uint64(fileInfo.Size()) / (maxParts - 1)
	}
	opts.PartSize = partSize

	ctx, cancel := s.ctx(s3TransferTimeout)
	defer cancel()
	if _, err := s.client.PutObject(ctx, s.bucket, target, file, fileInfo.Size(), opts); err != nil {
		return false, err
	}
	return true, nil
}

func (s s3Client) Download(src, target string) (bool, error) {
	if _, err := os.Stat(target); err == nil {
		_ = os.Remove(target)
	}
	ctx, cancel := s.ctx(s3TransferTimeout)
	defer cancel()
	if err := s.client.FGetObject(ctx, s.bucket, src, target, minio.GetObjectOptions{}); err != nil {
		os.Remove(target)
		return false, err
	}
	return true, nil
}

func (s *s3Client) ListObjects(prefix string) ([]string, error) {
	opts := minio.ListObjectsOptions{
		Recursive: true,
		Prefix:    prefix,
	}
	var result []string
	ctx, cancel := s.ctx(s3DefaultTimeout)
	defer cancel()
	for object := range s.client.ListObjects(ctx, s.bucket, opts) {
		if object.Err != nil {
			return result, object.Err
		}
		result = append(result, object.Key)
	}
	return result, nil
}
