package compose

import (
	"github.com/1Panel-dev/1Panel/backend/utils/cmd"
)

func Up(filePath string) (string, error) {
	stdout, err := cmd.Execf("docker-compose -f %s up -d", filePath)
	return stdout, err
}

func Down(filePath string) (string, error) {
	stdout, err := cmd.Execf("docker-compose -f %s down", filePath)
	return stdout, err
}

func Start(filePath string) (string, error) {
	stdout, err := cmd.Execf("docker-compose -f %s start", filePath)
	return stdout, err
}

func Stop(filePath string) (string, error) {
	stdout, err := cmd.Execf("docker-compose -f %s stop", filePath)
	return stdout, err
}

func Restart(filePath string) (string, error) {
	stdout, err := cmd.Execf("docker-compose -f %s restart", filePath)
	return stdout, err
}

func Operate(filePath, operation string) (string, error) {
	stdout, err := cmd.Execf("docker-compose -f %s %s", filePath, operation)
	return stdout, err
}
