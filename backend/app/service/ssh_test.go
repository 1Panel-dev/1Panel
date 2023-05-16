package service

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/utils/cmd"
)

func TestCa(t *testing.T) {
	var (
		fileList        []string
		datas           []history
		successfulCount int
		failedCount     int
	)
	baseDir := "/Users/slooop/Downloads"
	if err := filepath.Walk(baseDir, func(pathItem string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasPrefix(info.Name(), "secure") || strings.HasPrefix(info.Name(), "auth") {
			fileList = append(fileList, strings.ReplaceAll(pathItem, ".gz", ""))
		}
		return nil
	}); err != nil {
		fmt.Println(err)
	}
	for i := 0; i < len(fileList); i++ {
		if strings.HasPrefix(path.Base(fileList[i]), "secure") {
			dataItem := loadDatas2(fmt.Sprintf("cat %s | grep -a 'Failed password for' | grep -v 'invalid'", fileList[i]), 14, constant.StatusFailed)
			failedCount += len(dataItem)
			datas = append(datas, dataItem...)
		}
		if strings.HasPrefix(path.Base(fileList[i]), "auth.log") {
			dataItem := loadDatas2(fmt.Sprintf("cat %s | grep -a 'Connection closed by authenticating user' | grep -a 'preauth'", fileList[i]), 15, constant.StatusFailed)
			failedCount += len(dataItem)
			datas = append(datas, dataItem...)
		}
		dataItem := loadDatas2(fmt.Sprintf("cat %s | grep Accepted", fileList[i]), 14, constant.StatusSuccess)
		datas = append(datas, dataItem...)
	}
	successfulCount = len(datas) - failedCount
	fmt.Println(len(datas), successfulCount, failedCount)
}

func loadDatas2(command string, length int, status string) []history {
	var datas []history
	stdout2, err := cmd.Exec(command)
	if err == nil {
		lines := strings.Split(string(stdout2), "\n")
		for _, line := range lines {
			parts := strings.Fields(line)
			if len(parts) != length {
				continue
			}
			historyItem := history{
				Belong:   parts[3],
				User:     parts[8],
				AuthMode: parts[6],
				Address:  parts[10],
				Port:     parts[12],
				Status:   status,
			}
			dateStr := fmt.Sprintf("%d %s %s %s", time.Now().Year(), parts[0], parts[1], parts[2])
			historyItem.Date, _ = time.Parse("2006 Jan 2 15:04:05", dateStr)
			// if err != nil {
			// 	historyItem.Date, _ = time.Parse("2006 Jan 2 15:04:05", dateStr)
			// }
			fmt.Println(dateStr + "===>" + historyItem.Date.Format("2006.01.02 15:04:05"))
			datas = append(datas, historyItem)
		}
	}
	return datas
}

func TestCas(t *testing.T) {
	ss := "2023 May 9 14:48:28"
	kk, err := time.Parse("2006 Jan 2 15:04:05", ss)
	fmt.Println(kk, err)
}

type history struct {
	Date     time.Time
	Belong   string
	User     string
	AuthMode string
	Address  string
	Port     string
	Status   string
	Message  string
}
