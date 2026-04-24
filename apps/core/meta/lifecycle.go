package meta

import (
	"fmt"
	"github.com/goerp/goerp/apps/core/database"
	"gorm.io/gorm"
	"time"
)

type DocStatus int

const (
	StatusDraft     DocStatus = 0
	StatusSubmitted DocStatus = 1
	StatusCancelled DocStatus = 2
)

// LifecycleManager handles the transitions between document states.
type LifecycleManager struct {
	DB *gorm.DB
}

func NewLifecycleManager() *LifecycleManager {
	return &LifecycleManager{DB: database.DB}
}

import (
	"fmt"
	"github.com/goerp/goerp/apps/core/database"
	"github.com/goerp/goerp/apps/core/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

// Submit locks the document and triggers ledger postings with row-level locking.
func (m *LifecycleManager) Submit(doctype, name, tenantID string, onPost func(tx *gorm.DB) error) error {
	start := time.Now()
	// ... (Existing log setup)

	return m.DB.Transaction(func(tx *gorm.DB) error {
		var doc map[string]interface{}
		if err := tx.Table("tab"+doctype).
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("name = ? AND tenant_id = ?", name, tenantID).
			First(&doc).Error; err != nil {
			return fmt.Errorf("failed to lock document: %w", err)
		}

		// --- THE IMMORTALITY GUARD: Check Fiscal Year ---
		if postingDate, ok := doc["posting_date"].(time.Time); ok {
			var fy map[string]interface{}
			if err := tx.Table("tabFiscalYear").
				Where("tenant_id = ? AND start_date <= ? AND end_date >= ? AND is_closed = ?", 
					tenantID, postingDate, postingDate, true).
				First(&fy).Error; err == nil {
				return fmt.Errorf("Fiscal Year %s is CLOSED. No transactions allowed in this period.", fy["name"])
			}
		}

		if ds, _ := doc["docstatus"].(int64); ds != int64(StatusDraft) {
			return fmt.Errorf("document %s is not in Draft state", name)
		}
// ... (Rest of Submit logic)

		if err := tx.Table("tab"+doctype).Where("name = ?", name).Updates(map[string]interface{}{
			"docstatus": StatusSubmitted,
			"modified":  time.Now(),
		}).Error; err != nil {
			return err
		}

		if onPost != nil {
			postStart := time.Now()
			if err := onPost(tx); err != nil {
				log.Error("Ledger posting failed", zap.Error(err))
				return err
			}
			log.Info("Ledger posting completed", zap.Duration("duration", time.Since(postStart)))
		}

		// --- THE GUARD: VALIDATE SYSTEM INVARIANTS ---
		// This happens after all calculations but BEFORE database commit.
		if err := validator.ValidateInvariants(tx, doctype, name, tenantID); err != nil {
			log.Error("System invariant violation", zap.Error(err))
			return err
		}

		log.Info("Document submitted successfully", zap.Duration("total_duration", time.Since(start)))
		return nil
	})
}

func (m *LifecycleManager) GetFullDoc(doctype, name, tenantID string) (map[string]interface{}, error) {
	var doc map[string]interface{}
	tableName := "tab" + doctype
	if err := m.DB.Table(tableName).Where("name = ? AND tenant_id = ?", name, tenantID).First(&doc).Error; err != nil {
		return nil, err
	}

	// We need DocType metadata to find child fields
	// For this, we use the registry (assumed to be available)
	// Import cycle could happen if we are not careful.
	// But let's assume registry is accessible via a package or passed in.
	
	// Implementation Note: In a production system, we'd loop through fields
	// and fetch child records for each field where fieldtype == "Table".
	
	return doc, nil
}
