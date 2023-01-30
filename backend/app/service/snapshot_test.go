package service

import (
	"context"
	"fmt"
	"io/ioutil"
	"testing"
	"time"

	"github.com/1Panel-dev/1Panel/backend/init/log"
	"github.com/1Panel-dev/1Panel/backend/init/viper"
	"github.com/1Panel-dev/1Panel/backend/utils/files"
	"github.com/google/go-github/github"
)

func TestDi(t *testing.T) {
	viper.Init()
	log.Init()
	ti := time.Now().Format("20060102150405")
	oss := "https://1panel.oss-cn-hangzhou.aliyuncs.com"
	tmpPath := fmt.Sprintf("/opt/1Panel/data/tmp/%s_upgrade.tar.gz", ti)
	fileOp := files.NewFileOp()
	downloadPath := fmt.Sprintf("%s/releases/v1.0.1/v1.0.1.tar.gz", oss)
	if err := fileOp.DownloadFile(downloadPath, tmpPath); err != nil {
		fmt.Println(err)
	}
}

func TestGit(t *testing.T) {
	client := github.NewClient(nil)
	stats, _, err := client.Repositories.GetLatestRelease(context.Background(), "KubeOperator", "KubeOperator")
	fmt.Println(github.Timestamp(*stats.PublishedAt), err)
}

func TestSdasd(t *testing.T) {
	u := NewISnapshotService()
	var snapjson SnapshotJson
	snapjson, _ = u.readFromJson("/Users/slooop/Downloads/snapshot.json")
	fmt.Println(111, snapjson)
	// if err := ioutil.WriteFile("/Users/slooop/Downloads/snapshot.json", []byte("111xxxxx"), 0640); err != nil {
	// 	fmt.Println(err)
	// }
}

func TestCp(t *testing.T) {
	_, err := ioutil.ReadFile("/Users/slooop/Downloads/test/main")
	if err != nil {
		fmt.Println(err)
	}
	if err := ioutil.WriteFile("/Users/slooop/Downloads/test/main", []byte("sdadasd"), 0640); err != nil {
		fmt.Println(err)
	}
}
