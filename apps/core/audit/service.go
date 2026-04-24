package audit

import (
	"encoding/json"
	"fmt"
	"github.com/goerp/goerp/apps/core/database"
	"time"
)

func LogActivity(user, operation, ip, status, message, tenantID string) {
	data := map[string]interface{}{
		"name":       fmt.Sprintf("LOG-%d", time.Now().UnixNano()),
		"tenant_id":  tenantID,
		"user":       user,
		"operation":  operation,
		"ip_address": ip,
		"log_status": status,
		"message":    message,
		"creation":   time.Now(),
		"modified":   time.Now(),
	}
	database.DB.Table("tabActivityLog").Create(data)
}

func CreateVersion(doctype, docname, changedBy, tenantID string, oldDoc, newDoc interface{}) {
	changeData, _ := json.Marshal(map[string]interface{}{
		"before": oldDoc,
		"after":  newDoc,
	})

	data := map[string]interface{}{
		"name":         fmt.Sprintf("VER-%d", time.Now().UnixNano()),
		"tenant_id":    tenantID,
		"ref_doctype":  doctype,
		"docname":      docname,
		"data":         string(changeData),
		"changed_by":   changedBy,
		"creation":     time.Now(),
		"modified":     time.Now(),
	}
	database.DB.Table("tabVersion").Create(data)
}
