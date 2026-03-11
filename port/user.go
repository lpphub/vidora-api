// contract/user.go
package port

import "context"

// UserBiz 用户服务接口（供其他模块调用）
type UserBiz interface {
	Create(ctx context.Context, email, password string) (*UserDTO, error)
	Get(ctx context.Context, userID uint) (*UserDTO, error)
	GetByEmail(ctx context.Context, email string) (*UserDTO, error)
	GetByIDs(ctx context.Context, ids []uint) ([]UserDTO, error)
	ValidateLogin(ctx context.Context, email, password string) (*UserDTO, error)
}

// UserDTO 用户数据传输对象
type UserDTO struct {
	ID       uint
	Email    string
	Name     string
	Password string
}