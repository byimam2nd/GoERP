package meta

import (
	"encoding/csv"
	"fmt"
	"github.com/goerp/goerp/apps/core/database"
	"github.com/goerp/goerp/apps/core/meta/registry"
	"github.com/goerp/goerp/apps/core/meta/types"
	"io"
	"strings"
)

func GenerateCSVTemplate(doctypeName string, tenantID string) (string, error) {
	dt, exists := registry.DefaultRegistry.Get(doctypeName, tenantID)
	if !exists {
		return "", fmt.Errorf("DocType not found")
	}

	var headers []string
	for _, field := range dt.Fields {
		// Skip layout fields
		if field.FieldType == types.SectionBreak || field.FieldType == types.ColumnBreak {
			continue
		}
		headers = append(headers, field.Name)
	}

	return strings.Join(headers, ",") + "\n", nil
}

func ProcessCSVImport(doctypeName string, r io.Reader, tenantID string) (int, error) {
	reader := csv.NewReader(r)
	
	// 1. Read Header
	headers, err := reader.Read()
	if err != nil {
		return 0, err
	}

	count := 0
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return count, err
		}

		// 2. Map Record to Data
		docData := make(map[string]interface{})
		for i, h := range headers {
			if i < len(record) {
				docData[h] = record[i]
			}
		}

		// 3. Create Document (Simplified - reuse internal logic or direct DB)
		// Here we would ideally call HandleCreate logic but to keep it simple for now:
		name, _ := GetNextName(doctypeName, tenantID)
		docData["name"] = name
		docData["tenant_id"] = tenantID
		docData["docstatus"] = 0
		
		tableName := fmt.Sprintf("tab%s", doctypeName)
		if err := database.DB.Table(tableName).Create(docData).Error; err == nil {
			count++
		}
	}

	return count, nil
}
