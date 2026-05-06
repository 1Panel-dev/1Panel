package dto

type CreateUserRequest struct {
	Username string `json:"username" validate:"required,min=3,max=255"`
	Email    string `json:"email" validate:"email"`
	Password string `json:"password" validate:"required,min=6"`
	Role     string `json:"role" validate:"required,oneof=user reseller admin"`
	RealName string `json:"realName"`
	Phone    string `json:"phone"`
	Remark   string `json:"remark"`
}

type UpdateUserRequest struct {
	ID       uint   `json:"id" validate:"required"`
	Email    string `json:"email" validate:"email"`
	Role     string `json:"role" validate:"required,oneof=user reseller admin"`
	Status   string `json:"status" validate:"required,oneof=active inactive"`
	RealName string `json:"realName"`
	Phone    string `json:"phone"`
	Remark   string `json:"remark"`
}

type ChangePasswordRequest struct {
	UserID      uint   `json:"userId" validate:"required"`
	OldPassword string `json:"oldPassword" validate:"required"`
	NewPassword string `json:"newPassword" validate:"required,min=6"`
}

type ResetPasswordRequest struct {
	UserID      uint   `json:"userId" validate:"required"`
	NewPassword string `json:"newPassword" validate:"required,min=6"`
}

type UserResponse struct {
	ID        uint   `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	Status    string `json:"status"`
	RealName  string `json:"realName"`
	Phone     string `json:"phone"`
	LastLogin int64  `json:"lastLogin"`
	Remark    string `json:"remark"`
	CreatedAt int64  `json:"createdAt"`
	UpdatedAt int64  `json:"updatedAt"`
}

type UserListResponse struct {
	Total int64           `json:"total"`
	Items []UserResponse  `json:"items"`
}

type UserDetailResponse struct {
	User        UserResponse `json:"user"`
	Permissions []string     `json:"permissions"`
}

type UserPageRequest struct {
	PageNum   int    `json:"pageNum" validate:"required,min=1"`
	PageSize  int    `json:"pageSize" validate:"required,min=1,max=100"`
	Username  string `json:"username"`
	Role      string `json:"role"`
	Status    string `json:"status"`
}

type AssignPermissionRequest struct {
	UserID      uint     `json:"userId" validate:"required"`
	Permissions []string `json:"permissions" validate:"required"`
}

type UserLoginResponse struct {
	Name       string `json:"name"`
	Token      string `json:"token"`
	MfaStatus  string `json:"mfaStatus"`
	MfaSession string `json:"mfaSession"`
	Role       string `json:"role"`
}

type UserResourceAccess struct {
	HostID     uint   `json:"hostId"`
	HostName   string `json:"hostName"`
	Permission string `json:"permission"`
}

type UserResourceAccessRequest struct {
	UserID   uint                 `json:"userId" validate:"required"`
	Accesses []UserResourceAccess `json:"accesses" validate:"required"`
}
