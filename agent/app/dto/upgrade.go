package dto

type Upgrade struct {
	Version     string `json:"version"`
	UpgradePath string `json:"upgradePath"`
}

type Rollback struct {
	Version string `json:"version"`
}
