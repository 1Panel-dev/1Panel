package client

import (
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type s3Client struct {
	scType string
	bucket string
	Sess   session.Session
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
	sess, err := session.NewSession(&aws.Config{
		Credentials:      credentials.NewStaticCredentials(accessKey, secretKey, ""),
		Endpoint:         aws.String(endpoint),
		Region:           aws.String(region),
		DisableSSL:       aws.Bool(true),
		S3ForcePathStyle: aws.Bool(mode == "path"),
	})
	if err != nil {
		return nil, err
	}
	return &s3Client{scType: scType, bucket: bucket, Sess: *sess}, nil
}

func (s s3Client) ListBuckets() ([]interface{}, error) {
	var result []interface{}
	svc := s3.New(&s.Sess)
	res, err := svc.ListBuckets(nil)
	if err != nil {
		return nil, err
	}
	for _, b := range res.Buckets {
		result = append(result, b.Name)
	}
	return result, nil
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

	uploader := s3manager.NewUploader(&s.Sess)
	if fileInfo.Size() > s3manager.MaxUploadParts*s3manager.DefaultUploadPartSize {
		uploader.PartSize = fileInfo.Size() / (s3manager.MaxUploadParts - 1)
	}
	if _, err := uploader.Upload(&s3manager.UploadInput{
		Bucket:       aws.String(s.bucket),
		Key:          aws.String(target),
		Body:         file,
		StorageClass: &s.scType,
	}); err != nil {
		return false, err
	}
	return true, nil
}

func (s s3Client) Delete(path string) (bool, error) {
	svc := s3.New(&s.Sess)
	if _, err := svc.DeleteObject(&s3.DeleteObjectInput{Bucket: aws.String(s.bucket), Key: aws.String(path)}); err != nil {
		return false, err
	}
	if err := svc.WaitUntilObjectNotExists(&s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(path),
	}); err != nil {
		return false, err
	}
	return true, nil
}
