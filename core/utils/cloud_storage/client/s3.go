package client

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type s3Client struct {
	scType string
	bucket string
	client *s3.Client
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
	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
	)
	if err != nil {
		return nil, err
	}
	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = mode == "path"
		if endpoint != "" {
			o.BaseEndpoint = aws.String(normalizeEndpoint(endpoint))
		}
	})
	return &s3Client{scType: scType, bucket: bucket, client: client}, nil
}

func (s *s3Client) ListBuckets() ([]interface{}, error) {
	var result []interface{}
	res, err := s.client.ListBuckets(context.Background(), &s3.ListBucketsInput{})
	if err != nil {
		return nil, err
	}
	for _, b := range res.Buckets {
		result = append(result, b.Name)
	}
	return result, nil
}

func (s *s3Client) Upload(src, target string) (bool, error) {
	fileInfo, err := os.Stat(src)
	if err != nil {
		return false, err
	}
	file, err := os.Open(src)
	if err != nil {
		return false, err
	}
	defer file.Close()

	uploader := manager.NewUploader(s.client)
	maxUploadSize := int64(manager.MaxUploadParts) * manager.DefaultUploadPartSize
	if fileInfo.Size() > maxUploadSize {
		uploader.PartSize = fileInfo.Size() / (int64(manager.MaxUploadParts) - 1)
	}
	if _, err := uploader.Upload(context.Background(), &s3.PutObjectInput{
		Bucket:       aws.String(s.bucket),
		Key:          aws.String(target),
		Body:         file,
		StorageClass: types.StorageClass(s.scType),
	}); err != nil {
		return false, err
	}
	return true, nil
}

func (s *s3Client) Delete(path string) (bool, error) {
	if _, err := s.client.DeleteObject(context.Background(), &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(path),
	}); err != nil {
		return false, err
	}
	waiter := s3.NewObjectNotExistsWaiter(s.client)
	if err := waiter.Wait(context.Background(), &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(path),
	}, 30*time.Second); err != nil {
		return false, err
	}
	return true, nil
}

func normalizeEndpoint(endpoint string) string {
	if strings.HasPrefix(endpoint, "http://") || strings.HasPrefix(endpoint, "https://") {
		return endpoint
	}
	return "http://" + endpoint
}
