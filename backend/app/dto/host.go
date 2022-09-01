package dto

import (
	"time"
)

type HostOperate struct {
	GroupBelong string `json:"groupBelong" validate:"required"`
	Name        string `json:"name" validate:"required"`
	Addr        string `json:"addr" validate:"required,ip"`
	Port        uint   `json:"port" validate:"required,number,max=65535,min=1"`
	User        string `json:"user" validate:"required"`
	AuthMode    string `json:"authMode" validate:"oneof=password key"`
	PrivateKey  string `json:"privateKey"`
	Password    string `json:"password"`

	Description string `json:"description"`
}

type HostConnTest struct {
	Addr       string `json:"addr" validate:"required,ip"`
	Port       uint   `json:"port" validate:"required,number,max=65535,min=1"`
	User       string `json:"user" validate:"required"`
	AuthMode   string `json:"authMode" validate:"oneof=password key"`
	PrivateKey string `json:"privateKey"`
	Password   string `json:"password"`
}

type SearchForTree struct {
	Info string `json:"info"`
}

type HostInfo struct {
	ID          uint      `json:"id"`
	CreatedAt   time.Time `json:"createdAt"`
	GroupBelong string    `json:"groupBelong"`
	Name        string    `json:"name"`
	Addr        string    `json:"addr"`
	Port        uint      `json:"port"`
	User        string    `json:"user"`
	AuthMode    string    `json:"authMode"`

	Description string `json:"description"`
}

type HostTree struct {
	ID       uint        `json:"id"`
	Label    string      `json:"label"`
	Children []TreeChild `json:"children"`
}

type TreeChild struct {
	ID    uint   `json:"id"`
	Label string `json:"label"`
}
