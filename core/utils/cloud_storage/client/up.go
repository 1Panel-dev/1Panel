package client

import (
	"fmt"
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
		Bucket:    bucket,
		Operator:  operator,
		Password:  password,
		UserAgent: "1panel-son.test.upcdn.net",
	})

	return &upClient{bucket: bucket, client: client}, nil
}

func (o upClient) ListBuckets() ([]interface{}, error) {
	var result []interface{}
	return result, nil
}

func (s upClient) Upload(src, target string) (bool, error) {
	if _, err := s.client.GetInfo(path.Dir(target)); err != nil {
		if err := s.client.Mkdir(path.Dir(target)); err != nil {
			fmt.Println(err)
		}
	}
	if err := s.client.Put(&upyun.PutObjectConfig{
		Path:      target,
		LocalPath: src,
	}); err != nil {
		return false, err
	}
	return true, nil
}
