package service

import (
	"context"
	"fmt"
	"github.com/goerp/goerp/apps/core/database"
)

// GetCompanyDefaultAccount mengambil akun default (misal: Receivable, Payable, Income) dari metadata Company.
func GetCompanyDefaultAccount(tenantID string, companyName string, fieldName string) (string, error) {
	var accountName string
	err := database.DB.Table("tabCompany").
		Where("name = ? AND tenant_id = ?", companyName, tenantID).
		Select(fieldName).Scan(&accountName).Error
	
	if err != nil || accountName == "" {
		return "", fmt.Errorf("default account %s not configured for company %s", fieldName, companyName)
	}
	return accountName, nil
}

// ReversalImpact menyimpan ringkasan dokumen yang akan terkena dampak pembatalan.
type RevaluationImpact struct {
	DocType string
	Name    string
	Action  string // "Reverse", "Unlink", "Warning"
}

// CascadeReverseDocument melakukan pembatalan berantai secara rekursif.
func CascadeReverseDocument(tx *gorm.DB, tenantID string, dt string, name string) ([]RevaluationImpact, error) {
	var impact []RevaluationImpact

	// 1. Cari semua dokumen anak yang aktif (docstatus = 1)
	var children []struct {
		ChildDT   string
		ChildName string
	}
	tx.Table("tabDocLink").
		Where("tenant_id = ? AND parent_name = ? AND docstatus = 1", tenantID, name).
		Scan(&children)

	for _, child := range children {
		// 2. Rekursi ke bawah dulu (Bottom-up Reversal)
		childImpact, err := CascadeReverseDocument(tx, tenantID, child.ChildDT, child.ChildName)
		if err != nil {
			return nil, fmt.Errorf("failed to reverse child %s %s: %v", child.ChildDT, child.ChildName, err)
		}
		impact = append(impact, childImpact...)

		// 3. Eksekusi pembatalan spesifik per DocType (via Hook)
		// Di sini kita memanggil fungsi pembatalan generik
		if err := executeInternalCancel(tx, tenantID, child.ChildDT, child.ChildName); err != nil {
			return nil, err
		}
		impact = append(impact, RevaluationImpact{DocType: child.ChildDT, Name: child.ChildName, Action: "Reverse"})
	}

	return impact, nil
}

func executeInternalCancel(tx *gorm.DB, tenantID string, dt string, name string) error {
	// Update status dokumen menjadi Cancelled (2)
	tableName := "tab" + dt
	if err := tx.Table(tableName).Where("tenant_id = ? AND name = ?", tenantID, name).Update("docstatus", 2).Error; err != nil {
		return err
	}
	
	// Tandai DocLink sebagai tidak aktif
	return tx.Table("tabDocLink").Where("tenant_id = ? AND child_name = ?", tenantID, name).Update("docstatus", 2).Error
}
