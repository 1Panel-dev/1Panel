package client

import (
	"context"
	"strconv"
	"time"

	"github.com/1Panel-dev/1Panel/backend/utils/files"
	"github.com/qiniu/go-sdk/v7/auth"
	"github.com/qiniu/go-sdk/v7/storage"
)

type kodoClient struct {
	bucket  string
	domain  string
	timeout string
	auth    *auth.Credentials
	client  *storage.BucketManager
}

func NewKodoClient(vars map[string]interface{}) (*kodoClient, error) {
	accessKey := loadParamFromVars("accessKey", vars)
	secretKey := loadParamFromVars("secretKey", vars)
	bucket := loadParamFromVars("bucket", vars)
	domain := loadParamFromVars("domain", vars)
	timeout := loadParamFromVars("timeout", vars)
	if timeout == "" {
		timeout = "1"
	}
	conn := auth.New(accessKey, secretKey)
	cfg := storage.Config{
		UseHTTPS: false,
	}
	bucketManager := storage.NewBucketManager(conn, &cfg)

	return &kodoClient{client: bucketManager, auth: conn, bucket: bucket, domain: domain, timeout: timeout}, nil
}

func (k kodoClient) ListBuckets() ([]interface{}, error) {
	buckets, err := k.client.Buckets(true)
	if err != nil {
		return nil, err
	}
	var datas []interface{}
	for _, bucket := range buckets {
		datas = append(datas, bucket)
	}
	return datas, nil
}

func (k kodoClient) Exist(path string) (bool, error) {
	if _, err := k.client.Stat(k.bucket, path); err != nil {
		return false, err
	}
	return true, nil
}

func (k kodoClient) Size(path string) (int64, error) {
	file, err := k.client.Stat(k.bucket, path)
	if err != nil {
		return 0, err
	}
	return file.Fsize, nil
}

func (k kodoClient) Delete(path string) (bool, error) {
	if err := k.client.Delete(k.bucket, path); err != nil {
		return false, err
	}
	return true, nil
}

func (k kodoClient) Upload(src, target string) (bool, error) {

	int64Value, _ := strconv.ParseInt(k.timeout, 10, 64)
	unixTimestamp := int64Value * 3600

	putPolicy := storage.PutPolicy{
		Scope:   k.bucket,
		Expires: uint64(unixTimestamp),
	}
	upToken := putPolicy.UploadToken(k.auth)
	cfg := storage.Config{UseHTTPS: true, UseCdnDomains: false}
	resumeUploader := storage.NewResumeUploaderV2(&cfg)
	ret := storage.PutRet{}
	putExtra := storage.RputV2Extra{}
	if err := resumeUploader.PutFile(context.Background(), &ret, upToken, target, src, &putExtra); err != nil {
		return false, err
	}
	return true, nil
}

func (k kodoClient) Download(src, target string) (bool, error) {
	deadline := time.Now().Add(time.Second * 3600).Unix()
	privateAccessURL := storage.MakePrivateURL(k.auth, k.domain, src, deadline)

	fo := files.NewFileOp()
	if err := fo.DownloadFile(privateAccessURL, target); err != nil {
		return false, err
	}
	return true, nil
}

func (k kodoClient) ListObjects(prefix string) ([]string, error) {
	var result []string
	marker := ""
	for {
		entries, _, nextMarker, hashNext, err := k.client.ListFiles(k.bucket, prefix, "", marker, 1000)
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
