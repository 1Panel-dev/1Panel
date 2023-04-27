package files

import (
	"github.com/spf13/afero"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"sync"
)

func IsSymlink(mode os.FileMode) bool {
	return mode&os.ModeSymlink != 0
}

func GetMimeType(path string) string {
	file, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer file.Close()

	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil {
		return ""
	}
	mimeType := http.DetectContentType(buffer)
	return mimeType
}

func GetSymlink(path string) string {
	linkPath, err := os.Readlink(path)
	if err != nil {
		return ""
	}
	return linkPath
}

func GetUsername(uid uint32) string {
	usr, err := user.LookupId(strconv.Itoa(int(uid)))
	if err != nil {
		return ""
	}
	return usr.Username
}

func GetGroup(gid uint32) string {
	usr, err := user.LookupGroupId(strconv.Itoa(int(gid)))
	if err != nil {
		return ""
	}
	return usr.Name
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
