package compose

import "os/exec"

func Up(filePath string) (string, error) {
	cmd := exec.Command("docker-compose", "-f", filePath, "up", "-d")
	stdout, err := cmd.CombinedOutput()
	if err != nil {
		return string(stdout), err
	}
	return string(stdout), nil
}

func Down(filePath string) (string, error) {
	cmd := exec.Command("docker-compose", "-f", filePath, "down")
	stdout, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return string(stdout), nil
}

func Restart(filePath string) (string, error) {
	cmd := exec.Command("docker-compose", "-f", filePath, "restart")
	stdout, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return string(stdout), nil
}

func Rmf(filePath string) (string, error) {
	cmd := exec.Command("docker-compose", "-f", filePath, "rm", "-f")
	stdout, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return string(stdout), nil
}
