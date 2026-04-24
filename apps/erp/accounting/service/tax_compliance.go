package service

import (
	"fmt"
	"github.com/goerp/goerp/apps/core/database"
	"gorm.io/gorm"
	"time"
)

type ComplianceState string

const (
	TaxDraft    ComplianceState = "Draft"
	TaxReported ComplianceState = "Reported" // Dokumen sudah dilaporkan ke regulator (e-Faktur)
	TaxCorrected ComplianceState = "Corrected"
)

// MarkAsReported mengunci dokumen karena sudah dilaporkan ke otoritas pajak.
func MarkAsReported(tenantID string, dt string, name string, reference string) error {
	tableName := "tab" + dt
	return database.DB.Table(tableName).
		Where("tenant_id = ? AND name = ?", tenantID, name).
		Updates(map[string]interface{}{
			"tax_compliance_state": TaxReported,
			"tax_reference_no":     reference,
			"tax_reported_at":      time.Now(),
		}).Error
}

// ValidateCancellation mencegah pembatalan langsung jika dokumen sudah dilaporkan pajak.
func ValidateCancellation(tenantID string, dt string, name string) error {
	var doc struct {
		TaxComplianceState ComplianceState
	}
	tableName := "tab" + dt
	database.DB.Table(tableName).Where("tenant_id = ? AND name = ?", tenantID, name).Scan(&doc)

	if doc.TaxComplianceState == TaxReported {
		return fmt.Errorf("Compliance Error: Document %s has been reported to Tax Authority. You must issue a Credit Note/Correction instead of cancelling.", name)
	}
	return nil
}
