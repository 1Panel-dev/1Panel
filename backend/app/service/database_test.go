package service

import (
	"compress/gzip"
	"fmt"
	"os"
	"os/exec"
	"testing"

	_ "github.com/go-sql-driver/mysql"
)

func TestMysql(t *testing.T) {

	gzipFile, err := os.Open("/tmp/ko.sql.gz")
	if err != nil {
		fmt.Println(err)
	}
	defer gzipFile.Close()
	gzipReader, err := gzip.NewReader(gzipFile)
	if err != nil {
		fmt.Println(err)
	}
	defer gzipReader.Close()

	cmd := exec.Command("docker", "exec", "-i", "365", "mysql", "-uroot", "-pCalong@2012", "kubeoperator")
	cmd.Stdin = gzipReader
	stdout, err := cmd.CombinedOutput()
	fmt.Println(string(stdout), err)
}
