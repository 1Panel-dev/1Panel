package docker

import (
	"context"
	"fmt"
	"github.com/compose-spec/compose-go/v2/loader"
	"github.com/compose-spec/compose-go/v2/types"
	"github.com/docker/compose/v2/pkg/api"
	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
	"path"
	"regexp"
	"strings"
)

type ComposeService struct {
	api.Service
	project *types.Project
}

func GetComposeProject(projectName, workDir string, yml []byte, env []byte, skipNormalization bool) (*types.Project, error) {
	var configFiles []types.ConfigFile
	configFiles = append(configFiles, types.ConfigFile{
		Filename: "docker-compose.yml",
		Content:  yml},
	)
	envMap, err := godotenv.UnmarshalBytes(env)
	if err != nil {
		return nil, err
	}
	details := types.ConfigDetails{
		WorkingDir:  workDir,
		ConfigFiles: configFiles,
		Environment: envMap,
	}
	projectName = strings.ToLower(projectName)
	reg, _ := regexp.Compile(`[^a-z0-9_-]+`)
	projectName = reg.ReplaceAllString(projectName, "")
	project, err := loader.LoadWithContext(context.Background(), details, func(options *loader.Options) {
		options.SetProjectName(projectName, true)
		options.ResolvePaths = true
		options.SkipNormalization = skipNormalization
	})
	if err != nil {
		return nil, err
	}
	project.ComposeFiles = []string{path.Join(workDir, "docker-compose.yml")}
	return project, nil
}

type ComposeProject struct {
	Version  string
	Services map[string]Service `yaml:"services"`
}

type Service struct {
	Image string `yaml:"image"`
}

func replaceEnvVariables(input string, envVars map[string]string) string {
	for key, value := range envVars {
		placeholder := fmt.Sprintf("${%s}", key)
		input = strings.ReplaceAll(input, placeholder, value)
	}
	return input
}
func GetDockerComposeImagesV2(env, yml []byte) ([]string, error) {
	var (
		compose ComposeProject
		err     error
		images  []string
	)
	err = yaml.Unmarshal(yml, &compose)
	if err != nil {
		return nil, err
	}
	envMap, err := godotenv.UnmarshalBytes(env)
	if err != nil {
		return nil, err
	}
	for _, service := range compose.Services {
		image := replaceEnvVariables(service.Image, envMap)
		images = append(images, image)
	}
	return images, nil
}

func GetDockerComposeImages(projectName string, env, yml []byte) ([]string, error) {
	var (
		configFiles []types.ConfigFile
		images      []string
		imagesMap   = make(map[string]struct{})
	)
	configFiles = append(configFiles, types.ConfigFile{
		Filename: "docker-compose.yml",
		Content:  yml},
	)
	envMap, err := godotenv.UnmarshalBytes(env)
	if err != nil {
		return nil, err
	}
	details := types.ConfigDetails{
		ConfigFiles: configFiles,
		Environment: envMap,
	}

	project, err := loader.LoadWithContext(context.Background(), details, func(options *loader.Options) {
		options.SetProjectName(projectName, true)
		options.ResolvePaths = true
	})
	if err != nil {
		return nil, err
	}
	for _, service := range project.AllServices() {
		imagesMap[service.Image] = struct{}{}
	}
	for image := range imagesMap {
		images = append(images, image)
	}
	return images, nil
}
