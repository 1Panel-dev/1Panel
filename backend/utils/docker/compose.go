package docker

import (
	"context"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/compose-spec/compose-go/v2/loader"
	"github.com/compose-spec/compose-go/v2/types"
	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/flags"
	"github.com/docker/compose/v2/pkg/api"
	"github.com/docker/compose/v2/pkg/compose"
	"github.com/docker/docker/client"
	"github.com/joho/godotenv"
)

type ComposeService struct {
	api.Service
	project *types.Project
}

func UpComposeProject(project *types.Project) error {
	for i, s := range project.Services {
		s.CustomLabels = map[string]string{
			api.ProjectLabel:     project.Name,
			api.ServiceLabel:     s.Name,
			api.VersionLabel:     api.ComposeVersion,
			api.WorkingDirLabel:  project.WorkingDir,
			api.ConfigFilesLabel: strings.Join(project.ComposeFiles, ","),
			api.OneoffLabel:      "False",
		}
		project.Services[i] = s
	}

	apiClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}
	var ops []command.CLIOption
	ops = append(ops, command.WithAPIClient(apiClient), command.WithDefaultContextStoreConfig())
	cli, err := command.NewDockerCli(ops...)
	if err != nil {
		return err
	}
	cliOp := flags.NewClientOptions()
	if err = cli.Initialize(cliOp); err != nil {
		return err
	}
	service := compose.NewComposeService(cli)
	composeService := ComposeService{Service: service}

	return composeService.Up(context.Background(), project, api.UpOptions{
		Create: api.CreateOptions{
			Timeout: getComposeTimeout(),
		},
		Start: api.StartOptions{
			WaitTimeout: *getComposeTimeout(),
			Wait:        true,
		},
	})
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

func getComposeTimeout() *time.Duration {
	timeout := time.Minute * time.Duration(10)
	return &timeout
}

type ComposeProject struct {
	Version  string
	Services map[string]Service `yaml:"services"`
}

type Service struct {
	Image string `yaml:"image"`
}

func GetDockerComposeImages(projectName string, env, yml []byte) ([]string, error) {
	var (
		configFiles []types.ConfigFile
		images      []string
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
		images = append(images, service.Image)
	}
	return images, nil
}
