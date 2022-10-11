package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/1Panel-dev/1Panel/app/dto"
	"github.com/1Panel-dev/1Panel/constant"
	"github.com/1Panel-dev/1Panel/utils/docker"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/pkg/archive"
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

func TestBuild(t *testing.T) {
	client, err := docker.NewDockerClient()
	if err != nil {
		fmt.Println(err)
	}
	tar, err := archive.TarWithOptions("/Users/slooop/Documents/neeko/", &archive.TarOptions{})
	if err != nil {
		fmt.Println(err)
	}

	opts := types.ImageBuildOptions{
		Dockerfile: "Dockerfile",
		Tags:       []string{"neeko" + "/test"},
		Remove:     true,
	}
	res, err := client.ImageBuild(context.TODO(), tar, opts)
	if err != nil {
		fmt.Println(err)
	}
	defer res.Body.Close()
}

func TestDeam(t *testing.T) {
	file, err := ioutil.ReadFile(constant.DaemonJsonDir)
	if err != nil {
		fmt.Println(err)
	}
	deamonMap := make(map[string]interface{})
	err = json.Unmarshal(file, &deamonMap)
	fmt.Println(err)
	for k, v := range deamonMap {
		fmt.Println(k, v)
	}
	if _, ok := deamonMap["insecure-registries"]; ok {
		if k, v := deamonMap["insecure-registries"].(string); v {
			fmt.Println("string ", k)
		}
		if k, v := deamonMap["insecure-registries"].([]interface{}); v {
			fmt.Println("[]string ", k)
			k = append(k, "172.16.10.111:8085")
			deamonMap["insecure-registries"] = k
		}
	}
	newss, err := json.Marshal(deamonMap)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(newss))
	if err := ioutil.WriteFile(constant.DaemonJsonDir, newss, 0777); err != nil {
		fmt.Println(err)
	}
}

func TestNetwork(t *testing.T) {
	client, err := docker.NewDockerClient()
	if err != nil {
		fmt.Println(err)
	}
	var data []dto.Volume
	list, err := client.VolumeList(context.TODO(), filters.NewArgs())
	if err != nil {
		fmt.Println(err)
	}
	for _, item := range list.Volumes {
		tag := make([]string, 0)
		for _, val := range item.Labels {
			tag = append(tag, val)
		}
		createTime, _ := time.Parse("2006-01-02T15:04:05Z", item.CreatedAt)
		data = append(data, dto.Volume{
			CreatedAt:  createTime,
			Name:       item.Name,
			Driver:     item.Driver,
			Mountpoint: item.Mountpoint,
			Labels:     tag,
		})
	}
	fmt.Println(data)
}
