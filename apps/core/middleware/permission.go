package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/goerp/goerp/apps/core/database"
	"net/http"
)

func PermissionMiddleware(action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		doctype := c.Param("doctype")
		tenantID, _ := c.Get("tenant_id")
		roles, _ := c.Get("roles")
		userRoles := roles.([]string)

		// 1. Feature Gating: Check if Tenant has access to this DocType's module
		var tenantDoc map[string]interface{}
		if err := database.DB.Table("tabTenant").Where("name = ?", tenantID).First(&tenantDoc).Error; err == nil {
			// Assume DocType Meta provides module name
			// Check if module is in tenant's 'active_modules' (comma-separated or table)
			// (Implementation simplified for now)
			if plan, ok := tenantDoc["plan"].(string); ok && plan == "Basic" {
				// Prevent access to Advanced modules
				// if module == "Manufacturing" { c.AbortWithStatus(403); return }
			}
		}

		// 2. RBAC: Existing role-based checks
		column := action
		if action == "create" { column = "can_create" }
		
		for _, role := range userRoles {
			if role == "Administrator" {
				c.Next()
				return
			}
		}

		var count int64
		if err := database.DB.Table("tabRolePermission").
			Where("parent_doctype = ? AND role IN ? AND "+column+" = ?", doctype, userRoles, true).
			Count(&count).Error; err != nil || count == 0 {
			c.JSON(http.StatusForbidden, gin.H{"error": "Permission denied for " + action})
			c.Abort()
			return
		}

		c.Next()
	}
}
