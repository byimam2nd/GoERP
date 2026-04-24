package middleware

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func TenantMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := c.GetHeader("X-GoERP-Tenant")
		if tenantID == "" {
			// In production, you might want to resolve this from domain name
			// For now, we require the header
			c.JSON(http.StatusBadRequest, gin.H{"error": "X-GoERP-Tenant header is required"})
			c.Abort()
			return
		}

		c.Set("tenant_id", tenantID)
		c.Next()
	}
}
