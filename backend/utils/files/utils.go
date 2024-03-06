package files

import (
	"bufio"
	"github.com/spf13/afero"
	"io"
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

func ReadFileByLine(filename string, page, pageSize int) ([]string, bool, error) {
	if !NewFileOp().Stat(filename) {
		return nil, true, nil
	}
	file, err := os.Open(filename)
	if err != nil {
		return nil, false, err
	}
	defer file.Close()

	reader := bufio.NewReaderSize(file, 8192)

	var lines []string
	currentLine := 0
	startLine := (page - 1) * pageSize
	endLine := startLine + pageSize

	for {
		line, _, err := reader.ReadLine()
		if err == io.EOF {
			break
		}
		if currentLine >= startLine && currentLine < endLine {
			lines = append(lines, string(line))
		}
		currentLine++
		if currentLine >= endLine {
			break
		}
	}

	isEndOfFile := currentLine < endLine

	return lines, isEndOfFile, nil
}
