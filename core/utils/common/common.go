package common

import (
	"fmt"
	"github.com/gin-gonic/gin"
	mathRand "math/rand"
	"net"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/core/global"
	"github.com/1Panel-dev/1Panel/core/utils/cmd"
)

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

func ScanPort(port int) bool {
	ln, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		return true
	}
	defer ln.Close()
	return false
}

func ComparePanelVersion(version1, version2 string) bool {
	if version1 == version2 {
		return false
	}
	version1s := SplitStr(version1, ".", "-")
	version2s := SplitStr(version2, ".", "-")

	if len(version2s) > len(version1s) {
		for i := 0; i < len(version2s)-len(version1s); i++ {
			version1s = append(version1s, "0")
		}
	}
	if len(version1s) > len(version2s) {
		for i := 0; i < len(version1s)-len(version2s); i++ {
			version2s = append(version2s, "0")
		}
	}

	n := min(len(version1s), len(version2s))
	for i := 0; i < n; i++ {
		if version1s[i] == version2s[i] {
			continue
		} else {
			v1, err1 := strconv.Atoi(version1s[i])
			if err1 != nil {
				return version1s[i] > version2s[i]
			}
			v2, err2 := strconv.Atoi(version2s[i])
			if err2 != nil {
				return version1s[i] > version2s[i]
			}
			return v1 > v2
		}
	}
	return true
}

func SplitStr(str string, spi ...string) []string {
	lists := []string{str}
	var results []string
	for _, s := range spi {
		results = []string{}
		for _, list := range lists {
			results = append(results, strings.Split(list, s)...)
		}
		lists = results
	}
	return results
}

func LoadArch() (string, error) {
	std, err := cmd.Exec("uname -a")
	if err != nil {
		return "", fmt.Errorf("std: %s, err: %s", std, err.Error())
	}
	if strings.Contains(std, "x86_64") {
		return "amd64", nil
	}
	if strings.Contains(std, "arm64") || strings.Contains(std, "aarch64") {
		return "arm64", nil
	}
	if strings.Contains(std, "armv7l") {
		return "armv7", nil
	}
	if strings.Contains(std, "ppc64le") {
		return "ppc64le", nil
	}
	if strings.Contains(std, "s390x") {
		return "s390x", nil
	}
	return "", fmt.Errorf("unsupported such arch: %s", std)
}

func Clean(str []byte) {
	for i := 0; i < len(str); i++ {
		str[i] = 0
	}
}

func CreateDirWhenNotExist(isDir bool, pathItem string) (string, error) {
	checkPath := pathItem
	if !isDir {
		checkPath = path.Dir(pathItem)
	}
	if _, err := os.Stat(checkPath); err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(checkPath, os.ModePerm); err != nil {
			global.LOG.Errorf("mkdir %s failed, err: %v", checkPath, err)
			return pathItem, err
		}
	}
	return pathItem, nil
}

func GetLang(c *gin.Context) string {
	lang := c.GetHeader("Accept-Language")
	if lang == "" {
		lang = "en"
	}
	return lang
}

func CheckIpInCidr(cidr, checkIP string) bool {
	ip, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		global.LOG.Errorf("parse CIDR %s failed, err: %v", cidr, err)
		return false
	}
	for ip := ip.Mask(ipNet.Mask); ipNet.Contains(ip); incIP(ip) {
		if ip.String() == checkIP {
			return true
		}
	}
	return false
}

func incIP(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}
