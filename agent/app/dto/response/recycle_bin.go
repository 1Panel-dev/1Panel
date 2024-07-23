package response

import "time"

type RecycleBinDTO struct {
	Name       string    `json:"name"`
	Size       int       `json:"size"`
	Type       string    `json:"type"`
	DeleteTime time.Time `json:"deleteTime"`
	RName      string    `json:"rName"`
	SourcePath string    `json:"sourcePath"`
	IsDir      bool      `json:"isDir"`
	From       string    `json:"from"`
}
