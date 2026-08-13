package docker

import (
	"bytes"
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/compose-spec/compose-go/v2/dotenv"
	"github.com/compose-spec/compose-go/v2/loader"
	"github.com/compose-spec/compose-go/v2/template"
	"github.com/compose-spec/compose-go/v2/types"
	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"

	"github.com/1Panel-dev/1Panel/agent/utils/re"
)

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
	projectName = re.GetRegex(re.ComposeDisallowedCharsPattern).ReplaceAllString(projectName, "")
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
	Image       string      `yaml:"image"`
	Environment Environment `yaml:"environment"`
	Volumes     []string    `yaml:"volumes"`
	ExtraHosts  []string    `yaml:"extra_hosts"`
	Restart     string      `yaml:"restart"`
}

type Environment struct {
	Variables map[string]string
}

func (e *Environment) UnmarshalYAML(value *yaml.Node) error {
	e.Variables = make(map[string]string)
	switch value.Kind {
	case yaml.MappingNode:
		for i := 0; i < len(value.Content); i += 2 {
			key := value.Content[i].Value
			val := value.Content[i+1].Value
			e.Variables[key] = val
		}
	case yaml.SequenceNode:
		for _, item := range value.Content {
			var kv string
			if err := item.Decode(&kv); err != nil {
				return err
			}
			parts := strings.SplitN(kv, "=", 2)
			if len(parts) == 2 {
				e.Variables[parts[0]] = parts[1]
			} else {
				e.Variables[parts[0]] = ""
			}
		}
	default:
		return fmt.Errorf("unsupported environment format")
	}
	return nil
}

func GetImagesFromDockerCompose(env, yml []byte) ([]string, error) {
	envVars, err := dotenv.Parse(bytes.NewReader(env))
	if err != nil {
		return nil, fmt.Errorf("load env failed: %v", err)
	}

	var compose ComposeProject
	if err := yaml.Unmarshal(yml, &compose); err != nil {
		return nil, fmt.Errorf("parse docker-compose file failed: %v", err)
	}

	var images []string
	for _, service := range compose.Services {
		if service.Image != "" {
			resolvedImage, err := template.Substitute(service.Image, func(key string) (string, bool) {
				value, ok := envVars[key]
				return value, ok
			})
			if err != nil {
				return nil, fmt.Errorf("resolve image failed: %v", err)
			}
			images = append(images, resolvedImage)
		}
	}

	return images, nil
}
