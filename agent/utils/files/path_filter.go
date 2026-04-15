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
	return false
}
