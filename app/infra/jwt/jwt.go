package jwt

import (
	"fmt"
	"sync"
	"vidora-api/app/infra"

	"github.com/lpphub/goweb/pkg/jwt"
)

type (
	Claims    = jwt.Claims
	TokenPair = jwt.TokenPair
)

var (
	jm   *jwt.Manager
	once sync.Once
)

func useJwt() *jwt.Manager {
	once.Do(func() {
		jm, _ = jwt.NewManager(jwt.Config{
			Secret:           infra.Cfg.JWT.Secret,
			AccessExpireSec:  infra.Cfg.JWT.ExpireTime,
			RefreshExpireSec: infra.Cfg.JWT.RefreshExpireTime,
		})
	})
	return jm
}

// ParseAccessToken 解析 JWT token
func ParseAccessToken(tokenString string) (*Claims, error) {
	claims, err := useJwt().ParseToken(tokenString)
	if err != nil {
		return nil, err
	}

	if claims.Type != jwt.AccessTokenType {
		return nil, fmt.Errorf("invalid token type")
	}

	return claims, nil
}

// GenerateTokenPair 生成 access_token 和 refresh_token
func GenerateTokenPair(userID uint) (*TokenPair, error) {
	token, err := useJwt().GenerateTokenPair(userID)
	if err != nil {
		return nil, err
	}
	return token, nil
}

// RefreshToken 使用 refresh_token 刷新 token 对
func RefreshToken(refreshToken string) (*TokenPair, error) {
	token, err := useJwt().RefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}
	return token, nil
}
