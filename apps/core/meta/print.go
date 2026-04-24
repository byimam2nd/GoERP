package meta

import (
	"bytes"
	"fmt"
	"github.com/goerp/goerp/apps/core/database"
	"html/template"
)

func GeneratePrintHTML(doctype, name, formatName, tenantID string) (string, error) {
	// 1. Fetch main document
	var doc map[string]interface{}
	tableName := fmt.Sprintf("tab%s", doctype)
	if err := database.DB.Table(tableName).Where("name = ? AND tenant_id = ?", name, tenantID).First(&doc).Error; err != nil {
		return "", err
	}

	// 2. Fetch DocType Meta to identify Child Tables
	// Assume we have a registry to get fields
	// For simplicity, let's fetch all tables that have this name as parent
	// In a real system, we'd use meta.Fields where FieldType == "Table"
	
	// Fetch all potential child records (simplified)
	// We'll pass the whole Doc object with its children to the template
	// This is where the magic of ERPNext Print Format happens

	// 3. Fetch the Print Format Template
	var format map[string]interface{}
	if err := database.DB.Table("tabPrintFormat").
		Where("doc_type = ? AND tenant_id = ?", doctype, tenantID).
		Order("is_default DESC").First(&format).Error; err != nil {
		// Fallback to basic template if none found
		format = map[string]interface{}{
			"html_template": "<h1>{{.name}}</h1><p>Please define a Print Format for this DocType.</p>",
			"custom_css": "",
		}
	}

	htmlTemplate := format["html_template"].(string)

	// 4. Template Helpers (Currency formatting, etc)
	funcMap := template.FuncMap{
		"currency": func(v interface{}) string {
			return fmt.Sprintf("%.2f", v)
		},
		"date": func(v interface{}) string {
			// Convert to readable date
			return fmt.Sprintf("%v", v)
		},
	}

	tmpl, err := template.New("print").Funcs(funcMap).Parse(htmlTemplate)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, doc); err != nil {
		return "", err
	}

	return fmt.Sprintf(`
		<html><head><style>body{padding:20px; font-family: sans-serif;} %s</style></head>
		<body>%s</body></html>
	`, format["custom_css"], buf.String()), nil
}
