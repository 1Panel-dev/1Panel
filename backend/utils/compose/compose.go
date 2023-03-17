package compose

import (
	"github.com/1Panel-dev/1Panel/backend/utils/cmd"
	"github.com/compose-spec/compose-go/loader"
	"github.com/compose-spec/compose-go/types"
)

func Up(filePath string) (string, error) {
	stdout, err := cmd.Execf("docker-compose -f %s up -d --quiet-pull", filePath)
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
