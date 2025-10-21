package model

type TensorRTLLM struct {
	BaseModel
	Name          string `json:"name"`
	DockerCompose string `json:"dockerCompose"`
	ContainerName string `json:"containerName"`
	Message       string `json:"message"`
	//Port          int    `json:"port"`
	Status string `json:"status"`
	Env    string `json:"env"`
	TaskID string `json:"taskID"`
}
