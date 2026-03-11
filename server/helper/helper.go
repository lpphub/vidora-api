package helper

import (
	"github.com/gin-gonic/gin"
)

func MustBindJSON(c *gin.Context, req interface{}) bool {
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return false
	}
	return true
}

func MustBindQuery(c *gin.Context, req interface{}) bool {
	if err := c.ShouldBindQuery(req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return false
	}
	return true
}

func Respond(c *gin.Context, err error, data interface{}) {
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"data": data})
}
