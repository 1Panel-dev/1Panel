package dto

import "time"

type ScriptInfo struct {
	ID            uint      `json:"id"`
	Name          string    `json:"name"`
	IsInteractive bool      `json:"isInteractive"`
	Lable         string    `json:"lable"`
	Script        string    `json:"script"`
	GroupList     []uint    `json:"groupList"`
	GroupBelong   []string  `json:"groupBelong"`
	IsSystem      bool      `json:"isSystem"`
	Description   string    `json:"description"`
	CreatedAt     time.Time `json:"createdAt"`
}

type ScriptOperate struct {
	ID            uint   `json:"id"`
	IsInteractive bool   `json:"isInteractive"`
	Name          string `json:"name"`
	Script        string `json:"script"`
	Groups        string `json:"groups"`
	Description   string `json:"description"`
}
