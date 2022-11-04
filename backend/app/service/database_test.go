package service

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"testing"
)

func TestMysql(t *testing.T) {
	path := "/Users/slooop/go/src/github.com/1Panel/apps/mysql/5.7.39/conf/my.cnf"

	var lines []string
	lineBytes, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Println(err)
	} else {
		lines = strings.Split(string(lineBytes), "\n")
	}
	var newLines []string

	start := "[mysqld]"
	isOn := false
	hasKey := false
	regItem, _ := regexp.Compile(`^\[*\]`)
	i := 0
	for _, line := range lines {
		i++
		if strings.HasPrefix(line, start) {
			isOn = true
			newLines = append(newLines, line)
			continue
		}
		if !isOn {
			newLines = append(newLines, line)
			continue
		}
		if strings.HasPrefix(line, "user") || strings.HasPrefix(line, "# user") {
			newLines = append(newLines, "user="+"ON")
			hasKey = true
			continue
		}
		isDeadLine := regItem.Match([]byte(line))
		if !isDeadLine {
			newLines = append(newLines, line)
			continue
		}
		if !hasKey {
			newLines = append(newLines, "user="+"ON \n")
			newLines = append(newLines, line)
		}
	}

	file, err := os.OpenFile(path, os.O_WRONLY, 0666)
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()
	_, err = file.WriteString(strings.Join(newLines, "\n"))
	if err != nil {
		fmt.Println(err)
	}
}
