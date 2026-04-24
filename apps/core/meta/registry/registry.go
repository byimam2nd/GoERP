package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/goerp/goerp/apps/core/database"
	"github.com/goerp/goerp/apps/core/meta/types"
	"sync"
	"time"
)

type Registry struct {
	mu       sync.RWMutex
	doctypes map[string]*types.DocType
}

var DefaultRegistry = &Registry{
	doctypes: make(map[string]*types.DocType),
}

func (r *Registry) Register(dt *types.DocType) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.doctypes[dt.Name]; exists {
		return fmt.Errorf("DocType %s already registered", dt.Name)
	}
	r.doctypes[dt.Name] = dt
	return nil
}

func (r *Registry) Get(name string, tenantID string) (*types.DocType, bool) {
	// 1. Redis Cache Layer
	cacheKey := fmt.Sprintf("goerp:meta:%s:%s", tenantID, name)
	if database.Redis != nil {
		if val, err := database.Redis.Get(context.Background(), cacheKey).Result(); err == nil {
			var cachedDT types.DocType
			if err := json.Unmarshal([]byte(val), &cachedDT); err == nil {
				return &cachedDT, true
			}
		}
	}

	r.mu.RLock()
	dt, exists := r.doctypes[name]
	r.mu.RUnlock()

	var finalDocType *types.DocType

	if exists {
		// Clone original from memory
		cloned := *dt
		cloned.Fields = make([]types.DocField, len(dt.Fields))
		copy(cloned.Fields, dt.Fields)
		finalDocType = &cloned
	} else if tenantID != "" && database.DB != nil {
		// Try to find in Database (Dynamic DocType)
		var dbDoc map[string]interface{}
		if err := database.DB.Table("tabDocType").Where("name = ? AND tenant_id = ?", name, tenantID).First(&dbDoc).Error; err == nil {
			var dbFields []map[string]interface{}
			database.DB.Table("tabDocField").Where("parent = ? AND tenant_id = ?", name, tenantID).Order("creation ASC").Find(&dbFields)
			fields := make([]types.DocField, 0)
			for _, f := range dbFields {
				fields = append(fields, types.DocField{
					Name: f["fieldname"].(string), Label: f["label"].(string),
					FieldType: types.FieldType(f["fieldtype"].(string)), Options: f["options"].(string),
					Required: f["required"].(bool), Unique: f["unique"].(bool), InListView: f["in_list_view"].(bool),
				})
			}
			finalDocType = &types.DocType{Name: dbDoc["name"].(string), Module: dbDoc["module"].(string), IsSubmittable: dbDoc["is_submittable"].(bool), Fields: fields}
		}
	}

	if finalDocType == nil { return nil, false }

	// Merge with Custom Fields
	if tenantID != "" && database.DB != nil {
		var customFields []map[string]interface{}
		if err := database.DB.Table("tabCustomField").Where("dt = ? AND tenant_id = ?", name, tenantID).Find(&customFields).Error; err == nil {
			for _, cf := range customFields {
				newField := types.DocField{
					Name: cf["fieldname"].(string), Label: cf["label"].(string),
					FieldType: types.FieldType(cf["field_type"].(string)), Options: cf["options"].(string),
					Required: cf["is_required"].(bool), InListView: cf["in_list_view"].(bool),
				}
				insertAfter := cf["insert_after"].(string)
				inserted := false
				if insertAfter != "" {
					for i, f := range finalDocType.Fields {
						if f.Name == insertAfter {
							finalDocType.Fields = append(finalDocType.Fields[:i+1], append([]types.DocField{newField}, finalDocType.Fields[i+1:]...)...)
							inserted = true
							break
						}
					}
				}
				if !inserted { finalDocType.Fields = append(finalDocType.Fields, newField) }
			}
		}
	}

	// 2. Save to Cache upon success
	if finalDocType != nil && database.Redis != nil {
		jsonData, _ := json.Marshal(finalDocType)
		database.Redis.Set(context.Background(), cacheKey, jsonData, 1*time.Hour)
	}

	return finalDocType, true
}

func (r *Registry) ClearCache(name string, tenantID string) {
	cacheKey := fmt.Sprintf("goerp:meta:%s:%s", tenantID, name)
	if database.Redis != nil {
		database.Redis.Del(context.Background(), cacheKey)
	}
}

func (r *Registry) GetAll() []*types.DocType {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]*types.DocType, 0, len(r.doctypes))
	for _, dt := range r.doctypes { list = append(list, dt) }
	return list
}
