package files

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/1Panel-dev/1Panel/agent/constant"
)

const DefaultFileAIMaxItems = 500

type AISearchInventoryItem struct {
	RelPath string `json:"relPath"`
	IsDir   bool   `json:"isDir"`
	Size    int64  `json:"size"`
	ModTime string `json:"modTime"`
}

func CollectDirInventory(root string, containSub bool, maxItems int) (items []AISearchInventoryItem, truncated bool, err error) {
	if maxItems <= 0 {
		maxItems = DefaultFileAIMaxItems
	}
	root = filepath.Clean(root)
	st, err := os.Stat(root)
	if err != nil {
		return nil, false, err
	}
	if !st.IsDir() {
		return nil, false, errors.New("path is not a directory")
	}

	items = make([]AISearchInventoryItem, 0, maxItems)

	if !containSub {
		entries, err := os.ReadDir(root)
		if err != nil {
			return nil, false, err
		}
		for _, e := range entries {
			full := filepath.Join(root, e.Name())
			if ShouldFilterSensitivePath(full) {
				continue
			}
			if len(items) >= maxItems {
				truncated = true
				break
			}
			info, err := e.Info()
			if err != nil {
				continue
			}
			items = append(items, AISearchInventoryItem{
				RelPath: filepath.ToSlash(e.Name()),
				IsDir:   e.IsDir(),
				Size:    info.Size(),
				ModTime: info.ModTime().Format(constant.DateTimeLayout),
			})
		}
		return items, truncated, nil
	}

	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil
		}
		if ShouldFilterSensitivePath(path) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil || rel == "." {
			return nil
		}
		if len(items) >= maxItems {
			truncated = true
			return filepath.SkipAll
		}
		info, err := d.Info()
		if err != nil {
			return nil
		}
		items = append(items, AISearchInventoryItem{
			RelPath: filepath.ToSlash(rel),
			IsDir:   d.IsDir(),
			Size:    info.Size(),
			ModTime: info.ModTime().Format(constant.DateTimeLayout),
		})
		return nil
	})
	if err != nil {
		return nil, false, err
	}
	return items, truncated, nil
}
