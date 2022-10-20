package dto

import "time"

type MysqlDBInfo struct {
	ID          uint      `json:"id"`
	CreatedAt   time.Time `json:"createdAt"`
	Name        string    `json:"name"`
	Format      string    `json:"format"`
	Username    string    `json:"username"`
	Password    string    `json:"password"`
	Permission  string    `json:"permission"`
	Description string    `json:"description"`
}

type MysqlDBCreate struct {
	Name          string `json:"name" validate:"required"`
	Format        string `json:"format" validate:"required,oneof=utf8mb4 utf-8 gbk big5"`
	Username      string `json:"username" validate:"required"`
	Password      string `json:"password" validate:"required"`
	Permission    string `json:"permission" validate:"required,oneof=local all ip"`
	PermissionIPs string `json:"permissionIPs"`
	Description   string `json:"description"`
}
