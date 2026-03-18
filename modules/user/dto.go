// module/user/dto.go
package user

// CreateUserReq 创建用户请求
type CreateUserReq struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// UpdateProfileReq 更新资料请求
type UpdateProfileReq struct {
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
}

// ChangePasswordReq 修改密码请求
type ChangePasswordReq struct {
	OldPassword string `json:"oldPassword" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required,min=6"`
}

// UserResp 用户响应
type UserResp struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Avatar    string `json:"avatar"`
	Role      int8   `json:"role"`
	Status    int8   `json:"status"`
	CreatedAt string `json:"createdAt"`
}