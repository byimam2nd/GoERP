package http

import (
	"github.com/gin-gonic/gin"
	"github.com/goerp/goerp/apps/core/middleware"
	"github.com/goerp/goerp/apps/core/multitenant"
	"net/http"
)

func RegisterRoutes(r *gin.Engine) {
	// Public Group
	v1Public := r.Group("/api/v1")
	{
		v1Public.POST("/auth/login", HandleLogin)
		v1Public.POST("/provision", HandleProvision)
	}

	// Protected Group
	v1Protected := r.Group("/api/v1")
	v1Protected.Use(middleware.RequestIDMiddleware())
	v1Protected.Use(middleware.AuthMiddleware())
	v1Protected.Use(middleware.TenantMiddleware())
	{
		// Resource Routes (CRUD)
		v1Protected.POST("/resource/:doctype", middleware.PermissionMiddleware("create"), HandleCreate)
		v1Protected.GET("/resource/:doctype", middleware.PermissionMiddleware("read"), HandleList)
		v1Protected.GET("/resource/:doctype/:name", middleware.PermissionMiddleware("read"), HandleGet)
		v1Protected.PUT("/resource/:doctype/:name", middleware.PermissionMiddleware("write"), HandleUpdate)
		v1Protected.DELETE("/resource/:doctype/:name", middleware.PermissionMiddleware("delete"), HandleDelete)
		v1Protected.POST("/resource/:doctype/:name/amend", HandleAmend)
		v1Protected.GET("/resource/:doctype/tree", HandleTree)
		
		// Workflow Routes
		v1Protected.GET("/resource/:doctype/:name/workflow", HandleGetWorkflow)
		v1Protected.POST("/resource/:doctype/:name/workflow", HandleWorkflowAction)

		// Traceability Routes
		v1Protected.GET("/resource/:doctype/:name/links", HandleGetLinks)
		
		// Print Routes
		v1Protected.GET("/resource/:doctype/:name/print", HandlePrintPreview)
		
		// Import Routes
		v1Protected.GET("/import/template/:doctype", HandleDownloadTemplate)
		v1Protected.POST("/import/upload/:doctype", HandleUploadImport)

		// Mapping Routes
		v1Protected.GET("/map/:source_dt/:source_name/:target_dt", HandleMapDocument)
		
		// Metadata Routes
		v1Protected.GET("/meta/:doctype", HandleGetMeta)

		// Analytics Routes
		v1Protected.POST("/report", HandleReport)
		v1Protected.GET("/stats/:doctype", HandleStats)
		v1Protected.GET("/search", HandleGlobalSearch)
	}
}

type ProvisionRequest struct {
	TenantName    string `json:"tenant_name" binding:"required"`
	Domain        string `json:"domain"`
	AdminEmail    string `json:"admin_email" binding:"required"`
	AdminPassword string `json:"admin_password" binding:"required"`
}

func HandleProvision(c *gin.Context) {
	var req ProvisionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := multitenant.ProvisionTenant(req.TenantName, req.Domain, req.AdminEmail, req.AdminPassword); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "tenant provisioned successfully"})
}
