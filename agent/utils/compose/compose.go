package compose

import (
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"time"
)

func Up(filePath string) (string, error) {
	stdout, err := cmd.RunDefaultWithStdoutBashCfAndTimeOut("docker compose -f %s up -d", 20*time.Minute, filePath)
	return stdout, err
}

func Down(filePath string) (string, error) {
	stdout, err := cmd.RunDefaultWithStdoutBashCf("docker compose -f %s down --remove-orphans", filePath)
	return stdout, err
}

func Stop(filePath string) (string, error) {
	stdout, err := cmd.RunDefaultWithStdoutBashCf("docker compose -f %s stop", filePath)
	return stdout, err
}

func Restart(filePath string) (string, error) {
	stdout, err := cmd.RunDefaultWithStdoutBashCf("docker compose -f %s restart", filePath)
	return stdout, err
}

func Operate(filePath, operation string) (string, error) {
	stdout, err := cmd.RunDefaultWithStdoutBashCf("docker compose -f %s %s", filePath, operation)
	return stdout, err
}

func DownAndUp(filePath string) (string, error) {
	stdout, err := cmd.RunDefaultWithStdoutBashCf("docker compose -f %s down", filePath)
	if err != nil {
		return stdout, err
	}
	stdout, err = cmd.RunDefaultWithStdoutBashCf("docker compose -f %s up -d", filePath)
	return stdout, err
}
