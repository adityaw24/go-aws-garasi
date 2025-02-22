package utils

import (
	"log"

	"github.com/gin-gonic/gin"
)

func ErrorResp(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, gin.H{"status": "error", "message": message})
}

func SuccessResp(c *gin.Context, statusCode int, message string, data interface{}) {
	c.JSON(statusCode, gin.H{"status": "success", "message": message, "data": data})
}

func ErrorLog(layer string, funcName string, err error) {
	log.Printf("[Error] %s - %s: %v", layer, funcName, err)
}
