package service

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
)

func listLocalModelOptions(root string) ([]dto.LocalModelOption, error) {
	items := make([]dto.LocalModelOption, 0)
	entries, err := os.ReadDir(root)
	if os.IsNotExist(err) {
		return items, nil
	}
	if err != nil {
		return nil, err
	}
	var scan func(string, string) error
	scan = func(path, organization string) error {
		children, err := os.ReadDir(path)
		if err != nil {
			return err
		}
		for _, child := range children {
			if child.IsDir() || child.Type()&os.ModeSymlink != 0 || strings.HasPrefix(child.Name(), ".") {
				continue
			}
			name := child.Name()
			if name == "config.json" || name == "params.json" {
				items = append(items, dto.LocalModelOption{Name: filepath.Base(path), Organization: organization, Path: path})
				return nil
			}
		}
		if organization == "" {
			for _, child := range children {
				if child.IsDir() && !strings.HasPrefix(child.Name(), ".") {
					if err := scan(filepath.Join(path, child.Name()), filepath.Base(path)); err != nil {
						return err
					}
				}
			}
		}
		return nil
	}
	for _, entry := range entries {
		if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
			if err := scan(filepath.Join(root, entry.Name()), ""); err != nil {
				return nil, err
			}
		}
	}
	sort.Slice(items, func(i, j int) bool {
		left, right := strings.ToLower(items[i].Name), strings.ToLower(items[j].Name)
		if left == right {
			return items[i].Path < items[j].Path
		}
		return left < right
	})
	return items, nil
}
