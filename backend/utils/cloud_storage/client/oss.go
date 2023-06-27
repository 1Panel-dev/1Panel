package client

import (
	"github.com/1Panel-dev/1Panel/backend/constant"
	osssdk "github.com/aliyun/aliyun-oss-go-sdk/oss"
)

type ossClient struct {
	scType string
	Vars   map[string]interface{}
	client osssdk.Client
}

func NewOssClient(vars map[string]interface{}) (*ossClient, error) {
	var endpoint string
	var accessKey string
	var secretKey string
	var scType string
	if _, ok := vars["endpoint"]; ok {
		endpoint = vars["endpoint"].(string)
	} else {
		return nil, constant.ErrInvalidParams
	}
	if _, ok := vars["accessKey"]; ok {
		accessKey = vars["accessKey"].(string)
	} else {
		return nil, constant.ErrInvalidParams
	}
	if _, ok := vars["scType"]; ok {
		scType = vars["scType"].(string)
	} else {
		scType = "Standard"
	}
	if _, ok := vars["secretKey"]; ok {
		secretKey = vars["secretKey"].(string)
	} else {
		return nil, constant.ErrInvalidParams
	}
	client, err := osssdk.New(endpoint, accessKey, secretKey)
	if err != nil {
		return nil, err
	}
	return &ossClient{
		scType: scType,
		Vars:   vars,
		client: *client,
	}, nil
}

func (oss ossClient) ListBuckets() ([]interface{}, error) {
	response, err := oss.client.ListBuckets()
	if err != nil {
		return nil, err
	}
	var result []interface{}
	for _, bucket := range response.Buckets {
		result = append(result, bucket.Name)
	}
	return result, err
}

func (oss ossClient) Exist(path string) (bool, error) {
	bucket, err := oss.GetBucket()
	if err != nil {
		return false, err
	}
	return bucket.IsObjectExist(path)

}

func (oss ossClient) Delete(path string) (bool, error) {
	bucket, err := oss.GetBucket()
	if err != nil {
		return false, err
	}
	err = bucket.DeleteObject(path)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (oss ossClient) Upload(src, target string) (bool, error) {
	bucket, err := oss.GetBucket()
	if err != nil {
		return false, err
	}
	err = bucket.UploadFile(target, src, 200*1024*1024, osssdk.Routines(5), osssdk.Checkpoint(true, ""), osssdk.ObjectStorageClass(osssdk.StorageClassType(oss.scType)))
	if err != nil {
		return false, err
	}
	return true, nil
}

func (oss ossClient) Download(src, target string) (bool, error) {
	bucket, err := oss.GetBucket()
	if err != nil {
		return false, err
	}
	err = bucket.DownloadFile(src, target, 200*1024*1024, osssdk.Routines(5), osssdk.Checkpoint(true, ""))
	if err != nil {
		return false, err
	}
	return true, nil
}

func (oss *ossClient) GetBucket() (*osssdk.Bucket, error) {
	if _, ok := oss.Vars["bucket"]; ok {
		bucket, err := oss.client.Bucket(oss.Vars["bucket"].(string))
		if err != nil {
			return nil, err
		}
		return bucket, nil
	} else {
		return nil, constant.ErrInvalidParams
	}
}

func (oss *ossClient) ListObjects(prefix string) ([]interface{}, error) {
	bucket, err := oss.GetBucket()
	if err != nil {
		return nil, constant.ErrInvalidParams
	}
	lor, err := bucket.ListObjects(osssdk.Prefix(prefix))
	if err != nil {
		return nil, err
	}
	var result []interface{}
	for _, obj := range lor.Objects {
		result = append(result, obj.Key)
	}
	return result, nil
}
