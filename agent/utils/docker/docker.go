package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/task"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/docker/cli/cli/config"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/client"
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

func (c Client) ListContainersByName(names []string) ([]container.Summary, error) {
	var (
		options  container.ListOptions
		namesMap = make(map[string]bool)
		res      []container.Summary
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
func (c Client) ListAllContainers() ([]container.Summary, error) {
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
func (c Client) ImageExists(imageID string) (bool, error) {
	_, err := c.cli.ImageInspect(context.Background(), imageID)
	if err != nil {
		return false, err
	}
	return true, nil
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
		global.LOG.Warnf("init docker client error %s", err.Error())
		return err
	}

	defer cli.Close()
	if !cli.NetworkExist("1panel-network") {
		if err := cli.CreateNetwork("1panel-network"); err != nil {
			global.LOG.Warnf("create default docker network  error %s", err.Error())
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
			timeStr := time.Now().Format("2006/01/02 15:04:05")
			lines[index] = timeStr + " " + newLastLine
			exist = true
			break
		} else {
			lines[index] = strings.TrimSpace(lines[index])
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
		if msg, ok := progress["errorDetail"]; ok {
			return fmt.Errorf("image pull failed, err: %v", msg)
		}
		if msg, ok := progress["error"]; ok {
			return fmt.Errorf("image pull failed, err: %v", msg)
		}
		status, _ := progress["status"].(string)
		if status == "Downloading" || status == "Extracting" {
			logProcess(progress, task)
		}
		if status == "Pull complete" || status == "Download complete" {
			id, _ := progress["id"].(string)
			progressStr := fmt.Sprintf("%s %s", status, id)
			_ = setLog(id, progressStr, task)
		}
	}
	return nil
}

func (c Client) PushImageWithProcessAndOptions(task *task.Task, imageName string, options image.PushOptions) error {
	out, err := c.cli.ImagePush(context.Background(), imageName, options)
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
		if msg, ok := progress["errorDetail"]; ok {
			return fmt.Errorf("image push failed, err: %v", msg)
		}
		if msg, ok := progress["error"]; ok {
			return fmt.Errorf("image push failed, err: %v", msg)
		}
		status, _ := progress["status"].(string)
		switch status {
		case "Pushing":
			logProcess(progress, task)
		case "Pushed":
			id, _ := progress["id"].(string)
			progressStr := fmt.Sprintf("%s %s", status, id)
			_ = setLog(id, progressStr, task)
		default:
			progressStr, _ := json.Marshal(progress)
			task.Log(string(progressStr))
		}
	}
	return nil
}

func (c Client) BuildImageWithProcessAndOptions(task *task.Task, tar io.ReadCloser, options types.ImageBuildOptions) error {
	out, err := c.cli.ImageBuild(context.Background(), tar, options)
	if err != nil {
		return err
	}
	defer out.Body.Close()
	decoder := json.NewDecoder(out.Body)
	for {
		var progress map[string]interface{}
		if err = decoder.Decode(&progress); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		if msg, ok := progress["errorDetail"]; ok {
			return fmt.Errorf("image build failed, err: %v", msg)
		}
		if msg, ok := progress["error"]; ok {
			return fmt.Errorf("image build failed, err: %v", msg)
		}
		status, _ := progress["status"].(string)
		stream, _ := progress["stream"].(string)
		if len(status) == 0 && len(stream) != 0 {
			if stream != "\n" {
				task.Log(stream)
			}
			continue
		}
		switch status {
		case "Downloading", "Extracting":
			logProcess(progress, task)
		case "Pull complete", "Download complete", "Verifying Checksum":
			id, _ := progress["id"].(string)
			progressStr := fmt.Sprintf("%s %s", status, id)
			_ = setLog(id, progressStr, task)
		default:
			progressStr, _ := json.Marshal(progress)
			task.Log(string(progressStr))
		}
	}
	return nil
}

func (c Client) PullImageWithProcess(task *task.Task, imageName string) error {
	options := image.PullOptions{}
	if authStr, ok := loadRegistryAuthFromDockerConfig(imageName); ok {
		options.RegistryAuth = authStr
	}
	return c.PullImageWithProcessAndOptions(task, imageName, options)
}

func logProcess(progress map[string]interface{}, task *task.Task) {
	status, _ := progress["status"].(string)
	id, _ := progress["id"].(string)
	progressItem, _ := progress["progress"].(string)
	progressStr := ""
	progressStr = fmt.Sprintf("%s %s %s", status, id, progressItem)
	_ = setLog(id, progressStr, task)
}

func PullImage(imageName string) error {
	cli, err := NewDockerClient()
	if err != nil {
		return err
	}
	defer cli.Close()
	options := image.PullOptions{}
	if authStr, ok := loadRegistryAuthFromDockerConfig(imageName); ok {
		options.RegistryAuth = authStr
	}
	if _, err := cli.ImagePull(context.Background(), imageName, options); err != nil {
		return err
	}
	return nil
}

func loadRegistryAuthFromDockerConfig(imageName string) (string, bool) {
	registryHost, hasRegistry := extractRegistryHost(imageName)
	cfg := config.LoadDefaultConfigFile(io.Discard)
	if cfg == nil {
		return "", false
	}
	candidates := make([]string, 0)
	if hasRegistry {
		candidates = append(candidates, registryHost, "https://"+registryHost, "http://"+registryHost)
	}
	if !hasRegistry || isDockerHubRegistry(registryHost) {
		candidates = append(candidates,
			"https://index.docker.io/v1/",
			"index.docker.io",
			"docker.io",
			"registry-1.docker.io",
			"https://registry-1.docker.io",
		)
	}
	seen := make(map[string]struct{})
	for _, key := range candidates {
		if key == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		auth, err := cfg.GetAuthConfig(key)
		if err != nil {
			continue
		}
		if auth.Username == "" && auth.Password == "" && auth.Auth == "" && auth.IdentityToken == "" && auth.RegistryToken == "" {
			continue
		}
		authStr, err := registry.EncodeAuthConfig(registry.AuthConfig{
			Username:      auth.Username,
			Password:      auth.Password,
			Auth:          auth.Auth,
			ServerAddress: auth.ServerAddress,
			IdentityToken: auth.IdentityToken,
			RegistryToken: auth.RegistryToken,
		})
		if err != nil {
			return "", false
		}
		return authStr, true
	}
	return "", false
}

func isDockerHubRegistry(host string) bool {
	switch normalizeRegistryHost(host) {
	case "docker.io", "index.docker.io", "registry-1.docker.io":
		return true
	default:
		return false
	}
}

func extractRegistryHost(imageName string) (string, bool) {
	parts := strings.Split(imageName, "/")
	if len(parts) < 2 {
		return "", false
	}
	first := parts[0]
	if strings.Contains(first, ".") || strings.Contains(first, ":") || first == "localhost" {
		return normalizeRegistryHost(first), true
	}
	return "", false
}

func normalizeRegistryHost(registryKey string) string {
	key := strings.TrimSpace(registryKey)
	if key == "" {
		return ""
	}
	key = strings.TrimPrefix(key, "http://")
	key = strings.TrimPrefix(key, "https://")
	key = strings.Trim(key, "/")
	if strings.Contains(key, "/") {
		key = strings.SplitN(key, "/", 2)[0]
	}
	return strings.ToLower(key)
}
