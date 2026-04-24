package workflow

import (
	"errors"
	"fmt"
	"github.com/goerp/goerp/apps/core/database"
	"github.com/goerp/goerp/apps/core/logger"
	"github.com/goerp/goerp/apps/core/meta/registry"
	"github.com/goerp/goerp/apps/core/meta/types"
	"go.uber.org/zap"
	"time"
)

type WorkflowService struct{}

var DefaultService = &WorkflowService{}

func (s *WorkflowService) GetActiveWorkflow(doctype string) (map[string]interface{}, error) {
	var wf map[string]interface{}
	err := database.DB.Table("tabWorkflow").
		Where("target_doctype = ? AND is_active = ?", doctype, true).
		First(&wf).Error
	return wf, err
}

func (s *WorkflowService) ApplyAction(doctype string, docName string, action string, userRoles []string) error {
	wf, err := s.GetActiveWorkflow(doctype)
	if err != nil {
		return fmt.Errorf("no active workflow for %s", doctype)
	}

	wfName := wf["name"].(string)
	stateField := wf["document_state_field"].(string)

	// Get Current Document
	var doc map[string]interface{}
	tableName := fmt.Sprintf("tab%s", doctype)
	if err := database.DB.Table(tableName).Where("name = ?", docName).First(&doc).Error; err != nil {
		return err
	}

	currentState := "Draft" // Default
	if val, ok := doc[stateField]; ok && val != nil {
		currentState = val.(string)
	}

	// Find Transition
	var transition map[string]interface{}
	err = database.DB.Table("tabWorkflowTransition").
		Where("parent = ? AND state = ? AND action = ? AND allowed IN ?", wfName, currentState, action, userRoles).
		First(&transition).Error

	if err != nil {
		return errors.New("invalid transition or insufficient permissions")
	}

	nextState := transition["next_state"].(string)

	// Get Next State Metadata
	var stateMeta map[string]interface{}
	if err := database.DB.Table("tabWorkflowState").
		Where("parent = ? AND state = ?", wfName, nextState).
		First(&stateMeta).Error; err != nil {
		return fmt.Errorf("failed to fetch metadata for state %s", nextState)
	}

	// Update Document
	newDocStatus := int(stateMeta["doc_status"].(float64)) // JSON unmarshal might give float64
	updates := map[string]interface{}{
		stateField:  nextState,
		"docstatus": newDocStatus,
		"modified":  time.Now(),
	}

	if err := database.DB.Table(tableName).Where("name = ?", docName).Updates(updates).Error; err != nil {
		return err
	}

	// Trigger Hooks
	if newDocStatus == 1 {
		// Fetch full doc for hook context
		var updatedDoc map[string]interface{}
		database.DB.Table(tableName).Where("name = ?", docName).First(&updatedDoc)
		
		tenantID := updatedDoc["tenant_id"].(string)
		dt, _ := registry.DefaultRegistry.Get(doctype, tenantID)
		hookCtx := &types.HookContext{
			DocType:  dt,
			Doc:      updatedDoc,
			User:     "system", // Should ideally be the acting user
			HookType: types.OnSubmit,
		}
		registry.DefaultHookRegistry.Trigger(doctype, types.OnSubmit, hookCtx)
	}

	logger.Log.Info("Workflow action applied", 
		zap.String("doctype", doctype), 
		zap.String("doc", docName), 
		zap.String("action", action), 
		zap.String("next_state", nextState))

	return nil
}
