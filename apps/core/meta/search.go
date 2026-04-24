package meta

import (
	"fmt"
	"github.com/goerp/goerp/apps/core/database"
	"github.com/goerp/goerp/apps/core/meta/registry"
	"strings"
)

type SearchResult struct {
	DocType     string `json:"doctype"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Route       string `json:"route"`
}

func SearchAll(query string, tenantID string) ([]SearchResult, error) {
	results := []SearchResult{}
	if query == "" { return results, nil }

	// 1. Get all registered DocTypes
	allDTs := registry.DefaultRegistry.GetAll()
	
	// 2. Search across priority DocTypes
	// In a real system, we might only search "Master" data or "Transaction" headers
	searchableTypes := []string{"Customer", "Item", "SalesInvoice", "PurchaseOrder", "Employee", "Account", "User"}
	
	for _, dtName := range searchableTypes {
		// Check if DT exists in registry for this tenant
		if _, exists := registry.DefaultRegistry.Get(dtName, tenantID); !exists {
			continue
		}

		tableName := "tab" + dtName
		var rows []map[string]interface{}
		
		// Simple LIKE search on 'name' field
		// In production, we'd use Full-Text Search (TSVECTOR)
		err := database.DB.Table(tableName).
			Select("name").
			Where("tenant_id = ? AND name LIKE ?", tenantID, "%"+query+"%").
			Limit(5).
			Find(&rows).Error
		
		if err == nil {
			for _, row := range rows {
				name := row["name"].(string)
				results = append(results, SearchResult{
					DocType:     dtName,
					Name:        name,
					Description: fmt.Sprintf("%s: %s", dtName, name),
					Route:       fmt.Sprintf("/app/%s/%s", strings.ToLower(dtName), name),
				})
			}
		}
	}

	// 3. Add Navigation Search (Search for DocTypes themselves)
	for _, dt := range allDTs {
		if strings.Contains(strings.ToLower(dt.Name), strings.ToLower(query)) {
			results = append(results, SearchResult{
				DocType:     "DocType",
				Name:        dt.Name,
				Description: "Navigate to: " + dt.Name + " List",
				Route:       fmt.Sprintf("/app/%s", strings.ToLower(dt.Name)),
			})
		}
	}

	return results, nil
}
