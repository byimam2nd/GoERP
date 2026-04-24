package meta

import (
	"fmt"
	"github.com/goerp/goerp/apps/core/database"
	"time"
)

func CreateDocLink(parentDT, parentName, childDT, childName, tenantID string) error {
	// Check if already linked
	var count int64
	database.DB.Table("tabDocLink").
		Where("parent_name = ? AND child_name = ? AND tenant_id = ?", parentName, childName, tenantID).
		Count(&count)
	
	if count > 0 {
		return nil
	}

	linkData := map[string]interface{}{
		"name":            fmt.Sprintf("LINK-%d", time.Now().UnixNano()),
		"tenant_id":       tenantID,
		"parent_doctype":  parentDT,
		"parent_name":     parentName,
		"child_doctype":   childDT,
		"child_name":      childName,
		"link_type":       "Reference",
		"creation":        time.Now(),
		"modified":        time.Now(),
		"docstatus":       0,
	}

	return database.DB.Table("tabDocLink").Create(linkData).Error
}

type DocLinkInfo struct {
	Parents []map[string]interface{} `json:"parents"`
	Children []map[string]interface{} `json:"children"`
}

func GetDocLinks(doctype, name, tenantID string) (DocLinkInfo, error) {
	var parents []map[string]interface{}
	var children []map[string]interface{}

	// Upward links
	database.DB.Table("tabDocLink").
		Where("child_name = ? AND tenant_id = ?", name, tenantID).
		Find(&parents)

	// Downward links
	database.DB.Table("tabDocLink").
		Where("parent_name = ? AND tenant_id = ?", name, tenantID).
		Find(&children)

	return DocLinkInfo{Parents: parents, Children: children}, nil
}
