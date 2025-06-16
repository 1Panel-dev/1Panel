package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"regexp"

	cosSDK "github.com/tencentyun/cos-go-sdk-v5"
)

type cosClient struct {
	scType           string
	client           *cosSDK.Client
	clientWithBucket *cosSDK.Client
}

func NewCosClient(vars map[string]interface{}) (*cosClient, error) {
	region := loadParamFromVars("region", vars)
	endpoint := loadParamFromVars("endpoint", vars)
	accessKey := loadParamFromVars("accessKey", vars)
	secretKey := loadParamFromVars("secretKey", vars)
	bucket := loadParamFromVars("bucket", vars)
	scType := loadParamFromVars("scType", vars)
	if len(scType) == 0 {
		scType = "Standard"
	}

	endpointType := "cos"
	if len(endpoint) != 0 {
		re := regexp.MustCompile(`.*cos-dualstack\..*`)
		if re.MatchString(endpoint) {
			endpointType = "cos-dualstack"
		}
	}

	u, _ := url.Parse(fmt.Sprintf("https://%s.%s.myqcloud.com", endpointType, region))
	b := &cosSDK.BaseURL{BucketURL: u}
	client := cosSDK.NewClient(b, &http.Client{
		Transport: &cosSDK.AuthorizationTransport{
			SecretID:  accessKey,
			SecretKey: secretKey,
		},
	})

	if len(bucket) != 0 {
		u2, _ := url.Parse(fmt.Sprintf("https://%s.%s.%s.myqcloud.com", bucket, endpointType, region))
		b2 := &cosSDK.BaseURL{BucketURL: u2}
		clientWithBucket := cosSDK.NewClient(b2, &http.Client{
			Transport: &cosSDK.AuthorizationTransport{
				SecretID:  accessKey,
				SecretKey: secretKey,
			},
		})
		return &cosClient{client: client, clientWithBucket: clientWithBucket, scType: scType}, nil
	}

	return &cosClient{client: client, clientWithBucket: nil, scType: scType}, nil
}

func (c cosClient) ListBuckets() ([]interface{}, error) {
	buckets, _, err := c.client.Service.Get(context.Background())
	if err != nil {
		return nil, err
	}
	var datas []interface{}
	for _, bucket := range buckets.Buckets {
		datas = append(datas, bucket.Name)
	}
	return datas, nil
}

func (c cosClient) Upload(src, target string) (bool, error) {
	fileInfo, err := os.Stat(src)
	if err != nil {
		return false, err
	}
	if fileInfo.Size() > 5368709120 {
		opt := &cosSDK.MultiUploadOptions{
			OptIni: &cosSDK.InitiateMultipartUploadOptions{
				ACLHeaderOptions: nil,
				ObjectPutHeaderOptions: &cosSDK.ObjectPutHeaderOptions{
					XCosStorageClass: c.scType,
				},
			},
			PartSize: 200,
		}
		if _, _, err := c.clientWithBucket.Object.MultiUpload(
			context.Background(), target, src, opt,
		); err != nil {
			return false, err
		}
		return true, nil
	}
	if _, err := c.clientWithBucket.Object.PutFromFile(context.Background(), target, src, &cosSDK.ObjectPutOptions{
		ACLHeaderOptions: nil,
		ObjectPutHeaderOptions: &cosSDK.ObjectPutHeaderOptions{
			XCosStorageClass: c.scType,
		},
	}); err != nil {
		return false, err
	}
	return true, nil
}

func (c cosClient) Delete(path string) (bool, error) {
	if _, err := c.clientWithBucket.Object.Delete(context.Background(), path); err != nil {
		return false, err
	}
	return true, nil
}
