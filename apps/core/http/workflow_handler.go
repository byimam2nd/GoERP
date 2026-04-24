package http

import (
	"github.com/gin-gonic/gin"
	"github.com/goerp/goerp/apps/core/workflow"
	"net/http"
)

type WorkflowActionRequest struct {
	Action string `json:"action" binding:"required"`
}

func HandleWorkflowAction(c *gin.Context) {
	doctype := c.Param("doctype")
	name := c.Param("name")

	var req WorkflowActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	roles, _ := c.Get("roles")
	userRoles := roles.([]string)

	if err := workflow.DefaultService.ApplyAction(doctype, name, req.Action, userRoles); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}
