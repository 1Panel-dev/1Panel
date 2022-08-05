package dto

import (
	"1Panel/app/model"
	"time"
)

type UserCreate struct {
	Name     string `json:"name" validate:"name,required"`
	Password string `json:"password" validate:"password,required"`
	Email    string `json:"email" validate:"required,email"`
}

type UserUpdate struct {
	Email string `json:"email" validate:"required,email"`
}

type UserBack struct {
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"createdAt"`
}

func (u UserCreate) UserCreateToMo() model.User {
	return model.User{
		Name:     u.Name,
		Password: u.Password,
		Email:    u.Email,
	}
}
