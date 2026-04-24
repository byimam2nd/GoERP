package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/goerp/goerp/apps/core/logger"
	"go.uber.org/zap"
	"net/http"
)

func JSONRecovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger.Log.Error("Server Panic Recovered",
					zap.Any("error", err),
					zap.String("path", c.Request.URL.Path),
				)
				
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error": "Internal Server Error",
					"message": "An unexpected error occurred. Please contact support.",
				})
			}
		}()
		c.Next()
	}
}
