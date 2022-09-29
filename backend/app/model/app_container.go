package model

type AppContainer struct {
	BaseModel
	ServiceName   string `json:"serviceName"`
	ContainerName string `json:"containerName"`
	AppInstallId  uint   `json:"appInstallId"`
	Image         string `json:"image"`
}
