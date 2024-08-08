package client

import (
	"fmt"

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

func (o ossClient) Exist(path string) (bool, error) {
	bucket, err := o.client.Bucket(o.bucketStr)
	if err != nil {
		return false, err
	}
	return bucket.IsObjectExist(path)
}

func (o ossClient) Size(path string) (int64, error) {
	bucket, err := o.client.Bucket(o.bucketStr)
	if err != nil {
		return 0, err
	}
	lor, err := bucket.ListObjectsV2(osssdk.Prefix(path))
	if err != nil {
		return 0, err
	}
	if len(lor.Objects) == 0 {
		return 0, fmt.Errorf("no such file %s", path)
	}
	return lor.Objects[0].Size, nil
}

func (o ossClient) Delete(path string) (bool, error) {
	bucket, err := o.client.Bucket(o.bucketStr)
	if err != nil {
		return false, err
	}
	if err := bucket.DeleteObject(path); err != nil {
		return false, err
	}
	return true, nil
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

func (o ossClient) Download(src, target string) (bool, error) {
	bucket, err := o.client.Bucket(o.bucketStr)
	if err != nil {
		return false, err
	}
	if err := bucket.DownloadFile(src, target, 200*1024*1024, osssdk.Routines(5), osssdk.Checkpoint(true, "")); err != nil {
		return false, err
	}
	return true, nil
}

func (o *ossClient) ListObjects(prefix string) ([]string, error) {
	bucket, err := o.client.Bucket(o.bucketStr)
	if err != nil {
		return nil, err
	}
	lor, err := bucket.ListObjectsV2(osssdk.Prefix(prefix))
	if err != nil {
		return nil, err
	}
	var result []string
	for _, obj := range lor.Objects {
		result = append(result, obj.Key)
	}
	return result, nil
}
