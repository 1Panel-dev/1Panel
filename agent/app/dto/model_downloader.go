package dto

type ModelSearch struct {
	Query    string `json:"query"`
	Sort     string `json:"sort" validate:"omitempty,oneof=trending downloads updated created likes most_params least_params"`
	Page     int    `json:"page" validate:"omitempty,min=1"`
	PageSize int    `json:"pageSize" validate:"omitempty,min=1,max=100"`
	Limit    int    `json:"limit" validate:"omitempty,min=1,max=100"`
}

type ModelInfoReq struct {
	RepoID   string `json:"repoID" validate:"required"`
	Revision string `json:"revision"`
}

type ModelDownload struct {
	DownloadType string `json:"downloadType" validate:"omitempty,oneof=all tokenizer"`
	RepoID       string `json:"repoID" validate:"required"`
	Revision     string `json:"revision"`
}

type ModelDownloaderSyncSetting struct {
	Value string `json:"value" validate:"required,oneof=Enable Disable"`
}

type ModelDownloaderSetting struct {
	FeaturedModelsSync string `json:"featuredModelsSync"`
	ModelDir           string `json:"modelDir" validate:"required"`
	HFEndpoint         string `json:"hfEndpoint"`
	HFToken            string `json:"hfToken"`
	ModelScopeEndpoint string `json:"modelScopeEndpoint"`
	ModelScopeToken    string `json:"modelScopeToken"`
}

type ModelItem struct {
	RepoID        string `json:"repoID"`
	Name          string `json:"name"`
	Downloads     int64  `json:"downloads"`
	Likes         int64  `json:"likes"`
	Size          int64  `json:"size"`
	SizeFormatted string `json:"sizeFormatted"`
}

type ModelFile struct {
	Name          string `json:"name"`
	Size          int64  `json:"size"`
	SizeFormatted string `json:"sizeFormatted"`
}

type ModelInfo struct {
	ModelItem
	ModelCard string      `json:"modelCard"`
	Files     []ModelFile `json:"files"`
}

type ModelDownloadTaskSearch struct {
	IncludeLocal bool   `json:"includeLocal"`
	DownloadType string `json:"downloadType" validate:"omitempty,oneof=all tokenizer"`
	PageInfo
	Info   string `json:"info"`
	Status string `json:"status"`
}

type ModelDownloadTaskOperate struct {
	WithFile bool `json:"withFile"`
	ID       uint `json:"id" validate:"required"`
}

type ModelDownloadTaskRetry struct {
	ID uint `json:"id" validate:"required"`
}

type ModelDownloadTaskItem struct {
	LocalModel      *LocalModelItem `json:"localModel,omitempty"`
	DownloadType    string          `json:"downloadType"`
	ID              uint            `json:"id"`
	Source          string          `json:"source"`
	RepoID          string          `json:"repoID"`
	Revision        string          `json:"revision"`
	ModelName       string          `json:"modelName"`
	TargetDir       string          `json:"targetDir"`
	Status          string          `json:"status"`
	Progress        float64         `json:"progress"`
	TotalSize       int64           `json:"totalSize"`
	DownloadedSize  int64           `json:"downloadedSize"`
	ErrorMessage    string          `json:"errorMessage"`
	CreatedAt       string          `json:"createdAt"`
	StartedAt       string          `json:"startedAt"`
	CompletedAt     string          `json:"completedAt"`
	CanCancel       bool            `json:"canCancel"`
	CanRetry        bool            `json:"canRetry"`
	CanRemoveRecord bool            `json:"canRemoveRecord"`
}

type LocalModelSearch struct {
	DownloadType string `json:"downloadType" validate:"omitempty,oneof=all tokenizer"`
	PageInfo
	Info string `json:"info"`
}

type LocalModelItem struct {
	DownloadType  string `json:"downloadType"`
	RepoID        string `json:"repoID"`
	Source        string `json:"source"`
	Revision      string `json:"revision"`
	Name          string `json:"name"`
	Path          string `json:"path"`
	Size          int64  `json:"size"`
	SizeFormatted string `json:"sizeFormatted"`
	CreatedAt     string `json:"createdAt"`
}

type LocalModelDelete struct {
	WithFile     bool   `json:"withFile"`
	DownloadType string `json:"downloadType" validate:"omitempty,oneof=all tokenizer"`
	Name         string `json:"name" validate:"required"`
}

type FeaturedModel struct {
	Key            string            `json:"key" yaml:"key"`
	Sort           int               `json:"sort" yaml:"sort"`
	Family         string            `json:"family" yaml:"category"`
	RepoID         string            `json:"repoID" yaml:"name"`
	Sources        []string          `json:"sources" yaml:"sources"`
	SourceModelIDs map[string]string `json:"sourceModelIds" yaml:"sourceModelIds"`
	Architecture   string            `json:"architecture" yaml:"modelType"`
	BackboneParams string            `json:"backboneParams" yaml:"moeBackboneParameters"`
	ActiveParams   string            `json:"activeParams" yaml:"activeParameters"`
	Engram         string            `json:"engram" yaml:"engramParameters"`
	WeightGB       float64           `json:"weightGB" yaml:"approximateWeightSizeGB"`
	Description    map[string]string `json:"description" yaml:"description"`
}

type FeaturedModelSearch struct {
	PageInfo
	Family string `json:"family"`
	Info   string `json:"info"`
}

type FeaturedModelPage struct {
	PageResult
	Families []string `json:"families"`
}
