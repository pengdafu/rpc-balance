package service

import "github.com/gin-gonic/gin"

func Alive(c *gin.Context) {
  c.JSON(200, gin.H{
    "alive": true,
  })
}
