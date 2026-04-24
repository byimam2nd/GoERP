package meta

import (
	"fmt"
	"github.com/goerp/goerp/apps/core/database"
	"github.com/goerp/goerp/apps/core/meta/registry"
	"github.com/goerp/goerp/apps/core/meta/types"
)

func MapDocument(sourceDT, sourceName, targetDT, tenantID string) (map[string]interface{}, error) {
	// 1. Fetch Source Meta & Data
	sMeta, exists := registry.DefaultRegistry.Get(sourceDT, tenantID)
	if !exists { return nil, fmt.Errorf("Source DocType not found") }

	var sourceDoc map[string]interface{}
	if err := database.DB.Table("tab"+sourceDT).Where("name = ? AND tenant_id = ?", sourceName, tenantID).First(&sourceDoc).Error; err != nil {
		return nil, fmt.Errorf("Source document not found")
	}

	// 2. Fetch Target Meta
	tMeta, exists := registry.DefaultRegistry.Get(targetDT, tenantID)
	if !exists { return nil, fmt.Errorf("Target DocType not found") }

	// 3. Start Mapping Logic
	targetDoc := make(map[string]interface{})

	// Map Header Fields (Simple matching by name)
	for _, tField := range tMeta.Fields {
		if tField.FieldType == types.Table || tField.FieldType == types.SectionBreak || tField.FieldType == types.ColumnBreak {
			continue
		}

		// Try to find matching value in source
		if val, ok := sourceDoc[tField.Name]; ok {
			targetDoc[tField.Name] = val
		}
	}

	// Special Field: Reference to Source
	refFieldName := ""
	for _, f := range tMeta.Fields {
		if f.FieldType == types.Link && f.Options == sourceDT {
			refFieldName = f.Name
			break
		}
	}
	if refFieldName != "" {
		targetDoc[refFieldName] = sourceName
	}

	// 4. Map Child Tables
	for _, tField := range tMeta.Fields {
		if tField.FieldType == types.Table {
			// Find matching table in source by options (DocType)
			sourceTableField := ""
			for _, sField := range sMeta.Fields {
				if sField.FieldType == types.Table && sField.Options == tField.Options {
					sourceTableField = sField.Name
					break
				}
			}

			if sourceTableField != "" {
				// Fetch source rows
				var sourceRows []map[string]interface{}
				database.DB.Table("tab"+tField.Options).
					Where("parent = ? AND tenant_id = ?", sourceName, tenantID).
					Find(&sourceRows)

				targetRows := []map[string]interface{}{}
				for _, sRow := range sourceRows {
					tRow := make(map[string]interface{})
					// Copy columns by name
					for k, v := range sRow {
						if k != "name" && k != "parent" && k != "idx" {
							tRow[k] = v
						}
					}
					targetRows = append(targetRows, tRow)
				}
				targetDoc[tField.Name] = targetRows
			}
		}
	}

	return targetDoc, nil
}
