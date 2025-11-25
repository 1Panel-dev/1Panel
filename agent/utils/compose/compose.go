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
	"github.com/goccy/go-yaml"
)

func checkCmd() error {
	if global.CONF.DockerConfig.Command == "" {
		dockerComposCmd := common.GetDockerComposeCommand()
		if dockerComposCmd == "" {
			return buserr.New("ErrDockerComposeCmdNotFound")
		}
		global.CONF.DockerConfig.Command = dockerComposCmd
	}
	return nil
}

func Up(filePath string) (string, error) {
	if err := checkCmd(); err != nil {
		return "", err
	}
	stdout, err := cmd.RunDefaultWithStdoutBashCfAndTimeOut(global.CONF.DockerConfig.Command+" -f %s up -d", 20*time.Minute, filePath)
	return stdout, err
}

func UpWithTask(filePath string, task *task.Task) error {
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
	errMsg := ""
	for _, image := range images {
		task.Log(i18n.GetWithName("PullImageStart", image))
		if err = dockerCLi.PullImageWithProcess(task, image); err != nil {
			errOur := err.Error()
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
			task.LogFailedWithErr(i18n.GetMsgByKey("PullImage"), installErr)
			if exist, _ := dockerCLi.ImageExists(image); !exist {
				return installErr
			} else {
				task.Log(i18n.GetMsgByKey("UseExistImage"))
			}
		} else {
			task.Log(i18n.GetMsgByKey("PullImageSuccess"))
		}
	}

	dockerCommand := global.CONF.DockerConfig.Command
	if dockerCommand == "docker-compose" {
		return cmd.NewCommandMgr(cmd.WithTask(*task)).Run("docker-compose", "-f", filePath, "up", "-d")
	} else {
		return cmd.NewCommandMgr(cmd.WithTask(*task)).Run("docker", "compose", "-f", filePath, "up", "-d")
	}
}

func Down(filePath string) (string, error) {
	if err := checkCmd(); err != nil {
		return "", err
	}
	stdout, err := cmd.RunDefaultWithStdoutBashCfAndTimeOut(global.CONF.DockerConfig.Command+" -f %s down --remove-orphans", 20*time.Minute, filePath)
	return stdout, err
}

func Stop(filePath string) (string, error) {
	if err := checkCmd(); err != nil {
		return "", err
	}
	stdout, err := cmd.RunDefaultWithStdoutBashCf(global.CONF.DockerConfig.Command+" -f %s stop", filePath)
	return stdout, err
}

func Restart(filePath string) (string, error) {
	if err := checkCmd(); err != nil {
		return "", err
	}
	stdout, err := cmd.RunDefaultWithStdoutBashCf(global.CONF.DockerConfig.Command+" -f %s restart", filePath)
	return stdout, err
}

func Operate(filePath, operation string) (string, error) {
	if err := checkCmd(); err != nil {
		return "", err
	}
	stdout, err := cmd.RunDefaultWithStdoutBashCf(global.CONF.DockerConfig.Command+" -f %s %s", filePath, operation)
	return stdout, err
}

func DownAndUp(filePath string) (string, error) {
	if err := checkCmd(); err != nil {
		return "", err
	}
	stdout, err := cmd.RunDefaultWithStdoutBashCf(global.CONF.DockerConfig.Command+" -f %s down", filePath)
	if err != nil {
		return stdout, err
	}
	stdout, err = cmd.RunDefaultWithStdoutBashCf(global.CONF.DockerConfig.Command+" -f %s up -d", filePath)
	return stdout, err
}
