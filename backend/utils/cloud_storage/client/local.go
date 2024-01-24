package client

import (
	"fmt"
	"os"
	"path"

	"github.com/1Panel-dev/1Panel/backend/utils/cmd"
)

type localClient struct {
	dir string
}

func NewLocalClient(vars map[string]interface{}) (*localClient, error) {
	dir := loadParamFromVars("dir", true, vars)
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

	stdout, err := cmd.Execf("\\cp -f %s %s", src, path.Join(c.dir, target))
	if err != nil {
		return false, fmt.Errorf("cp file failed, stdout: %v, err: %v", stdout, err)
	}
	return true, nil
}

func (c localClient) Download(src, target string) (bool, error) {
	return true, nil
}

func (c localClient) ListObjects(prefix string) ([]string, error) {
	return nil, nil
}
