package compose

import (
	"github.com/compose-spec/compose-go/loader"
	"github.com/compose-spec/compose-go/types"
	"os/exec"
)

func Up(filePath string) (string, error) {
	cmd := exec.Command("docker-compose", "-f", filePath, "up", "-d")
	stdout, err := cmd.CombinedOutput()
	return string(stdout), err
}

func Down(filePath string) (string, error) {
	cmd := exec.Command("docker-compose", "-f", filePath, "stop")
	stdout, err := cmd.CombinedOutput()
	return string(stdout), err
}

func Restart(filePath string) (string, error) {
	cmd := exec.Command("docker-compose", "-f", filePath, "restart")
	stdout, err := cmd.CombinedOutput()
	return string(stdout), err
}

func Rmf(filePath string) (string, error) {
	cmd := exec.Command("docker-compose", "-f", filePath, "rm", "-f")
	stdout, err := cmd.CombinedOutput()
	return string(stdout), err
}

func GetComposeProject(yml []byte, env map[string]string) (*types.Project, error) {
	var configFiles []types.ConfigFile
	configFiles = append(configFiles, types.ConfigFile{
		Filename: "docker-compose.yml",
		Content:  yml},
	)
	details := types.ConfigDetails{
		WorkingDir:  "",
		ConfigFiles: configFiles,
		Environment: env,
	}

	project, err := loader.Load(details, func(options *loader.Options) {

	})
	if err != nil {
		return nil, err
	}
	return project, nil
}
