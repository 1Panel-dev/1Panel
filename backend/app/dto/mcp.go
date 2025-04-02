package dto

type DockerComposeService struct {
	Image         string   `yaml:"image"`
	ContainerName string   `yaml:"container_name"`
	Restart       string   `yaml:"restart"`
	Ports         []string `yaml:"ports"`
	Environment   []string `yaml:"environment"`
	Command       []string `yaml:"command"`
}
