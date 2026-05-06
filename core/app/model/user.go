package model

import "time"

type User struct {
	BaseModel
	Username  string    `gorm:"column:username;not null;uniqueIndex;size:255" json:"username"`
	Email     string    `gorm:"column:email;size:255" json:"email"`
	Password  string    `gorm:"column:password;not null" json:"password"`
	Role      string    `gorm:"column:role;not null;default:'user'" json:"role"`
	Status    string    `gorm:"column:status;not null;default:'active'" json:"status"`
	ParentID  uint      `gorm:"column:parent_id;index" json:"parentId"`
	RealName  string    `gorm:"column:real_name;size:255" json:"realName"`
	Phone     string    `gorm:"column:phone;size:20" json:"phone"`
	LastLogin *time.Time `gorm:"column:last_login" json:"lastLogin"`
	Remark    string    `gorm:"column:remark;type:text" json:"remark"`
}

func (User) TableName() string {
	return "users"
}

type UserPermission struct {
	BaseModel
	UserID     uint   `gorm:"column:user_id;not null;index" json:"userId"`
	Permission string `gorm:"column:permission;not null;size:255" json:"permission"`
}

func (UserPermission) TableName() string {
	return "user_permissions"
}

type UserLoginHistory struct {
	BaseModel
	UserID   uint      `gorm:"column:user_id;not null;index" json:"userId"`
	IP       string    `gorm:"column:ip;size:50" json:"ip"`
	Address  string    `gorm:"column:address;size:255" json:"address"`
	Agent    string    `gorm:"column:agent;type:text" json:"agent"`
	Status   string    `gorm:"column:status;not null" json:"status"`
	Message  string    `gorm:"column:message;type:text" json:"message"`
	LoginAt  time.Time `gorm:"column:login_at" json:"loginAt"`
}

func (UserLoginHistory) TableName() string {
	return "user_login_histories"
}
