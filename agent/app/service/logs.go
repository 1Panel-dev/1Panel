package service

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/global"
)

type LogService struct{}

type ILogService interface {
	ListSystemLogFile() ([]string, error)
	LoadSystemLog(name string) (string, error)
}

func NewILogService() ILogService {
	return &LogService{}
}

func (u *LogService) ListSystemLogFile() ([]string, error) {
	var files []string
	if err := filepath.Walk(global.Dir.LogDir, func(pathItem string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasPrefix(info.Name(), "1Panel") {
			if info.Name() == "1Panel.log" {
				files = append(files, time.Now().Format("2006-01-02"))
				return nil
			}
			itemFileName := strings.TrimPrefix(info.Name(), "1Panel-")
			itemFileName = strings.TrimSuffix(itemFileName, ".gz")
			itemFileName = strings.TrimSuffix(itemFileName, ".log")
			files = append(files, itemFileName)
			return nil
		}
		return nil
	}); err != nil {
		return nil, err
	}

	if len(files) < 2 {
		return files, nil
	}
	sort.Slice(files, func(i, j int) bool {
		return files[i] > files[j]
	})

	return files, nil
}

func (u *LogService) LoadSystemLog(name string) (string, error) {
	if name == time.Now().Format("2006-01-02") {
		name = "1Panel.log"
	} else {
		name = "1Panel-" + name + ".log"
	}
	filePath := path.Join(global.Dir.LogDir, name)
	if _, err := os.Stat(filePath); err != nil {
		fileGzPath := path.Join(global.Dir.LogDir, name+".gz")
		if _, err := os.Stat(fileGzPath); err != nil {
			return "", buserr.New("ErrHttpReqNotFound")
		}
		if err := handleGunzip(fileGzPath); err != nil {
			return "", fmt.Errorf("handle ungzip file %s failed, err: %v", fileGzPath, err)
		}
	}
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(content), nil
}
