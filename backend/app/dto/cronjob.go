package dto

type CronjobCreate struct {
	Name     string `json:"name" validate:"required"`
	Type     string `json:"type" validate:"required"`
	SpecType string `json:"specType" validate:"required"`
	Week     int    `json:"week" validate:"number,max=7,min=1"`
	Day      int    `json:"day" validate:"number,max=31,min=1"`
	Hour     int    `json:"hour" validate:"number,max=23,min=0"`
	Minute   int    `json:"minute" validate:"number,max=59,min=0"`

	Script         string `json:"script"`
	Website        string `json:"website"`
	ExclusionRules string `json:"exclusionRules"`
	Database       string `json:"database"`
	URL            string `json:"url"`
	TargetDirID    int    `json:"targetDirID"`
	RetainCopies   int    `json:"retainCopies" validate:"number,min=1"`
}

type CronjobUpdate struct {
	Name     string `json:"name" validate:"required"`
	Type     string `json:"type" validate:"required"`
	SpecType string `json:"specType" validate:"required"`
	Week     int    `json:"week" validate:"number,max=7,min=1"`
	Day      int    `json:"day" validate:"number,max=31,min=1"`
	Hour     int    `json:"hour" validate:"number,max=23,min=0"`
	Minute   int    `json:"minute" validate:"number,max=60,min=1"`

	Script         string `json:"script"`
	Website        string `json:"website"`
	ExclusionRules string `json:"exclusionRules"`
	Database       string `json:"database"`
	URL            string `json:"url"`
	TargetDirID    int    `json:"targetDirID" validate:"number,min=1"`
	RetainCopies   int    `json:"retainCopies" validate:"number,min=1"`

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
	Website        string `json:"website"`
	ExclusionRules string `json:"exclusionRules"`
	Database       string `json:"database"`
	URL            string `json:"url"`
	TargetDir      string `json:"targetDir"`
	TargetDirID    int    `json:"targetDirID"`
	RetainCopies   int    `json:"retainCopies"`

	Status string `json:"status"`
}
