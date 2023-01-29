package service

import (
	"context"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/google/go-github/github"
)

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
