package utils

import (
	"github.com/gin-gonic/gin"
)

func JSON(c *gin.Context, status int, v any) {
	c.Header("Content-Type", "application/json")
	c.JSON(status, v)
}

func Error(c *gin.Context, status int, message string) {
	JSON(c, status, gin.H{"message": message})
}
