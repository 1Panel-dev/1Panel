package service

import (
	"context"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/1Panel-dev/1Panel/backend/init/db"
	"github.com/1Panel-dev/1Panel/backend/init/viper"
	"github.com/google/go-github/github"
)

func TestDw(t *testing.T) {
	viper.Init()
	db.Init()

	backup, err := backupRepo.Get(commonRepo.WithByType("OSS"))
	if err != nil {
		fmt.Println(err)
	}
	client, err := NewIBackupService().NewClient(&backup)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(client.Download("system_snapshot/1panel_snapshot_20230112135640.tar.gz", "/opt/1Panel/data/backup/system/test.tar.gz"))
}

func TestDi(t *testing.T) {
	docker := "var/lib/docker"
	fmt.Println(docker[strings.LastIndex(docker, "/"):])
	fmt.Println(docker[:strings.LastIndex(docker, "/")])
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
