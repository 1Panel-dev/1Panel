package client

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/1Panel-dev/1Panel/backend/utils/common"
)

type localClient struct {
	dir string
}

func NewLocalClient(vars map[string]interface{}) (*localClient, error) {
	dir := loadParamFromVars("dir", vars)
	return &localClient{dir: dir}, nil
}

func (c localClient) ListBuckets() ([]interface{}, error) {
	return nil, nil
}

func (c localClient) Exist(file string) (bool, error) {
	_, err := os.Stat(path.Join(c.dir, file))
	return err == nil, err
}

func (c localClient) Size(file string) (int64, error) {
	fileInfo, err := os.Stat(path.Join(c.dir, file))
	if err != nil {
		return 0, err
	}
	return fileInfo.Size(), nil
}

func (c localClient) Delete(file string) (bool, error) {
	if err := os.RemoveAll(path.Join(c.dir, file)); err != nil {
		return false, err
	}
	return true, nil
}

func (c localClient) Upload(src, target string) (bool, error) {
	targetFilePath := path.Join(c.dir, target)
	if _, err := os.Stat(path.Dir(targetFilePath)); err != nil {
		if os.IsNotExist(err) {
			if err = os.MkdirAll(path.Dir(targetFilePath), os.ModePerm); err != nil {
				return false, err
			}
		} else {
			return false, err
		}
	}

	if err := common.CopyFile(src, targetFilePath); err != nil {
		return false, fmt.Errorf("cp file failed, err: %v", err)
	}
	return true, nil
}

func (c localClient) Download(src, target string) (bool, error) {
	localPath := path.Join(c.dir, src)
	if _, err := os.Stat(path.Dir(target)); err != nil {
		if os.IsNotExist(err) {
			if err = os.MkdirAll(path.Dir(target), os.ModePerm); err != nil {
				return false, err
			}
		} else {
			return false, err
		}
	}

	if err := common.CopyFile(localPath, target); err != nil {
		return false, fmt.Errorf("cp file failed, err: %v", err)
	}
	return true, nil
}

func (c localClient) ListObjects(prefix string) ([]string, error) {
	var files []string
	itemPath := path.Join(c.dir, prefix)
	if _, err := os.Stat(itemPath); err != nil {
		return files, nil
	}
	if err := filepath.Walk(itemPath, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			files = append(files, info.Name())
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return files, nil
}
