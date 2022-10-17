package cloud_storage

import (
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/utils/cloud_storage/client"
)

type CloudStorageClient interface {
	ListBuckets() ([]interface{}, error)
	ListObjects(prefix string) ([]interface{}, error)
	Exist(path string) (bool, error)
	Delete(path string) (bool, error)
	Upload(src, target string) (bool, error)
	Download(src, target string) (bool, error)
}

func NewCloudStorageClient(vars map[string]interface{}) (CloudStorageClient, error) {
	if vars["type"] == constant.S3 {
		return client.NewS3Client(vars)
	}
	if vars["type"] == constant.OSS {
		return client.NewOssClient(vars)
	}
	if vars["type"] == constant.Sftp {
		return client.NewSftpClient(vars)
	}
	if vars["type"] == constant.MinIo {
		return client.NewMinIoClient(vars)
	}
	return nil, constant.ErrNotSupportType
}
