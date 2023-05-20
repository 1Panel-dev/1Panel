package client

import (
	"context"
	"time"

	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/utils/files"
	"github.com/qiniu/go-sdk/v7/auth"
	"github.com/qiniu/go-sdk/v7/auth/qbox"
	"github.com/qiniu/go-sdk/v7/storage"
)

type kodoClient struct {
	accessKey string
	secretKey string
	Vars      map[string]interface{}
	client    *storage.BucketManager
}

func NewKodoClient(vars map[string]interface{}) (*kodoClient, error) {
	var accessKey string
	var secretKey string
	if _, ok := vars["accessKey"]; ok {
		accessKey = vars["accessKey"].(string)
	} else {
		return nil, constant.ErrInvalidParams
	}
	if _, ok := vars["secretKey"]; ok {
		secretKey = vars["secretKey"].(string)
	} else {
		return nil, constant.ErrInvalidParams
	}

	conn := qbox.NewMac(accessKey, secretKey)
	cfg := storage.Config{
		UseHTTPS: false,
	}
	bucketManager := storage.NewBucketManager(conn, &cfg)

	return &kodoClient{Vars: vars, client: bucketManager, accessKey: accessKey, secretKey: secretKey}, nil
}

func (kodo kodoClient) ListBuckets() ([]interface{}, error) {
	buckets, err := kodo.client.Buckets(true)
	if err != nil {
		return nil, err
	}
	var datas []interface{}
	for _, bucket := range buckets {
		datas = append(datas, bucket)
	}
	return datas, nil
}

func (kodo kodoClient) Exist(path string) (bool, error) {
	bucket, err := kodo.GetBucket()
	if err != nil {
		return false, err
	}
	if _, err := kodo.client.Stat(bucket, path); err != nil {
		return false, err
	}
	return true, nil
}

func (kodo kodoClient) Delete(path string) (bool, error) {
	bucket, err := kodo.GetBucket()
	if err != nil {
		return false, err
	}
	if err := kodo.client.Delete(bucket, path); err != nil {
		return false, err
	}
	return true, nil
}

func (kodo kodoClient) Upload(src, target string) (bool, error) {
	bucket, err := kodo.GetBucket()
	if err != nil {
		return false, err
	}
	putPolicy := storage.PutPolicy{
		Scope: bucket,
	}
	mac := qbox.NewMac(kodo.accessKey, kodo.secretKey)
	upToken := putPolicy.UploadToken(mac)
	cfg := storage.Config{UseHTTPS: true, UseCdnDomains: false}
	resumeUploader := storage.NewResumeUploaderV2(&cfg)
	ret := storage.PutRet{}
	putExtra := storage.RputV2Extra{}
	if err := resumeUploader.PutFile(context.Background(), &ret, upToken, target, src, &putExtra); err != nil {
		return false, err
	}
	return true, nil
}

func (kodo kodoClient) Download(src, target string) (bool, error) {
	mac := auth.New(kodo.accessKey, kodo.secretKey)
	if _, ok := kodo.Vars["domain"]; !ok {
		return false, constant.ErrInvalidParams
	}
	domain := kodo.Vars["domain"].(string)
	deadline := time.Now().Add(time.Second * 3600).Unix()
	privateAccessURL := storage.MakePrivateURL(mac, domain, src, deadline)

	fo := files.NewFileOp()
	if err := fo.DownloadFile(privateAccessURL, target); err != nil {
		return false, err
	}
	return true, nil
}

func (kodo *kodoClient) GetBucket() (string, error) {
	if _, ok := kodo.Vars["bucket"]; ok {
		return kodo.Vars["bucket"].(string), nil
	} else {
		return "", constant.ErrInvalidParams
	}
}

func (kodo kodoClient) ListObjects(prefix string) ([]interface{}, error) {
	bucket, err := kodo.GetBucket()
	if err != nil {
		return nil, constant.ErrInvalidParams
	}

	var result []interface{}
	marker := ""
	for {
		entries, _, nextMarker, hashNext, err := kodo.client.ListFiles(bucket, prefix, "", marker, 1000)
		if err != nil {
			return nil, err
		}
		for _, entry := range entries {
			result = append(result, entry.Key)
		}
		if hashNext {
			marker = nextMarker
		} else {
			break
		}
	}
	return result, nil
}
