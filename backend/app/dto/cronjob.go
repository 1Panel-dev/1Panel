package dto

type CronjobCreate struct {
	Name     string `json:"name" validate:"required"`
	Type     string `json:"type" validate:"required"`
	SpecType string `json:"specType" validate:"required"`
	Week     int    `json:"week"`
	Day      int    `json:"day"`
	Hour     int    `json:"hour"`
	Minute   int    `json:"minute"`

	Script         string `json:"script"`
	WebSite        string `json:"webSite"`
	ExclusionRules string `json:"exclusionRules"`
	Database       string `json:"database"`
	URL            string `json:"url"`
	TargetDir      string `json:"targetDir"`
	RetainCopies   string `json:"retainCopies"`
}

type CronjobUpdate struct {
	Name     string `json:"name" validate:"required"`
	Type     string `json:"type" validate:"required"`
	SpecType string `json:"specType" validate:"required"`
	Week     int    `json:"week"`
	Day      int    `json:"day"`
	Hour     int    `json:"hour"`
	Minute   int    `json:"minute"`

	Script         string `json:"script"`
	WebSite        string `json:"webSite"`
	ExclusionRules string `json:"exclusionRules"`
	Database       string `json:"database"`
	URL            string `json:"url"`
	TargetDir      string `json:"targetDir"`
	RetainCopies   string `json:"retainCopies"`

	Status string `json:"status"`
}

type CronjobInfo struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	SpecType string `json:"specType"`
	Week     int    `json:"week"`
	Day      int    `json:"day"`
	Hour     int    `json:"hour"`
	Minute   int    `json:"minute"`

	Script         string `json:"script"`
	WebSite        string `json:"webSite"`
	ExclusionRules string `json:"exclusionRules"`
	Database       string `json:"database"`
	URL            string `json:"url"`
	TargetDir      string `json:"targetDir"`
	RetainCopies   string `json:"retainCopies"`

	Status string `json:"status"`
}
