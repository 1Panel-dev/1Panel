package client

import (
	"context"
	"os"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type s3Client struct {
	scType string
	bucket string
	client *minio.Client
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

	endpoint = strings.TrimPrefix(endpoint, "https://")
	endpoint = strings.TrimPrefix(endpoint, "http://")

	client, err := minio.New(endpoint, &minio.Options{
		Creds:        credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure:       false,
		Region:       region,
		BucketLookup: lookupStyle,
	})
	if err != nil {
		return nil, err
	}
	return &s3Client{scType: scType, bucket: bucket, client: client}, nil
}

func (s s3Client) ListBuckets() ([]interface{}, error) {
	buckets, err := s.client.ListBuckets(context.Background())
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
	_, err := s.client.StatObject(context.Background(), s.bucket, path, minio.StatObjectOptions{})
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
	info, err := s.client.StatObject(context.Background(), s.bucket, path, minio.StatObjectOptions{})
	if err != nil {
		return 0, err
	}
	return info.Size, nil
}

func (s s3Client) Delete(path string) (bool, error) {
	if err := s.client.RemoveObject(context.Background(), s.bucket, path, minio.RemoveObjectOptions{}); err != nil {
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

	if _, err := s.client.PutObject(context.Background(), s.bucket, target, file, fileInfo.Size(), opts); err != nil {
		return false, err
	}
	return true, nil
}

func (s s3Client) Download(src, target string) (bool, error) {
	if _, err := os.Stat(target); err == nil {
		_ = os.Remove(target)
	}
	if err := s.client.FGetObject(context.Background(), s.bucket, src, target, minio.GetObjectOptions{}); err != nil {
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
	for object := range s.client.ListObjects(context.Background(), s.bucket, opts) {
		if object.Err != nil {
			return result, object.Err
		}
		result = append(result, object.Key)
	}
	return result, nil
}
