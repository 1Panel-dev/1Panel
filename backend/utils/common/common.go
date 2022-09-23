package common

import (
	"regexp"
	"strconv"
	"strings"
)

func CompareVersion(version1 string, version2 string) bool {
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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
