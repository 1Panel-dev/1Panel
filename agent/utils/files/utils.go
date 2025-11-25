package files

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

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
)

const (
	MaxReadFileSize = 512 * 1024 * 1024
	tailBufSize     = int64(32768)
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

var readerPool = sync.Pool{
	New: func() interface{} {
		return bufio.NewReaderSize(nil, 8192)
	},
}

var tailBufPool = sync.Pool{
	New: func() interface{} {
		buf := make([]byte, tailBufSize)
		return &buf
	},
}

func readLineTrimmed(reader *bufio.Reader) (string, error) {
	line, err := reader.ReadString('\n')
	if err == io.EOF {
		if len(line) == 0 {
			return "", io.EOF
		}
		err = nil
	}
	if err != nil {
		return "", err
	}
	line = strings.TrimSuffix(line, "\n")
	line = strings.TrimSuffix(line, "\r")
	return line, nil
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

	bufPtr := tailBufPool.Get().(*[]byte)
	buf := *bufPtr
	defer tailBufPool.Put(bufPtr)

	var result []string
	var leftover string

	for offset := fileSize; offset > 0 && len(result) < lines; {
		readSize := tailBufSize
		if offset < tailBufSize {
			readSize = offset
		}
		offset -= readSize

		_, err := file.ReadAt(buf[:readSize], offset)
		if err != nil && err != io.EOF {
			return nil, err
		}

		data := string(buf[:readSize]) + leftover
		linesInChunk := strings.Split(data, "\n")

		if offset > 0 {
			leftover = linesInChunk[0]
			linesInChunk = linesInChunk[1:]
		} else {
			leftover = ""
		}

		for i := len(linesInChunk) - 1; i >= 0; i-- {
			if len(result) >= lines {
				break
			}
			if i == len(linesInChunk)-1 && linesInChunk[i] == "" && len(result) == 0 {
				continue
			}
			// 反插数据
			result = append(result, linesInChunk[i])
		}
	}

	if leftover != "" && len(result) < lines {
		result = append(result, leftover)
	}

	if len(result) > lines {
		result = result[:lines]
	}

	// 反转数据
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	return result, nil
}

func ReadFileByLine(filename string, page, pageSize int, latest bool) (res *dto.LogFileRes, err error) {
	if !NewFileOp().Stat(filename) {
		return
	}
	if pageSize <= 0 {
		err = fmt.Errorf("pageSize must be positive")
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

	if fi.Size() > MaxReadFileSize {
		err = buserr.New("ErrLogFileToLarge")
		return
	}

	res = &dto.LogFileRes{}
	reader := readerPool.Get().(*bufio.Reader)
	reader.Reset(file)
	defer readerPool.Put(reader)

	if latest {
		ringBuf := make([]string, pageSize)
		writeIdx := 0
		totalLines := 0

		for {
			line, readErr := readLineTrimmed(reader)
			if readErr == io.EOF {
				break
			}
			if readErr != nil {
				err = readErr
				return
			}
			ringBuf[writeIdx%pageSize] = line
			writeIdx++
			totalLines++
		}

		if totalLines == 0 {
			res.Lines = []string{}
			res.TotalLines = 0
			res.TotalPages = 0
			res.IsEndOfFile = true
			return
		}

		total := (totalLines + pageSize - 1) / pageSize
		res.TotalPages = total
		res.TotalLines = totalLines

		lastPageSize := totalLines % pageSize
		if lastPageSize == 0 {
			lastPageSize = pageSize
		}
		if lastPageSize > totalLines {
			lastPageSize = totalLines
		}

		result := make([]string, 0, lastPageSize)
		startIdx := writeIdx - lastPageSize
		for i := 0; i < lastPageSize; i++ {
			idx := (startIdx + i) % pageSize
			result = append(result, ringBuf[idx])
		}
		res.Lines = result
		res.IsEndOfFile = true
	} else {
		startLine := (page - 1) * pageSize
		endLine := startLine + pageSize
		currentLine := 0
		lines := make([]string, 0, pageSize)

		for {
			line, readErr := readLineTrimmed(reader)
			if readErr == io.EOF {
				break
			}
			if readErr != nil {
				err = readErr
				return
			}

			if currentLine >= startLine && currentLine < endLine {
				lines = append(lines, line)
			}
			currentLine++
		}

		res.Lines = lines
		res.TotalLines = currentLine
		total := (currentLine + pageSize - 1) / pageSize
		res.TotalPages = total
		res.IsEndOfFile = page >= total
	}

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
