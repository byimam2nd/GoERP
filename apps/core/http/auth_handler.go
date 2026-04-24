package http

import (
	"github.com/gin-gonic/gin"
	"github.com/goerp/goerp/apps/core/audit"
	"github.com/goerp/goerp/apps/core/auth"
	"github.com/goerp/goerp/apps/core/database"
	"net/http"
)

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func HandleLogin(c *gin.Context) {
	tenantID := c.GetHeader("X-GoERP-Tenant")
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user struct {
		Password string
		Username string
	}
	if err := database.DB.Table("tabUser").Where("username = ? AND tenant_id = ?", req.Username, tenantID).First(&user).Error; err != nil {
		audit.LogActivity(req.Username, "Login", c.ClientIP(), "Failure", "User not found", tenantID)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	if !auth.CheckPasswordHash(req.Password, user.Password) {
		audit.LogActivity(req.Username, "Login", c.ClientIP(), "Failure", "Incorrect password", tenantID)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	roles := []string{"Administrator"}

	token, err := auth.GenerateToken(req.Username, roles)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	audit.LogActivity(req.Username, "Login", c.ClientIP(), "Success", "", tenantID)

	c.JSON(http.StatusOK, gin.H{
		"token": token,
	})
}
