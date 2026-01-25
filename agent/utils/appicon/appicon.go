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

var SupportedContentTypes = map[string]string{
	"image/png":                "png",
	"image/jpeg":               "jpg",
	"image/jpg":                "jpg",
	"image/svg+xml":            "svg",
	"image/x-icon":             "ico",
	"image/vnd.microsoft.icon": "ico",
	"image/webp":               "webp",
}

var ExtToContentType = map[string]string{
	"png":  "image/png",
	"jpg":  "image/jpeg",
	"jpeg": "image/jpeg",
	"svg":  "image/svg+xml",
	"ico":  "image/x-icon",
	"webp": "image/webp",
}

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

func DetectExtFromContentType(contentType string) string {
	ct := strings.TrimSpace(strings.Split(contentType, ";")[0])
	ct = strings.ToLower(ct)
	if ext, ok := SupportedContentTypes[ct]; ok {
		return ext
	}
	return ""
}

func GetContentTypeFromExt(fileName string) string {
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(fileName), "."))
	if ct, ok := ExtToContentType[ext]; ok {
		return ct
	}
	return "image/png"
}

func WriteIconFile(appKey, ext string, data []byte) (fileName string, err error) {
	fileName = BuildIconFileName(appKey, ext)
	filePath := GetIconFilePath(fileName)

	_ = CleanOldIconFiles(appKey, ext)

	err = os.WriteFile(filePath, data, 0644)
	return
}

func CleanOldIconFiles(appKey, keepExt string) error {
	pattern := path.Join(global.Dir.IconCacheDir, fmt.Sprintf("%s%s.*", IconPrefix, appKey))
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return err
	}
	keepFileName := BuildIconFileName(appKey, keepExt)
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
