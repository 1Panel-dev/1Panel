package model

type FeaturedModel struct {
	BaseModel
	Key            string            `json:"key"`
	Sort           int               `json:"sort"`
	Family         string            `json:"family"`
	RepoID         string            `json:"repoID"`
	WeightProvider string            `json:"weightProvider"`
	Sources        []string          `json:"sources" gorm:"serializer:json"`
	SourceModelIDs map[string]string `json:"sourceModelIds" gorm:"serializer:json"`
	Architecture   string            `json:"architecture"`
	BackboneParams string            `json:"backboneParams"`
	ActiveParams   string            `json:"activeParams"`
	Engram         string            `json:"engram"`
	WeightGB       float64           `json:"weightGB"`
	Description    string            `json:"description"`
}
