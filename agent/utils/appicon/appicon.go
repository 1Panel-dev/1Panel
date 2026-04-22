package appicon

import (
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/1Panel-dev/1Panel/agent/global"
)

const IconPrefix = "app_"

func IsIconFile(icon string) bool {
	return strings.HasPrefix(icon, IconPrefix)
}

func ParseIconField(icon string) (fileName, etag string) {
	if !IsIconFile(icon) {
		return "", ""
	}
	parts := strings.SplitN(icon, "?", 2)
	fileName = parts[0]
	if len(parts) == 2 {
		values, err := url.ParseQuery(parts[1])
		if err == nil {
			etag = values.Get("etag")
		}
	}
	return
}

func BuildIconField(fileName, etag string) string {
	if etag == "" {
		return fileName
	}
	return fmt.Sprintf("%s?etag=%s", fileName, url.QueryEscape(etag))
}

func GetIconFilePath(fileName string) string {
	return path.Join(global.Dir.IconCacheDir, fileName)
}

func BuildIconFileName(appKey, ext string) string {
	return fmt.Sprintf("%s%s.%s", IconPrefix, appKey, ext)
}

const ContentTypePNG = "image/png"

func WriteIconFile(appKey string, data []byte) (fileName string, err error) {
	fileName = BuildIconFileName(appKey, "png")
	filePath := GetIconFilePath(fileName)

	_ = CleanOldIconFiles(appKey)

	err = os.WriteFile(filePath, data, 0644)
	return
}

func CleanOldIconFiles(appKey string) error {
	pattern := path.Join(global.Dir.IconCacheDir, fmt.Sprintf("%s%s.*", IconPrefix, appKey))
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return err
	}
	keepFileName := BuildIconFileName(appKey, "png")
	for _, match := range matches {
		baseName := filepath.Base(match)
		if baseName == keepFileName {
			continue
		}
		_ = os.Remove(match)
	}
	return nil
}

func ReadIconFile(fileName string) ([]byte, error) {
	filePath := GetIconFilePath(fileName)
	return os.ReadFile(filePath)
}

func IconFileExists(fileName string) bool {
	filePath := GetIconFilePath(fileName)
	_, err := os.Stat(filePath)
	return err == nil
}

func GetETagFromIconField(icon string) string {
	_, etag := ParseIconField(icon)
	return etag
}
