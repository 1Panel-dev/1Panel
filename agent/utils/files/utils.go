package files

import (
	"bufio"
	"fmt"
	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/utils/req_helper"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/encoding/korean"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/traditionalchinese"
	"golang.org/x/text/encoding/unicode"
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

func TailFromEnd(filename string, lines int) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}
	fileSize := stat.Size()

	bufSize := int64(4096)
	var result []string
	var leftover string

	for offset := fileSize; offset > 0 && len(result) < lines; {
		readSize := bufSize
		if offset < bufSize {
			readSize = offset
		}
		offset -= readSize

		buf := make([]byte, readSize)
		_, err := file.ReadAt(buf, offset)
		if err != nil && err != io.EOF {
			return nil, err
		}

		data := string(buf) + leftover
		linesInChunk := strings.Split(data, "\n")

		if offset > 0 {
			leftover = linesInChunk[0]
			linesInChunk = linesInChunk[1:]
		} else {
			leftover = ""
		}

		for i := len(linesInChunk) - 1; i >= 0; i-- {
			if len(result) < lines {
				if !(i == len(linesInChunk)-1 && linesInChunk[i] == "" && len(result) == 0) {
					result = append([]string{linesInChunk[i]}, result...)
				}
			}
		}
	}

	if leftover != "" && len(result) < lines {
		result = append([]string{leftover}, result...)
	}

	if len(result) > lines {
		result = result[len(result)-lines:]
	}

	return result, nil
}

func ReadFileByLine(filename string, page, pageSize int, latest bool) (res *dto.LogFileRes, err error) {
	if !NewFileOp().Stat(filename) {
		return
	}
	file, err := os.Open(filename)
	if err != nil {
		return
	}
	defer file.Close()

	fi, err := file.Stat()
	if err != nil {
		return
	}

	if fi.Size() > 500*1024*1024 {
		err = buserr.New("ErrLogFileToLarge")
		return
	}

	totalLines, err := countLines(filename)
	if err != nil {
		return
	}
	res = &dto.LogFileRes{}
	total := (totalLines + pageSize - 1) / pageSize
	res.TotalPages = total
	res.TotalLines = totalLines
	reader := bufio.NewReaderSize(file, 8192)

	if latest {
		page = total
	}
	currentLine := 0
	startLine := (page - 1) * pageSize
	endLine := startLine + pageSize
	lines := make([]string, 0, pageSize)
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
	res.Lines = lines
	res.IsEndOfFile = currentLine < endLine
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
	resp, cancel, err := req_helper.RequestFile(url, http.MethodGet, constant.TimeOut5m)
	if err != nil {
		return err
	}
	defer cancel()
	defer resp.Close()

	out, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("create download file [%s] error, err %s", dst, err.Error())
	}
	defer out.Close()

	if _, err = io.Copy(out, resp); err != nil {
		return fmt.Errorf("save download file [%s] error, err %s", dst, err.Error())
	}
	return nil
}

func GetDecoderByName(name string) encoding.Encoding {
	switch strings.ToLower(name) {
	case "gbk":
		return simplifiedchinese.GBK
	case "gb18030":
		return simplifiedchinese.GB18030
	case "big5":
		return traditionalchinese.Big5
	case "euc-jp":
		return japanese.EUCJP
	case "iso-2022-jp":
		return japanese.ISO2022JP
	case "shift_jis":
		return japanese.ShiftJIS
	case "euc-kr":
		return korean.EUCKR
	case "utf-16be":
		return unicode.UTF16(unicode.BigEndian, unicode.ExpectBOM)
	case "utf-16le":
		return unicode.UTF16(unicode.LittleEndian, unicode.ExpectBOM)
	case "windows-1250":
		return charmap.Windows1250
	case "windows-1251":
		return charmap.Windows1251
	case "windows-1252":
		return charmap.Windows1252
	case "windows-1253":
		return charmap.Windows1253
	case "windows-1254":
		return charmap.Windows1254
	case "windows-1255":
		return charmap.Windows1255
	case "windows-1256":
		return charmap.Windows1256
	case "windows-1257":
		return charmap.Windows1257
	case "windows-1258":
		return charmap.Windows1258
	case "iso-8859-1":
		return charmap.ISO8859_1
	case "iso-8859-2":
		return charmap.ISO8859_2
	case "iso-8859-3":
		return charmap.ISO8859_3
	case "iso-8859-4":
		return charmap.ISO8859_4
	case "iso-8859-5":
		return charmap.ISO8859_5
	case "iso-8859-6":
		return charmap.ISO8859_6
	case "iso-8859-7":
		return charmap.ISO8859_7
	case "iso-8859-8":
		return charmap.ISO8859_8
	case "iso-8859-9":
		return charmap.ISO8859_9
	case "iso-8859-13":
		return charmap.ISO8859_13
	case "iso-8859-15":
		return charmap.ISO8859_15
	default:
		return encoding.Nop
	}
}
