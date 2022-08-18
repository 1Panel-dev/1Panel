package dto

import (
	"time"
)

type UserCreate struct {
	Name     string `json:"name" validate:"name,required"`
	Password string `json:"password" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
}

type UserPage struct {
	PageInfo
	Name string `json:"name"  validate:"required"`
}

type CaptchaResponse struct {
	CaptchaID string `json:"captchaID"`
	ImagePath string `json:"imagePath"`
}

type UserUpdate struct {
	ID    uint   `json:"id" validate:"required"`
	Name  string `json:"name" validate:"required"`
	Email string `json:"email" validate:"required,email"`
}

type UserInfo struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"createdAt"`
}

type UserLoginInfo struct {
	Name  string `json:"name"`
	Token string `json:"token"`
}
