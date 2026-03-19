// modules/auth/service.go
package auth

import (
	"context"
	"vidora-api/infra/jwt"
	"vidora-api/shared/contracts"

	"github.com/jinzhu/copier"
)

// Service 认证服务
type Service struct {
	userSvc contracts.UserBiz
}

// NewService 创建认证服务
func NewService(userSvc contracts.UserBiz) *Service {
	return &Service{userSvc: userSvc}
}

// Register 用户注册
func (s *Service) Register(ctx context.Context, email, password string) (*AuthResp, error) {
	newUser, err := s.userSvc.Create(ctx, email, password)
	if err != nil {
		return nil, err
	}

	tokenPair, err := jwt.GenerateTokenPair(newUser.ID)
	if err != nil {
		return nil, err
	}

	var authUser AuthUser
	_ = copier.Copy(&authUser, newUser)

	return &AuthResp{
		User:         &authUser,
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
	}, nil
}

// Login 用户登录
func (s *Service) Login(ctx context.Context, email, password string) (*AuthResp, error) {
	loginUser, err := s.userSvc.ValidateLogin(ctx, email, password)
	if err != nil {
		return nil, err
	}

	tokenPair, err := jwt.GenerateTokenPair(loginUser.ID)
	if err != nil {
		return nil, err
	}

	var authUser AuthUser
	_ = copier.Copy(&authUser, loginUser)

	return &AuthResp{
		User:         &authUser,
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
	}, nil
}

// RefreshToken 刷新令牌
func (s *Service) RefreshToken(_ context.Context, refreshToken string) (*AuthResp, error) {
	tokenPair, err := jwt.RefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	return &AuthResp{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
	}, nil
}
