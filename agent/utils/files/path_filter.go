package files

import (
	"path/filepath"
	"strings"
)

var sensitivePanelPaths = []string{
	"/.1panel_clash",
}

func ShouldFilterSensitivePath(path string) bool {
	cleanedPath := filepath.Clean(path)
	for _, filteredPath := range sensitivePanelPaths {
		cleanedFilteredPath := filepath.Clean(filteredPath)
		if cleanedFilteredPath == cleanedPath || strings.HasPrefix(cleanedPath, cleanedFilteredPath+string(filepath.Separator)) {
			return true
		}
	}
	return ShouldDenySensitiveFileRead(path)
}

func ShouldDenySensitiveFileRead(path string) bool {
	if isSensitiveSSHPath(path) {
		return true
	}
	realPath, err := filepath.EvalSymlinks(path)
	if err != nil {
		return false
	}
	return isSensitiveSSHPath(realPath)
}

func isSensitiveSSHPath(path string) bool {
	cleanedPath := filepath.Clean(path)
	parts := strings.Split(filepath.ToSlash(cleanedPath), "/")
	for i, part := range parts {
		if part != ".ssh" {
			continue
		}
		if i == len(parts)-1 {
			return true
		}
		return !strings.HasSuffix(parts[len(parts)-1], ".pub")
	}
	return false
}
