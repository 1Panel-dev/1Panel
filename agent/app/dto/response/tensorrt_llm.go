package response

import "github.com/1Panel-dev/1Panel/agent/app/model"

type TensorRTLLMsRes struct {
	Items []TensorRTLLMDTO `json:"items"`
	Total int64            `json:"total"`
}

type TensorRTLLMDTO struct {
	model.TensorRTLLM
	Version  string `json:"version"`
	Model    string `json:"model"`
	Dir      string `json:"dir"`
	ModelDir string `json:"modelDir"`
}
