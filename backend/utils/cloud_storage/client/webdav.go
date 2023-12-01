package client

import (
	"fmt"
	"io"
	"os"

	"github.com/studio-b12/gowebdav"
)

type webDAVClient struct {
	Bucket string
	client *gowebdav.Client
}

func NewWebDAVClient(vars map[string]interface{}) (*webDAVClient, error) {
	address := loadParamFromVars("address", true, vars)
	port := loadParamFromVars("port", false, vars)
	password := loadParamFromVars("password", true, vars)
	username := loadParamFromVars("username", true, vars)
	bucket := loadParamFromVars("bucket", true, vars)

	client := gowebdav.NewClient(fmt.Sprintf("%s:%s", address, port), username, password)
	if err := client.Connect(); err != nil {
		return nil, err
	}
	return &webDAVClient{Bucket: bucket, client: client}, nil
}

func (s webDAVClient) Upload(src, target string) (bool, error) {
	targetFilePath := s.Bucket + "/" + target
	fileInfo, err := os.Stat(src)
	if err != nil {
		return false, err
	}
	// 50M
	if fileInfo.Size() > 52428800 {
		bytes, _ := os.ReadFile(src)
		if err := s.client.Write(targetFilePath, bytes, 0644); err != nil {
			return false, err
		}
		return true, nil
	}
	file, _ := os.Open(src)
	defer file.Close()

	if err := s.client.WriteStream(targetFilePath, file, 0644); err != nil {
		return false, err
	}
	return true, nil
}

func (s webDAVClient) ListBuckets() ([]interface{}, error) {
	var result []interface{}
	return result, nil
}

func (s webDAVClient) Download(src, target string) (bool, error) {
	srcPath := s.Bucket + "/" + src
	info, err := s.client.Stat(srcPath)
	if err != nil {
		return false, err
	}

	file, err := os.Create(target)
	if err != nil {
		return false, err
	}
	defer file.Close()
	// 50M
	if info.Size() > 52428800 {
		reader, _ := s.client.ReadStream(srcPath)
		if _, err := io.Copy(file, reader); err != nil {
			return false, err
		}
	}

	bytes, _ := s.client.Read(srcPath)
	if err := os.WriteFile(target, bytes, 0644); err != nil {
		return false, err
	}
	return true, err
}

func (s webDAVClient) Exist(path string) (bool, error) {
	if _, err := s.client.Stat(s.Bucket + "/" + path); err != nil {
		return false, err
	}
	return true, nil
}

func (s webDAVClient) Size(path string) (int64, error) {
	file, err := s.client.Stat(s.Bucket + "/" + path)
	if err != nil {
		return 0, err
	}
	return file.Size(), nil
}

func (s webDAVClient) Delete(filePath string) (bool, error) {
	if err := s.client.Remove(s.Bucket + "/" + filePath); err != nil {
		return false, err
	}
	return true, nil
}

func (s webDAVClient) ListObjects(prefix string) ([]string, error) {
	files, err := s.client.ReadDir(s.Bucket + "/" + prefix)
	if err != nil {
		return nil, err
	}
	var result []string
	for _, file := range files {
		result = append(result, file.Name())
	}
	return result, nil
}
