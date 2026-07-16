package client

import (
	"context"
	"net/http"
	"path"

	"github.com/upyun/go-sdk/upyun"
)

type upClient struct {
	bucket string
	client *upyun.UpYun
}

func NewUpClient(vars map[string]interface{}) (*upClient, error) {
	operator := loadParamFromVars("operator", vars)
	password := loadParamFromVars("password", vars)
	bucket := loadParamFromVars("bucket", vars)
	client := upyun.NewUpYun(&upyun.UpYunConfig{
		Bucket:   bucket,
		Operator: operator,
		Password: password,
	})

	return &upClient{bucket: bucket, client: client}, nil
}

func (o upClient) ListBuckets() ([]interface{}, error) {
	var result []interface{}
	return result, nil
}

func (s upClient) Upload(ctx context.Context, src, target string) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	client := s.client
	if _, ok := ctx.Deadline(); ok {
		client = upyun.NewUpYun(&s.client.UpYunConfig)
		client.SetHTTPClient(&http.Client{Transport: deadlineTransport{ctx: ctx, base: http.DefaultTransport}})
	}
	if _, err := client.GetInfo(path.Dir(src)); err != nil {
		_ = client.Mkdir(path.Dir(target))
	}
	if err := client.Put(&upyun.PutObjectConfig{
		Path:            target,
		LocalPath:       src,
		UseResumeUpload: true,
	}); err != nil {
		return false, err
	}
	return true, nil
}

func (s upClient) Size(path string) (int64, error) {
	fileInfo, err := s.client.GetInfo(path)
	if err != nil {
		return 0, err
	}
	return fileInfo.Size, nil
}

func (s upClient) Delete(path string) (bool, error) {
	if err := s.client.Delete(&upyun.DeleteObjectConfig{
		Path: path,
	}); err != nil {
		return false, err
	}
	return true, nil
}

func (s upClient) Exist(filePath string) (bool, error) {
	if _, err := s.client.GetInfo(filePath); err != nil {
		return false, err
	}
	return true, nil
}

func (s upClient) Download(src, target string) (bool, error) {
	if _, err := s.client.Get(&upyun.GetObjectConfig{
		Path:      src,
		LocalPath: target,
	}); err != nil {
		return false, err
	}
	return true, nil
}

func (s *upClient) ListObjects(prefix string) ([]string, error) {
	objsChan := make(chan *upyun.FileInfo, 10)
	if err := s.client.List(&upyun.GetObjectsConfig{
		Path:         prefix,
		ObjectsChan:  objsChan,
		MaxListTries: 1,
	}); err != nil {
		return nil, err
	}
	var files []string
	for obj := range objsChan {
		files = append(files, obj.Name)
	}
	return files, nil
}
