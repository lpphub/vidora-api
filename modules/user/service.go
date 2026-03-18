// module/user/service.go
package user

import (
	"context"
	"errors"
	"time"

	"vidora-api/contract"
	"vidora-api/shared/errs"
	"vidora-api/shared/strutils"

	"gorm.io/gorm"
)

// 确保实现 contract.UserBiz 接口
var _ contract.UserBiz = (*Service)(nil)

// Service 用户服务
type Service struct {
	repo *Repository
}

// NewService 创建用户服务
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// Create 创建用户
func (s *Service) Create(ctx context.Context, email, password string) (*contract.UserDTO, error) {
	exists, _ := s.repo.ExistsByEmail(ctx, email)
	if exists {
		return nil, errs.ErrUserExists
	}

	user := &User{
		Name:      strutils.ExtractNameFromEmail(email),
		Email:     email,
		Status:    StatusActive,
		CreatedAt: time.Now(),
	}

	if err := user.SetPassword(password); err != nil {
		return nil, errs.ErrInvalidPassword
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	return toDTO(user), nil
}

// Get 获取用户
func (s *Service) Get(ctx context.Context, userID uint) (*contract.UserDTO, error) {
	user, err := s.repo.First(ctx, userID)
	if err != nil {
		return nil, err
	}
	return toDTO(user), nil
}

// GetByEmail 根据邮箱获取用户
func (s *Service) GetByEmail(ctx context.Context, email string) (*contract.UserDTO, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	return toDTO(user), nil
}

// GetByIDs 批量获取用户
func (s *Service) GetByIDs(ctx context.Context, ids []uint) ([]contract.UserDTO, error) {
	users, err := s.repo.FindByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}

	dtos := make([]contract.UserDTO, len(users))
	for i, u := range users {
		dtos[i] = *toDTO(&u)
	}
	return dtos, nil
}

// ValidateLogin 验证登录
func (s *Service) ValidateLogin(ctx context.Context, email, password string) (*contract.UserDTO, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.ErrLoginFailed
		}
		return nil, err
	}

	if !user.IsActive() {
		return nil, errs.ErrUserDisabled
	}

	if err := user.ValidatePassword(password); err != nil {
		return nil, errs.ErrLoginFailed
	}

	return toDTO(user), nil
}

// UpdateProfile 更新用户资料
func (s *Service) UpdateProfile(ctx context.Context, userID uint, name, avatar string) error {
	user, err := s.repo.First(ctx, userID)
	if err != nil {
		return err
	}

	user.UpdateProfile(name, avatar)

	return s.repo.Update(ctx, userID, map[string]interface{}{
		"name":       user.Name,
		"avatar":     user.Avatar,
		"updated_at": time.Now(),
	})
}

// ChangePassword 修改密码
func (s *Service) ChangePassword(ctx context.Context, userID uint, oldPassword, newPassword string) error {
	user, err := s.repo.First(ctx, userID)
	if err != nil {
		return err
	}

	if err = user.ValidatePassword(oldPassword); err != nil {
		return errs.ErrLoginFailed
	}

	if err = user.SetPassword(newPassword); err != nil {
		return errs.ErrInvalidPassword
	}

	return s.repo.Update(ctx, userID, map[string]interface{}{
		"password":   user.Password,
		"updated_at": time.Now(),
	})
}

// toDTO 转换为 DTO
func toDTO(user *User) *contract.UserDTO {
	return &contract.UserDTO{
		ID:       user.ID,
		Email:    user.Email,
		Name:     user.Name,
		Password: user.Password,
	}
}
