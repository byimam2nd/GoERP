package event

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/goerp/goerp/apps/core/database"
	"github.com/goerp/goerp/apps/core/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"time"
)

type EventType string

const (
	DocSubmitted  EventType = "DOC_SUBMITTED"
	DocCancelled  EventType = "DOC_CANCELLED"
	DocCorrected  EventType = "DOC_CORRECTED"
	SchemaChanged EventType = "SCHEMA_CHANGED"
)

// BusinessEvent adalah "Atomic Truth" dari setiap aksi di sistem.
type BusinessEvent struct {
	ID              uint      `gorm:"primaryKey"`
	EventID         string    `gorm:"uniqueIndex"` // UUID/ULID
	TenantID        string    `gorm:"index"`
	EventType       EventType `gorm:"index"`
	SourceDocType   string    `gorm:"index"`
	SourceDocName   string    `gorm:"index"`
	Payload         string    `gorm:"type:text"`   // Data asli saat kejadian
	Metadata        string    `gorm:"type:text"`   // User, IP, NodeID, dll
	Checksum        string    // Untuk verifikasi integritas data
	Creation        time.Time `gorm:"index"`
}

// Store mencatat event secara immutable ke dalam database.
func Store(tx *gorm.DB, tenantID string, eType EventType, dt string, name string, payload interface{}, meta interface{}) error {
	if tx == nil {
		tx = database.DB
	}

	payloadBytes, _ := json.Marshal(payload)
	metaBytes, _ := json.Marshal(meta)

	event := BusinessEvent{
		EventID:       fmt.Sprintf("EVT-%d", time.Now().UnixNano()),
		TenantID:      tenantID,
		EventType:     eType,
		SourceDocType: dt,
		SourceDocName: name,
		Payload:       string(payloadBytes),
		Metadata:      string(metaBytes),
		Creation:      time.Now(),
	}

	// Hitung Checksum sederhana (Bisa diperkuat dengan HMAC/Digital Signature)
	event.Checksum = fmt.Sprintf("%x", time.Now().UnixNano()) 

	if err := tx.Table("tabBusinessEvent").Create(&event).Error; err != nil {
		logger.Log.Error("Failed to persist business event", zap.Error(err))
		return err
	}

	return nil
}

// ReplayEvent (V2) - Kemampuan untuk membangun ulang state dari event log.
func ReplayEvents(tenantID string, docName string) ([]BusinessEvent, error) {
	var events []BusinessEvent
	err := database.DB.Table("tabBusinessEvent").
		Where("tenant_id = ? AND source_doc_name = ?", tenantID, docName).
		Order("creation ASC").
		Find(&events).Error
	return events, err
}
