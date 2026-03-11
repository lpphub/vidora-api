// pkg/mod/module.go
package mod

import "github.com/gin-gonic/gin"

// Module 模块接口
type Module interface {
	RegisterRoutes(r *gin.RouterGroup)
}