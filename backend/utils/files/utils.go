package files

import (
	"github.com/gabriel-vasile/mimetype"
	"github.com/spf13/afero"
	"os"
	"path/filepath"
	"sync"
)

func IsSymlink(mode os.FileMode) bool {
	return mode&os.ModeSymlink != 0
}

func GetMimeType(path string) string {
	mime, err := mimetype.DetectFile(path)
	if err != nil {
		return ""
	}
	return mime.String()
}

func GetSymlink(path string) string {
	linkPath, err := os.Readlink(path)
	if err != nil {
		return ""
	}
	return linkPath
}

func ScanDir(fs afero.Fs, path string, dirMap *sync.Map, wg *sync.WaitGroup) {
	afs := &afero.Afero{Fs: fs}
	files, _ := afs.ReadDir(path)
	for _, f := range files {
		if f.IsDir() {
			wg.Add(1)
			go ScanDir(fs, filepath.Join(path, f.Name()), dirMap, wg)
		} else {
			if f.Size() > 0 {
				dirMap.Store(filepath.Join(path, f.Name()), float64(f.Size()))
			}
		}
	}
	defer wg.Done()
}

const dotCharacter = 46

func IsHidden(path string) bool {
	return path[0] == dotCharacter
}
