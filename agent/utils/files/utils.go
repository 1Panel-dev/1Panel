package files

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/utils/req_helper"
	"io"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
)

func IsSymlink(mode os.FileMode) bool {
	return mode&os.ModeSymlink != 0
}

func IsBlockDevice(mode os.FileMode) bool {
	return mode&os.ModeDevice != 0 && mode&os.ModeCharDevice == 0
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

const dotCharacter = 46

func IsHidden(path string) bool {
	base := filepath.Base(path)
	return len(base) > 1 && base[0] == dotCharacter
}

func countLines(path string) (int, error) {
	file, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	count := 0
	for {
		_, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				if count > 0 {
					count++
				}
				return count, nil
			}
			return count, err
		}
		count++
	}
}

func ReadFileByLine(filename string, page, pageSize int, latest bool) (lines []string, isEndOfFile bool, total int, err error) {
	if !NewFileOp().Stat(filename) {
		return
	}
	file, err := os.Open(filename)
	if err != nil {
		return
	}
	defer file.Close()

	totalLines, err := countLines(filename)
	if err != nil {
		return
	}
	total = (totalLines + pageSize - 1) / pageSize
	reader := bufio.NewReaderSize(file, 8192)

	if latest {
		page = total
	}
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

	isEndOfFile = currentLine < endLine
	return
}

func GetParentMode(path string) (os.FileMode, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return 0, err
	}

	for {
		fileInfo, err := os.Stat(absPath)
		if err == nil {
			return fileInfo.Mode() & os.ModePerm, nil
		}
		if !os.IsNotExist(err) {
			return 0, err
		}

		parentDir := filepath.Dir(absPath)
		if parentDir == absPath {
			return 0, fmt.Errorf("no existing directory found in the path: %s", path)
		}
		absPath = parentDir
	}
}

func IsInvalidChar(name string) bool {
	return strings.Contains(name, "&")
}

func IsEmptyDir(dir string) bool {
	f, err := os.Open(dir)
	if err != nil {
		return false
	}
	defer f.Close()
	_, err = f.Readdirnames(1)
	return err == io.EOF
}

func DownloadFileWithProxy(url, dst string) error {
	_, resp, err := req_helper.HandleRequest(url, http.MethodGet, constant.TimeOut5m)
	if err != nil {
		return err
	}

	out, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("create download file [%s] error, err %s", dst, err.Error())
	}
	defer out.Close()

	reader := bytes.NewReader(resp)
	if _, err = io.Copy(out, reader); err != nil {
		return fmt.Errorf("save download file [%s] error, err %s", dst, err.Error())
	}
	return nil
}
