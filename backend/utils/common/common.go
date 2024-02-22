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
	"time"
	"unicode"

	"golang.org/x/net/idna"

	"github.com/1Panel-dev/1Panel/backend/utils/cmd"
)

func CompareVersion(version1, version2 string) bool {
	v1s := extractNumbers(version1)
	v2s := extractNumbers(version2)

	maxLen := max(len(v1s), len(v2s))
	v1s = append(v1s, make([]string, maxLen-len(v1s))...)
	v2s = append(v2s, make([]string, maxLen-len(v2s))...)

	for i := 0; i < maxLen; i++ {
		v1, err1 := strconv.Atoi(v1s[i])
		v2, err2 := strconv.Atoi(v2s[i])
		if err1 != nil {
			v1 = 0
		}
		if err2 != nil {
			v2 = 0
		}
		if v1 != v2 {
			return v1 > v2
		}
	}
	return false
}

func extractNumbers(version string) []string {
	var numbers []string
	start := -1
	for i, r := range version {
		if isDigit(r) {
			if start == -1 {
				start = i
			}
		} else {
			if start != -1 {
				numbers = append(numbers, version[start:i])
				start = -1
			}
		}
	}
	if start != -1 {
		numbers = append(numbers, version[start:])
	}
	return numbers
}

func isDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

func max(x, y int) int {
	if x > y {
		return x
	}
	return y
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

func RandStrAndNum(n int) string {
	source := mathRand.NewSource(time.Now().UnixNano())
	randGen := mathRand.New(source)
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = charset[randGen.Intn(len(charset)-1)]
	}
	return (string(b))
}

func ScanPort(port int) bool {
	ln, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		return true
	}
	defer ln.Close()
	return false
}

func ScanUDPPort(port int) bool {
	ln, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: port})
	if err != nil {
		return true
	}
	defer ln.Close()
	return false
}

func ScanPortWithProto(port int, proto string) bool {
	if proto == "udp" {
		return ScanUDPPort(port)
	}
	return ScanPort(port)
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

func LoadSizeUnit2F(value float64) string {
	if value > 1073741824 {
		return fmt.Sprintf("%.2fG", value/1073741824)
	}
	if value > 1048576 {
		return fmt.Sprintf("%.2fM", value/1048576)
	}
	if value > 1024 {
		return fmt.Sprintf("%.2fK", value/1024)
	}
	return fmt.Sprintf("%.2f", value)
}

func LoadTimeZone() string {
	loc := time.Now().Location()
	if _, err := time.LoadLocation(loc.String()); err != nil {
		return "Asia/Shanghai"
	}
	return loc.String()
}
func LoadTimeZoneByCmd() string {
	loc := time.Now().Location().String()
	if _, err := time.LoadLocation(loc); err != nil {
		loc = "Asia/Shanghai"
	}
	std, err := cmd.Exec("timedatectl | grep 'Time zone'")
	if err != nil {
		return loc
	}
	fields := strings.Fields(string(std))
	if len(fields) != 5 {
		return loc
	}
	if _, err := time.LoadLocation(fields[2]); err != nil {
		return loc
	}
	return fields[2]
}

func IsValidDomain(domain string) bool {
	pattern := `^([\w\p{Han}\-\*]{1,100}\.){1,10}([\w\p{Han}\-]{1,24}|[\w\p{Han}\-]{1,24}\.[\w\p{Han}\-]{1,24})(:\d{1,5})?$`
	match, err := regexp.MatchString(pattern, domain)
	if err != nil {
		return false
	}
	return match
}

func ContainsChinese(text string) bool {
	for _, char := range text {
		if unicode.Is(unicode.Han, char) {
			return true
		}
	}
	return false
}

func PunycodeEncode(text string) (string, error) {
	encoder := idna.New()
	ascii, err := encoder.ToASCII(text)
	if err != nil {
		return "", err
	}
	return ascii, nil
}
