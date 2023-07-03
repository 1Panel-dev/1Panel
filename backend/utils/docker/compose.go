package docker

import (
	"context"
	"github.com/compose-spec/compose-go/loader"
	"github.com/compose-spec/compose-go/types"
	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/flags"
	"github.com/docker/compose/v2/pkg/api"
	"github.com/docker/compose/v2/pkg/compose"
	"github.com/docker/docker/client"
	"github.com/joho/godotenv"
	"path"
	"regexp"
	"strings"
	"time"
)

type ComposeService struct {
	api.Service
	project *types.Project
}

func NewComposeService(ops ...command.DockerCliOption) (*ComposeService, error) {
	apiClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	ops = append(ops, command.WithAPIClient(apiClient), command.WithDefaultContextStoreConfig())
	cli, err := command.NewDockerCli(ops...)
	if err != nil {
		return nil, err
	}
	cliOp := flags.NewClientOptions()
	if err := cli.Initialize(cliOp); err != nil {
		return nil, err
	}
	service := compose.NewComposeService(cli)
	return &ComposeService{service, nil}, nil
}

func (s *ComposeService) SetProject(project *types.Project) {
	s.project = project
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
}

func (s *ComposeService) ComposeUp() error {
	return s.Up(context.Background(), s.project, api.UpOptions{
		Create: api.CreateOptions{
			Timeout: getComposeTimeout(),
		},
		Start: api.StartOptions{
			WaitTimeout: *getComposeTimeout(),
		},
	})
}

func (s *ComposeService) ComposeDown() error {
	return s.Down(context.Background(), s.project.Name, api.DownOptions{})
}

func (s *ComposeService) ComposeStart() error {
	return s.Start(context.Background(), s.project.Name, api.StartOptions{})
}

func (s *ComposeService) ComposeRestart() error {
	return s.Restart(context.Background(), s.project.Name, api.RestartOptions{})
}

func (s *ComposeService) ComposeStop() error {
	return s.Stop(context.Background(), s.project.Name, api.StopOptions{})
}

func (s *ComposeService) ComposeCreate() error {
	return s.Create(context.Background(), s.project, api.CreateOptions{})
}

func (s *ComposeService) ComposeBuild() error {
	return s.Build(context.Background(), s.project, api.BuildOptions{})
}

func (s *ComposeService) ComposePull() error {
	return s.Pull(context.Background(), s.project, api.PullOptions{})
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
	project, err := loader.Load(details, func(options *loader.Options) {
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
