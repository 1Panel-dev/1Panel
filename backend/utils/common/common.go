package common

import (
	"crypto/rand"
	"fmt"
	"io"
	mathRand "math/rand"
	"net"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/1Panel-dev/1Panel/backend/utils/cmd"
)

func CompareVersion(version1 string, version2 string) bool {
	if version1 == version2 {
		return false
	}
	version1s := strings.Split(version1, ".")
	version2s := strings.Split(version2, ".")

	n := min(len(version1s), len(version2s))
	re := regexp.MustCompile("[0-9]+")
	for i := 0; i < n; i++ {
		sVersion1s := re.FindAllString(version1s[i], -1)
		sVersion2s := re.FindAllString(version2s[i], -1)
		if len(sVersion1s) == 0 {
			return false
		}
		if len(sVersion2s) == 0 {
			return false
		}
		v1num, _ := strconv.Atoi(sVersion1s[0])
		v2num, _ := strconv.Atoi(sVersion2s[0])
		if v1num == v2num {
			continue
		} else {
			return v1num > v2num
		}
	}
	return true
}

func GetSortedVersions(versions []string) []string {
	sort.Slice(versions, func(i, j int) bool {
		return CompareVersion(versions[i], versions[j])
	})
	return versions
}

func IsCrossVersion(version1, version2 string) bool {
	version1s := strings.Split(version1, ".")
	version2s := strings.Split(version2, ".")
	v1num, _ := strconv.Atoi(version1s[0])
	v2num, _ := strconv.Atoi(version2s[0])
	return v2num > v1num
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func GetUuid() string {
	b := make([]byte, 16)
	_, _ = io.ReadFull(rand.Reader, b)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")

func RandStr(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[mathRand.Intn(len(letters))]
	}
	return string(b)
}

func ScanPort(port int) bool {

	ln, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		return true
	}
	defer ln.Close()
	return false
}

func ScanPortWithProtocol(port, Protocol string) bool {
	command := "netstat -ntpl"
	if Protocol == "udp" {
		command = "netstat -nupl"
	}
	stdout, err := cmd.Execf("%s | awk '{print $4}' ", command)
	if err != nil {
		return false
	}
	lines := strings.Split(stdout, "\n")
	if len(lines) == 0 {
		return false
	}
	for _, line := range lines {
		if strings.HasSuffix(line, ":"+port) {
			return true
		}
	}
	return false
}

func ExistWithStrArray(str string, arr []string) bool {
	for _, a := range arr {
		if strings.Contains(a, str) {
			return true
		}
	}
	return false
}

func IsNum(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}

func RemoveRepeatElement(a interface{}) (ret []interface{}) {
	va := reflect.ValueOf(a)
	for i := 0; i < va.Len(); i++ {
		if i > 0 && reflect.DeepEqual(va.Index(i-1).Interface(), va.Index(i).Interface()) {
			continue
		}
		ret = append(ret, va.Index(i).Interface())
	}
	return ret
}

func LoadSizeUnit(value float64) string {
	if value > 1048576 {
		return fmt.Sprintf("%vM", value/1048576)
	}
	if value > 1024 {
		return fmt.Sprintf("%vK", value/1024)
	}
	return fmt.Sprintf("%v", value)
}
