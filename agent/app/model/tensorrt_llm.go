package model

type TensorRTLLM struct {
	BaseModel
	Name          string `json:"name"`
	DockerCompose string `json:"dockerCompose"`
	ContainerName string `json:"containerName"`
	Message       string `json:"message"`
	Status        string `json:"status"`
	Env           string `json:"env"`
	TaskID        string `json:"taskID"`
	ModelType     string `json:"modelType"`
	ModelSpeedup  bool   `json:"modelSpeedup"`
}
