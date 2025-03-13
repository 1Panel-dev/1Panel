package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/task"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"io"
	"os"
	"strings"
	"time"
)

func NewDockerClient() (*client.Client, error) {
	var settingItem model.Setting
	_ = global.DB.Where("key = ?", "DockerSockPath").First(&settingItem).Error
	if len(settingItem.Value) == 0 {
		settingItem.Value = "unix:///var/run/docker.sock"
	}
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithHost(settingItem.Value), client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	return cli, nil
}

func NewClient() (Client, error) {
	var settingItem model.Setting
	_ = global.DB.Where("key = ?", "DockerSockPath").First(&settingItem).Error
	if len(settingItem.Value) == 0 {
		settingItem.Value = "unix:///var/run/docker.sock"
	}
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithHost(settingItem.Value), client.WithAPIVersionNegotiation())
	if err != nil {
		return Client{}, err
	}

	return Client{
		cli: cli,
	}, nil
}

func NewClientWithExist(cli *client.Client) Client {
	return Client{
		cli: cli,
	}
}

type Client struct {
	cli *client.Client
}

func (c Client) Close() {
	_ = c.cli.Close()
}

func (c Client) ListContainersByName(names []string) ([]types.Container, error) {
	var (
		options  container.ListOptions
		namesMap = make(map[string]bool)
		res      []types.Container
	)
	options.All = true
	if len(names) > 0 {
		var array []filters.KeyValuePair
		for _, n := range names {
			namesMap["/"+n] = true
			array = append(array, filters.Arg("name", n))
		}
		options.Filters = filters.NewArgs(array...)
	}
	containers, err := c.cli.ContainerList(context.Background(), options)
	if err != nil {
		return nil, err
	}
	for _, con := range containers {
		if _, ok := namesMap[con.Names[0]]; ok {
			res = append(res, con)
		}
	}
	return res, nil
}
func (c Client) ListAllContainers() ([]types.Container, error) {
	var (
		options container.ListOptions
	)
	options.All = true
	containers, err := c.cli.ContainerList(context.Background(), options)
	if err != nil {
		return nil, err
	}
	return containers, nil
}

func (c Client) CreateNetwork(name string) error {
	_, err := c.cli.NetworkCreate(context.Background(), name, network.CreateOptions{
		Driver:     "bridge",
		EnableIPv6: new(bool),
	})
	return err
}

func (c Client) DeleteImage(imageID string) error {
	if _, err := c.cli.ImageRemove(context.Background(), imageID, image.RemoveOptions{Force: true}); err != nil {
		return err
	}
	return nil
}

func (c Client) GetImageIDByName(imageName string) (string, error) {
	filter := filters.NewArgs()
	filter.Add("reference", imageName)
	list, err := c.cli.ImageList(context.Background(), image.ListOptions{
		Filters: filter,
	})
	if err != nil {
		return "", err
	}
	if len(list) > 0 {
		return list[0].ID, nil
	}
	return "", nil
}

func (c Client) NetworkExist(name string) bool {
	var options network.ListOptions
	options.Filters = filters.NewArgs(filters.Arg("name", name))
	networks, err := c.cli.NetworkList(context.Background(), options)
	if err != nil {
		return false
	}
	return len(networks) > 0
}

func CreateDefaultDockerNetwork() error {
	cli, err := NewClient()
	if err != nil {
		global.LOG.Errorf("init docker client error %s", err.Error())
		return err
	}

	defer cli.Close()
	if !cli.NetworkExist("1panel-network") {
		if err := cli.CreateNetwork("1panel-network"); err != nil {
			global.LOG.Errorf("create default docker network  error %s", err.Error())
			return err
		}
	}
	return nil
}

func setLog(id, newLastLine string, task *task.Task) error {
	data, err := os.ReadFile(task.Task.LogFile)
	if err != nil {
		return fmt.Errorf("failed to read file: %v", err)
	}
	lines := strings.Split(string(data), "\n")
	exist := false
	for index, line := range lines {
		if strings.Contains(line, id) {
			lines[index] = newLastLine
			exist = true
			break
		}
	}
	if !exist {
		task.Log(newLastLine)
		return nil
	}
	output := strings.Join(lines, "\n")
	_ = os.WriteFile(task.Task.LogFile, []byte(output), os.ModePerm)
	return nil
}

func (c Client) PullImageWithProcessAndOptions(task *task.Task, imageName string, options image.PullOptions) error {
	out, err := c.cli.ImagePull(context.Background(), imageName, options)
	if err != nil {
		return err
	}
	defer out.Close()
	decoder := json.NewDecoder(out)
	for {
		var progress map[string]interface{}
		if err = decoder.Decode(&progress); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		timeStr := time.Now().Format("2006/01/02 15:04:05")
		status, _ := progress["status"].(string)
		if status == "Downloading" || status == "Extracting" {
			id, _ := progress["id"].(string)
			progressDetail, _ := progress["progressDetail"].(map[string]interface{})
			current, _ := progressDetail["current"].(float64)
			progressStr := ""
			total, ok := progressDetail["total"].(float64)
			if ok {
				progressStr = fmt.Sprintf("%s %s [%s] --- %.2f%%", timeStr, status, id, (current/total)*100)
			} else {
				progressStr = fmt.Sprintf("%s %s [%s] --- %.2f%%", timeStr, status, id, current)
			}

			_ = setLog(id, progressStr, task)
		}
		if status == "Pull complete" || status == "Download complete" {
			id, _ := progress["id"].(string)
			progressStr := fmt.Sprintf("%s %s [%s] --- %.2f%%", timeStr, status, id, 100.0)
			_ = setLog(id, progressStr, task)
		}
	}
	return nil
}

func (c Client) PullImageWithProcess(task *task.Task, imageName string) error {
	return c.PullImageWithProcessAndOptions(task, imageName, image.PullOptions{})
}
