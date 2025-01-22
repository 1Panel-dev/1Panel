package client

import (
	"fmt"
	"os"
	"path"

	"github.com/1Panel-dev/1Panel/core/utils/files"
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

	if err := files.CopyFile(src, targetFilePath, false); err != nil {
		return false, fmt.Errorf("cp file failed, err: %v", err)
	}
	return true, nil
}

func (c localClient) Delete(file string) (bool, error) {
	if err := os.RemoveAll(file); err != nil {
		return false, err
	}
	return true, nil
}
