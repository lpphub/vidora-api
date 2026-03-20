package middleware

import (
	"net/http"
	"strings"
	"vidora-api/app/infra/jwt"
	"vidora-api/app/shared/errs"

	"github.com/gin-gonic/gin"
	"github.com/lpphub/goweb/base"
)

const (
	AuthorizationHeader = "Authorization"
	AuthorizationPrefix = "Bearer"
	UserIDKey           = "user_id"
)

func JwtAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader(AuthorizationHeader)
		if authHeader == "" {
			base.Fail(c, errs.ErrNoToken)
			return
		}
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != AuthorizationPrefix {
			c.AbortWithStatusJSON(http.StatusUnauthorized, errs.ErrNoToken)
			return
		}

		claims, err := jwt.ParseAccessToken(parts[1])
		if err != nil {
			base.Fail(c, errs.ErrInvalidToken)
			return
		}

		c.Set(UserIDKey, claims.UserID)
		c.Next()
	}
}

func GetUserID(c *gin.Context) (uint, bool) {
	value, exists := c.Get(UserIDKey)
	if !exists {
		return 0, false
	}
	userID, ok := value.(uint)
	return userID, ok
}
