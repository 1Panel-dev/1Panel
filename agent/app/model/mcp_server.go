package model

type McpServer struct {
	BaseModel
	Name          string `json:"name"`
	DockerCompose string `json:"dockerCompose"`
	Command       string `json:"command"`
	ContainerName string `json:"containerName"`
	Message       string `json:"message"`
	Port          int    `json:"port"`
	Status        string `json:"status"`
	Env           string `json:"env"`
	BaseURL       string `json:"baseUrl"`
	SsePath       string `json:"ssePath"`
	WebsiteID     int    `json:"websiteID"`
	Dir           string `json:"dir"`
	HostIP        string `json:"hostIP"`
}
