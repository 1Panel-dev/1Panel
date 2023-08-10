package dto

import "time"

type RemoteDBSearch struct {
	PageInfo
	Info    string `json:"info"`
	Type    string `json:"type"`
	OrderBy string `json:"orderBy"`
	Order   string `json:"order"`
}

type RemoteDBInfo struct {
	ID          uint      `json:"id"`
	CreatedAt   time.Time `json:"createdAt"`
	Name        string    `json:"name" validate:"max=256"`
	From        string    `json:"from"`
	Version     string    `json:"version"`
	Address     string    `json:"address"`
	Port        uint      `json:"port"`
	Username    string    `json:"username"`
	Password    string    `json:"password"`
	Description string    `json:"description"`
}

type RemoteDBOption struct {
	ID      uint   `json:"id"`
	Name    string `json:"name"`
	Address string `json:"address"`
}

type RemoteDBCreate struct {
	Name        string `json:"name" validate:"required,max=256"`
	Type        string `json:"type" validate:"required,oneof=mysql"`
	From        string `json:"from" validate:"required,oneof=local remote"`
	Version     string `json:"version" validate:"required"`
	Address     string `json:"address"`
	Port        uint   `json:"port"`
	Username    string `json:"username" validate:"required"`
	Password    string `json:"password" validate:"required"`
	Description string `json:"description"`
}

type RemoteDBUpdate struct {
	ID          uint   `json:"id"`
	Version     string `json:"version" validate:"required"`
	Address     string `json:"address"`
	Port        uint   `json:"port"`
	Username    string `json:"username" validate:"required"`
	Password    string `json:"password" validate:"required"`
	Description string `json:"description"`
}
