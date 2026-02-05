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
	"gopkg.in/yaml.v3"
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
	return cmd.NewCommandMgr(cmd.WithTimeout(20*time.Minute)).RunWithStdoutBashCf("%s %s up -d", global.CONF.DockerConfig.Command, loadFiles(filePath))
}

func UpWithTask(filePath string, task *task.Task, forcePull bool) error {
	if err := pullComposeImages(filePath, forcePull, task); err != nil {
		return err
	}
	return cmd.NewCommandMgr(cmd.WithTask(*task)).RunBashCf("%s %s up -d", global.CONF.DockerConfig.Command, loadFiles(filePath))
}

func pullComposeImages(filePath string, forcePull bool, task *task.Task) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	env, _ := os.ReadFile(path.Join(path.Dir(filePath), ".env"))
	var compose docker.ComposeProject
	if err := yaml.Unmarshal(content, &compose); err != nil {
		return fmt.Errorf("parse docker-compose file failed: %v", err)
	}
	images, err := docker.GetImagesFromDockerCompose(env, content)
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

func Down(filePath string) (string, error) {
	if err := checkCmd(); err != nil {
		return "", err
	}
	return cmd.NewCommandMgr(cmd.WithTimeout(20*time.Minute)).RunWithStdoutBashCf("%s %s down --remove-orphans", global.CONF.DockerConfig.Command, loadFiles(filePath))
}

func Stop(filePath string) (string, error) {
	if err := checkCmd(); err != nil {
		return "", err
	}
	return cmd.NewCommandMgr(cmd.WithTimeout(20*time.Minute)).RunWithStdoutBashCf("%s %s stop", global.CONF.DockerConfig.Command, loadFiles(filePath))
}

func Restart(filePath string) (string, error) {
	if err := checkCmd(); err != nil {
		return "", err
	}
	return cmd.NewCommandMgr(cmd.WithTimeout(20*time.Minute)).RunWithStdoutBashCf("%s %s restart", global.CONF.DockerConfig.Command, loadFiles(filePath))
}

func Operate(filePath, operation string) (string, error) {
	if err := checkCmd(); err != nil {
		return "", err
	}
	return cmd.NewCommandMgr(cmd.WithTimeout(20*time.Minute)).RunWithStdoutBashCf("%s %s %s", global.CONF.DockerConfig.Command, loadFiles(filePath), operation)
}

func DownAndUp(filePath string) (string, error) {
	if err := checkCmd(); err != nil {
		return "", err
	}
	cmdMgr := cmd.NewCommandMgr(cmd.WithTimeout(20 * time.Minute))
	stdout, err := cmdMgr.RunWithStdoutBashCf("%s %s down", global.CONF.DockerConfig.Command, loadFiles(filePath))
	if err != nil {
		return stdout, err
	}
	stdout, err = cmdMgr.RunWithStdoutBashCf("%s %s up -d", global.CONF.DockerConfig.Command, loadFiles(filePath))
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
