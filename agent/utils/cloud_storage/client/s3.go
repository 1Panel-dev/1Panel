package client

import (
	"context"
	"errors"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
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

func (s *s3Client) Exist(path string) (bool, error) {
	_, err := s.client.HeadObject(context.Background(), &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(path),
	})
	if err != nil {
		var apiErr smithy.APIError
		if errors.As(err, &apiErr) {
			switch apiErr.ErrorCode() {
			case "NotFound", "NoSuchKey":
				return false, nil
			}
		}
		return false, err
	}
	return true, nil
}

func (s *s3Client) Size(path string) (int64, error) {
	file, err := s.client.GetObject(context.Background(), &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(path),
	})
	if err != nil {
		return 0, err
	}
	defer file.Body.Close()
	return aws.ToInt64(file.ContentLength), nil
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

func (s *s3Client) Download(src, target string) (bool, error) {
	if _, err := os.Stat(target); err == nil {
		_ = os.Remove(target)
	}
	file, err := os.Create(target)
	if err != nil {
		return false, err
	}
	defer file.Close()
	downloader := manager.NewDownloader(s.client)
	if _, err = downloader.Download(context.Background(), file, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(src),
	}); err != nil {
		os.Remove(target)
		return false, err
	}
	return true, nil
}

func (s *s3Client) ListObjects(prefix string) ([]string, error) {
	var result []string
	outputs, err := s.client.ListObjectsV2(context.Background(), &s3.ListObjectsV2Input{
		Bucket: aws.String(s.bucket),
		Prefix: aws.String(prefix),
	})
	if err != nil {
		return result, err
	}
	for _, item := range outputs.Contents {
		result = append(result, aws.ToString(item.Key))
	}
	return result, nil
}

func normalizeEndpoint(endpoint string) string {
	if strings.HasPrefix(endpoint, "http://") || strings.HasPrefix(endpoint, "https://") {
		return endpoint
	}
	return "http://" + endpoint
}
