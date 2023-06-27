package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/1Panel-dev/1Panel/backend/constant"
	cosSDK "github.com/tencentyun/cos-go-sdk-v5"
)

type cosClient struct {
	region    string
	accessKey string
	secretKey string
	scType    string
	Vars      map[string]interface{}
	client    *cosSDK.Client
}

func NewCosClient(vars map[string]interface{}) (*cosClient, error) {
	var accessKey string
	var secretKey string
	var scType string
	var region string
	if _, ok := vars["region"]; ok {
		region = vars["region"].(string)
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

	u, _ := url.Parse(fmt.Sprintf("https://cos.%s.myqcloud.com", region))
	b := &cosSDK.BaseURL{BucketURL: u}
	client := cosSDK.NewClient(b, &http.Client{
		Transport: &cosSDK.AuthorizationTransport{
			SecretID:  accessKey,
			SecretKey: secretKey,
		},
	})

	return &cosClient{Vars: vars, client: client, accessKey: accessKey, secretKey: secretKey, scType: scType, region: region}, nil
}

func (cos cosClient) ListBuckets() ([]interface{}, error) {
	buckets, _, err := cos.client.Service.Get(context.Background())
	if err != nil {
		return nil, err
	}
	var datas []interface{}
	for _, bucket := range buckets.Buckets {
		datas = append(datas, bucket.Name)
	}
	return datas, nil
}

func (cos cosClient) Exist(path string) (bool, error) {
	client, err := cos.newClientWithBucket()
	if err != nil {
		return false, err
	}
	exist, err := client.Object.IsExist(context.Background(), path)
	if err != nil {
		return false, err
	}
	return exist, nil
}

func (cos cosClient) Delete(path string) (bool, error) {
	client, err := cos.newClientWithBucket()
	if err != nil {
		return false, err
	}
	if _, err := client.Object.Delete(context.Background(), path); err != nil {
		return false, err
	}
	return true, nil
}

func (cos cosClient) Upload(src, target string) (bool, error) {
	client, err := cos.newClientWithBucket()
	if err != nil {
		return false, err
	}
	if _, err := client.Object.PutFromFile(context.Background(), target, src, &cosSDK.ObjectPutOptions{
		ACLHeaderOptions: nil,
		ObjectPutHeaderOptions: &cosSDK.ObjectPutHeaderOptions{
			XCosStorageClass: cos.scType,
		},
	}); err != nil {
		return false, err
	}
	return true, nil
}

func (cos cosClient) Download(src, target string) (bool, error) {
	client, err := cos.newClientWithBucket()
	if err != nil {
		return false, err
	}
	if _, err := client.Object.Download(context.Background(), src, target, &cosSDK.MultiDownloadOptions{}); err != nil {
		return false, err
	}
	return true, nil
}

func (cos *cosClient) GetBucket() (string, error) {
	if _, ok := cos.Vars["bucket"]; ok {
		return cos.Vars["bucket"].(string), nil
	} else {
		return "", constant.ErrInvalidParams
	}
}

func (cos cosClient) ListObjects(prefix string) ([]interface{}, error) {
	client, err := cos.newClientWithBucket()
	if err != nil {
		return nil, err
	}
	datas, _, err := client.Bucket.Get(context.Background(), &cosSDK.BucketGetOptions{Prefix: prefix})
	if err != nil {
		return nil, err
	}

	var result []interface{}
	for _, item := range datas.Contents {
		result = append(result, item.Key)
	}
	return result, nil
}

func (cos cosClient) newClientWithBucket() (*cosSDK.Client, error) {
	bucket, err := cos.GetBucket()
	if err != nil {
		return nil, err
	}
	u, _ := url.Parse(fmt.Sprintf("https://%s.cos.%s.myqcloud.com", bucket, cos.region))
	b := &cosSDK.BaseURL{BucketURL: u}
	client := cosSDK.NewClient(b, &http.Client{
		Transport: &cosSDK.AuthorizationTransport{
			SecretID:  cos.accessKey,
			SecretKey: cos.secretKey,
		},
	})
	return client, nil
}
