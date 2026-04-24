package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/goerp/goerp/apps/core/database"
	"github.com/goerp/goerp/apps/core/job"
	"github.com/goerp/goerp/apps/core/logger"
	"github.com/hibiken/asynq"
	"go.uber.org/zap"
	"net/http"
	"time"
)

const TypeWebhookDispatch = "webhook:dispatch"

type WebhookPayload struct {
	URL      string                 `json:"url"`
	Event    string                 `json:"event"`
	DocType  string                 `json:"doctype"`
	Data     map[string]interface{} `json:"data"`
	Secret   string                 `json:"secret"` // For signature
	TenantID string                 `json:"tenant_id"`
}

func InitIntegration() {
	job.RegisterHandler(TypeWebhookDispatch, HandleWebhookDispatch)
}

// TriggerWebhook finds and enqueues all webhooks for a given event
func TriggerWebhook(doctype string, event string, doc map[string]interface{}, tenantID string) {
	var webhooks []map[string]interface{}
	database.DB.Table("tabWebhook").
		Where("webhook_doctype = ? AND event = ? AND is_active = ? AND tenant_id = ?", doctype, event, true, tenantID).
		Find(&webhooks)

	for _, wh := range webhooks {
		payload := WebhookPayload{
			URL:      wh["request_url"].(string),
			Event:    event,
			DocType:  doctype,
			Data:     doc,
			Secret:   wh["secret"].(string),
			TenantID: tenantID,
		}
		
		// Use Asynq for reliable background delivery
		job.Enqueue(TypeWebhookDispatch, payload)
	}
}

func HandleWebhookDispatch(ctx context.Context, t *asynq.Task) error {
	var p WebhookPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return err
	}

	jsonData, err := json.Marshal(p)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", p.URL, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-GoERP-Event", p.Event)
	req.Header.Set("X-GoERP-Tenant", p.TenantID)
	// In production: add X-GoERP-Signature using p.Secret

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err // Asynq will retry based on config
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook failed: %d", resp.StatusCode)
	}

	logger.Log.Info("Webhook delivered", zap.String("url", p.URL), zap.String("event", p.Event))
	return nil
}
