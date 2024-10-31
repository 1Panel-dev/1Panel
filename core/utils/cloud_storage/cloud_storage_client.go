package cloud_storage

import (
	"github.com/1Panel-dev/1Panel/core/constant"
	"github.com/1Panel-dev/1Panel/core/utils/cloud_storage/client"
)

type CloudStorageClient interface {
	ListBuckets() ([]interface{}, error)
	Upload(src, target string) (bool, error)
}

func NewCloudStorageClient(backupType string, vars map[string]interface{}) (CloudStorageClient, error) {
	switch backupType {
	case constant.Local:
		return client.NewLocalClient(vars)
	case constant.S3:
		return client.NewS3Client(vars)
	case constant.OSS:
		return client.NewOssClient(vars)
	case constant.Sftp:
		return client.NewSftpClient(vars)
	case constant.WebDAV:
		return client.NewWebDAVClient(vars)
	case constant.MinIo:
		return client.NewMinIoClient(vars)
	case constant.Cos:
		return client.NewCosClient(vars)
	case constant.Kodo:
		return client.NewKodoClient(vars)
	case constant.OneDrive:
		return client.NewOneDriveClient(vars)
	case constant.UPYUN:
		return client.NewUpClient(vars)
	default:
		return nil, constant.ErrNotSupportType
	}
}
