package service

import (
	"context"
	"fmt"
	"github.com/goerp/goerp/apps/core/database"
	"github.com/goerp/goerp/apps/core/meta/migrator"
	"github.com/goerp/goerp/apps/core/meta/registry"
	"github.com/goerp/goerp/apps/core/meta/types"
	"time"
)

// AddCustomField creates a custom field record and triggers the database migration.
func AddCustomField(ctx context.Context, tenantID string, dtName string, field types.DocField, insertAfter string) error {
	// 1. Persist metadata to tabCustomField
	data := map[string]interface{}{
		"name":         fmt.Sprintf("CF-%s-%s", dtName, field.Name),
		"tenant_id":    tenantID,
		"dt":           dtName,
		"label":        field.Label,
		"fieldname":    field.Name,
		"field_type":   string(field.FieldType),
		"options":      field.Options,
		"is_required":  field.Required,
		"in_list_view": field.InListView,
		"insert_after": insertAfter,
		"creation":     time.Now(),
		"modified":     time.Now(),
	}

	if err := database.DB.Table("tabCustomField").Create(data).Error; err != nil {
		return fmt.Errorf("failed to save custom field metadata: %v", err)
	}

	// 2. Prepare Migration Plan
	plan := migrator.MigrationPlan{
		DocType: dtName,
		Version: int(time.Now().Unix()), // Simplistic versioning
		Steps: []migrator.MigrationStep{
			{
				Type:      migrator.AddColumn,
				TableName: dtName,
				Field:     field,
			},
		},
	}

	// 3. Execute Database Migration
	if err := migrator.ExecutePlan(ctx, plan); err != nil {
		return fmt.Errorf("failed to migrate database: %v", err)
	}

	// 4. Invalidate Cache
	registry.DefaultRegistry.ClearCache(dtName, tenantID)

	return nil
}
