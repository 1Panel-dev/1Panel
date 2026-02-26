package compose

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/task"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/i18n"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/1Panel-dev/1Panel/agent/utils/common"
	"github.com/1Panel-dev/1Panel/agent/utils/docker"
)

func checkCmd() error {
	if global.CONF.DockerConfig.Command == "" {
		dockerComposeCmd := common.GetDockerComposeCommand()
		if dockerComposeCmd == "" {
			return buserr.New("ErrDockerComposeCmdNotFound")
		}
		global.CONF.DockerConfig.Command = dockerComposeCmd
	}
	return nil
}

func Up(filePath string) (string, error) {
	if err := checkCmd(); err != nil {
		return "", err
	}
	return cmd.NewCommandMgr(cmd.WithTimeout(20*time.Minute)).RunWithStdout(global.CONF.DockerConfig.Command, loadFiles(filePath), "up", "-d")
}

func UpWithTask(filePath string, task *task.Task, forcePull bool) error {
	if err := pullComposeImages(filePath, forcePull, task); err != nil {
		return err
	}
	return cmd.NewCommandMgr(cmd.WithTask(*task)).Run(global.CONF.DockerConfig.Command, loadFiles(filePath), "up", "-d")
}

func pullComposeImages(filePath string, forcePull bool, task *task.Task) error {
	images, err := GetComposeImages(filePath)
	if err != nil {
		return err
	}
	dockerCLi, err := docker.NewClient()
	if err != nil {
		return err
	}
	for _, image := range images {
		if !forcePull {
			if exist, _ := dockerCLi.ImageExists(image); exist {
				if task != nil {
					task.Log(i18n.GetMsgByKey("UseExistImage"))
				}
				continue
			}
		}

		if task != nil {
			task.Log(i18n.GetWithName("PullImageStart", image))
		}
		pullErr := error(nil)
		if task != nil {
			pullErr = dockerCLi.PullImageWithProcess(task, image)
		} else {
			pullErr = docker.PullImage(image)
		}
		if pullErr != nil {
			errMsg := ""
			errOur := pullErr.Error()
			if errOur != "" {
				if strings.Contains(errOur, "no such host") {
					errMsg = i18n.GetMsgByKey("ErrNoSuchHost") + ":"
				}
				if strings.Contains(errOur, "Error response from daemon") {
					errMsg = i18n.GetMsgByKey("PullImageTimeout") + ":"
				}
			}
			message := errMsg + errOur
			installErr := errors.New(message)
			if task != nil {
				task.LogFailedWithErr(i18n.GetMsgByKey("PullImage"), installErr)
			}
			if exist, _ := dockerCLi.ImageExists(image); !exist {
				return installErr
			}
			if task != nil {
				task.Log(i18n.GetMsgByKey("UseExistImage"))
			}
		} else if task != nil {
			task.Log(i18n.GetMsgByKey("PullImageSuccess"))
		}
	}

	return nil
}

func GetComposeImages(filePath string) ([]string, error) {
	images, err := getComposeImagesByCommand(filePath)
	if err == nil {
		return images, nil
	}

	content, readErr := os.ReadFile(filePath)
	if readErr != nil {
		return nil, readErr
	}
	env, _ := os.ReadFile(path.Join(path.Dir(filePath), ".env"))
	images, parseErr := docker.GetImagesFromDockerCompose(env, content)
	if parseErr != nil {
		return nil, fmt.Errorf("get compose images failed, cmd err: %v, parse err: %v", err, parseErr)
	}
	return images, nil
}

func getComposeImagesByCommand(filePath string) ([]string, error) {
	if err := checkCmd(); err != nil {
		return nil, err
	}
	stdout, err := cmd.NewCommandMgr(cmd.WithTimeout(5*time.Minute)).
		RunWithStdout(global.CONF.DockerConfig.Command, loadFiles(filePath), "config", "--images")
	if err != nil {
		return nil, fmt.Errorf("run compose config --images failed, std: %s, err: %v", stdout, err)
	}

	var images []string
	seen := make(map[string]struct{})
	for _, line := range strings.Split(stdout, "\n") {
		image := strings.TrimSpace(line)
		if len(image) == 0 {
			continue
		}
		if _, ok := seen[image]; ok {
			continue
		}
		seen[image] = struct{}{}
		images = append(images, image)
	}
	if len(images) == 0 {
		return nil, errors.New("no images found from compose config")
	}
	return images, nil
}

func Down(filePath string) (string, error) {
	if err := checkCmd(); err != nil {
		return "", err
	}
	return cmd.NewCommandMgr(cmd.WithTimeout(20*time.Minute)).RunWithStdout(global.CONF.DockerConfig.Command, loadFiles(filePath), "down", "--remove-orphans")
}

func Stop(filePath string) (string, error) {
	if err := checkCmd(); err != nil {
		return "", err
	}
	return cmd.NewCommandMgr(cmd.WithTimeout(20*time.Minute)).RunWithStdout(global.CONF.DockerConfig.Command, loadFiles(filePath), "stop")
}

func Restart(filePath string) (string, error) {
	if err := checkCmd(); err != nil {
		return "", err
	}
	return cmd.NewCommandMgr(cmd.WithTimeout(20*time.Minute)).RunWithStdout(global.CONF.DockerConfig.Command, loadFiles(filePath), "restart")
}

func Operate(filePath, operation string) (string, error) {
	if err := checkCmd(); err != nil {
		return "", err
	}
	return cmd.NewCommandMgr(cmd.WithTimeout(20*time.Minute)).RunWithStdout(global.CONF.DockerConfig.Command, loadFiles(filePath), operation)
}

func DownAndUp(filePath string) (string, error) {
	if err := checkCmd(); err != nil {
		return "", err
	}
	cmdMgr := cmd.NewCommandMgr(cmd.WithTimeout(20 * time.Minute))
	stdout, err := cmdMgr.RunWithStdout(global.CONF.DockerConfig.Command, loadFiles(filePath), "down")
	if err != nil {
		return stdout, err
	}
	stdout, err = cmdMgr.RunWithStdout(global.CONF.DockerConfig.Command, loadFiles(filePath), "up", "-d")
	return stdout, err
}

func loadFiles(filePath string) string {
	var fileItem []string
	for _, item := range strings.Split(filePath, ",") {
		if len(item) != 0 {
			fileItem = append(fileItem, fmt.Sprintf("-f %s", item))
		}
	}
	return strings.Join(fileItem, " ")
}
