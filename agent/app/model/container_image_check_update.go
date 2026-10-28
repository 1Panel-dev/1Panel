package model

import "time"

// ContainerImageCheckUpdate is a cache tied to the exact image used by a container.
type ContainerImageCheckUpdate struct {
	ContainerID   string    `gorm:"primaryKey" json:"containerID"`
	ImageID       string    `json:"imageID"`
	ImageName     string    `json:"imageName"`
	LatestVersion string    `json:"latestVersion"`
	TargetImage   string    `json:"targetImage"`
	TargetImageID string    `json:"-"`
	HasUpdate     bool      `gorm:"not null;default:false" json:"hasUpdate"`
	CheckedAt     time.Time `json:"checkedAt"`
	CheckStatus   string    `json:"checkStatus"`
	CheckError    string    `json:"checkError"`
}

// TableName preserves the table created by the original migration.
func (ContainerImageCheckUpdate) TableName() string { return "container_image_updates" }
