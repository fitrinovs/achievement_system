package helper

import "github.com/gin-gonic/gin"

func SendUnauthorized(c *gin.Context, msg string) {
	c.JSON(401, gin.H{"error": msg})
}
