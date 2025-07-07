package compose

import (
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/1Panel-dev/1Panel/agent/utils/common"
	"time"
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
