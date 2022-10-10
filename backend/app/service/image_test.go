package service

import (
	"context"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/1Panel-dev/1Panel/utils/docker"
)

func TestImage(t *testing.T) {
	file, err := os.OpenFile(("/tmp/nginx.tar"), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0666)
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()

	client, err := docker.NewDockerClient()
	if err != nil {
		fmt.Println(err)
	}
	out, err := client.ImageSave(context.TODO(), []string{"nginx:1.14.2"})
	fmt.Println(err)
	defer out.Close()
	if _, err = io.Copy(file, out); err != nil {
		fmt.Println(err)
	}
}
