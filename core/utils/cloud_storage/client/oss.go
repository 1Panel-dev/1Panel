package client

import (
	osssdk "github.com/aliyun/aliyun-oss-go-sdk/oss"
)

type ossClient struct {
	scType    string
	bucketStr string
	client    osssdk.Client
}

func NewOssClient(vars map[string]interface{}) (*ossClient, error) {
	endpoint := loadParamFromVars("endpoint", vars)
	accessKey := loadParamFromVars("accessKey", vars)
	secretKey := loadParamFromVars("secretKey", vars)
	bucketStr := loadParamFromVars("bucket", vars)
	scType := loadParamFromVars("scType", vars)
	if len(scType) == 0 {
		scType = "Standard"
	}
	client, err := osssdk.New(endpoint, accessKey, secretKey)
	if err != nil {
		return nil, err
	}

	return &ossClient{scType: scType, bucketStr: bucketStr, client: *client}, nil
}

func (o ossClient) ListBuckets() ([]interface{}, error) {
	response, err := o.client.ListBuckets()
	if err != nil {
		return nil, err
	}
	var result []interface{}
	for _, bucket := range response.Buckets {
		result = append(result, bucket.Name)
	}
	return result, err
}

func (o ossClient) Upload(src, target string) (bool, error) {
	bucket, err := o.client.Bucket(o.bucketStr)
	if err != nil {
		return false, err
	}
	if err := bucket.UploadFile(target, src,
		200*1024*1024,
		osssdk.Routines(5),
		osssdk.Checkpoint(true, ""),
		osssdk.ObjectStorageClass(osssdk.StorageClassType(o.scType))); err != nil {
		return false, err
	}
	return true, nil
}
