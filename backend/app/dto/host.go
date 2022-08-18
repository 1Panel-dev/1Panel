package dto

import "time"

type HostCreate struct {
	Name       string `json:"name" validate:"required,name"`
	Addr       string `json:"addr" validate:"required,ip"`
	Port       uint   `json:"port" validate:"required,number,max=65535,min=1"`
	User       string `json:"user" validate:"required"`
	AuthMode   string `json:"authMode" validate:"oneof=password key"`
	PrivateKey string `json:"privateKey"`
	Password   string `json:"password"`

	Description string `json:"description"`
}

type HostInfo struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	Name      string    `json:"name"`
	Addr      string    `json:"addr"`
	Port      uint      `json:"port"`
	User      string    `json:"user"`
	AuthMode  string    `json:"authMode"`

	Description string `json:"description"`
}

type HostUpdate struct {
	Name       string `json:"name" validate:"required,name"`
	Addr       string `json:"addr" validate:"required,ip"`
	Port       uint   `json:"port" validate:"required,number,max=65535,min=1"`
	User       string `json:"user" validate:"required"`
	AuthMode   string `json:"authMode" validate:"oneof=password key"`
	PrivateKey string `json:"privateKey"`
	Password   string `json:"password"`

	Description string `json:"description"`
}
