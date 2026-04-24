package notification

import (
	"bytes"
	"github.com/goerp/goerp/apps/core/database"
	"github.com/goerp/goerp/apps/core/job"
	"github.com/goerp/goerp/apps/core/logger"
	"go.uber.org/zap"
	"html/template"
)

type EmailPayload struct {
	Recipient string `json:"recipient"`
	Subject   string `json:"subject"`
	Body      string `json:"body"`
	TenantID  string `json:"tenant_id"`
}

func TriggerNotifications(doctype, event, tenantID string, doc map[string]interface{}) {
	var rules []map[string]interface{}
	err := database.DB.Table("tabNotification").
		Where("ref_doctype = ? AND event = ? AND enabled = ? AND tenant_id = ?", doctype, event, true, tenantID).
		Find(&rules).Error

	if err != nil || len(rules) == 0 {
		return
	}

	for _, rule := range rules {
		// 1. Process Template
		subjectTmpl, _ := template.New("sub").Parse(rule["subject"].(string))
		bodyTmpl, _ := template.New("body").Parse(rule["message_template"].(string))

		var subBuf, bodyBuf bytes.Buffer
		subjectTmpl.Execute(&subBuf, doc)
		bodyTmpl.Execute(&bodyBuf, doc)

		// 2. Resolve Recipient (Simplified: fetch users by role)
		role := rule["send_to_role"].(string)
		var recipients []string
		if role != "" {
			// In a real system, join User with UserRole
			database.DB.Table("tabUser").Where("tenant_id = ?", tenantID).Pluck("email", &recipients)
		}

		// 3. Queue Email Jobs
		for _, email := range recipients {
			payload := EmailPayload{
				Recipient: email,
				Subject:   subBuf.String(),
				Body:      bodyBuf.String(),
				TenantID:  tenantID,
			}
			if err := job.Enqueue("email:send", payload); err != nil {
				logger.Log.
Error("Failed to queue notification", zap.Error(err))
			}
		}
	}
}
