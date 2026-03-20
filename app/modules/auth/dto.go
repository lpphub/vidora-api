// modules/auth/dto.go
package auth

// AuthReq 认证请求
type AuthReq struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// AuthResp 认证响应
type AuthResp struct {
	User         *AuthUser `json:"user"`
	AccessToken  string    `json:"accessToken"`
	RefreshToken string    `json:"refreshToken"`
}

// AuthUser 认证用户
type AuthUser struct {
	ID     uint   `json:"id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Avatar string `json:"avatar"`
	Role   int8   `json:"role"`
}
